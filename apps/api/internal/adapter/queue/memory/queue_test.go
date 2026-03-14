package memory

import (
	"context"
	"testing"

	"github.com/salmanabdurrahman/bisakerja.com/apps/api/internal/domain/notification"
)

func TestQueue_JobEventsFIFO(t *testing.T) {
	queue := NewQueue()

	_ = queue.EnqueueJobEvent(context.Background(), notification.JobEvent{JobID: "job_1"})
	_ = queue.EnqueueJobEvent(context.Background(), notification.JobEvent{JobID: "job_2"})

	events, err := queue.DequeueJobEvents(context.Background(), 1)
	if err != nil {
		t.Fatalf("dequeue job events: %v", err)
	}
	if len(events) != 1 || events[0].JobID != "job_1" {
		t.Fatalf("expected first event job_1, got %+v", events)
	}

	remaining, err := queue.DequeueJobEvents(context.Background(), 10)
	if err != nil {
		t.Fatalf("dequeue remaining events: %v", err)
	}
	if len(remaining) != 1 || remaining[0].JobID != "job_2" {
		t.Fatalf("expected remaining event job_2, got %+v", remaining)
	}
}

func TestQueue_DeliveryTasksFIFO(t *testing.T) {
	queue := NewQueue()
	_ = queue.EnqueueDeliveryTask(context.Background(), notification.DeliveryTask{NotificationID: "notif_1"})
	_ = queue.EnqueueDeliveryTask(context.Background(), notification.DeliveryTask{NotificationID: "notif_2"})

	tasks, err := queue.DequeueDeliveryTasks(context.Background(), 1)
	if err != nil {
		t.Fatalf("dequeue delivery tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].NotificationID != "notif_1" {
		t.Fatalf("expected notif_1 first, got %+v", tasks)
	}

	remaining, err := queue.DequeueDeliveryTasks(context.Background(), 2)
	if err != nil {
		t.Fatalf("dequeue delivery tasks: %v", err)
	}
	if len(remaining) != 1 || remaining[0].NotificationID != "notif_2" {
		t.Fatalf("expected notif_2 remaining, got %+v", remaining)
	}
}
