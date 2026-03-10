package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"go-project-2/internal/domain"
)

func (t *Tx) CreateResource(ctx context.Context, resource domain.Resource) error {
	_, err := t.tx.Exec(ctx, `
		insert into resources(id, name, created_at)
		values ($1, $2, $3)
	`, resource.ID, resource.Name, resource.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert resource: %w", err)
	}
	return nil
}

func (t *Tx) CreateUser(ctx context.Context, user domain.User) error {
	_, err := t.tx.Exec(ctx, `
		insert into users(id, email, name, created_at)
		values ($1, $2, $3, $4)
	`, user.ID, user.Email, user.Name, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (t *Tx) CreateSlots(ctx context.Context, slots []domain.Slot) error {
	for _, slot := range slots {
		_, err := t.tx.Exec(ctx, `
			insert into slots(id, resource_id, start_time, end_time, status, version, created_at)
			values ($1, $2, $3, $4, $5, $6, $7)
		`, slot.ID, slot.ResourceID, slot.StartTime, slot.EndTime, slot.Status, slot.Version, slot.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert slot: %w", err)
		}
	}
	return nil
}

func (t *Tx) GetSlotForUpdate(ctx context.Context, slotID string) (domain.Slot, error) {
	row := t.tx.QueryRow(ctx, `
		select id, resource_id, start_time, end_time, status, version, created_at
		from slots
		where id = $1
		for update
	`, slotID)
	var slot domain.Slot
	err := row.Scan(
		&slot.ID,
		&slot.ResourceID,
		&slot.StartTime,
		&slot.EndTime,
		&slot.Status,
		&slot.Version,
		&slot.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Slot{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Slot{}, fmt.Errorf("scan slot: %w", err)
	}
	return slot, nil
}

func (t *Tx) MarkSlotBooked(ctx context.Context, slotID string, expectedVersion int64) error {
	cmd, err := t.tx.Exec(ctx, `
		update slots
		set status = 'booked',
		    version = version + 1
		where id = $1
		  and status = 'free'
		  and version = $2
	`, slotID, expectedVersion)
	if err != nil {
		return fmt.Errorf("update slot booked: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrSlotAlreadyBooked
	}
	return nil
}

func (t *Tx) MarkSlotFree(ctx context.Context, slotID string) error {
	_, err := t.tx.Exec(ctx, `
		update slots
		set status = 'free',
		    version = version + 1
		where id = $1
	`, slotID)
	if err != nil {
		return fmt.Errorf("update slot free: %w", err)
	}
	return nil
}

func (t *Tx) CreateBooking(ctx context.Context, booking domain.Booking) error {
	_, err := t.tx.Exec(ctx, `
		insert into bookings(id, user_id, resource_id, slot_id, status, created_at, cancelled_at)
		values ($1, $2, $3, $4, $5, $6, $7)
	`, booking.ID, booking.UserID, booking.ResourceID, booking.SlotID, booking.Status, booking.CreatedAt, booking.CancelledAt)
	if err != nil {
		return fmt.Errorf("insert booking: %w", err)
	}
	return nil
}

func (t *Tx) GetBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	row := t.tx.QueryRow(ctx, `
		select id, user_id, resource_id, slot_id, status, created_at, cancelled_at
		from bookings
		where id = $1
		for update
	`, bookingID)
	var booking domain.Booking
	err := row.Scan(
		&booking.ID,
		&booking.UserID,
		&booking.ResourceID,
		&booking.SlotID,
		&booking.Status,
		&booking.CreatedAt,
		&booking.CancelledAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Booking{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Booking{}, fmt.Errorf("scan booking: %w", err)
	}
	return booking, nil
}

func (t *Tx) CancelBooking(ctx context.Context, bookingID string) error {
	now := time.Now().UTC()
	_, err := t.tx.Exec(ctx, `
		update bookings
		set status = 'cancelled',
		    cancelled_at = $2
		where id = $1
	`, bookingID, now)
	if err != nil {
		return fmt.Errorf("cancel booking: %w", err)
	}
	return nil
}

func (t *Tx) CreateOutboxEvent(ctx context.Context, event domain.OutboxEvent) error {
	_, err := t.tx.Exec(ctx, `
		insert into outbox_events(
			id, event_type, aggregate_id, payload, status, attempt, next_attempt_at, created_at, published_at
		)
		values($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, event.ID, event.EventType, event.AggregateID, event.Payload, event.Status, event.Attempt, event.NextAttemptAt, event.CreatedAt, event.PublishedAt)
	if err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}
