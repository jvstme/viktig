package queue

type Queue[T any] struct {
	ch chan T
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{ch: make(chan T)}
}

func (q *Queue[T]) Put(x T) {
	q.ch <- x
}

func (q *Queue[T]) Take() T {
	return <-q.ch
}

func (q *Queue[T]) AsChan() chan T {
	return q.ch
}
