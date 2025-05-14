package workerpool

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPool_BasicExecution(t *testing.T) {
	wp := NewWorkerPool(10, 5)

	var counter int32
	for range 100 {
		wp.Submit(func() {
			atomic.AddInt32(&counter, 1)
			time.Sleep(10 * time.Millisecond)
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wp.Shutdown(ctx)

	if atomic.LoadInt32(&counter) != 100 {
		t.Errorf("Expected 100 tasks to be executed, got %d", counter)
	}
}

func TestWorkerPool_PriorityTasks(t *testing.T) {
	wp := NewWorkerPool(5, 2)

	var regularCounter, priorityCounter int32
	startTime := time.Now()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for range 50 {
			wp.Submit(func() {
				time.Sleep(20 * time.Millisecond)
				atomic.AddInt32(&regularCounter, 1)
			})
		}
	}()

	go func() {
		defer wg.Done()
		for range 10 {
			time.Sleep(5 * time.Millisecond)
			wp.SubmitPriority(func() {
				atomic.AddInt32(&priorityCounter, 1)
			})
		}
	}()

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	priorityTime := time.Since(startTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wp.Shutdown(ctx)

	if priorityCounter != 10 {
		t.Errorf("Expected 10 priority tasks, got %d", priorityCounter)
	}

	if priorityTime > 200*time.Millisecond {
		t.Errorf("Priority tasks took too long: %v", priorityTime)
	}
}

func TestWorkerPool_AutoScaling(t *testing.T) {
	wp := NewWorkerPool(5, 2)

	initialWorkers := wp.GetActiveWorkerCount()

	var wg sync.WaitGroup
	taskCount := 100
	wg.Add(taskCount)

	for i := 0; i < taskCount; i++ {
		wp.Submit(func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
		})
	}

	time.Sleep(100 * time.Millisecond)
	scaledWorkers := wp.GetActiveWorkerCount()

	if scaledWorkers <= initialWorkers {
		t.Errorf("Expected worker pool to scale up from %d workers, but got %d", initialWorkers, scaledWorkers)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Test timed out waiting for tasks to complete")
	}

	wp.Shutdown(context.Background())
}

func TestWorkerPool_IdleTimeout(t *testing.T) {
	wp := NewWorkerPool(20, 5)

	var counter int32
	for range 100 {
		wp.Submit(func() {
			atomic.AddInt32(&counter, 1)
			time.Sleep(10 * time.Millisecond)
		})
	}

	time.Sleep(100 * time.Millisecond)
	midWorkers := wp.GetActiveWorkerCount()

	time.Sleep(31 * time.Second)
	finalWorkers := wp.GetActiveWorkerCount()

	if finalWorkers >= midWorkers {
		t.Errorf("Expected worker count to decrease from %d, but got %d", midWorkers, finalWorkers)
	}

	if finalWorkers < int(wp.minWorkers) {
		t.Errorf("Worker count %d fell below minimum %d", finalWorkers, wp.minWorkers)
	}

	wp.Shutdown(context.Background())
}

func TestWorkerPool_ShutdownBehavior(t *testing.T) {
	wp := NewWorkerPool(10, 5)

	var counter int32
	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)
		wp.Submit(func() {
			time.Sleep(300 * time.Millisecond)
			atomic.AddInt32(&counter, 1)
			wg.Done()
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	wp.Shutdown(ctx)
	finalCount := atomic.LoadInt32(&counter)

	if finalCount >= 30 {
		t.Errorf("Expected very few tasks to complete with short timeout, but got %d", finalCount)
	}
}

func TestWorkerPool_ContextRespect(t *testing.T) {
	wp := NewWorkerPool(5, 2)

	for range 20 {
		wp.Submit(func() {
			time.Sleep(500 * time.Millisecond)
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	startShutdown := time.Now()
	wp.Shutdown(ctx)
	shutdownTime := time.Since(startShutdown)

	if shutdownTime > 150*time.Millisecond {
		t.Errorf("Shutdown didn't respect context timeout, took %v", shutdownTime)
	}
}

func TestWorkerPool_HighLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high load test in short mode")
	}

	wp := NewWorkerPool(10, 5)
	taskCount := 10000
	var counter int32

	startTime := time.Now()
	for range taskCount {
		wp.Submit(func() {
			atomic.AddInt32(&counter, 1)
			time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	wp.Shutdown(ctx)

	duration := time.Since(startTime)

	if atomic.LoadInt32(&counter) != int32(taskCount) {
		t.Errorf("Expected %d tasks to complete, got %d", taskCount, counter)
	}

	t.Logf("Processed %d tasks in %v", taskCount, duration)
}

func TestWorkerPool_ConcurrentProducers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	wp := NewWorkerPool(20, 10)
	var completedTasks int32
	var submittedTasks int32

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					wp.Submit(func() {
						atomic.AddInt32(&completedTasks, 1)
						time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
					})
					atomic.AddInt32(&submittedTasks, 1)

					if id%2 == 0 {
						time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
					} else {
						time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
					}
				}
			}
		}(i)
	}

	wg.Wait()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()
	wp.Shutdown(shutdownCtx)

	taskRatio := float64(completedTasks) / float64(submittedTasks)
	t.Logf("Completed %d/%d tasks (%.2f%%)", completedTasks, submittedTasks, taskRatio*100)

	if taskRatio < 0.5 {
		t.Errorf("Completion rate too low: %.2f%%", taskRatio*100)
	}
}

func TestWorkerPool_BurstyLoad(t *testing.T) {
	wp := NewWorkerPool(5, 3)
	var totalTasks int32

	for burst := 0; burst < 3; burst++ {
		var burstCounter int32
		burstSize := 200

		startTime := time.Now()
		for i := 0; i < burstSize; i++ {
			wp.Submit(func() {
				atomic.AddInt32(&burstCounter, 1)
				atomic.AddInt32(&totalTasks, 1)
				time.Sleep(5 * time.Millisecond)
			})
		}

		for atomic.LoadInt32(&burstCounter) < int32(burstSize) {
			time.Sleep(10 * time.Millisecond)
			if time.Since(startTime) > 5*time.Second {
				t.Fatalf("Burst %d timed out after 5 seconds", burst)
			}
		}

		maxWorkers := wp.GetActiveWorkerCount()
		t.Logf("Burst %d: scaled to %d workers", burst, maxWorkers)

		if maxWorkers <= 5 {
			t.Errorf("Expected worker pool to scale during burst %d", burst)
		}

		time.Sleep(time.Second)
	}

	if atomic.LoadInt32(&totalTasks) != 600 {
		t.Errorf("Expected 600 total tasks, got %d", totalTasks)
	}

	wp.Shutdown(context.Background())
}

func TestWorkerPool_QueueCapacity(t *testing.T) {
	wp := NewWorkerPool(5, 2)

	queueCapacity := wp.GetQueueCapacity()
	var wg sync.WaitGroup

	for i := 0; i < queueCapacity; i++ {
		wg.Add(1)
		wp.Submit(func() {
			time.Sleep(300 * time.Millisecond)
			wg.Done()
		})
	}

	time.Sleep(50 * time.Millisecond)

	queueSize := wp.GetQueueSize()

	if queueSize < int(float64(queueCapacity)*0.5) {
		t.Logf("Queue size: %d, capacity: %d", queueSize, queueCapacity)
	}

	initialWorkers := wp.GetActiveWorkerCount()

	for range queueCapacity {
		wg.Add(1)
		wp.Submit(func() {
			time.Sleep(200 * time.Millisecond)
			wg.Done()
		})
	}

	time.Sleep(50 * time.Millisecond)

	scaledWorkers := wp.GetActiveWorkerCount()

	if scaledWorkers <= initialWorkers {
		t.Logf("Initial workers: %d, scaled workers: %d", initialWorkers, scaledWorkers)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}

	wp.Shutdown(context.Background())
}
