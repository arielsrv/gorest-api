package tpl_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/src/app/tpl"
)

func TestNew(t *testing.T) {
	pool := tpl.New().WithMaxGoroutines(2)

	var accum int
	numbers := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	rChan := make(chan tpl.Task[int])

	pool.Submit(func() {
		for i := 0; i < len(numbers); i++ {
			task := <-rChan
			if task.Err != nil {
				t.Error(task.Err)
			}
			accum += task.Result
		}
	})

	pool.Submit(func() {
		tpl.ForEach(numbers, func(i *int) {
			rChan <- tpl.ToTask[int](func() (int, error) {
				return *i, nil
			})
		}, len(numbers))
	})

	pool.Wait()

	assert.Equal(t, 45, accum)
}
