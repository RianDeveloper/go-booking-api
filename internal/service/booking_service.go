package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"go-project-2/internal/domain"
	"go-project-2/internal/ports"
)

type Service struct {
	store ports.BookingStore
}

func New(store ports.BookingStore) *Service {
	return &Service{store: store}
}

type CreateResourceInput struct {
	Name string
}

func (s *Service) CreateResource(ctx context.Context, in CreateResourceInput) (domain.Resource, error) {
	resource := domain.Resource{
		ID:        uuid.NewString(),
		Name:      in.Name,
		CreatedAt: time.Now().UTC(),
	}
	err := s.store.WithTx(ctx, func(tx ports.Tx) error {
		return tx.CreateResource(ctx, resource)
	})
	if err != nil {
		return domain.Resource{}, err
	}
	return resource, nil
}

type CreateUserInput struct {
	Email string
	Name  string
}

func (s *Service) CreateUser(ctx context.Context, in CreateUserInput) (domain.User, error) {
	user := domain.User{
		ID:        uuid.NewString(),
		Email:     in.Email,
		Name:      in.Name,
		CreatedAt: time.Now().UTC(),
	}
	err := s.store.WithTx(ctx, func(tx ports.Tx) error {
		return tx.CreateUser(ctx, user)
	})
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

type CreateSlotsInput struct {
	ResourceID string
	Slots      []TimeRange
}

type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}

func (s *Service) CreateSlots(ctx context.Context, in CreateSlotsInput) ([]domain.Slot, error) {
	now := time.Now().UTC()
	slots := make([]domain.Slot, 0, len(in.Slots))
	for _, tr := range in.Slots {
		if !tr.EndTime.After(tr.StartTime) {
			return nil, domain.ErrInvalidTimeRange
		}
		slots = append(slots, domain.Slot{
			ID:         uuid.NewString(),
			ResourceID: in.ResourceID,
			StartTime:  tr.StartTime.UTC(),
			EndTime:    tr.EndTime.UTC(),
			Status:     domain.SlotStatusFree,
			Version:    1,
			CreatedAt:  now,
		})
	}
	err := s.store.WithTx(ctx, func(tx ports.Tx) error {
		return tx.CreateSlots(ctx, slots)
	})
	if err != nil {
		return nil, err
	}
	return slots, nil
}

type CreateBookingInput struct {
	UserID string
	SlotID string
}

func (s *Service) CreateBooking(ctx context.Context, in CreateBookingInput) (domain.Booking, error) {
	now := time.Now().UTC()
	booking := domain.Booking{
		ID:        uuid.NewString(),
		UserID:    in.UserID,
		Status:    domain.BookingStatusCreated,
		CreatedAt: now,
	}
	err := s.store.WithTx(ctx, func(tx ports.Tx) error {
		slot, err := tx.GetSlotForUpdate(ctx, in.SlotID)
		if err != nil {
			return err
		}
		if slot.Status != domain.SlotStatusFree {
			return domain.ErrSlotAlreadyBooked
		}
		err = tx.MarkSlotBooked(ctx, slot.ID, slot.Version)
		if err != nil {
			return err
		}
		booking.SlotID = slot.ID
		booking.ResourceID = slot.ResourceID
		err = tx.CreateBooking(ctx, booking)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(map[string]any{
			"booking_id":  booking.ID,
			"user_id":     booking.UserID,
			"resource_id": booking.ResourceID,
			"slot_id":     booking.SlotID,
			"status":      booking.Status,
			"created_at":  booking.CreatedAt,
		})
		if err != nil {
			return fmt.Errorf("marshal event payload: %w", err)
		}
		event := domain.OutboxEvent{
			ID:            uuid.NewString(),
			EventType:     "booking.created",
			AggregateID:   booking.ID,
			Payload:       payload,
			Status:        "pending",
			Attempt:       0,
			NextAttemptAt: now,
			CreatedAt:     now,
		}
		return tx.CreateOutboxEvent(ctx, event)
	})
	if err != nil {
		return domain.Booking{}, err
	}
	return booking, nil
}

func (s *Service) CancelBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	var booking domain.Booking
	now := time.Now().UTC()
	err := s.store.WithTx(ctx, func(tx ports.Tx) error {
		current, err := tx.GetBooking(ctx, bookingID)
		if err != nil {
			return err
		}
		if current.Status == domain.BookingStatusCancelled {
			return domain.ErrBookingAlreadyEnded
		}
		err = tx.CancelBooking(ctx, bookingID)
		if err != nil {
			return err
		}
		err = tx.MarkSlotFree(ctx, current.SlotID)
		if err != nil {
			return err
		}
		current.Status = domain.BookingStatusCancelled
		current.CancelledAt = &now
		payload, err := json.Marshal(map[string]any{
			"booking_id":   current.ID,
			"user_id":      current.UserID,
			"resource_id":  current.ResourceID,
			"slot_id":      current.SlotID,
			"status":       current.Status,
			"cancelled_at": current.CancelledAt,
		})
		if err != nil {
			return fmt.Errorf("marshal event payload: %w", err)
		}
		event := domain.OutboxEvent{
			ID:            uuid.NewString(),
			EventType:     "booking.cancelled",
			AggregateID:   current.ID,
			Payload:       payload,
			Status:        "pending",
			Attempt:       0,
			NextAttemptAt: now,
			CreatedAt:     now,
		}
		err = tx.CreateOutboxEvent(ctx, event)
		if err != nil {
			return err
		}
		booking = current
		return nil
	})
	if err != nil {
		return domain.Booking{}, err
	}
	return booking, nil
}

func (s *Service) GetBooking(ctx context.Context, bookingID string) (domain.Booking, error) {
	return s.store.GetBooking(ctx, bookingID)
}

func (s *Service) GetResourceSchedule(ctx context.Context, resourceID string, from, to time.Time) ([]domain.Slot, error) {
	if !to.After(from) {
		return nil, domain.ErrInvalidTimeRange
	}
	return s.store.GetResourceSchedule(ctx, resourceID, from.UTC(), to.UTC())
}
