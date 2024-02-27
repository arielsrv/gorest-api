package tpl

import (
	"runtime"

	"go.uber.org/multierr"

	"github.com/sourcegraph/conc/iter"
	"github.com/sourcegraph/conc/pool"
)

type Pool struct {
	*pool.Pool
}

func New() *Pool {
	return &Pool{
		Pool: pool.New(),
	}
}

func (r *Pool) WithMaxGoroutines(maxGoroutines int) *Pool {
	r.Pool.WithMaxGoroutines(maxGoroutines)
	return r
}

func (r *Pool) Submit(task func()) {
	r.Pool.Go(task)
}

type Task[T any] struct {
	Result T
	Err    error
}

func ToTask[T any](f func() (T, error)) Task[T] {
	result, err := f()

	return Task[T]{
		Result: result,
		Err:    err,
	}
}

func ForEach[T any](input []T, f func(*T), maxGoroutines ...int) {
	it := iter.Iterator[T]{
		MaxGoroutines: runtime.NumCPU() - 1,
	}

	if len(maxGoroutines) > 0 && maxGoroutines[0] > 0 {
		it.MaxGoroutines = maxGoroutines[0]
	}

	it.ForEach(input, f)
}

type Pool41[TInput any, T1 any, T2 any, T3 any] struct {
	*pool.Pool
}

func NewWorkerPool41[TInput any, T1 any, T2 any, T3 any]() *Pool41[TInput, T1, T2, T3] {
	return &Pool41[TInput, T1, T2, T3]{
		Pool: pool.New().WithMaxGoroutines(3),
	}
}

func (r *Pool41[TInput, T1, T2, T3]) Zip(
	elements []TInput,
	f1 func(elements []TInput) ([]T1, error),
	f2 func(elements []TInput) ([]T2, error),
	f3 func(elements []TInput) ([]T3, error),
	f func(r1 []T1, r2 []T2, r3 []T3, err error) ([]T1, error)) ([]T1, error) {
	var aggErr error

	var r1 []T1
	r.Pool.Go(func() {
		c1, err := f1(elements)
		if err != nil {
			aggErr = multierr.Append(aggErr, err)
			return
		}
		r1 = append(r1, c1...)
	})

	var r2 []T2
	r.Pool.Go(func() {
		c2, err := f2(elements)
		if err != nil {
			aggErr = multierr.Append(aggErr, err)
			return
		}
		r2 = append(r2, c2...)
	})

	var r3 []T3
	r.Pool.Go(func() {
		c3, err := f3(elements)
		if err != nil {
			aggErr = multierr.Append(aggErr, err)
			return
		}
		r3 = append(r3, c3...)
	})

	r.Pool.Wait()

	result, err := f(r1, r2, r3, aggErr)
	if err != nil {
		return nil, err
	}

	return result, aggErr
}
