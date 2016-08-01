// Package multiq is a relaxed concurrent priority queues
/*
   https://arxiv.org/abs/1411.1209
*/
package multiq

import (
	"container/heap"
	"math"
	"runtime"
	"sync/atomic"

	"github.com/dgryski/go-multiq/internal/pq"
)

type Q struct {
	qs []pq.PriorityQueue

	locks []lock

	// minimum priority for each queue
	mins []int32
}

func New(c int) *Q {

	mins := make([]int32, c)

	for i := range mins {
		mins[i] = math.MaxInt32
	}
	return &Q{
		qs:    make([]pq.PriorityQueue, c),
		locks: make([]lock, c),
		mins:  mins,
	}
}

// Insert adds an element to the queue
func (q *Q) Insert(v interface{}, prio int32) {

	rng := rdrand()

	// Implementing our own lock and spinning here is probably bad,
	// but sync.Mutex has no TryLock()
	// https://github.com/golang/go/issues/6123
	var c uint32
	var iter int
	for {
		rng = xorshiftMult64(rng)
		c = reduce(uint32(rng), len(q.locks))
		gotlock := q.locks[c].trylock()
		if gotlock {
			break
		}
		iter++
		if iter >= len(q.locks) {
			runtime.Gosched()
		}
	}

	// insert the item into priority queue c
	heap.Push(&q.qs[c], &pq.Item{Value: v, Priority: prio})

	// update the stored minimum
	atomic.StoreInt32(&q.mins[c], q.qs[c][0].Priority)

	// unlock
	q.locks[c].unlock()
}

func (q *Q) DeleteMin() (v interface{}, prio int32) {

	rng := rdrand()

	var i, j uint32

	// From the paper:
	// "It is an interesting question how to actually and reliably detect a globally empty queue."
	const maxAttempts = 10

	for attempt := 0; attempt < maxAttempts; attempt++ {
		rng = xorshiftMult64(rng)
		i = reduce(uint32(rng), len(q.locks))

		rng = xorshiftMult64(rng)
		j = reduce(uint32(rng), len(q.locks))

		mini := atomic.LoadInt32(&q.mins[i])
		minj := atomic.LoadInt32(&q.mins[j])

		if mini > minj {
			i, j = j, i
		}

		if mini == math.MaxInt32 {
			continue
		}

		gotlock := q.locks[i].trylock()
		if gotlock {
			if len(q.qs[i]) == 0 {
				// queue was empty -- unlock and try again
				q.locks[i].unlock()
				continue
			}

			e := heap.Pop(&q.qs[i])

			if len(q.qs[i]) > 0 {
				atomic.StoreInt32(&q.mins[i], q.qs[i][0].Priority)
			} else {
				atomic.StoreInt32(&q.mins[i], math.MaxInt32)
			}

			// unlock
			q.locks[i].unlock()

			item := e.(*pq.Item)

			return item.Value, item.Priority
		}

		if attempt > len(q.locks) {
			runtime.Gosched()
		}
	}

	return nil, math.MaxInt32
}

type lock uint32

func (l *lock) trylock() bool {
	return atomic.CompareAndSwapUint32((*uint32)(l), 0, 1)
}

func (l *lock) unlock() {
	atomic.StoreUint32((*uint32)(l), 0)
}

// http://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
func reduce(x uint32, n int) uint32 {
	return uint32((uint64(x) * uint64(n)) >> 32)
}

// 64-bit xorshift multiply rng from http://vigna.di.unimi.it/ftp/papers/xorshift.pdf
func xorshiftMult64(x uint64) uint64 {
	x ^= x >> 12 // a
	x ^= x << 25 // b
	x ^= x >> 27 // c
	return x * 2685821657736338717
}
