package main

import (
	"github.com/sourcegraph/conc/iter"
	"github.com/sourcegraph/conc/pool"
	"log"
)

func main() {
	numberChan := make(chan int)

	buffer := pool.New()

	values1 := []int{1, 2, 3, 4, 5}
	values2 := []int{6, 7, 8, 9, 10}

	buffer.Go(func() {
		iter.ForEach[int](values1, func(i *int) {
			numberChan <- *i
		})
	})

	buffer.Go(func() {
		iter.ForEach[int](values2, func(i *int) {
			numberChan <- *i
		})
	})

	buffer.Go(func() {
		iter.ForEach[int](values1, func(_ *int) {
			number := <-numberChan
			log.Printf("Got %d\n", number)
		})
	})

	buffer.Go(func() {
		iter.ForEach[int](values2, func(_ *int) {
			log.Printf("Got %d\n", <-numberChan)
		})
	})

	buffer.Wait()
}
