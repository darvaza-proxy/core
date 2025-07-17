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

	if result != tc.expected {
		t.Errorf("TryLock() = %v, expected %v", result, tc.expected)
	}
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
		var wg sync.WaitGroup
		acquired := make(chan bool, 2)

		// First goroutine acquires lock
		wg.Add(1)
		go func() {
			defer wg.Done()
			sl.Lock()
			acquired <- true
			time.Sleep(10 * time.Millisecond)
			sl.Unlock()
		}()

		// Second goroutine waits for lock
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(5 * time.Millisecond)
			sl.Lock()
			acquired <- true
			sl.Unlock()
		}()

		wg.Wait()
		close(acquired)

		// Both should have acquired the lock
		count := 0
		for range acquired {
			count++
		}
		if count != 2 {
			t.Errorf("Expected 2 lock acquisitions, got %d", count)
		}
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

//revive:disable-next-line:cognitive-complexity
func (tc spinLockUnlockTestCase) test(t *testing.T) {
	t.Helper()

	sl := tc.setup()

	if tc.shouldPanic {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic but none occurred")
			} else if r != "invalid SpinLock.Unlock" {
				t.Errorf("Expected panic message 'invalid SpinLock.Unlock', got %v", r)
			}
		}()
	}

	sl.Unlock()

	if tc.shouldPanic {
		t.Error("Expected panic but Unlock completed normally")
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
	if ptr := sl.ptr(); ptr != nil {
		t.Errorf("Expected nil ptr(), got %p", ptr)
	}
}

func testSpinLockNilTryLock(t *testing.T) {
	t.Helper()
	var sl *SpinLock
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on nil TryLock")
		}
	}()
	sl.TryLock()
}

func testSpinLockNilUnlock(t *testing.T) {
	t.Helper()
	var sl *SpinLock
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on nil Unlock")
		}
	}()
	sl.Unlock()
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
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				sl.Lock()
				counter++
				sl.Unlock()
				runtime.Gosched()
			}
		}()
	}

	wg.Wait()

	expected := int64(numGoroutines * numIterations)
	if counter != expected {
		t.Errorf("Expected counter %d, got %d", expected, counter)
	}
}

func BenchmarkSpinLockUncontended(b *testing.B) {
	var sl SpinLock
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sl.Lock()
		// Do minimal work while holding the lock
		_ = runtime.NumGoroutine()
		sl.Unlock()
	}
}

func BenchmarkSpinLockContended(b *testing.B) {
	var sl SpinLock
	var wg sync.WaitGroup

	b.ResetTimer()

	numWorkers := runtime.NumCPU()
	iterations := b.N / numWorkers

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				sl.Lock()
				// Do minimal work while holding the lock
				_ = runtime.NumGoroutine()
				sl.Unlock()
			}
		}()
	}

	wg.Wait()
}
