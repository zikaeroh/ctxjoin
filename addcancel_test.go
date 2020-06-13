package ctxjoin_test

import (
	"context"
	"testing"
	"time"

	"github.com/zikaeroh/ctxjoin"
	"gotest.tools/v3/assert"
)

type contextKey string

func TestAddCancelDeadline(t *testing.T) {
	test := func(main, extra context.Context, deadline time.Time, ok bool) func(*testing.T) {
		return func(t *testing.T) {
			ctx, cancel := ctxjoin.AddCancel(main, extra)
			defer cancel()

			gotDeadline, gotOk := ctx.Deadline()
			assert.Equal(t, gotDeadline, deadline)
			assert.Equal(t, gotOk, ok)
		}
	}

	main := context.Background()
	hasCancel, cancel := context.WithCancel(main)
	defer cancel()

	now := time.Now()
	soon := now.Add(time.Minute)
	far := now.Add(time.Hour)

	hasDeadlineSoon, cancel := context.WithDeadline(main, soon)
	defer cancel()

	hasDeadlineFar, cancel := context.WithDeadline(main, far)
	defer cancel()

	t.Run("No deadline", test(main, main, time.Time{}, false))
	t.Run("Only cancel", test(hasCancel, main, time.Time{}, false))
	t.Run("Extra sooner", test(hasDeadlineFar, hasDeadlineSoon, soon, true))
	t.Run("Main sooner", test(hasDeadlineSoon, hasDeadlineFar, soon, true))
	t.Run("Extra only", test(main, hasDeadlineSoon, soon, true))
}

func TestAddCancelValue(t *testing.T) {
	main := context.WithValue(context.Background(), contextKey("main"), 1234)
	main = context.WithValue(main, contextKey("main2"), "value")
	extra := context.WithValue(context.Background(), contextKey("extra"), true)
	extra = context.WithValue(extra, contextKey("main"), 7890)
	extra, cancel := context.WithCancel(extra)
	defer cancel()

	ctx, cancel := ctxjoin.AddCancel(main, extra)
	defer cancel()

	assert.Equal(t, ctx.Value(contextKey("main")), 1234)
	assert.Equal(t, ctx.Value(contextKey("main2")), "value")
	assert.Equal(t, ctx.Value(contextKey("extra")), nil)
}

func TestAddCancelDoneErr(t *testing.T) {
	t.Run("No cancel", func(t *testing.T) {
		a, cancelA := context.WithCancel(context.Background())
		b, cancelB := context.WithCancel(context.Background())

		ctx, cancel := ctxjoin.AddCancel(a, b)
		defer cancel()

		assert.NilError(t, ctx.Err())
		cancelA()
		cancelB()
	})

	t.Run("Main canceled", func(t *testing.T) {
		a, cancelA := context.WithCancel(context.Background())
		b, cancelB := context.WithCancel(context.Background())

		ctx, cancel := ctxjoin.AddCancel(a, b)
		defer cancel()

		done := ctx.Done()
		cancelA()

		<-done
		assert.Equal(t, ctx.Err(), context.Canceled)
		cancelB()
	})

	t.Run("Extra canceled", func(t *testing.T) {
		a, cancelA := context.WithCancel(context.Background())
		b, cancelB := context.WithCancel(context.Background())

		ctx, cancel := ctxjoin.AddCancel(a, b)
		defer cancel()

		done := ctx.Done()
		cancelB()

		<-done
		assert.Equal(t, ctx.Err(), context.Canceled)
		cancelA()
	})

	t.Run("Top canceled", func(t *testing.T) {
		a := context.Background()
		b := context.Background()

		ctx, cancel := ctxjoin.AddCancel(a, b)

		done := ctx.Done()
		cancel()

		<-done
		assert.Equal(t, ctx.Err(), context.Canceled)
	})

	t.Run("Top without calling Done", func(t *testing.T) {
		a := context.Background()
		b := context.Background()

		ctx, cancel := ctxjoin.AddCancel(a, b)
		cancel()

		assert.Equal(t, ctx.Err(), context.Canceled)
	})
}
