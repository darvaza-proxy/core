package core

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = waitGroupGoTestCase{}
	_ TestCase = waitGroupGoCatchTestCase{}
	_ TestCase = waitGroupOnErrorTestCase{}
)

// waitGroupGoTestCase states what Wait reports once a worker is done:
// the error it ended with, nil where it ended cleanly. A worker that
// panics is reported through the error it panicked with.
type waitGroupGoTestCase struct {
	fn      func() error
	wantErr error
	name    string
}

// Factory function for waitGroupGoTestCase
func newWaitGroupGoTestCase(name string, fn func() error, wantErr error) waitGroupGoTestCase {
	return waitGroupGoTestCase{
		fn:      fn,
		wantErr: wantErr,
		name:    name,
	}
}

func waitGroupGoTestCases() []waitGroupGoTestCase {
	workerErr := errors.New("worker error")
	panicErr := errors.New("worker panic")

	return S(
		newWaitGroupGoTestCase("successful worker", func() error {
			return nil
		}, nil),
		newWaitGroupGoTestCase("worker with error", func() error {
			return workerErr
		}, workerErr),
		newWaitGroupGoTestCase("worker with panic", func() error {
			panic(panicErr)
		}, panicErr),
		newWaitGroupGoTestCase("nil function", nil, nil),
	)
}

func (tc waitGroupGoTestCase) Name() string {
	return tc.name
}

func (tc waitGroupGoTestCase) Test(t *testing.T) {
	t.Helper()

	var wg WaitGroup
	wg.Go(tc.fn)

	AssertErrorIs(t, wg.Wait(), tc.wantErr, "error")
}

func TestWaitGroupGo(t *testing.T) {
	RunTestCases(t, waitGroupGoTestCases())
}

// A worker that panics is reported as a *PanicError carrying what it
// panicked with and the stack it was caught on.
func TestWaitGroupGoPanic(t *testing.T) {
	var wg WaitGroup
	wg.Go(func() error {
		// False positive: the rule guards sync.WaitGroup, and this
		// WaitGroup recovering the panic is what is under test.
		//revive:disable-next-line:forbidden-call-in-wg-go
		panic("worker panic")
	})

	pe := AssertMustErrorAs[*PanicError](t, wg.Wait(), "error")
	payload := AssertMustTypeIs[error](t, pe.Recovered(), "payload is an error")
	AssertEqual(t, "worker panic", payload.Error(), "payload")
	AssertTrue(t, len(pe.CallStack()) > 0, "stack captured")
}

// waitGroupGoCatchTestCase states what the catch function leaves for
// Wait to report: nil where it dismissed the worker's error, and the
// error it returned or panicked with otherwise.
type waitGroupGoCatchTestCase struct {
	fn      func() error
	catch   func(error) error
	wantErr error
	name    string
}

// Factory function for waitGroupGoCatchTestCase
func newWaitGroupGoCatchTestCase(name string, fn func() error, catch func(error) error,
	wantErr error) waitGroupGoCatchTestCase {
	return waitGroupGoCatchTestCase{
		fn:      fn,
		catch:   catch,
		wantErr: wantErr,
		name:    name,
	}
}

func waitGroupGoCatchTestCases() []waitGroupGoCatchTestCase {
	workerErr := errors.New("worker error")
	transformedErr := errors.New("transformed error")
	panicErr := errors.New("catch panic")

	return S(
		newWaitGroupGoCatchTestCase("successful worker with catch", func() error {
			return nil
		}, func(_ error) error {
			return nil
		}, nil),
		newWaitGroupGoCatchTestCase("worker error handled by catch", func() error {
			return workerErr
		}, func(_ error) error {
			return nil // catch dismisses the error
		}, nil),
		newWaitGroupGoCatchTestCase("worker error transformed by catch", func() error {
			return workerErr
		}, func(_ error) error {
			return transformedErr
		}, transformedErr),
		newWaitGroupGoCatchTestCase("catch function panics", func() error {
			return workerErr
		}, func(_ error) error {
			panic(panicErr)
		}, panicErr),
		newWaitGroupGoCatchTestCase("nil function with catch", nil, func(err error) error {
			return err
		}, nil),
		newWaitGroupGoCatchTestCase("worker error with nil catch", func() error {
			return workerErr
		}, nil, workerErr),
	)
}

func (tc waitGroupGoCatchTestCase) Name() string {
	return tc.name
}

func (tc waitGroupGoCatchTestCase) Test(t *testing.T) {
	t.Helper()

	var wg WaitGroup
	wg.GoCatch(tc.fn, tc.catch)

	AssertErrorIs(t, wg.Wait(), tc.wantErr, "error")
}

func TestWaitGroupGoCatch(t *testing.T) {
	RunTestCases(t, waitGroupGoCatchTestCases())
}

// A catch function that panics is reported the same way as a worker
// that does.
func TestWaitGroupGoCatchPanic(t *testing.T) {
	var wg WaitGroup
	wg.GoCatch(func() error {
		return errors.New("worker error")
	}, func(_ error) error {
		panic("catch panic")
	})

	pe := AssertMustErrorAs[*PanicError](t, wg.Wait(), "error")
	payload := AssertMustTypeIs[error](t, pe.Recovered(), "payload is an error")
	AssertEqual(t, "catch panic", payload.Error(), "payload")
	AssertTrue(t, len(pe.CallStack()) > 0, "stack captured")
}

// waitGroupOnErrorTestCase states what the OnError filter leaves for
// Wait to report across a set of workers: nil where it dismissed what
// it was handed, and the error it returned otherwise.
type waitGroupOnErrorTestCase struct {
	onErrorHandler func(error) error
	wantErr        error
	name           string
	workers        []func() error
}

// Factory function for waitGroupOnErrorTestCase
func newWaitGroupOnErrorTestCase(name string, workers []func() error,
	onErrorHandler func(error) error, wantErr error) waitGroupOnErrorTestCase {
	return waitGroupOnErrorTestCase{
		onErrorHandler: onErrorHandler,
		wantErr:        wantErr,
		name:           name,
		workers:        workers,
	}
}

func waitGroupOnErrorTestCases() []waitGroupOnErrorTestCase {
	workerErr := errors.New("worker error")
	filteredErr := errors.New("filtered error")

	return S(
		newWaitGroupOnErrorTestCase("successful workers with onError", S(
			func() error { return nil },
			func() error { return nil },
		), func(err error) error {
			return err
		}, nil),
		newWaitGroupOnErrorTestCase("error dismissed by onError filter", S(
			func() error { return workerErr },
		), func(_ error) error {
			return nil // onError filter dismisses the error
		}, nil),
		newWaitGroupOnErrorTestCase("error transformed by onError filter", S(
			func() error { return workerErr },
		), func(_ error) error {
			return filteredErr
		}, filteredErr),
	)
}

func (tc waitGroupOnErrorTestCase) Name() string {
	return tc.name
}

func (tc waitGroupOnErrorTestCase) Test(t *testing.T) {
	t.Helper()

	var wg WaitGroup
	wg.OnError(tc.onErrorHandler)

	for _, worker := range tc.workers {
		wg.Go(worker)
	}

	AssertErrorIs(t, wg.Wait(), tc.wantErr, "error")
}

func TestWaitGroupOnError(t *testing.T) {
	RunTestCases(t, waitGroupOnErrorTestCases())
}

func TestWaitGroupDone(t *testing.T) {
	var wg WaitGroup

	// Test with successful workers
	wg.Go(func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	wg.Go(func() error {
		time.Sleep(20 * time.Millisecond)
		return nil
	})

	done := wg.Done()

	// Should not be closed yet
	select {
	case <-done:
		AssertTrue(t, false, "done channel timing")
	case <-time.After(5 * time.Millisecond):
		// Expected
	}

	// Wait for completion
	select {
	case <-done:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Done channel never closed")
	}

	// Verify Wait() also works
	if err := wg.Wait(); err != nil {
		t.Errorf("Expected no error from Wait(), got: %v", err)
	}
}

func testWaitGroupErrNoError(t *testing.T) {
	t.Helper()
	var wg WaitGroup
	wg.Go(func() error { return nil })
	AssertNoError(t, wg.Wait(), "wait")
	AssertNil(t, wg.Err(), "error")
}

func testWaitGroupErrWithError(t *testing.T) {
	t.Helper()
	var wg WaitGroup
	expectedErr := errors.New("test error")
	wg.Go(func() error { return expectedErr })
	AssertError(t, wg.Wait(), "wait error")
	AssertError(t, wg.Err(), "error")
	AssertEqual(t, expectedErr.Error(), wg.Err().Error(), "error message")
}

func TestWaitGroupErr(t *testing.T) {
	t.Run("no error", testWaitGroupErrNoError)
	t.Run("with error", testWaitGroupErrWithError)
}

func TestWaitGroupConcurrency(t *testing.T) {
	const numWorkers = 50
	const numIterations = 100

	var wg WaitGroup
	var counter int64
	var mu sync.Mutex

	for range numWorkers {
		wg.Go(func() error {
			for range numIterations {
				mu.Lock()
				counter++
				mu.Unlock()
			}
			return nil
		})
	}

	err := wg.Wait()
	AssertNoError(t, err, "concurrent workers")

	expected := int64(numWorkers * numIterations)
	AssertEqual(t, expected, counter, "counter value")
}

// The first error stored survives: the workers failing after it, however
// many at once and whatever their type, do not replace it. The first
// worker runs to Wait on its own so that which error is first is not
// left to timing.
func TestWaitGroupFirstErrorWins(t *testing.T) {
	first := errors.New("first")
	later := errors.New("later")

	var wg WaitGroup
	wg.Go(func() error { return first })
	AssertErrorIs(t, wg.Wait(), first, "first error")

	for _, err := range S[error](
		later,
		StringError("later"),
		Wrap(later, "wrapped"),
		NewTimeoutError(later),
		NewCompoundError(later),
		NewPanicError(0, later),
	) {
		wg.Go(func() error { return err })
	}
	AssertErrorIs(t, wg.Wait(), first, "first error kept")
}

func TestWaitGroupWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var wg WaitGroup

	// Worker that takes longer than context timeout
	wg.Go(func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	// Worker that checks context
	wg.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return nil
		}
	})

	err := wg.Wait()
	if err == nil {
		t.Error("Expected context timeout error")
	}
}

func TestWaitGroupOnErrorCalledForAllErrors(t *testing.T) {
	var wg WaitGroup
	var callCount int
	var mu sync.Mutex

	// Set up onError handler that counts calls
	wg.OnError(func(err error) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return err
	})

	// Create 3 workers that all return errors
	wg.Go(func() error {
		return errors.New("error 1")
	})
	wg.Go(func() error {
		return errors.New("error 2")
	})
	wg.Go(func() error {
		return errors.New("error 3")
	})

	// Wait for all workers to complete
	err := wg.Wait()
	if err == nil {
		t.Error("Expected an error but got nil")
	}

	// Check how many times onError was called
	mu.Lock()
	count := callCount
	mu.Unlock()

	if count != 3 {
		t.Errorf("Expected onError to be called 3 times, but it was called %d times", count)
	}
}

func TestWaitGroupOnErrorCalledForMixedResults(t *testing.T) {
	var wg WaitGroup
	var callCount int
	var errorMessages []string
	var mu sync.Mutex

	// Set up onError handler that counts calls and collects error messages
	wg.OnError(func(err error) error {
		mu.Lock()
		callCount++
		errorMessages = append(errorMessages, err.Error())
		mu.Unlock()
		return err
	})

	// Create workers with mixed results
	wg.Go(func() error {
		return nil // success
	})
	wg.Go(func() error {
		return errors.New("error A")
	})
	wg.Go(func() error {
		return nil // success
	})
	wg.Go(func() error {
		return errors.New("error B")
	})

	// Wait for all workers to complete
	err := wg.Wait()
	if err == nil {
		t.Error("Expected an error but got nil")
	}

	// Check how many times onError was called
	mu.Lock()
	count := callCount
	mu.Unlock()

	if count != 2 {
		t.Errorf("Expected onError to be called 2 times (for 2 errors), but it was called %d times", count)
	}
}
