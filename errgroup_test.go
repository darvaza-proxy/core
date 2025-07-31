package core

import (
	"context"
	"errors"
	"testing"
	"time"
)

type errGroupSetDefaultsTestCase struct {
	expectedParent context.Context
	setup          func() *ErrGroup
	name           string
}

var errGroupSetDefaultsTestCases = []errGroupSetDefaultsTestCase{
	{
		name: "nil parent context",
		setup: func() *ErrGroup {
			return &ErrGroup{}
		},
		expectedParent: context.Background(),
	},
	{
		name: "custom parent context",
		setup: func() *ErrGroup {
			type testKey string
			ctx := context.WithValue(context.Background(), testKey("test"), "value")
			return &ErrGroup{Parent: ctx}
		},
		expectedParent: nil, // Will be set by test
	},
}

func (tc errGroupSetDefaultsTestCase) test(t *testing.T) {
	t.Helper()

	eg := tc.setup()
	expectedParent := tc.expectedParent
	if tc.name == "custom parent context" {
		expectedParent = eg.Parent
	}

	eg.SetDefaults()

	if eg.Parent != expectedParent {
		t.Errorf("Expected Parent %v, got %v", expectedParent, eg.Parent)
	}

	if eg.ctx == nil {
		t.Error("Expected ctx to be initialized")
	}

	if eg.cancel == nil {
		t.Error("Expected cancel function to be initialized")
	}
}

func TestErrGroupSetDefaults(t *testing.T) {
	for _, tc := range errGroupSetDefaultsTestCases {
		t.Run(tc.name, tc.test)
	}
}

type errGroupGoTestCase struct {
	runFunc      func(context.Context) error
	shutdownFunc func() error
	name         string
	expectError  bool
	expectCancel bool
}

var errGroupGoTestCases = []errGroupGoTestCase{
	{
		name: "successful worker",
		runFunc: func(_ context.Context) error {
			return nil
		},
		shutdownFunc: nil,
		expectError:  false,
		expectCancel: false,
	},
	{
		name: "worker with error",
		runFunc: func(_ context.Context) error {
			return errors.New("worker error")
		},
		shutdownFunc: nil,
		expectError:  true,
		expectCancel: true,
	},
	{
		name: "worker with panic",
		runFunc: func(_ context.Context) error {
			panic("worker panic")
		},
		shutdownFunc: nil,
		expectError:  true,
		expectCancel: true,
	},
	{
		name: "successful worker with shutdown",
		runFunc: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		},
		shutdownFunc: func() error {
			return nil
		},
		expectError:  false,
		expectCancel: true, // Manual cancellation will happen
	},
	{
		name: "worker with shutdown error",
		runFunc: func(_ context.Context) error {
			// Worker runs for a short time then completes
			time.Sleep(5 * time.Millisecond)
			return nil
		},
		shutdownFunc: func() error {
			// Shutdown immediately returns error
			return errors.New("shutdown error")
		},
		expectError:  false,
		expectCancel: true, // Manual cancellation will happen
	},
}

func (tc errGroupGoTestCase) test(t *testing.T) {
	t.Helper()

	var eg ErrGroup
	eg.Go(tc.runFunc, tc.shutdownFunc)

	tc.handleShutdownTests(&eg)
	err := eg.Wait()
	tc.checkError(t, err)
	tc.checkCancellation(t, &eg)
}

func (tc errGroupGoTestCase) handleShutdownTests(eg *ErrGroup) {
	if tc.name == "successful worker with shutdown" || tc.name == "worker with shutdown error" {
		go func() {
			time.Sleep(10 * time.Millisecond)
			eg.Cancel(errors.New("manual cancellation"))
		}()
	}
}

func (tc errGroupGoTestCase) checkError(t *testing.T, err error) {
	t.Helper()
	if tc.expectError {
		AssertError(t, err, "error")
	} else if errors.Is(err, context.Canceled) {
		t.Log("context cancelled as expected")
	} else {
		AssertNoError(t, err, "no error")
	}
}

func (tc errGroupGoTestCase) checkCancellation(t *testing.T, eg *ErrGroup) {
	t.Helper()
	if tc.expectCancel && !eg.IsCancelled() {
		t.Error("Expected group to be cancelled")
	}
}

func TestErrGroupGo(t *testing.T) {
	for _, tc := range errGroupGoTestCases {
		t.Run(tc.name, tc.test)
	}
}

type errGroupGoCatchTestCase struct {
	runFunc     func(context.Context) error
	catchFunc   func(context.Context, error) error
	name        string
	expectError bool
}

var errGroupGoCatchTestCases = []errGroupGoCatchTestCase{
	{
		name: "successful worker with catch",
		runFunc: func(_ context.Context) error {
			return nil
		},
		catchFunc: func(_ context.Context, _ error) error {
			// This should never be called for successful workers
			return errors.New("catch should not be called")
		},
		expectError: false,
	},
	{
		name: "worker error handled by catch",
		runFunc: func(_ context.Context) error {
			return errors.New("worker error")
		},
		catchFunc: func(_ context.Context, _ error) error {
			return nil // dismiss error
		},
		expectError: false,
	},
	{
		name: "worker error transformed by catch",
		runFunc: func(_ context.Context) error {
			return errors.New("original error")
		},
		catchFunc: func(_ context.Context, _ error) error {
			return errors.New("transformed error")
		},
		expectError: true, // Transformed error should propagate
	},
	{
		name:    "nil run function",
		runFunc: nil,
		catchFunc: func(_ context.Context, err error) error {
			return err
		},
		expectError: true, // Should panic and be caught
	},
}

func (tc errGroupGoCatchTestCase) test(t *testing.T) {
	t.Helper()

	var eg ErrGroup

	if tc.runFunc == nil {
		tc.testNilFunction(t, &eg)
		return
	}

	eg.GoCatch(tc.runFunc, tc.catchFunc)
	err := eg.Wait()
	tc.checkTestResult(t, err)
}

func (tc errGroupGoCatchTestCase) testNilFunction(t *testing.T, eg *ErrGroup) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nil run function")
		}
	}()
	eg.GoCatch(tc.runFunc, tc.catchFunc)
}

func (tc errGroupGoCatchTestCase) checkTestResult(t *testing.T, err error) {
	t.Helper()
	if tc.expectError {
		if err == nil {
			t.Errorf("Test case '%s': Expected error but got nil", tc.name)
		}
	} else if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
}

func TestErrGroupGoCatch(t *testing.T) {
	for _, tc := range errGroupGoCatchTestCases {
		t.Run(tc.name, tc.test)
	}
}

func testErrGroupFirstCancellation(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	cause := errors.New("test error")
	isFirst := eg.Cancel(cause)

	AssertTrue(t, isFirst, "first cancellation")
	AssertTrue(t, eg.IsCancelled(), "cancelled")
	AssertEqual(t, cause, eg.Err(), "error")
}

func testErrGroupSubsequentCancellation(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	// First cancellation
	cause1 := errors.New("first error")
	eg.Cancel(cause1)

	// Second cancellation
	cause2 := errors.New("second error")
	isFirst := eg.Cancel(cause2)

	if isFirst {
		t.Error("Expected subsequent cancellation to return false")
	}

	// Should keep the first error
	if err := eg.Err(); err != cause1 {
		t.Errorf("Expected first error %v, got %v", cause1, err)
	}
}

func testErrGroupNilCause(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	isFirst := eg.Cancel(nil)

	if !isFirst {
		t.Error("Expected first cancellation to return true")
	}

	if err := eg.Err(); err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

func TestErrGroupCancel(t *testing.T) {
	t.Run("first cancellation", testErrGroupFirstCancellation)
	t.Run("subsequent cancellation", testErrGroupSubsequentCancellation)
	t.Run("nil cause", testErrGroupNilCause)
}

func TestErrGroupOnError(t *testing.T) {
	var eg ErrGroup
	var errorReceived error

	eg.OnError(func(err error) {
		errorReceived = err
	})

	testErr := errors.New("test error")
	eg.Cancel(testErr)

	// Give time for onError to be called
	time.Sleep(1 * time.Millisecond)

	if errorReceived != testErr {
		t.Errorf("Expected onError to receive %v, got %v", testErr, errorReceived)
	}
}

func TestErrGroupContext(t *testing.T) {
	var eg ErrGroup

	ctx := eg.Context()
	if ctx == nil {
		t.Error("Expected non-nil context")
	}

	// Context should not be done initially
	select {
	case <-ctx.Done():
		t.Error("Context should not be done initially")
	default:
		// Expected
	}

	// Cancel the group
	eg.Cancel(errors.New("test"))

	// Context should now be done
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(1 * time.Millisecond):
		t.Error("Context should be done after cancellation")
	}
}

func TestErrGroupCancelled(t *testing.T) {
	var eg ErrGroup

	cancelled := eg.Cancelled()
	if cancelled == nil {
		t.Error("Expected non-nil cancelled channel")
	}

	// Should not be cancelled initially
	select {
	case <-cancelled:
		t.Error("Should not be cancelled initially")
	default:
		// Expected
	}

	// Cancel the group
	eg.Cancel(errors.New("test"))

	// Should now be cancelled
	select {
	case <-cancelled:
		// Expected
	case <-time.After(1 * time.Millisecond):
		t.Error("Should be cancelled after Cancel()")
	}
}

func TestErrGroupDone(t *testing.T) {
	var eg ErrGroup

	eg.Go(func(_ context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}, nil)

	done := eg.Done()

	// Should not be done yet
	select {
	case <-done:
		t.Error("Done channel closed too early")
	case <-time.After(5 * time.Millisecond):
		// Expected
	}

	// Wait for completion
	select {
	case <-done:
		// Expected
	case <-time.After(50 * time.Millisecond):
		t.Error("Done channel never closed")
	}
}

func TestErrGroupConcurrency(t *testing.T) {
	const numWorkers = 10

	var eg ErrGroup
	startConcurrentWorkers(t, &eg, numWorkers)

	err := eg.Wait()
	validateConcurrencyResult(t, &eg, err)
}

func startConcurrentWorkers(t *testing.T, eg *ErrGroup, numWorkers int) {
	t.Helper()
	for i := 0; i < numWorkers; i++ {
		worker := i
		eg.Go(createConcurrentWorker(worker), nil)
	}
}

func createConcurrentWorker(worker int) func(context.Context) error {
	return func(ctx context.Context) error {
		// Worker 5 fails quickly, others run longer but should be cancelled
		if worker == 5 {
			time.Sleep(5 * time.Millisecond)
			return errors.New("worker 5 error")
		}

		// Other workers wait for cancellation or timeout
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil // Should not reach here if cancellation works
		}
	}
}

func validateConcurrencyResult(t *testing.T, eg *ErrGroup, err error) {
	t.Helper()
	AssertError(t, err, "worker 5 error")
	AssertTrue(t, eg.IsCancelled(), "group cancelled")
}

func testErrGroupCatcherErrorWhenNotCancelled(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	testErr := errors.New("test error")
	result := eg.defaultErrGroupCatcher(testErr)

	if result != testErr {
		t.Errorf("Expected %v, got %v", testErr, result)
	}
}

func testErrGroupCatcherErrorWhenCancelled(t *testing.T) {
	t.Helper()
	var eg ErrGroup
	eg.Cancel(errors.New("cancellation error"))

	testErr := errors.New("test error")
	result := eg.defaultErrGroupCatcher(testErr)

	if result != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", result)
	}
}

func testErrGroupCatcherNilError(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	result := eg.defaultErrGroupCatcher(nil)

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestErrGroupDefaultErrGroupCatcher(t *testing.T) {
	t.Run("error when not cancelled", testErrGroupCatcherErrorWhenNotCancelled)
	t.Run("error when cancelled", testErrGroupCatcherErrorWhenCancelled)
	t.Run("nil error", testErrGroupCatcherNilError)
}

func TestErrGroupWithCustomParent(t *testing.T) {
	parentCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	eg := ErrGroup{Parent: parentCtx}

	eg.Go(func(ctx context.Context) error {
		// Worker should be cancelled by parent timeout
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return errors.New("should not reach here")
		}
	}, nil)

	err := eg.Wait()
	if err == nil {
		t.Error("Expected timeout error from parent context")
	}
}
