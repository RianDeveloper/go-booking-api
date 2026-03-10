package ports

import (
	"context"
	"time"

	"go-project-2/internal/domain"
)

type Tx interface {
	CreateResource(ctx context.Context, resource domain.Resource) error
	CreateUser(ctx context.Context, user domain.User) error
	CreateSlots(ctx context.Context, slots []domain.Slot) error
	GetSlotForUpdate(ctx context.Context, slotID string) (domain.Slot, error)
	MarkSlotBooked(ctx context.Context, slotID string, expectedVersion int64) error
	MarkSlotFree(ctx context.Context, slotID string) error
	CreateBooking(ctx context.Context, booking domain.Booking) error
	GetBooking(ctx context.Context, bookingID string) (domain.Booking, error)
	CancelBooking(ctx context.Context, bookingID string) error
	CreateOutboxEvent(ctx context.Context, event domain.OutboxEvent) error
}

type BookingStore interface {
	WithTx(ctx context.Context, fn func(tx Tx) error) error
	GetBooking(ctx context.Context, bookingID string) (domain.Booking, error)
	GetResourceSchedule(ctx context.Context, resourceID string, from, to time.Time) ([]domain.Slot, error)
}
