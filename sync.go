package core

import (
	"context"
	"sync"
	"sync/atomic"
)

// WaitGroup is a safer way to run workers
type WaitGroup struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	err     atomic.Value
	errCh   chan error
	onError func(error) error
}

func (wg *WaitGroup) init() {
	wg.mu.Lock()
	if wg.errCh == nil {
		wg.errCh = make(chan error)
		go wg.watchErrCh()
	}
	wg.mu.Unlock()
}

// OnError sets a helper that will be called when
// a worker returns an error or panics
func (wg *WaitGroup) OnError(fn func(error) error) {
	wg.onError = fn
}

func (wg *WaitGroup) watchErrCh() {
	defer close(wg.errCh)

	for {
		err, ok := <-wg.errCh
		switch {
		case !ok:
			// wtf
			return
		case wg.onError != nil:
			// process
			err = wg.onError(err)
		}

		switch {
		case err == nil:
			// error dismissed
		case wg.err.CompareAndSwap(nil, err):
			// first, we are done.
			return
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
	wg.init()

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
			// ignore if errCh is closed
			_ = recover()
		}()

		wg.errCh <- err
	}()
}

// Wait waits until all workers have finished, and returns
// the first error
func (wg *WaitGroup) Wait() error {
	wg.wg.Wait()
	return wg.Err()
}

// Done returns a channel that gets closed when all workers
// have finished.
func (wg *WaitGroup) Done() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.wg.Wait()
	}()
	return done
}

// Err returns the first error
func (wg *WaitGroup) Err() error {
	if err, ok := wg.err.Load().(error); ok {
		return err
	}
	return nil
}

// ErrGroup handles a group of workers where all are canceled once one fails.
// As it's based on [WaitGroup] it also catches panics.
type ErrGroup struct {
	wg        WaitGroup
	ctx       context.Context
	cancel    context.CancelCauseFunc
	cancelled atomic.Bool
	onError   func(error)

	Parent context.Context
}

// SetDefaults fills gaps in the config and initializes
// the internal structure.
func (eg *ErrGroup) SetDefaults() {
	if eg.Parent == nil {
		eg.Parent = context.Background()
	}

	if eg.ctx == nil {
		ctx, cancel := context.WithCancelCause(eg.Parent)

		eg.ctx = ctx
		eg.cancel = cancel

		eg.wg.OnError(eg.wgError)
	}
}

func (eg *ErrGroup) init() {
	eg.wg.mu.Lock()
	defer eg.wg.mu.Unlock()

	if eg.ctx == nil {
		// once
		eg.SetDefaults()
	}
}

// OnError sets a helper that will be called when
// a worker returns an error or panics
func (eg *ErrGroup) OnError(fn func(error)) {
	eg.onError = fn
}

// Cancel initiates a shutdown of the group. The returned
// value indicates if it was the first time.
func (eg *ErrGroup) Cancel(cause error) bool {
	eg.init()

	if cause == nil {
		cause = context.Canceled
	}

	return eg.doCancel(cause)
}

func (eg *ErrGroup) doCancel(cause error) bool {
	var first bool

	if eg.cancelled.CompareAndSwap(false, true) {
		// cancel once
		eg.cancel(cause)

		first = true
	}

	// and notify others
	if fn := eg.onError; fn != nil {
		fn(cause)
	}

	return first
}

func (eg *ErrGroup) wgError(err error) error {
	if eg.doCancel(err) {
		// first
		return err
	}
	return nil
}

// Context returns the cancellable context used with the workers
func (eg *ErrGroup) Context() context.Context {
	eg.init()

	return eg.ctx
}

// Cancelled returns a channel marker to know when the Group has
// been cancelled and the shutdown has been initiated.
//
// Cancelled() doesn't indicate all workers have finished, for that
// call [ErrGroup.Wait] or [ErrGroup.Done].
func (eg *ErrGroup) Cancelled() <-chan struct{} {
	eg.init()

	return eg.ctx.Done()
}

// Done returns a channel that gets closed when all workers
// have finished.
func (eg *ErrGroup) Done() <-chan struct{} {
	eg.init()

	return eg.wg.Done()
}

// IsCancelled tells the [ErrGroup] has been cancelled
func (eg *ErrGroup) IsCancelled() bool {
	return eg.cancelled.Load()
}

// Wait waits until all workers in the group have finished.
func (eg *ErrGroup) Wait() error {
	return eg.wg.Wait()
}

// Err returns the error that initiated the group's shutdown.
func (eg *ErrGroup) Err() error {
	return eg.wg.Err()
}

// Go spawns a worker and an optional shutdown routine to be invoked
// when the [ErrGroup] is cancelled, otherwise the provided context needs
// to be monitored and shutdown called.
func (eg *ErrGroup) Go(run func(context.Context) error, shutdown func() error) {
	// run with default error catcher
	eg.GoCatch(run, nil)

	if shutdown != nil {
		// shutdown
		s2 := func() error {
			<-eg.ctx.Done()
			return shutdown()
		}

		// don't intercept shutdown's return error
		eg.wg.GoCatch(s2, nil)
	}
}

// GoCatch runs a worker on the Group, with a custom error handler.
func (eg *ErrGroup) GoCatch(run func(context.Context) error,
	catch func(context.Context, error) error) {
	//
	var r2 func() error
	var c2 func(error) error

	if run == nil {
		PanicWrap(ErrInvalid, "%s", "run function not specified")
	}

	eg.init()

	// wrap runner
	r2 = func() error {
		return run(eg.ctx)
	}

	if catch != nil {
		// wrap catcher
		c2 = func(err error) error {
			return catch(eg.ctx, err)
		}
	} else {
		// use default error catcher
		c2 = eg.defaultErrGroupCatcher
	}

	// always intercepting errors
	eg.wg.GoCatch(r2, c2)
}

func (eg *ErrGroup) defaultErrGroupCatcher(err error) error {
	if err != nil && eg.IsCancelled() {
		err = context.Canceled
	}
	return err
}
