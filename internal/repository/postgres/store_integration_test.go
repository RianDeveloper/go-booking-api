package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-project-2/internal/domain"
	"go-project-2/internal/ports"
)

func TestStoreBookingFlowIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not set")
	}
	ctx := context.Background()
	store, err := New(ctx, dsn)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	upPath := filepath.Join("..", "..", "..", "migrations", "000001_init.up.sql")
	schema, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := store.pool.Exec(ctx, string(schema)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	resource := domain.Resource{
		ID:        "00000000-0000-0000-0000-000000000001",
		Name:      "room",
		CreatedAt: time.Now().UTC(),
	}
	user := domain.User{
		ID:        "00000000-0000-0000-0000-000000000002",
		Email:     "user@test.local",
		Name:      "User",
		CreatedAt: time.Now().UTC(),
	}
	slot := domain.Slot{
		ID:         "00000000-0000-0000-0000-000000000003",
		ResourceID: resource.ID,
		StartTime:  time.Now().UTC().Add(time.Hour),
		EndTime:    time.Now().UTC().Add(2 * time.Hour),
		Status:     domain.SlotStatusFree,
		Version:    1,
		CreatedAt:  time.Now().UTC(),
	}
	booking := domain.Booking{
		ID:         "00000000-0000-0000-0000-000000000004",
		UserID:     user.ID,
		ResourceID: resource.ID,
		SlotID:     slot.ID,
		Status:     domain.BookingStatusCreated,
		CreatedAt:  time.Now().UTC(),
	}
	err = store.WithTx(ctx, func(tx ports.Tx) error {
		if err := tx.CreateResource(ctx, resource); err != nil {
			return err
		}
		if err := tx.CreateUser(ctx, user); err != nil {
			return err
		}
		if err := tx.CreateSlots(ctx, []domain.Slot{slot}); err != nil {
			return err
		}
		if err := tx.CreateBooking(ctx, booking); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx flow: %v", err)
	}
	got, err := store.GetBooking(ctx, booking.ID)
	if err != nil {
		t.Fatalf("get booking: %v", err)
	}
	if got.ID != booking.ID {
		t.Fatalf("unexpected booking id: %s", got.ID)
	}
}
