package gpars

import (
	"runtime"

	"github.com/sourcegraph/conc/iter"
	"github.com/sourcegraph/conc/pool"
)

type Pool struct {
	*pool.Pool
}

func New() *Pool {
	return &Pool{
		pool.New(),
	}
}

func (r *Pool) WithMaxGoroutines(maxGoroutines int) *Pool {
	r.Pool.WithMaxGoroutines(maxGoroutines)
	return r
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
