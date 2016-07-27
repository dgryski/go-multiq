package multiq

import (
	"container/heap"
	"math"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dgryski/go-multiq/internal/pq"
)

func TestMultiQueue(t *testing.T) {

	q := New(4)

	for i := 0; i < 100; i++ {
		q.Insert(i, int32(i))
	}

	var vs []int
	for i := 0; i < 100; i++ {
		t.Logf("mins=%v", q.mins)
		v, p := q.DeleteMin()
		if p == math.MaxInt32 {
			continue
		}
		vs = append(vs, v.(int))
	}

	t.Log(vs)
}

type mpq struct {
	pq.PriorityQueue
	sync.Mutex
}

func (mq *mpq) Insert(v interface{}, prio int32) {
	mq.Lock()
	heap.Push(mq, &pq.Item{v, prio})
	mq.Unlock()
}

func (mq *mpq) DeleteMin() (interface{}, int32) {
	mq.Lock()
	e := heap.Pop(mq)
	mq.Unlock()
	item := e.(*pq.Item)
	return item.Value, item.Priority
}

type queue interface {
	Insert(v interface{}, prio int32)
	DeleteMin() (interface{}, int32)
}

var total int32

func benchmarkPQ(b *testing.B, q queue, g int) {

	var wg sync.WaitGroup

	for gg := 0; gg < g; gg++ {
		wg.Add(1)
		go func() {
			var t int32
			for i := 0; i < b.N; i++ {
				q.Insert(i, int32(i))
			}

			for i := 0; i < b.N; i++ {
				v, p := q.DeleteMin()
				if p == math.MaxInt32 {
					continue
				}
				t += int32(v.(int))
			}

			atomic.AddInt32(&total, t)
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkPQ_1(b *testing.B) { benchmarkPQ(b, &mpq{}, 1) }
func BenchmarkPQ_2(b *testing.B) { benchmarkPQ(b, &mpq{}, 2) }
func BenchmarkPQ_4(b *testing.B) { benchmarkPQ(b, &mpq{}, 4) }
func BenchmarkPQ_8(b *testing.B) { benchmarkPQ(b, &mpq{}, 8) }

func BenchmarkMQ_1(b *testing.B) { benchmarkPQ(b, New(4), 1) }
func BenchmarkMQ_2(b *testing.B) { benchmarkPQ(b, New(4), 2) }
func BenchmarkMQ_4(b *testing.B) { benchmarkPQ(b, New(4), 4) }
func BenchmarkMQ_8(b *testing.B) { benchmarkPQ(b, New(4), 8) }
