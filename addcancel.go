package ctxjoin

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

var closedChan = make(chan struct{})

func init() {
	close(closedChan)
}

type cancelContext struct {
	main       context.Context
	cancelMain context.CancelFunc
	extra      context.Context

	doneOnce sync.Once
	done     chan struct{}

	err atomic.Value
}

// AddCancel creates a context with the same values and cancellation as the main context,
// but which will additionally be canceled when the extra context is canceled.
func AddCancel(main context.Context, extra context.Context) (context.Context, context.CancelFunc) {
	main, cancel := context.WithCancel(main)
	j := &cancelContext{
		main:       main,
		cancelMain: cancel,
		extra:      extra,
	}

	return j, j.cancel
}

func (c *cancelContext) Deadline() (deadline time.Time, ok bool) {
	deadline, ok = c.main.Deadline()
	if !ok {
		return c.extra.Deadline()
	}

	extra, ok := c.extra.Deadline()
	if ok && extra.Before(deadline) {
		return extra, true
	}

	return deadline, true
}

func (c *cancelContext) Done() <-chan struct{} {
	// Lazily start the goroutine when this channel is requested.
	// We can create the done chan from the parent contexts at any time.
	c.doneOnce.Do(c.runCloser)
	return c.done
}

func (c *cancelContext) cancel() {
	c.doneOnce.Do(func() {
		c.done = closedChan
		c.err.Store(context.Canceled)
	})
	c.cancelMain()
}

func (c *cancelContext) runCloser() {
	c.done = make(chan struct{})
	go func() {
		defer close(c.done)

		var err error
		select {
		case <-c.main.Done():
			err = c.main.Err()
		case <-c.extra.Done():
			err = c.extra.Err()
		}

		c.err.Store(err)
	}()
}

func (c *cancelContext) Err() error {
	if err := c.err.Load(); err != nil {
		return err.(error)
	}
	return nil
}

func (c *cancelContext) Value(key interface{}) interface{} {
	return c.main.Value(key)
}
