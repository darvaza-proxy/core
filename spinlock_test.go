package core_test

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ core.TestCase = spinLockTryLockTestCase{}
	_ core.TestCase = spinLockLockTestCase{}
	_ core.TestCase = spinLockUnlockTestCase{}
)

type spinLockTryLockTestCase struct {
	setup    func() *core.SpinLock
	name     string
	expected bool
}

// Factory function for spinLockTryLockTestCase
func newSpinLockTryLockTestCase(name string, setup func() *core.SpinLock, expected bool) spinLockTryLockTestCase {
	return spinLockTryLockTestCase{
		name:     name,
		setup:    setup,
		expected: expected,
	}
}

var spinLockTryLockTestCases = []spinLockTryLockTestCase{
	newSpinLockTryLockTestCase("unlocked spinlock", func() *core.SpinLock {
		return new(core.SpinLock)
	}, true),
	newSpinLockTryLockTestCase("already locked spinlock", func() *core.SpinLock {
		sl := new(core.SpinLock)
		sl.TryLock()
		return sl
	}, false),
}

func (tc spinLockTryLockTestCase) Name() string {
	return tc.name
}

func (tc spinLockTryLockTestCase) Test(t *testing.T) {
	sl := tc.setup()
	result := sl.TryLock()

	core.AssertEqual(t, tc.expected, result, "TryLock")
}

func TestSpinLockTryLock(t *testing.T) {
	core.RunTestCases(t, spinLockTryLockTestCases)
}

type spinLockLockTestCase struct {
	setup func() *core.SpinLock
	name  string
}

// Factory function for spinLockLockTestCase
func newSpinLockLockTestCase(name string, setup func() *core.SpinLock) spinLockLockTestCase {
	return spinLockLockTestCase{
		name:  name,
		setup: setup,
	}
}

var spinLockLockTestCases = []spinLockLockTestCase{
	newSpinLockLockTestCase("unlocked spinlock", func() *core.SpinLock {
		return new(core.SpinLock)
	}),
	newSpinLockLockTestCase("contended spinlock", func() *core.SpinLock {
		return new(core.SpinLock)
	}),
}

func (tc spinLockLockTestCase) Name() string {
	return tc.name
}

func (tc spinLockLockTestCase) Test(t *testing.T) {
	sl := tc.setup()

	if tc.name == "contended spinlock" {
		// Test concurrent access
		acquired := make(chan bool, 2)

		err := core.RunConcurrentTest(t, 2, func(id int) error {
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

		core.AssertNoError(t, err, "concurrent lock test")
		close(acquired)

		// Both should have acquired the lock
		count := 0
		for range acquired {
			count++
		}
		core.AssertEqual(t, 2, count, "lock count")
	} else {
		// Simple case
		sl.Lock()
		// Do some work while holding the lock
		_ = runtime.NumGoroutine()
		sl.Unlock()
	}
}

func TestSpinLockLock(t *testing.T) {
	core.RunTestCases(t, spinLockLockTestCases)
}

type spinLockUnlockTestCase struct {
	setup       func() *core.SpinLock
	name        string
	shouldPanic bool
}

// Factory function for spinLockUnlockTestCase
func newSpinLockUnlockTestCase(name string, setup func() *core.SpinLock, shouldPanic bool) spinLockUnlockTestCase {
	return spinLockUnlockTestCase{
		name:        name,
		setup:       setup,
		shouldPanic: shouldPanic,
	}
}

var spinLockUnlockTestCases = []spinLockUnlockTestCase{
	newSpinLockUnlockTestCase("locked spinlock", func() *core.SpinLock {
		sl := new(core.SpinLock)
		sl.Lock()
		return sl
	}, false),
	newSpinLockUnlockTestCase("unlocked spinlock", func() *core.SpinLock {
		return new(core.SpinLock)
	}, true),
}

func (tc spinLockUnlockTestCase) Name() string {
	return tc.name
}

func (tc spinLockUnlockTestCase) Test(t *testing.T) {
	sl := tc.setup()

	if tc.shouldPanic {
		core.AssertPanic(t, func() { sl.Unlock() }, "invalid SpinLock.Unlock", "Unlock unlocked")
	} else {
		core.AssertNoPanic(t, func() { sl.Unlock() }, "Unlock locked")
	}
}

func TestSpinLockUnlock(t *testing.T) {
	core.RunTestCases(t, spinLockUnlockTestCases)
}

func testSpinLockNilTryLock(t *testing.T) {
	var sl *core.SpinLock
	core.AssertPanic(t, func() { sl.TryLock() }, nil, "nil TryLock")
}

func testSpinLockNilUnlock(t *testing.T) {
	var sl *core.SpinLock
	core.AssertPanic(t, func() { sl.Unlock() }, nil, "nil Unlock")
}

func TestSpinLockNilReceiver(t *testing.T) {
	t.Run("nil TryLock", testSpinLockNilTryLock)
	t.Run("nil Unlock panic", testSpinLockNilUnlock)
}

func TestSpinLockConcurrency(t *testing.T) {
	const numGoroutines = 100
	const numIterations = 1000

	var sl core.SpinLock
	var counter int64

	err := core.RunConcurrentTest(t, numGoroutines, func(_ int) error {
		for range numIterations {
			sl.Lock()
			counter++
			sl.Unlock()
			runtime.Gosched()
		}
		return nil
	})

	core.AssertNoError(t, err, "concurrent test")

	expected := int64(numGoroutines * numIterations)
	core.AssertEqual(t, expected, counter, "counter value")
}

func BenchmarkSpinLockUncontended(b *testing.B) {
	core.RunBenchmark(b, func() any {
		return new(core.SpinLock)
	}, func(data any) {
		sl := core.MustT[*core.SpinLock](data)
		for b.Loop() {
			sl.Lock()
			// Do minimal work while holding the lock
			_ = runtime.NumGoroutine()
			sl.Unlock()
		}
	})
}

func BenchmarkSpinLockContended(b *testing.B) {
	core.RunBenchmark(b, func() any {
		return new(core.SpinLock)
	}, func(data any) {
		sl := core.MustT[*core.SpinLock](data)
		runContentionBenchmark(b, sl)
	})
}

func runContentionBenchmark(b *testing.B, sl *core.SpinLock) {
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	iterations := b.N / numWorkers

	for range numWorkers {
		wg.Go(func() {
			for range iterations {
				sl.Lock()
				_ = runtime.NumGoroutine()
				sl.Unlock()
			}
		})
	}
	wg.Wait()
}
