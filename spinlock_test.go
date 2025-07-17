package core

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

type spinLockTryLockTestCase struct {
	setup    func() *SpinLock
	name     string
	expected bool
}

var spinLockTryLockTestCases = []spinLockTryLockTestCase{
	{
		name: "unlocked spinlock",
		setup: func() *SpinLock {
			return new(SpinLock)
		},
		expected: true,
	},
	{
		name: "already locked spinlock",
		setup: func() *SpinLock {
			sl := new(SpinLock)
			sl.TryLock()
			return sl
		},
		expected: false,
	},
}

func (tc spinLockTryLockTestCase) test(t *testing.T) {
	t.Helper()

	sl := tc.setup()
	result := sl.TryLock()

	AssertEqual(t, tc.expected, result, "TryLock() result mismatch")
}

func TestSpinLockTryLock(t *testing.T) {
	for _, tc := range spinLockTryLockTestCases {
		t.Run(tc.name, tc.test)
	}
}

type spinLockLockTestCase struct {
	setup func() *SpinLock
	name  string
}

var spinLockLockTestCases = []spinLockLockTestCase{
	{
		name: "unlocked spinlock",
		setup: func() *SpinLock {
			return new(SpinLock)
		},
	},
	{
		name: "contended spinlock",
		setup: func() *SpinLock {
			return new(SpinLock)
		},
	},
}

func (tc spinLockLockTestCase) test(t *testing.T) {
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

		AssertError(t, err, false, "concurrent lock test should not fail")
		close(acquired)

		// Both should have acquired the lock
		count := 0
		for range acquired {
			count++
		}
		AssertEqual(t, 2, count, "Expected 2 lock acquisitions")
	} else {
		// Simple case
		sl.Lock()
		// Do some work while holding the lock
		_ = runtime.NumGoroutine()
		sl.Unlock()
	}
}

func TestSpinLockLock(t *testing.T) {
	for _, tc := range spinLockLockTestCases {
		t.Run(tc.name, tc.test)
	}
}

type spinLockUnlockTestCase struct {
	setup       func() *SpinLock
	name        string
	shouldPanic bool
}

var spinLockUnlockTestCases = []spinLockUnlockTestCase{
	{
		name: "locked spinlock",
		setup: func() *SpinLock {
			sl := new(SpinLock)
			sl.Lock()
			return sl
		},
		shouldPanic: false,
	},
	{
		name: "unlocked spinlock",
		setup: func() *SpinLock {
			return new(SpinLock)
		},
		shouldPanic: true,
	},
}

func (tc spinLockUnlockTestCase) test(t *testing.T) {
	t.Helper()

	sl := tc.setup()

	if tc.shouldPanic {
		AssertPanic(t, func() { sl.Unlock() }, "invalid SpinLock.Unlock", "Unlock should panic on unlocked spinlock")
	} else {
		AssertNoPanic(t, func() { sl.Unlock() }, "Unlock should not panic on locked spinlock")
	}
}

func TestSpinLockUnlock(t *testing.T) {
	for _, tc := range spinLockUnlockTestCases {
		t.Run(tc.name, tc.test)
	}
}

func testSpinLockNilPtr(t *testing.T) {
	t.Helper()
	var sl *SpinLock
	ptr := sl.ptr()
	AssertEqual(t, (*uint32)(nil), ptr, "nil SpinLock ptr() should return nil")
}

func testSpinLockNilTryLock(t *testing.T) {
	t.Helper()
	var sl *SpinLock
	AssertPanic(t, func() { sl.TryLock() }, nil, "nil SpinLock TryLock should panic")
}

func testSpinLockNilUnlock(t *testing.T) {
	t.Helper()
	var sl *SpinLock
	AssertPanic(t, func() { sl.Unlock() }, nil, "nil SpinLock Unlock should panic")
}

func TestSpinLockNilReceiver(t *testing.T) {
	t.Run("nil ptr() method", testSpinLockNilPtr)
	t.Run("nil TryLock", testSpinLockNilTryLock)
	t.Run("nil Unlock panic", testSpinLockNilUnlock)
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

	AssertError(t, err, false, "concurrent test should not fail")

	expected := int64(numGoroutines * numIterations)
	AssertEqual(t, expected, counter, "counter should match expected value")
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
