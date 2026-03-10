package domain

import "time"

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Resource struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type SlotStatus string

const (
	SlotStatusFree   SlotStatus = "free"
	SlotStatusBooked SlotStatus = "booked"
)

type Slot struct {
	ID         string     `json:"id"`
	ResourceID string     `json:"resource_id"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    time.Time  `json:"end_time"`
	Status     SlotStatus `json:"status"`
	Version    int64      `json:"version"`
	CreatedAt  time.Time  `json:"created_at"`
}

type BookingStatus string

const (
	BookingStatusCreated   BookingStatus = "created"
	BookingStatusCancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID         string        `json:"id"`
	UserID     string        `json:"user_id"`
	ResourceID string        `json:"resource_id"`
	SlotID     string        `json:"slot_id"`
	Status     BookingStatus `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	CancelledAt *time.Time   `json:"cancelled_at,omitempty"`
}

type OutboxEvent struct {
	ID            string     `json:"id"`
	EventType     string     `json:"event_type"`
	AggregateID   string     `json:"aggregate_id"`
	Payload       []byte     `json:"payload"`
	Status        string     `json:"status"`
	Attempt       int        `json:"attempt"`
	NextAttemptAt time.Time  `json:"next_attempt_at"`
	CreatedAt     time.Time  `json:"created_at"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
}
