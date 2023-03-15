package core

import (
	"sync"
	"sync/atomic"
)

// WaitGroup is a safer way to run workers
type WaitGroup struct {
	wg      sync.WaitGroup
	err     atomic.Value
	errCh   chan error
	onError func(error) error
}

// OnError sets a helper that will be called when
// a workers returns an error or panics
func (wg *WaitGroup) OnError(fn func(error) error) {
	wg.onError = fn
}

func (wg *WaitGroup) watchErrCh() {
	for {
		err, ok := <-wg.errCh
		if !ok {
			break
		}
		if wg.onError != nil {
			err = wg.onError(err)
		}
		if err != nil {
			wg.err.CompareAndSwap(nil, err)
		}
	}
}

// Go spawns a supervised goroutine
func (wg *WaitGroup) Go(fn func() error) {
	wg.GoCatch(fn, nil)
}

// GoCatch spawns a supervised goroutine, and uses a given function
// to intercept the returned error
func (wg *WaitGroup) GoCatch(fn func() error, catch func(error) error) {
	if wg.errCh == nil {
		wg.errCh = make(chan error)
		go wg.watchErrCh()
	}

	if fn != nil {
		wg.wg.Add(1)

		go func() {
			defer wg.wg.Done()

			wg.run(fn, catch)
		}()
	}
}

func (wg *WaitGroup) run(fn func() error, catch func(error) error) {
	var c1, c2 Catcher

	err := c1.Do(fn)
	if err != nil && catch != nil {
		err = c2.Do(func() error {
			return catch(err)
		})
	}

	if err != nil {
		wg.tryReportError(err)
	}
}

func (wg *WaitGroup) tryReportError(err error) {
	wg.wg.Add(1)

	go func() {
		defer wg.wg.Done()
		defer func() {
			recover()
		}()

		wg.errCh <- err
	}()
}

func (wg *WaitGroup) tryCloseErrCh() {
	defer func() {
		recover()
	}()

	close(wg.errCh)
}

// Wait waits until all workers have finished, and returns
// the first error
func (wg *WaitGroup) Wait() error {
	wg.wg.Wait()
	defer wg.tryCloseErrCh()

	return wg.Err()
}

// Err returns the first error
func (wg *WaitGroup) Err() error {
	if err, ok := wg.err.Load().(error); ok {
		return err
	}
	return nil
}
