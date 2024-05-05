package ctxjoin

import (
	"context"
	"time"
)

type cancelContext struct {
	ctx   context.Context
	extra context.Context
}

var _ context.Context = (*cancelContext)(nil)

// AddCancel creates a context with the same values and cancellation as the main context,
// but which will additionally be canceled when the extra context is canceled.
func AddCancel(main context.Context, extra context.Context) (context.Context, context.CancelFunc) {
	// This is similar to the "Merge" example for context.WithCancelCause, handling deadlines.
	ctx, cancelWithCause := context.WithCancelCause(main)
	stop := context.AfterFunc(extra, func() { cancelWithCause(context.Cause(extra)) })
	cancel := func() {
		stop()
		cancelWithCause(context.Canceled)
	}

	j := &cancelContext{
		ctx:   ctx,
		extra: extra,
	}

	return j, cancel
}

func (c *cancelContext) Deadline() (deadline time.Time, ok bool) {
	deadline, ok = c.ctx.Deadline()
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
	return c.ctx.Done()
}

func (c *cancelContext) Err() error {
	return c.ctx.Err()
}

func (c *cancelContext) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}
