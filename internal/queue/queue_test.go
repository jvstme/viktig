package queue

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	t.Run("order", func(t *testing.T) {
		q := NewQueue[int]()

		go func() {
			q.Put(1)
			q.Put(2)
			q.Put(3)
		}()

		assert.Equal(t, 1, q.Take())
		assert.Equal(t, 2, q.Take())
		assert.Equal(t, 3, q.Take())
	})

	t.Run("correctness", func(t *testing.T) {
		q := NewQueue[int]()
		var x1, x2 int
		wg := sync.WaitGroup{}

		wg.Add(3)
		go func() {
			defer wg.Done()
			x1 = q.Take()
		}()
		go func() {
			defer wg.Done()
			x2 = q.Take()
		}()
		go func() {
			defer wg.Done()
			q.Put(1)
			q.Put(2)
		}()
		wg.Wait()

		assert.True(t, (x1 == 1 && x2 == 2) || (x1 == 2 && x2 == 1))
	})

	t.Run("blocks as chan", func(t *testing.T) {
		q := NewQueue[int]()

		go func() {
			q.Put(1)
			q.Put(2)
			q.Put(3)
		}()
		<-q.AsChan()
		<-q.AsChan()
		<-q.AsChan()
		blocks := false
		select {
		case <-q.AsChan():
		default:
			blocks = true
		}

		assert.True(t, blocks)
	})

	t.Run("blocks", func(t *testing.T) {
		q := NewQueue[int]()

		go func() {
			q.Put(1)
			q.Put(2)
			q.Put(3)
		}()
		<-q.AsChan()
		<-q.AsChan()
		<-q.AsChan()

		ch := make(chan any)
		go func() {
			ch <- q.Take()
		}()

		blocks := false
		select {
		case <-ch:
		default:
			blocks = true
		}

		assert.True(t, blocks)
	})
}
