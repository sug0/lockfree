// Copyright 2020 The golang.design Initiative authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.

package lockfree_test

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sug0/lockfree"
)

func ptr[T any](x T) *T {
	return &x
}

func TestQueueDequeueEmpty(t *testing.T) {
	q := lockfree.NewQueue[int]()
	if q.Dequeue() != nil {
		t.Fatalf("dequeue empty queue returns non-nil")
	}
}

func TestQueue_Length(t *testing.T) {
	q := lockfree.NewQueue[int]()
	if q.Length() != 0 {
		t.Fatalf("empty queue has non-zero length")
	}

	q.Enqueue(ptr(1))
	if q.Length() != 1 {
		t.Fatalf("count of enqueue wrong, want %d, got %d.", 1, q.Length())
	}

	q.Dequeue()
	if q.Length() != 0 {
		t.Fatalf("count of dequeue wrong, want %d, got %d", 0, q.Length())
	}
}

func ExampleQueue() {
	q := lockfree.NewQueue[string]()

	q.Enqueue(ptr("1st item"))
	q.Enqueue(ptr("2nd item"))
	q.Enqueue(ptr("3rd item"))

	fmt.Println(*q.Dequeue())
	fmt.Println(*q.Dequeue())
	fmt.Println(*q.Dequeue())

	// Output:
	// 1st item
	// 2nd item
	// 3rd item
}

type queueInterface[T any] interface {
	Enqueue(*T)
	Dequeue() *T
}

type mutexQueue[T any] struct {
	v  []*T
	mu sync.Mutex
}

func newMutexQueue[T any]() *mutexQueue[T] {
	return &mutexQueue[T]{v: make([]*T, 0)}
}

func (q *mutexQueue[T]) Enqueue(v *T) {
	q.mu.Lock()
	q.v = append(q.v, v)
	q.mu.Unlock()
}

func (q *mutexQueue[T]) Dequeue() *T {
	q.mu.Lock()
	if len(q.v) == 0 {
		q.mu.Unlock()
		return nil
	}
	v := q.v[0]
	q.v = q.v[1:]
	q.mu.Unlock()
	return v
}

func BenchmarkQueue(b *testing.B) {
	length := 1 << 12
	inputs := make([]int, length)
	for i := 0; i < length; i++ {
		inputs = append(inputs, rand.Int())
	}
	q, mq := lockfree.NewQueue[int](), newMutexQueue[int]()
	b.ResetTimer()

	for _, q := range [...]queueInterface[int]{q, mq} {
		b.Run(fmt.Sprintf("%T", q), func(b *testing.B) {
			var c int64
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					i := int(atomic.AddInt64(&c, 1)-1) % length
					v := inputs[i]
					if v >= 0 {
						q.Enqueue(ptr(v))
					} else {
						q.Dequeue()
					}
				}
			})
		})
	}
}
