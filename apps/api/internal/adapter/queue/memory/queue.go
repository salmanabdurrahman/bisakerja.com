package memory

import (
	"context"
	"sync"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

// Queue represents queue.
type Queue struct {
	mu            sync.Mutex
	jobEvents     []notification.JobEvent
	deliveryTasks []notification.DeliveryTask
}

// NewQueue creates a new queue instance.
func NewQueue() *Queue {
	return &Queue{
		jobEvents:     []notification.JobEvent{},
		deliveryTasks: []notification.DeliveryTask{},
	}
}

// EnqueueJobEvent handles enqueue job event.
func (q *Queue) EnqueueJobEvent(_ context.Context, event notification.JobEvent) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobEvents = append(q.jobEvents, event)
	return nil
}

// DequeueJobEvents handles dequeue job events.
func (q *Queue) DequeueJobEvents(_ context.Context, limit int) ([]notification.JobEvent, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := dequeueCount(limit, len(q.jobEvents))
	events := append([]notification.JobEvent(nil), q.jobEvents[:count]...)
	q.jobEvents = append([]notification.JobEvent(nil), q.jobEvents[count:]...)
	return events, nil
}

// EnqueueDeliveryTask handles enqueue delivery task.
func (q *Queue) EnqueueDeliveryTask(_ context.Context, task notification.DeliveryTask) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.deliveryTasks = append(q.deliveryTasks, task)
	return nil
}

// DequeueDeliveryTasks handles dequeue delivery tasks.
func (q *Queue) DequeueDeliveryTasks(_ context.Context, limit int) ([]notification.DeliveryTask, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := dequeueCount(limit, len(q.deliveryTasks))
	tasks := append([]notification.DeliveryTask(nil), q.deliveryTasks[:count]...)
	q.deliveryTasks = append([]notification.DeliveryTask(nil), q.deliveryTasks[count:]...)
	return tasks, nil
}

func dequeueCount(limit, length int) int {
	if length == 0 {
		return 0
	}
	if limit <= 0 || limit > length {
		return length
	}
	return limit
}
