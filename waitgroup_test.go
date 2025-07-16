package core

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type waitGroupGoTestCase struct {
	fn          func() error
	errorMsg    string
	name        string
	expectError bool
}

var waitGroupGoTestCases = []waitGroupGoTestCase{
	{
		name: "successful worker",
		fn: func() error {
			return nil
		},
		expectError: false,
	},
	{
		name: "worker with error",
		fn: func() error {
			return errors.New("worker error")
		},
		expectError: true,
		errorMsg:    "worker error",
	},
	{
		name: "worker with panic",
		fn: func() error {
			panic("worker panic")
		},
		expectError: true,
		errorMsg:    "", // Panic should be caught and converted to error
	},
	{
		name:        "nil function",
		fn:          nil,
		expectError: false,
	},
}

//revive:disable-next-line:cognitive-complexity
func (tc waitGroupGoTestCase) test(t *testing.T) {
	t.Helper()

	var wg WaitGroup

	wg.Go(tc.fn)
	err := wg.Wait()

	// Give a small delay for error reporting in case of async processing
	if tc.expectError && err == nil {
		time.Sleep(1 * time.Millisecond)
		err = wg.Err()
	}

	if tc.expectError {
		if err == nil {
			t.Error("Expected error but got nil")
		} else if tc.errorMsg != "" && err.Error() != tc.errorMsg {
			t.Errorf("Expected error message '%s', got '%s'", tc.errorMsg, err.Error())
		}
	} else {
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	}
}

func TestWaitGroupGo(t *testing.T) {
	for _, tc := range waitGroupGoTestCases {
		t.Run(tc.name, tc.test)
	}
}

type waitGroupGoCatchTestCase struct {
	fn          func() error
	catch       func(error) error
	name        string
	errorMsg    string
	expectError bool
}

var waitGroupGoCatchTestCases = []waitGroupGoCatchTestCase{
	{
		name: "successful worker with catch",
		fn: func() error {
			return nil
		},
		catch: func(_ error) error {
			return nil
		},
		expectError: false,
	},
	{
		name: "worker error handled by catch",
		fn: func() error {
			return errors.New("worker error")
		},
		catch: func(_ error) error {
			return nil // catch dismisses the error
		},
		expectError: false,
	},
	{
		name: "worker error transformed by catch",
		fn: func() error {
			return errors.New("original error")
		},
		catch: func(_ error) error {
			return errors.New("transformed error")
		},
		expectError: true,
		errorMsg:    "transformed error",
	},
	{
		name: "catch function panics",
		fn: func() error {
			return errors.New("worker error")
		},
		catch: func(_ error) error {
			panic("catch panic")
		},
		expectError: true,
		errorMsg:    "", // Don't check exact message as panic handling may vary
	},
	{
		name: "nil function with catch",
		fn:   nil,
		catch: func(err error) error {
			return err
		},
		expectError: false,
	},
	{
		name: "worker error with nil catch",
		fn: func() error {
			return errors.New("worker error")
		},
		catch:       nil,
		expectError: true,
		errorMsg:    "worker error",
	},
}

//revive:disable-next-line:cognitive-complexity
func (tc waitGroupGoCatchTestCase) test(t *testing.T) {
	t.Helper()

	var wg WaitGroup

	wg.GoCatch(tc.fn, tc.catch)
	err := wg.Wait()

	if tc.expectError {
		if err == nil {
			t.Error("Expected error but got nil")
		} else if tc.errorMsg != "" && err.Error() != tc.errorMsg {
			t.Errorf("Expected error message '%s', got '%s'", tc.errorMsg, err.Error())
		}
	} else {
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	}
}

func TestWaitGroupGoCatch(t *testing.T) {
	for _, tc := range waitGroupGoCatchTestCases {
		t.Run(tc.name, tc.test)
	}
}

type waitGroupOnErrorTestCase struct {
	onErrorHandler func(error) error
	errorMsg       string
	name           string
	workers        []func() error
	expectError    bool
}

var waitGroupOnErrorTestCases = []waitGroupOnErrorTestCase{
	{
		name: "successful workers with onError",
		workers: []func() error{
			func() error { return nil },
			func() error { return nil },
		},
		onErrorHandler: func(err error) error {
			return err
		},
		expectError: false,
	},
	{
		name: "error dismissed by onError filter",
		workers: []func() error{
			func() error { return errors.New("worker error") },
		},
		onErrorHandler: func(_ error) error {
			return nil // onError filter dismisses the error
		},
		expectError: false,
	},
	{
		name: "error transformed by onError filter",
		workers: []func() error{
			func() error { return errors.New("original error") },
		},
		onErrorHandler: func(_ error) error {
			return errors.New("filtered error")
		},
		expectError: true,
		errorMsg:    "filtered error",
	},
}

//revive:disable-next-line:cognitive-complexity
func (tc waitGroupOnErrorTestCase) test(t *testing.T) {
	t.Helper()

	var wg WaitGroup
	wg.OnError(tc.onErrorHandler)

	for _, worker := range tc.workers {
		wg.Go(worker)
	}

	err := wg.Wait()

	// Give a small delay for error processing in case of async handling
	if tc.expectError && err == nil {
		time.Sleep(1 * time.Millisecond)
		err = wg.Err()
	}

	if tc.expectError {
		if err == nil {
			t.Error("Expected error but got nil")
		} else if tc.errorMsg != "" && err.Error() != tc.errorMsg {
			t.Errorf("Expected error message '%s', got '%s'", tc.errorMsg, err.Error())
		}
	} else {
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	}
}

func TestWaitGroupOnError(t *testing.T) {
	for _, tc := range waitGroupOnErrorTestCases {
		t.Run(tc.name, tc.test)
	}
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
		t.Error("Done channel closed too early")
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

//revive:disable-next-line:cognitive-complexity
func TestWaitGroupErr(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		var wg WaitGroup
		wg.Go(func() error { return nil })
		if err := wg.Wait(); err != nil {
			t.Errorf("Expected no error from Wait(), got: %v", err)
		}

		if err := wg.Err(); err != nil {
			t.Errorf("Expected nil error, got: %v", err)
		}
	})

	t.Run("with error", func(t *testing.T) {
		var wg WaitGroup
		expectedErr := errors.New("test error")
		wg.Go(func() error { return expectedErr })
		if err := wg.Wait(); err == nil {
			t.Error("Expected error from Wait() but got nil")
		}

		if err := wg.Err(); err == nil {
			t.Error("Expected error but got nil")
		} else if err.Error() != expectedErr.Error() {
			t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
		}
	})
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
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := int64(numWorkers * numIterations)
	if counter != expected {
		t.Errorf("Expected counter %d, got %d", expected, counter)
	}
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
