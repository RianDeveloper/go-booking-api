package service

import (
	"context"
	"testing"
	"time"

	"go-project-2/internal/domain"
	"go-project-2/internal/ports"
)

type fakeStore struct {
	txImpl ports.Tx
	booking domain.Booking
	slots []domain.Slot
}

func (f *fakeStore) WithTx(ctx context.Context, fn func(tx ports.Tx) error) error {
	return fn(f.txImpl)
}

func (f *fakeStore) GetBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	if f.booking.ID == bookingID {
		return f.booking, nil
	}
	return domain.Booking{}, domain.ErrNotFound
}

func (f *fakeStore) GetResourceSchedule(ctx context.Context, resourceID string, from, to time.Time) ([]domain.Slot, error) {
	return f.slots, nil
}

type fakeTx struct {
	slot domain.Slot
	booking domain.Booking
}

func (f *fakeTx) CreateResource(ctx context.Context, resource domain.Resource) error { return nil }
func (f *fakeTx) CreateUser(ctx context.Context, user domain.User) error { return nil }
func (f *fakeTx) CreateSlots(ctx context.Context, slots []domain.Slot) error { return nil }
func (f *fakeTx) GetSlotForUpdate(ctx context.Context, slotID string) (domain.Slot, error) { return f.slot, nil }
func (f *fakeTx) MarkSlotBooked(ctx context.Context, slotID string, expectedVersion int64) error {
	if f.slot.Status != domain.SlotStatusFree {
		return domain.ErrSlotAlreadyBooked
	}
	return nil
}
func (f *fakeTx) MarkSlotFree(ctx context.Context, slotID string) error { return nil }
func (f *fakeTx) CreateBooking(ctx context.Context, booking domain.Booking) error { return nil }
func (f *fakeTx) GetBooking(ctx context.Context, bookingID string) (domain.Booking, error) { return f.booking, nil }
func (f *fakeTx) CancelBooking(ctx context.Context, bookingID string) error { return nil }
func (f *fakeTx) CreateOutboxEvent(ctx context.Context, event domain.OutboxEvent) error { return nil }

func TestCreateSlotsInvalidRange(t *testing.T) {
	svc := New(&fakeStore{txImpl: &fakeTx{}})
	_, err := svc.CreateSlots(context.Background(), CreateSlotsInput{
		ResourceID: "r1",
		Slots: []TimeRange{
			{StartTime: time.Now(), EndTime: time.Now().Add(-time.Minute)},
		},
	})
	if err != domain.ErrInvalidTimeRange {
		t.Fatalf("expected invalid range, got %v", err)
	}
}

func TestCreateBookingConflict(t *testing.T) {
	tx := &fakeTx{
		slot: domain.Slot{
			ID:      "s1",
			Status:  domain.SlotStatusBooked,
			Version: 1,
		},
	}
	svc := New(&fakeStore{txImpl: tx})
	_, err := svc.CreateBooking(context.Background(), CreateBookingInput{UserID: "u1", SlotID: "s1"})
	if err != domain.ErrSlotAlreadyBooked {
		t.Fatalf("expected slot already booked, got %v", err)
	}
}

func TestCancelBookingAlreadyCancelled(t *testing.T) {
	now := time.Now().UTC()
	tx := &fakeTx{
		booking: domain.Booking{
			ID:     "b1",
			Status: domain.BookingStatusCancelled,
			CancelledAt: &now,
		},
	}
	svc := New(&fakeStore{txImpl: tx})
	_, err := svc.CancelBooking(context.Background(), "b1")
	if err != domain.ErrBookingAlreadyEnded {
		t.Fatalf("expected already cancelled, got %v", err)
	}
}
