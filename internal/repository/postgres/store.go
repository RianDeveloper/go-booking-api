package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-project-2/internal/domain"
	"go-project-2/internal/ports"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) WithTx(ctx context.Context, fn func(tx ports.Tx) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	repoTx := &Tx{tx: tx}
	if err := fn(repoTx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback tx: %w", rbErr)
		}
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (s *Store) GetBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	row := s.pool.QueryRow(ctx, `
		select id, user_id, resource_id, slot_id, status, created_at, cancelled_at
		from bookings
		where id = $1
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

func (s *Store) GetResourceSchedule(ctx context.Context, resourceID string, from, to time.Time) ([]domain.Slot, error) {
	rows, err := s.pool.Query(ctx, `
		select id, resource_id, start_time, end_time, status, version, created_at
		from slots
		where resource_id = $1
		  and start_time >= $2
		  and end_time <= $3
		order by start_time asc
	`, resourceID, from, to)
	if err != nil {
		return nil, fmt.Errorf("query schedule: %w", err)
	}
	defer rows.Close()
	out := make([]domain.Slot, 0)
	for rows.Next() {
		var slot domain.Slot
		if err := rows.Scan(
			&slot.ID,
			&slot.ResourceID,
			&slot.StartTime,
			&slot.EndTime,
			&slot.Status,
			&slot.Version,
			&slot.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan slot: %w", err)
		}
		out = append(out, slot)
	}
	return out, rows.Err()
}

type Tx struct {
	tx pgx.Tx
}
