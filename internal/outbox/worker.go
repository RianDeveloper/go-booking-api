package outbox

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"

	"go-project-2/internal/domain"
)

type EventStore interface {
	LockPendingOutboxEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkOutboxEventPublished(ctx context.Context, eventID string) error
	MarkOutboxEventRetry(ctx context.Context, eventID string, nextAttemptAt time.Time) error
}

type Worker struct {
	store        EventStore
	writer       *kafka.Writer
	batchSize    int
	pollInterval time.Duration
}

func NewWorker(store EventStore, writer *kafka.Writer, batchSize int, pollInterval time.Duration) *Worker {
	return &Worker{
		store:        store,
		writer:       writer,
		batchSize:    batchSize,
		pollInterval: pollInterval,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.flush(ctx); err != nil {
				continue
			}
		}
	}
}

func (w *Worker) flush(ctx context.Context) error {
	events, err := w.store.LockPendingOutboxEvents(ctx, w.batchSize)
	if err != nil {
		return err
	}
	for _, ev := range events {
		msg := kafka.Message{
			Key:   []byte(ev.AggregateID),
			Value: ev.Payload,
			Time:  time.Now().UTC(),
			Headers: []kafka.Header{
				{Key: "event_type", Value: []byte(ev.EventType)},
				{Key: "event_id", Value: []byte(ev.ID)},
			},
		}
		err := w.writer.WriteMessages(ctx, msg)
		if err != nil {
			next := time.Now().UTC().Add(time.Duration(ev.Attempt+1) * time.Second)
			_ = w.store.MarkOutboxEventRetry(ctx, ev.ID, next)
			continue
		}
		if err := w.store.MarkOutboxEventPublished(ctx, ev.ID); err != nil {
			continue
		}
	}
	return nil
}
