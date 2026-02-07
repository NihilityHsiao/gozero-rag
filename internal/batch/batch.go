package batch

// 1
import (
	"log"
	"time"
)

const kTimeout = 120 * time.Second

type BatchItem struct {
	SpuId string
}

type OutResult[T any] struct {
	Items    []T
	BatchNum int
}

type Batch[T any] struct {
	size int
	in   chan T
	out  chan OutResult[T]
	err  error
}

func NewBatch[T any](size int) *Batch[T] {
	b := &Batch[T]{
		size: size,
		in:   make(chan T, size),
		out:  make(chan OutResult[T], size),
	}
	go func() {
		defer func() {
			close(b.out)
			log.Println("batch out channel 关闭")
		}()
		items := make([]T, 0)
		batchNum := 0
		for item := range b.in {
			items = append(items, item)
			if len(items) >= size {
				batchNum++
				select {
				case b.out <- OutResult[T]{items, batchNum}:
					items = make([]T, 0)
				case <-time.After(kTimeout):
					log.Println("out channel 超时 没消息发出，主动关闭")
					return
				}
			}
		}
		if len(items) > 0 {
			batchNum++
			select {
			case b.out <- OutResult[T]{items, batchNum}:
				items = make([]T, size)
			case <-time.After(kTimeout):
				log.Println("out channel 超时 没消息发出，主动关闭")
				return
			}
		}
	}()
	return b
}

func (receiver *Batch[T]) PushFunc(f func(in chan T) error) {
	go func() {
		defer func() {
			close(receiver.in)
			log.Println("batch in channel 关闭")
		}()
		err := f(receiver.in)
		if err != nil {
			receiver.err = err
		}
	}()
}

func (receiver *Batch[T]) ReadChannel() chan OutResult[T] {
	return receiver.out
}

func (receiver *Batch[T]) Error() error {
	return receiver.err
}
