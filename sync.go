package core

import (
	"sync"
	"sync/atomic"
)

// WaitGroup is a safer way to run workers
type WaitGroup struct {
	wg      sync.WaitGroup
	err     atomic.Value
	onError func(error)
}

// OnError sets a helper that will be called when
// the first worker returns an error or panics
func (wg *WaitGroup) OnError(fn func(error)) {
	wg.onError = fn
}

// Go spawns a supervised goroutine
func (wg *WaitGroup) Go(fn func() error) {
	wg.GoCatch(fn, nil)
}

// GoCatch spawns a supervised goroutine, and uses a given function
// to intercept the returned error
func (wg *WaitGroup) GoCatch(fn func() error, catch func(error) error) {
	if fn != nil {
		wg.wg.Add(1)

		go func() {
			defer wg.wg.Done()

			wg.run(fn, catch)
		}()
	}
}

func (wg *WaitGroup) run(fn func() error, catch func(error) error) {
	var c Catcher

	if catch != nil {
		fx := func() error {
			var cx Catcher

			return catch(cx.Do(fn))
		}
		fn = fx
	}

	if err := c.Do(fn); err != nil {
		if wg.err.CompareAndSwap(nil, err) {
			wg.callOnError(err)
		}
	}
}

func (wg *WaitGroup) callOnError(err error) {
	if fn := wg.onError; fn != nil {
		var c Catcher

		wg.wg.Add(1)
		go func() {
			defer wg.wg.Done()

			_ = c.Do(func() error {
				fn(err)
				return nil
			})
		}()
	}
}

// Wait waits until all workers have finished, and returns
// the first error
func (wg *WaitGroup) Wait() error {
	wg.wg.Wait()

	if err, ok := wg.err.Load().(error); ok {
		return err
	}
	return nil
}
