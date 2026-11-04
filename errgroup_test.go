package core_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ core.TestCase = errGroupGoTestCase{}
	_ core.TestCase = errGroupGoCatchTestCase{}
	_ core.TestCase = errGroupErrTestCase{}
)

// errGroupGoTestCase states what a worker's outcome does to the group:
// whether Wait reports an error, and whether the group ends cancelled.
// cancelManually marks the rows cancelled from outside rather than by
// their own worker failing.
type errGroupGoTestCase struct {
	runFunc        func(context.Context) error
	shutdownFunc   func() error
	name           string
	expectError    bool
	expectCancel   bool
	cancelManually bool
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

		cancelManually: false,
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

// newErrGroupGoTestCaseCancel creates a test case cancelled from outside
// the group part-way through, rather than by its own worker failing.
func newErrGroupGoTestCaseCancel(name string, runFunc func(context.Context) error,
	shutdownFunc func() error) errGroupGoTestCase {
	return errGroupGoTestCase{
		name:         name,
		runFunc:      runFunc,
		shutdownFunc: shutdownFunc,
		expectError:  false,
		expectCancel: true,

		cancelManually: true,
	}
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

	var eg core.ErrGroup
	eg.Go(tc.runFunc, tc.shutdownFunc)

	if tc.cancelManually {
		go func() {
			time.Sleep(10 * time.Millisecond)
			eg.Cancel(errors.New("manual cancellation"))
		}()
	}

	err := eg.Wait()
	if tc.expectError {
		core.AssertError(t, err, "error")
	} else {
		core.AssertNoError(t, err, "no error")
	}

	core.AssertEqual(t, tc.expectCancel, eg.IsCancelled(), "cancelled")
}

func TestErrGroupGo(t *testing.T) {
	core.RunTestCases(t, errGroupGoTestCases)
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
}

func (tc errGroupGoCatchTestCase) Name() string {
	return tc.name
}

func (tc errGroupGoCatchTestCase) Test(t *testing.T) {
	t.Helper()

	var eg core.ErrGroup
	eg.GoCatch(tc.runFunc, tc.catchFunc)

	err := eg.Wait()
	if tc.expectError {
		core.AssertError(t, err, "error")
		return
	}
	core.AssertNoError(t, err, "error")
}

func TestErrGroupGoCatch(t *testing.T) {
	core.RunTestCases(t, errGroupGoCatchTestCases)
}

// GoCatch refuses a nil run function outright rather than spawning a
// worker that cannot run.
func TestErrGroupGoCatchNilFunction(t *testing.T) {
	var eg core.ErrGroup

	core.AssertPanic(t, func() {
		eg.GoCatch(nil, func(_ context.Context, err error) error { return err })
	}, core.ErrInvalid, "panic")
}

func testErrGroupFirstCancellation(t *testing.T) {
	t.Helper()
	var eg core.ErrGroup

	cause := errors.New("test error")
	isFirst := eg.Cancel(cause)

	core.AssertTrue(t, isFirst, "first cancellation")
	core.AssertTrue(t, eg.IsCancelled(), "cancelled")
	core.AssertEqual(t, cause, eg.Err(), "error")
}

func testErrGroupSubsequentCancellation(t *testing.T) {
	t.Helper()
	var eg core.ErrGroup

	// First cancellation
	cause1 := errors.New("first error")
	eg.Cancel(cause1)

	// Second cancellation
	cause2 := errors.New("second error")
	isFirst := eg.Cancel(cause2)
	core.AssertFalse(t, isFirst, "subsequent cancellation")

	// Should keep the first error
	core.AssertSame(t, cause1, eg.Err(), "error instance")
}

func testErrGroupNilCause(t *testing.T) {
	t.Helper()
	var eg core.ErrGroup

	isFirst := eg.Cancel(nil)
	core.AssertTrue(t, isFirst, "first cancellation")
	core.AssertSame(t, context.Canceled, eg.Err(), "error instance")
}

func TestErrGroupCancel(t *testing.T) {
	t.Run("first cancellation", testErrGroupFirstCancellation)
	t.Run("subsequent cancellation", testErrGroupSubsequentCancellation)
	t.Run("nil cause", testErrGroupNilCause)
}

func TestErrGroupOnError(t *testing.T) {
	var eg core.ErrGroup
	var errorReceived error

	eg.OnError(func(err error) {
		errorReceived = err
	})

	testErr := errors.New("test error")
	eg.Cancel(testErr)

	// Give time for onError to be called
	time.Sleep(1 * time.Millisecond)

	core.AssertSame(t, testErr, errorReceived, "error instance")
}

func TestErrGroupContext(t *testing.T) {
	var eg core.ErrGroup

	ctx := eg.Context()
	core.AssertMustNotNil(t, ctx, "context")

	core.AssertOpen(t, ctx.Done(), 0, "done before cancel")
	eg.Cancel(errors.New("test"))
	core.AssertClosed(t, ctx.Done(), time.Millisecond, "done after cancel")
}

func TestErrGroupCancelled(t *testing.T) {
	var eg core.ErrGroup

	cancelled := eg.Cancelled()
	core.AssertMustNotNil(t, cancelled, "cancelled channel")

	core.AssertOpen(t, cancelled, 0, "cancelled before Cancel")
	eg.Cancel(errors.New("test"))
	core.AssertClosed(t, cancelled, time.Millisecond, "cancelled after Cancel")
}

func TestErrGroupDone(t *testing.T) {
	var eg core.ErrGroup

	eg.Go(func(_ context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}, nil)

	done := eg.Done()
	core.AssertOpen(t, done, 5*time.Millisecond, "done while the worker runs")
	core.AssertClosed(t, done, 50*time.Millisecond, "done")
}

func TestErrGroupConcurrency(t *testing.T) {
	const numWorkers = 10

	var eg core.ErrGroup
	startConcurrentWorkers(t, &eg, numWorkers)

	err := eg.Wait()
	validateConcurrencyResult(t, &eg, err)
}

func startConcurrentWorkers(t *testing.T, eg *core.ErrGroup, numWorkers int) {
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

func validateConcurrencyResult(t *testing.T, eg *core.ErrGroup, err error) {
	t.Helper()
	core.AssertError(t, err, "worker 5 error")
	core.AssertTrue(t, eg.IsCancelled(), "group cancelled")
}

func TestErrGroupWithCustomParent(t *testing.T) {
	parentCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	eg := core.ErrGroup{Parent: parentCtx}

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
	core.AssertError(t, err, "parent timeout")
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
	var eg core.ErrGroup
	if tc.workerErr != nil {
		eg.Go(func(_ context.Context) error { return tc.workerErr }, nil)
		_ = eg.Wait()
	}
	if tc.workerErr == nil {
		core.AssertNil(t, eg.Err(), tc.name)
		return
	}
	core.AssertSame(t, tc.workerErr, eg.Err(), tc.name)
}

func errGroupErrTestCases() []errGroupErrTestCase {
	return []errGroupErrTestCase{
		newErrGroupErrTestCase("pristine", nil),
		newErrGroupErrTestCase("worker error", errors.New("worker")),
	}
}

func TestErrGroupErr(t *testing.T) {
	core.RunTestCases(t, errGroupErrTestCases())
}
