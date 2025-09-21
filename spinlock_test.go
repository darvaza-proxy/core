package core

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = spinLockTryLockTestCase{}
	_ TestCase = spinLockLockTestCase{}
	_ TestCase = spinLockUnlockTestCase{}
)

type spinLockTryLockTestCase struct {
	setup    func() *SpinLock
	name     string
	expected bool
}

// Factory function for spinLockTryLockTestCase
func newSpinLockTryLockTestCase(name string, setup func() *SpinLock, expected bool) TestCase {
	return spinLockTryLockTestCase{
		name:     name,
		setup:    setup,
		expected: expected,
	}
}

func makeSpinLockTryLockTestCases() []TestCase {
	return S(
		newSpinLockTryLockTestCase("unlocked spinlock", func() *SpinLock {
			return new(SpinLock)
		}, true),
		newSpinLockTryLockTestCase("already locked spinlock", func() *SpinLock {
			sl := new(SpinLock)
			sl.TryLock()
			return sl
		}, false),
	)
}

func (tc spinLockTryLockTestCase) Name() string {
	return tc.name
}

func (tc spinLockTryLockTestCase) Test(t *testing.T) {
	t.Helper()

	sl := tc.setup()
	result := sl.TryLock()

	AssertEqual(t, tc.expected, result, "TryLock")
}

func TestSpinLockTryLock(t *testing.T) {
	RunTestCases(t, makeSpinLockTryLockTestCases())
}

type spinLockLockTestCase struct {
	setup func() *SpinLock
	name  string
}

// Factory function for spinLockLockTestCase
func newSpinLockLockTestCase(name string, setup func() *SpinLock) TestCase {
	return spinLockLockTestCase{
		name:  name,
		setup: setup,
	}
}

func makeSpinLockLockTestCases() []TestCase {
	return S(
		newSpinLockLockTestCase("unlocked spinlock", func() *SpinLock {
			return new(SpinLock)
		}),
		newSpinLockLockTestCase("contended spinlock", func() *SpinLock {
			return new(SpinLock)
		}),
	)
}

func (tc spinLockLockTestCase) Name() string {
	return tc.name
}

func (tc spinLockLockTestCase) Test(t *testing.T) {
	t.Helper()

	sl := tc.setup()

	if tc.name == "contended spinlock" {
		// Test concurrent access
		acquired := make(chan bool, 2)

		err := RunConcurrentTest(t, 2, func(id int) error {
			if id == 0 {
				// First goroutine acquires lock
				sl.Lock()
				acquired <- true
				time.Sleep(10 * time.Millisecond)
				sl.Unlock()
			} else {
				// Second goroutine waits for lock
				time.Sleep(5 * time.Millisecond)
				sl.Lock()
				acquired <- true
				sl.Unlock()
			}
			return nil
		})

		AssertNoError(t, err, "concurrent lock test")
		close(acquired)

		// Both should have acquired the lock
		count := 0
		for range acquired {
			count++
		}
		AssertEqual(t, 2, count, "lock count")
	} else {
		// Simple case
		sl.Lock()
		// Do some work while holding the lock
		_ = runtime.NumGoroutine()
		sl.Unlock()
	}
}

func TestSpinLockLock(t *testing.T) {
	RunTestCases(t, makeSpinLockLockTestCases())
}

type spinLockUnlockTestCase struct {
	setup       func() *SpinLock
	name        string
	shouldPanic bool
}

// Factory function for spinLockUnlockTestCase
func newSpinLockUnlockTestCase(name string, setup func() *SpinLock, shouldPanic bool) TestCase {
	return spinLockUnlockTestCase{
		name:        name,
		setup:       setup,
		shouldPanic: shouldPanic,
	}
}

func makeSpinLockUnlockTestCases() []TestCase {
	return S(
		newSpinLockUnlockTestCase("locked spinlock", func() *SpinLock {
			sl := new(SpinLock)
			sl.Lock()
			return sl
		}, false),
		newSpinLockUnlockTestCase("unlocked spinlock", func() *SpinLock {
			return new(SpinLock)
		}, true),
	)
}

func (tc spinLockUnlockTestCase) Name() string {
	return tc.name
}

func (tc spinLockUnlockTestCase) Test(t *testing.T) {
	t.Helper()

	sl := tc.setup()

	if tc.shouldPanic {
		AssertPanic(t, func() { sl.Unlock() }, "invalid SpinLock.Unlock", "Unlock unlocked")
	} else {
		AssertNoPanic(t, func() { sl.Unlock() }, "Unlock locked")
	}
}

func TestSpinLockUnlock(t *testing.T) {
	RunTestCases(t, makeSpinLockUnlockTestCases())
}

func runTestSpinLockNilPtr(t *testing.T) {
	t.Helper()
	var sl *SpinLock
	ptr := sl.ptr()
	AssertEqual(t, (*uint32)(nil), ptr, "nil ptr")
}

func runTestSpinLockNilTryLock(t *testing.T) {
	t.Helper()
	var sl *SpinLock
	AssertPanic(t, func() { sl.TryLock() }, nil, "nil TryLock")
}

func runTestSpinLockNilUnlock(t *testing.T) {
	t.Helper()
	var sl *SpinLock
	AssertPanic(t, func() { sl.Unlock() }, nil, "nil Unlock")
}

func TestSpinLockNilReceiver(t *testing.T) {
	t.Run("nil ptr() method", runTestSpinLockNilPtr)
	t.Run("nil TryLock", runTestSpinLockNilTryLock)
	t.Run("nil Unlock panic", runTestSpinLockNilUnlock)
}

func TestSpinLockConcurrency(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 1000

	var sl SpinLock
	var counter int64

	err := RunConcurrentTest(t, numGoroutines, func(_ int) error {
		for j := 0; j < numIterations; j++ {
			sl.Lock()
			counter++
			sl.Unlock()
			runtime.Gosched()
		}
		return nil
	})

	AssertNoError(t, err, "concurrent test")

	expected := int64(numGoroutines * numIterations)
	AssertEqual(t, expected, counter, "counter value")
}

func BenchmarkSpinLockUncontended(b *testing.B) {
	RunBenchmark(b, func() any {
		return new(SpinLock)
	}, func(data any) {
		sl, ok := data.(*SpinLock)
		if !ok {
			b.Fatal("invalid data type")
		}
		for i := 0; i < b.N; i++ {
			sl.Lock()
			// Do minimal work while holding the lock
			_ = runtime.NumGoroutine()
			sl.Unlock()
		}
	})
}

func BenchmarkSpinLockContended(b *testing.B) {
	RunBenchmark(b, func() any {
		return new(SpinLock)
	}, func(data any) {
		sl, ok := data.(*SpinLock)
		if !ok {
			b.Fatal("invalid data type")
		}
		runContentionBenchmark(b, sl)
	})
}

func runContentionBenchmark(b *testing.B, sl *SpinLock) {
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	iterations := b.N / numWorkers

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				sl.Lock()
				_ = runtime.NumGoroutine()
				sl.Unlock()
			}
		}()
	}
	wg.Wait()
}
