package batch

import (
	"log"
	"testing"
	"time"
)

func TestBatch(t *testing.T) {
	b := NewBatch[int](4)
	b.PushFunc(func(in chan int) error {
		for _, v := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9} {
			in <- v
		}
		return nil
	})

	for numbers := range b.ReadChannel() {
		log.Println(numbers)
	}

	err := b.Error()
	if err != nil {
		log.Println(err)
	}
	time.Sleep(time.Hour)
}
