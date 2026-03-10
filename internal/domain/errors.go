package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrSlotAlreadyBooked   = errors.New("slot already booked")
	ErrInvalidTimeRange    = errors.New("invalid time range")
	ErrBookingAlreadyEnded = errors.New("booking already cancelled")
)
