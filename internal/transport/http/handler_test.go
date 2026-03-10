package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-project-2/internal/domain"
	"go-project-2/internal/service"
)

type mockService struct {
	createBookingFn func(ctx context.Context, in service.CreateBookingInput) (domain.Booking, error)
}

func (m *mockService) CreateResource(ctx context.Context, in service.CreateResourceInput) (domain.Resource, error) {
	return domain.Resource{}, nil
}
func (m *mockService) CreateUser(ctx context.Context, in service.CreateUserInput) (domain.User, error) {
	return domain.User{}, nil
}
func (m *mockService) CreateSlots(ctx context.Context, in service.CreateSlotsInput) ([]domain.Slot, error) {
	return nil, nil
}
func (m *mockService) CreateBooking(ctx context.Context, in service.CreateBookingInput) (domain.Booking, error) {
	return m.createBookingFn(ctx, in)
}
func (m *mockService) CancelBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	return domain.Booking{}, nil
}
func (m *mockService) GetBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	return domain.Booking{}, nil
}
func (m *mockService) GetResourceSchedule(ctx context.Context, resourceID string, from, to time.Time) ([]domain.Slot, error) {
	return nil, nil
}

func TestCreateBookingOK(t *testing.T) {
	h := NewHandler(&mockService{
		createBookingFn: func(ctx context.Context, in service.CreateBookingInput) (domain.Booking, error) {
			return domain.Booking{
				ID:         "b1",
				UserID:     in.UserID,
				SlotID:     in.SlotID,
				ResourceID: "r1",
				Status:     domain.BookingStatusCreated,
				CreatedAt:  time.Now().UTC(),
			}, nil
		},
	})
	body, _ := json.Marshal(map[string]string{
		"user_id": "u1",
		"slot_id": "s1",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/bookings", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
}

func TestCreateBookingConflict(t *testing.T) {
	h := NewHandler(&mockService{
		createBookingFn: func(ctx context.Context, in service.CreateBookingInput) (domain.Booking, error) {
			return domain.Booking{}, domain.ErrSlotAlreadyBooked
		},
	})
	body, _ := json.Marshal(map[string]string{
		"user_id": "u1",
		"slot_id": "s1",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/bookings", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rec.Code)
	}
}
