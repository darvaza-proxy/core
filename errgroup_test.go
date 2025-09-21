package core

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = errGroupSetDefaultsTestCase{}
	_ TestCase = errGroupGoTestCase{}
	_ TestCase = errGroupGoCatchTestCase{}
)

type errGroupSetDefaultsTestCase struct {
	expectedParent context.Context
	setup          func() *ErrGroup
	name           string
}

// newErrGroupSetDefaultsTestCase creates a new errGroupSetDefaultsTestCase
func newErrGroupSetDefaultsTestCase(expectedParent context.Context, name string,
	setup func() *ErrGroup) TestCase {
	return errGroupSetDefaultsTestCase{
		name:           name,
		setup:          setup,
		expectedParent: expectedParent,
	}
}

func makeErrGroupSetDefaultsTestCases() []TestCase {
	var nilCtx context.Context

	return S(
		newErrGroupSetDefaultsTestCase(context.Background(), "nil parent context", func() *ErrGroup {
			return &ErrGroup{}
		}),
		newErrGroupSetDefaultsTestCase(nilCtx, "custom parent context", func() *ErrGroup {
			type testKey string
			ctx := context.WithValue(context.Background(), testKey("test"), "value")
			return &ErrGroup{Parent: ctx}
		}), // Will be set by test
	)
}

func (tc errGroupSetDefaultsTestCase) Name() string {
	return tc.name
}

func (tc errGroupSetDefaultsTestCase) Test(t *testing.T) {
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

type errGroupGoTestCase struct {
	runFunc      func(context.Context) error
	shutdownFunc func() error
	name         string
	expectError  bool
	expectCancel bool
}

// newErrGroupGoTestCase creates a new errGroupGoTestCase
func newErrGroupGoTestCase(name string, runFunc func(context.Context) error,
	shutdownFunc func() error, expectError, expectCancel bool) TestCase {
	return errGroupGoTestCase{
		name:         name,
		runFunc:      runFunc,
		shutdownFunc: shutdownFunc,
		expectError:  expectError,
		expectCancel: expectCancel,
	}
}

// newErrGroupGoTestCaseSuccess creates a test case expecting successful completion
func newErrGroupGoTestCaseSuccess(name string, runFunc func(context.Context) error,
	shutdownFunc func() error) TestCase {
	return newErrGroupGoTestCase(name, runFunc, shutdownFunc, false, false)
}

// newErrGroupGoTestCaseError creates a test case expecting error and cancellation
func newErrGroupGoTestCaseError(name string, runFunc func(context.Context) error,
	shutdownFunc func() error) TestCase {
	return newErrGroupGoTestCase(name, runFunc, shutdownFunc, true, true)
}

// newErrGroupGoTestCaseCancel creates a test case expecting cancellation but no error
func newErrGroupGoTestCaseCancel(name string, runFunc func(context.Context) error,
	shutdownFunc func() error) TestCase {
	return newErrGroupGoTestCase(name, runFunc, shutdownFunc, false, true)
}

func makeErrGroupGoTestCases() []TestCase {
	return S(
		newErrGroupGoTestCaseSuccess("successful worker", func(_ context.Context) error {
			return nil
		}, nil),
		newErrGroupGoTestCaseError("worker with error", func(_ context.Context) error {
			return errors.New("worker error")
		}, nil),
		newErrGroupGoTestCaseError("worker with panic", func(_ context.Context) error {
			panic("worker panic")
		}, nil),
		newErrGroupGoTestCaseCancel("successful worker with shutdown", func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		}, func() error {
			return nil
		}), // Manual cancellation will happen
		newErrGroupGoTestCaseCancel("worker with shutdown error", func(_ context.Context) error {
			// Worker runs for a short time then completes
			time.Sleep(5 * time.Millisecond)
			return nil
		}, func() error {
			// Shutdown immediately returns error
			return errors.New("shutdown error")
		}), // Manual cancellation will happen
	)
}

func (tc errGroupGoTestCase) Name() string {
	return tc.name
}

func (tc errGroupGoTestCase) Test(t *testing.T) {
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
	if tc.expectCancel {
		AssertTrue(t, eg.IsCancelled(), "group cancelled")
	}
}

type errGroupGoCatchTestCase struct {
	runFunc     func(context.Context) error
	catchFunc   func(context.Context, error) error
	name        string
	expectError bool
}

// newErrGroupGoCatchTestCase creates a new errGroupGoCatchTestCase
func newErrGroupGoCatchTestCase(name string, runFunc func(context.Context) error,
	catchFunc func(context.Context, error) error, expectError bool) TestCase {
	return errGroupGoCatchTestCase{
		name:        name,
		runFunc:     runFunc,
		catchFunc:   catchFunc,
		expectError: expectError,
	}
}

// newErrGroupGoCatchTestCaseSuccess creates a test case expecting successful completion
func newErrGroupGoCatchTestCaseSuccess(name string, runFunc func(context.Context) error,
	catchFunc func(context.Context, error) error) TestCase {
	return newErrGroupGoCatchTestCase(name, runFunc, catchFunc, false)
}

// newErrGroupGoCatchTestCaseError creates a test case expecting error
func newErrGroupGoCatchTestCaseError(name string, runFunc func(context.Context) error,
	catchFunc func(context.Context, error) error) TestCase {
	return newErrGroupGoCatchTestCase(name, runFunc, catchFunc, true)
}

func makeErrGroupGoCatchTestCases() []TestCase {
	return S(
		newErrGroupGoCatchTestCaseSuccess("successful worker with catch", func(_ context.Context) error {
			return nil
		}, func(_ context.Context, _ error) error {
			// This should never be called for successful workers
			return errors.New("catch should not be called")
		}),
		newErrGroupGoCatchTestCaseSuccess("worker error handled by catch", func(_ context.Context) error {
			return errors.New("worker error")
		}, func(_ context.Context, _ error) error {
			return nil // dismiss error
		}),
		newErrGroupGoCatchTestCaseError("worker error transformed by catch", func(_ context.Context) error {
			return errors.New("original error")
		}, func(_ context.Context, _ error) error {
			return errors.New("transformed error")
		}), // Transformed error should propagate
		newErrGroupGoCatchTestCaseError("nil run function", nil, func(_ context.Context, err error) error {
			return err
		}), // Should panic and be caught
	)
}

func (tc errGroupGoCatchTestCase) Name() string {
	return tc.name
}

func (tc errGroupGoCatchTestCase) Test(t *testing.T) {
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

func TestErrGroup(t *testing.T) {
	t.Run("set defaults", func(t *testing.T) {
		RunTestCases(t, makeErrGroupSetDefaultsTestCases())
	})
	t.Run("go", func(t *testing.T) {
		RunTestCases(t, makeErrGroupGoTestCases())
	})
	t.Run("go catch", func(t *testing.T) {
		RunTestCases(t, makeErrGroupGoCatchTestCases())
	})
	t.Run("cancel", runTestErrGroupCancel)
	t.Run("on error", runTestErrGroupOnError)
	t.Run("context", runTestErrGroupContext)
	t.Run("cancelled", runTestErrGroupCancelled)
	t.Run("done", runTestErrGroupDone)
	t.Run("concurrency", runTestErrGroupConcurrency)
	t.Run("default catcher", runTestErrGroupDefaultErrGroupCatcher)
	t.Run("custom parent", runTestErrGroupWithCustomParent)
}

func runTestErrGroupFirstCancellation(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	cause := errors.New("test error")
	isFirst := eg.Cancel(cause)

	AssertTrue(t, isFirst, "first cancellation")
	AssertTrue(t, eg.IsCancelled(), "cancelled")
	AssertEqual(t, cause, eg.Err(), "error")
}

func runTestErrGroupSubsequentCancellation(t *testing.T) {
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
	AssertSame(t, cause1, eg.Err(), "error instance")
}

func runTestErrGroupNilCause(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	isFirst := eg.Cancel(nil)

	if !isFirst {
		t.Error("Expected first cancellation to return true")
	}

	AssertSame(t, context.Canceled, eg.Err(), "error instance")
}

func runTestErrGroupCancel(t *testing.T) {
	t.Helper()
	t.Run("first cancellation", runTestErrGroupFirstCancellation)
	t.Run("subsequent cancellation", runTestErrGroupSubsequentCancellation)
	t.Run("nil cause", runTestErrGroupNilCause)
}

func runTestErrGroupOnError(t *testing.T) {
	t.Helper()
	var eg ErrGroup
	var errorReceived error

	eg.OnError(func(err error) {
		errorReceived = err
	})

	testErr := errors.New("test error")
	eg.Cancel(testErr)

	// Give time for onError to be called
	time.Sleep(1 * time.Millisecond)

	AssertSame(t, testErr, errorReceived, "error instance")
}

func runTestErrGroupContext(t *testing.T) {
	t.Helper()
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

func runTestErrGroupCancelled(t *testing.T) {
	t.Helper()
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

func runTestErrGroupDone(t *testing.T) {
	t.Helper()
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

func runTestErrGroupConcurrency(t *testing.T) {
	t.Helper()
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

func runTestErrGroupCatcherErrorWhenNotCancelled(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	testErr := errors.New("test error")
	result := eg.defaultErrGroupCatcher(testErr)

	AssertSame(t, testErr, result, "error instance")
}

func runTestErrGroupCatcherErrorWhenCancelled(t *testing.T) {
	t.Helper()
	var eg ErrGroup
	eg.Cancel(errors.New("cancellation error"))

	testErr := errors.New("test error")
	result := eg.defaultErrGroupCatcher(testErr)

	if result != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", result)
	}
}

func runTestErrGroupCatcherNilError(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	result := eg.defaultErrGroupCatcher(nil)

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func runTestErrGroupDefaultErrGroupCatcher(t *testing.T) {
	t.Helper()
	t.Run("error when not cancelled", runTestErrGroupCatcherErrorWhenNotCancelled)
	t.Run("error when cancelled", runTestErrGroupCatcherErrorWhenCancelled)
	t.Run("nil error", runTestErrGroupCatcherNilError)
}

func runTestErrGroupWithCustomParent(t *testing.T) {
	t.Helper()
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
