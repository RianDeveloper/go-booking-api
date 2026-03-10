package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"go-project-2/internal/domain"
)

var pgxTxRepeatableRead = pgx.TxOptions{
	IsoLevel: pgx.RepeatableRead,
}

func (s *Store) LockPendingOutboxEvents(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	tx, err := s.pool.BeginTx(ctx, pgxTxRepeatableRead)
	if err != nil {
		return nil, fmt.Errorf("begin outbox tx: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		select id, event_type, aggregate_id, payload, status, attempt, next_attempt_at, created_at, published_at
		from outbox_events
		where status = 'pending'
		  and next_attempt_at <= now()
		order by created_at
		limit $1
		for update skip locked
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query pending outbox: %w", err)
	}
	defer rows.Close()

	events := make([]domain.OutboxEvent, 0)
	for rows.Next() {
		var ev domain.OutboxEvent
		if err := rows.Scan(
			&ev.ID,
			&ev.EventType,
			&ev.AggregateID,
			&ev.Payload,
			&ev.Status,
			&ev.Attempt,
			&ev.NextAttemptAt,
			&ev.CreatedAt,
			&ev.PublishedAt,
		); err != nil {
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}
		events = append(events, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit outbox tx: %w", err)
	}
	return events, nil
}

func (s *Store) MarkOutboxEventPublished(ctx context.Context, eventID string) error {
	_, err := s.pool.Exec(ctx, `
		update outbox_events
		set status = 'published',
		    published_at = now()
		where id = $1
	`, eventID)
	if err != nil {
		return fmt.Errorf("mark event published: %w", err)
	}
	return nil
}

func (s *Store) MarkOutboxEventRetry(ctx context.Context, eventID string, nextAttemptAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		update outbox_events
		set attempt = attempt + 1,
		    next_attempt_at = $2
		where id = $1
	`, eventID, nextAttemptAt)
	if err != nil {
		return fmt.Errorf("mark event retry: %w", err)
	}
	return nil
}
