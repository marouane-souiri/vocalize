package interfaces

import (
	"context"

	"github.com/marouane-souiri/vocalize/internal/domain"
)

type WorkerPool interface {
	Submit(task domain.Task)
	SubmitPriority(task domain.Task)
	Shutdown(ctx context.Context)
	GetActiveWorkerCount() int
	GetMinWorkersCount() int
	GetQueueSize() int
	GetQueueCapacity() int
}
