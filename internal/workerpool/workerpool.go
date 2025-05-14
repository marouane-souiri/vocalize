package workerpool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Task func()

type WorkerPool struct {
	tasks         chan Task
	wg            sync.WaitGroup
	activeWorkers int32
	minWorkers    int32
	mu            sync.Mutex
	done          chan struct{}
	shutdownOnce  sync.Once
}

func NewWorkerPool(initialWorkers int, minWorkers int) *WorkerPool {
	if initialWorkers <= 0 {
		initialWorkers = 10
	}
	if minWorkers <= 0 {
		minWorkers = 5
	}
	wp := &WorkerPool{
		tasks:         make(chan Task, initialWorkers),
		activeWorkers: 0,
		minWorkers:    int32(minWorkers),
		done:          make(chan struct{}),
	}
	for range initialWorkers {
		wp.addWorker()
	}
	return wp
}

func (wp *WorkerPool) Submit(task Task) {
	queueLoad := float64(len(wp.tasks)) / float64(cap(wp.tasks))
	if queueLoad >= 0.7 {
		workersToAdd := max(cap(wp.tasks)/2, 1)
		for range workersToAdd {
			wp.addWorker()
		}
	}
	wp.tasks <- task
}

func (wp *WorkerPool) SubmitPriority(task Task) {
	wp.wg.Add(1)
	atomic.AddInt32(&wp.activeWorkers, 1)
	go func() {
		defer wp.wg.Done()
		defer atomic.AddInt32(&wp.activeWorkers, -1)
		task()
	}()
}

func (wp *WorkerPool) addWorker() {
	wp.wg.Add(1)
	atomic.AddInt32(&wp.activeWorkers, 1)

	go func() {
		defer wp.wg.Done()
		defer atomic.AddInt32(&wp.activeWorkers, -1)

		for {
			select {
			case task, ok := <-wp.tasks:
				if !ok {
					return
				}
				task()

			case <-time.After(30 * time.Second):
				if atomic.LoadInt32(&wp.activeWorkers) > wp.minWorkers {
					return
				}

			case <-wp.done:
				return
			}
		}
	}()
}

func (wp *WorkerPool) Shutdown(ctx context.Context) {
	close(wp.done)

	wp.shutdownOnce.Do(func() {
		close(wp.tasks)
	})

	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
}

func (wp *WorkerPool) GetActiveWorkerCount() int {
	return int(atomic.LoadInt32(&wp.activeWorkers))
}

func (wp *WorkerPool) GetQueueSize() int {
	return len(wp.tasks)
}

func (wp *WorkerPool) GetQueueCapacity() int {
	return cap(wp.tasks)
}
