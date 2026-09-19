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
	_ TestCase = errGroupErrTestCase{}
)

type errGroupSetDefaultsTestCase struct {
	expectedParent context.Context
	setup          func() *ErrGroup
	name           string
}

// newErrGroupSetDefaultsTestCase creates a new errGroupSetDefaultsTestCase
func newErrGroupSetDefaultsTestCase(expectedParent context.Context, name string,
	setup func() *ErrGroup) errGroupSetDefaultsTestCase {
	return errGroupSetDefaultsTestCase{
		name:           name,
		setup:          setup,
		expectedParent: expectedParent,
	}
}

func errGroupSetDefaultsTestCases() []errGroupSetDefaultsTestCase {
	var nilCtx context.Context

	return []errGroupSetDefaultsTestCase{
		newErrGroupSetDefaultsTestCase(context.Background(), "nil parent context", func() *ErrGroup {
			return &ErrGroup{}
		}),
		newErrGroupSetDefaultsTestCase(nilCtx, "custom parent context", func() *ErrGroup {
			type testKey string
			ctx := context.WithValue(context.Background(), testKey("test"), "value")
			return &ErrGroup{Parent: ctx}
		}), // Will be set by test
	}
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

	AssertNotNil(t, eg.ctx, "ctx")
	AssertNotNil(t, eg.cancel, "cancel")
}

func TestErrGroupSetDefaults(t *testing.T) {
	RunTestCases(t, errGroupSetDefaultsTestCases())
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
	shutdownFunc func() error, expectError, expectCancel bool) errGroupGoTestCase {
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
	shutdownFunc func() error) errGroupGoTestCase {
	return newErrGroupGoTestCase(name, runFunc, shutdownFunc, false, false)
}

// newErrGroupGoTestCaseError creates a test case expecting error and cancellation
func newErrGroupGoTestCaseError(name string, runFunc func(context.Context) error,
	shutdownFunc func() error) errGroupGoTestCase {
	return newErrGroupGoTestCase(name, runFunc, shutdownFunc, true, true)
}

// newErrGroupGoTestCaseCancel creates a test case expecting cancellation but no error
func newErrGroupGoTestCaseCancel(name string, runFunc func(context.Context) error,
	shutdownFunc func() error) errGroupGoTestCase {
	return newErrGroupGoTestCase(name, runFunc, shutdownFunc, false, true)
}

var errGroupGoTestCases = []errGroupGoTestCase{
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

func TestErrGroupGo(t *testing.T) {
	RunTestCases(t, errGroupGoTestCases)
}

type errGroupGoCatchTestCase struct {
	runFunc     func(context.Context) error
	catchFunc   func(context.Context, error) error
	name        string
	expectError bool
}

// newErrGroupGoCatchTestCase creates a new errGroupGoCatchTestCase
func newErrGroupGoCatchTestCase(name string, runFunc func(context.Context) error,
	catchFunc func(context.Context, error) error, expectError bool) errGroupGoCatchTestCase {
	return errGroupGoCatchTestCase{
		name:        name,
		runFunc:     runFunc,
		catchFunc:   catchFunc,
		expectError: expectError,
	}
}

// newErrGroupGoCatchTestCaseSuccess creates a test case expecting successful completion
func newErrGroupGoCatchTestCaseSuccess(name string, runFunc func(context.Context) error,
	catchFunc func(context.Context, error) error) errGroupGoCatchTestCase {
	return newErrGroupGoCatchTestCase(name, runFunc, catchFunc, false)
}

// newErrGroupGoCatchTestCaseError creates a test case expecting error
func newErrGroupGoCatchTestCaseError(name string, runFunc func(context.Context) error,
	catchFunc func(context.Context, error) error) errGroupGoCatchTestCase {
	return newErrGroupGoCatchTestCase(name, runFunc, catchFunc, true)
}

var errGroupGoCatchTestCases = []errGroupGoCatchTestCase{
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
		AssertError(t, err, "error")
		return
	}
	AssertNoError(t, err, "error")
}

func TestErrGroupGoCatch(t *testing.T) {
	RunTestCases(t, errGroupGoCatchTestCases)
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
	AssertFalse(t, isFirst, "subsequent cancellation")

	// Should keep the first error
	AssertSame(t, cause1, eg.Err(), "error instance")
}

func testErrGroupNilCause(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	isFirst := eg.Cancel(nil)
	AssertTrue(t, isFirst, "first cancellation")
	AssertSame(t, context.Canceled, eg.Err(), "error instance")
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

	AssertSame(t, testErr, errorReceived, "error instance")
}

func TestErrGroupContext(t *testing.T) {
	var eg ErrGroup

	ctx := eg.Context()
	AssertMustNotNil(t, ctx, "context")

	AssertOpen(t, ctx.Done(), 0, "done before cancel")
	eg.Cancel(errors.New("test"))
	AssertClosed(t, ctx.Done(), time.Millisecond, "done after cancel")
}

func TestErrGroupCancelled(t *testing.T) {
	var eg ErrGroup

	cancelled := eg.Cancelled()
	AssertMustNotNil(t, cancelled, "cancelled channel")

	AssertOpen(t, cancelled, 0, "cancelled before Cancel")
	eg.Cancel(errors.New("test"))
	AssertClosed(t, cancelled, time.Millisecond, "cancelled after Cancel")
}

func TestErrGroupDone(t *testing.T) {
	var eg ErrGroup

	eg.Go(func(_ context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}, nil)

	done := eg.Done()
	AssertOpen(t, done, 5*time.Millisecond, "done while the worker runs")
	AssertClosed(t, done, 50*time.Millisecond, "done")
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
	for worker := range numWorkers {
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

	AssertSame(t, testErr, result, "error instance")
}

func testErrGroupCatcherErrorWhenCancelled(t *testing.T) {
	t.Helper()
	var eg ErrGroup
	eg.Cancel(errors.New("cancellation error"))

	testErr := errors.New("test error")
	result := eg.defaultErrGroupCatcher(testErr)
	AssertSame(t, context.Canceled, result, "error instance")
}

func testErrGroupCatcherNilError(t *testing.T) {
	t.Helper()
	var eg ErrGroup

	result := eg.defaultErrGroupCatcher(nil)
	AssertNil(t, result, "result")
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
	AssertError(t, err, "parent timeout")
}

// errGroupErrTestCase covers ErrGroup.Err() across the pristine and
// worker-error branches. A nil workerErr means no worker is launched
// and Err() should return nil; a non-nil workerErr is returned by the
// launched worker and Err() should propagate it.
type errGroupErrTestCase struct {
	workerErr error
	name      string
}

func newErrGroupErrTestCase(name string, workerErr error) errGroupErrTestCase {
	return errGroupErrTestCase{name: name, workerErr: workerErr}
}

func (tc errGroupErrTestCase) Name() string { return tc.name }

func (tc errGroupErrTestCase) Test(t *testing.T) {
	t.Helper()
	var eg ErrGroup
	if tc.workerErr != nil {
		eg.Go(func(_ context.Context) error { return tc.workerErr }, nil)
		_ = eg.Wait()
	}
	if tc.workerErr == nil {
		AssertNil(t, eg.Err(), tc.name)
		return
	}
	AssertSame(t, tc.workerErr, eg.Err(), tc.name)
}

func errGroupErrTestCases() []errGroupErrTestCase {
	return []errGroupErrTestCase{
		newErrGroupErrTestCase("pristine", nil),
		newErrGroupErrTestCase("worker error", errors.New("worker")),
	}
}

func TestErrGroupErr(t *testing.T) {
	RunTestCases(t, errGroupErrTestCases())
}
