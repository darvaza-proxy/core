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

type waitGroupGoTestCase struct {
	fn          func() error
	errorMsg    string
	name        string
	expectError bool
}

// Factory function for waitGroupGoTestCase
func newWaitGroupGoTestCase(name string, fn func() error, expectError bool, errorMsg string) waitGroupGoTestCase {
	return waitGroupGoTestCase{
		name:        name,
		fn:          fn,
		errorMsg:    errorMsg,
		expectError: expectError,
	}
}

var waitGroupGoTestCases = []waitGroupGoTestCase{
	newWaitGroupGoTestCase("successful worker", func() error {
		return nil
	}, false, ""),
	newWaitGroupGoTestCase("worker with error", func() error {
		return errors.New("worker error")
	}, true, "worker error"),
	newWaitGroupGoTestCase("worker with panic", func() error {
		panic("worker panic")
	}, true, ""), // Panic should be caught and converted to error
	newWaitGroupGoTestCase("nil function", nil, false, ""),
}

func (tc waitGroupGoTestCase) Name() string {
	return tc.name
}

func (tc waitGroupGoTestCase) Test(t *testing.T) {
	t.Helper()

	var wg WaitGroup
	wg.Go(tc.fn)
	err := wg.Wait()

	tc.handleAsyncError(t, &wg, &err)
	tc.validateResult(t, err)
}

func (tc waitGroupGoTestCase) handleAsyncError(t *testing.T, wg *WaitGroup, err *error) {
	t.Helper()
	// Give a small delay for error reporting in case of async processing
	if tc.expectError && *err == nil {
		time.Sleep(1 * time.Millisecond)
		*err = wg.Err()
	}
}

func (tc waitGroupGoTestCase) validateResult(t *testing.T, err error) {
	t.Helper()
	if tc.expectError {
		AssertError(t, err, "wait group error")
		if tc.errorMsg != "" {
			AssertEqual(t, tc.errorMsg, err.Error(), "error message")
		}
	} else {
		AssertNoError(t, err, "wait group")
	}
}

func TestWaitGroupGo(t *testing.T) {
	RunTestCases(t, waitGroupGoTestCases)
}

type waitGroupGoCatchTestCase struct {
	fn          func() error
	catch       func(error) error
	name        string
	errorMsg    string
	expectError bool
}

// Factory function for waitGroupGoCatchTestCase
func newWaitGroupGoCatchTestCase(name string, fn func() error, catch func(error) error,
	expectError bool, errorMsg string) waitGroupGoCatchTestCase {
	return waitGroupGoCatchTestCase{
		name:        name,
		fn:          fn,
		catch:       catch,
		errorMsg:    errorMsg,
		expectError: expectError,
	}
}

var waitGroupGoCatchTestCases = []waitGroupGoCatchTestCase{
	newWaitGroupGoCatchTestCase("successful worker with catch", func() error {
		return nil
	}, func(_ error) error {
		return nil
	}, false, ""),
	newWaitGroupGoCatchTestCase("worker error handled by catch", func() error {
		return errors.New("worker error")
	}, func(_ error) error {
		return nil // catch dismisses the error
	}, false, ""),
	newWaitGroupGoCatchTestCase("worker error transformed by catch", func() error {
		return errors.New("original error")
	}, func(_ error) error {
		return errors.New("transformed error")
	}, true, "transformed error"),
	newWaitGroupGoCatchTestCase("catch function panics", func() error {
		return errors.New("worker error")
	}, func(_ error) error {
		panic("catch panic")
	}, true, ""), // Don't check exact message as panic handling may vary
	newWaitGroupGoCatchTestCase("nil function with catch", nil, func(err error) error {
		return err
	}, false, ""),
	newWaitGroupGoCatchTestCase("worker error with nil catch", func() error {
		return errors.New("worker error")
	}, nil, true, "worker error"),
}

func (tc waitGroupGoCatchTestCase) Name() string {
	return tc.name
}

func (tc waitGroupGoCatchTestCase) Test(t *testing.T) {
	t.Helper()

	var wg WaitGroup
	wg.GoCatch(tc.fn, tc.catch)
	err := wg.Wait()

	tc.validateGoCatchResult(t, err)
}

func (tc waitGroupGoCatchTestCase) validateGoCatchResult(t *testing.T, err error) {
	t.Helper()
	if tc.expectError {
		AssertError(t, err, "wait group catch error")
		if tc.errorMsg != "" {
			AssertEqual(t, tc.errorMsg, err.Error(), "error message")
		}
	} else {
		AssertNoError(t, err, "wait group catch")
	}
}

func TestWaitGroupGoCatch(t *testing.T) {
	RunTestCases(t, waitGroupGoCatchTestCases)
}

type waitGroupOnErrorTestCase struct {
	onErrorHandler func(error) error
	errorMsg       string
	name           string
	workers        []func() error
	expectError    bool
}

// Factory function for waitGroupOnErrorTestCase
func newWaitGroupOnErrorTestCase(name string, workers []func() error,
	onErrorHandler func(error) error, expectError bool, errorMsg string) waitGroupOnErrorTestCase {
	return waitGroupOnErrorTestCase{
		name:           name,
		workers:        workers,
		onErrorHandler: onErrorHandler,
		errorMsg:       errorMsg,
		expectError:    expectError,
	}
}

var waitGroupOnErrorTestCases = []waitGroupOnErrorTestCase{
	newWaitGroupOnErrorTestCase("successful workers with onError", []func() error{
		func() error { return nil },
		func() error { return nil },
	}, func(err error) error {
		return err
	}, false, ""),
	newWaitGroupOnErrorTestCase("error dismissed by onError filter", []func() error{
		func() error { return errors.New("worker error") },
	}, func(_ error) error {
		return nil // onError filter dismisses the error
	}, false, ""),
	newWaitGroupOnErrorTestCase("error transformed by onError filter", []func() error{
		func() error { return errors.New("original error") },
	}, func(_ error) error {
		return errors.New("filtered error")
	}, true, "filtered error"),
}

func (tc waitGroupOnErrorTestCase) Name() string {
	return tc.name
}

func (tc waitGroupOnErrorTestCase) Test(t *testing.T) {
	t.Helper()

	var wg WaitGroup
	wg.OnError(tc.onErrorHandler)

	tc.runWorkers(t, &wg)
	err := wg.Wait()
	tc.handleOnErrorAsync(t, &wg, &err)
	tc.validateOnErrorResult(t, err)
}

func (tc waitGroupOnErrorTestCase) runWorkers(t *testing.T, wg *WaitGroup) {
	t.Helper()
	for _, worker := range tc.workers {
		wg.Go(worker)
	}
}

func (tc waitGroupOnErrorTestCase) handleOnErrorAsync(t *testing.T, wg *WaitGroup, err *error) {
	t.Helper()
	// Give a small delay for error processing in case of async handling
	if tc.expectError && *err == nil {
		time.Sleep(1 * time.Millisecond)
		*err = wg.Err()
	}
}

func (tc waitGroupOnErrorTestCase) validateOnErrorResult(t *testing.T, err error) {
	t.Helper()
	if tc.expectError {
		AssertError(t, err, "wait group on error")
		if tc.errorMsg != "" {
			AssertEqual(t, tc.errorMsg, err.Error(), "error message")
		}
	} else {
		AssertNoError(t, err, "wait group on error")
	}
}

func TestWaitGroupOnError(t *testing.T) {
	RunTestCases(t, waitGroupOnErrorTestCases)
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

	for i := 0; i < numWorkers; i++ {
		wg.Go(func() error {
			for j := 0; j < numIterations; j++ {
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

func TestWaitGroupFirstErrorWins(t *testing.T) {
	var wg WaitGroup

	// Start multiple workers with errors
	wg.Go(func() error {
		time.Sleep(10 * time.Millisecond)
		return errors.New("error 1")
	})
	wg.Go(func() error {
		time.Sleep(5 * time.Millisecond)
		return errors.New("error 2")
	})
	wg.Go(func() error {
		time.Sleep(15 * time.Millisecond)
		return errors.New("error 3")
	})

	err := wg.Wait()
	if err == nil {
		t.Error("Expected error but got nil")
	}

	// The exact error returned depends on timing, but it should be one of them
	errMsg := err.Error()
	if errMsg != "error 1" && errMsg != "error 2" && errMsg != "error 3" {
		t.Errorf("Unexpected error message: %s", errMsg)
	}
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
