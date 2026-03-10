package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"go-project-2/internal/domain"
	"go-project-2/internal/service"
)

type BookingService interface {
	CreateResource(ctx context.Context, in service.CreateResourceInput) (domain.Resource, error)
	CreateUser(ctx context.Context, in service.CreateUserInput) (domain.User, error)
	CreateSlots(ctx context.Context, in service.CreateSlotsInput) ([]domain.Slot, error)
	CreateBooking(ctx context.Context, in service.CreateBookingInput) (domain.Booking, error)
	CancelBooking(ctx context.Context, bookingID string) (domain.Booking, error)
	GetBooking(ctx context.Context, bookingID string) (domain.Booking, error)
	GetResourceSchedule(ctx context.Context, resourceID string, from, to time.Time) ([]domain.Slot, error)
}

type Handler struct {
	svc BookingService
}

func NewHandler(svc BookingService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Post("/v1/users", h.createUser)
	r.Post("/v1/resources", h.createResource)
	r.Post("/v1/resources/{id}/slots", h.createSlots)
	r.Get("/v1/resources/{id}/schedule", h.getSchedule)
	r.Post("/v1/bookings", h.createBooking)
	r.Post("/v1/bookings/{id}/cancel", h.cancelBooking)
	r.Get("/v1/bookings/{id}", h.getBooking)
	return r
}

type createUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	user, err := h.svc.CreateUser(r.Context(), service.CreateUserInput{
		Email: req.Email,
		Name:  req.Name,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

type createResourceRequest struct {
	Name string `json:"name"`
}

func (h *Handler) createResource(w http.ResponseWriter, r *http.Request) {
	var req createResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	resource, err := h.svc.CreateResource(r.Context(), service.CreateResourceInput{Name: req.Name})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resource)
}

type createSlotsRequest struct {
	Slots []struct {
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	} `json:"slots"`
}

func (h *Handler) createSlots(w http.ResponseWriter, r *http.Request) {
	resourceID := chi.URLParam(r, "id")
	var req createSlotsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	ranges := make([]service.TimeRange, 0, len(req.Slots))
	for _, slot := range req.Slots {
		ranges = append(ranges, service.TimeRange{
			StartTime: slot.StartTime,
			EndTime:   slot.EndTime,
		})
	}
	out, err := h.svc.CreateSlots(r.Context(), service.CreateSlotsInput{
		ResourceID: resourceID,
		Slots:      ranges,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (h *Handler) getSchedule(w http.ResponseWriter, r *http.Request) {
	resourceID := chi.URLParam(r, "id")
	fromRaw := r.URL.Query().Get("from")
	toRaw := r.URL.Query().Get("to")
	from, err := time.Parse(time.RFC3339, fromRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	to, err := time.Parse(time.RFC3339, toRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	schedule, err := h.svc.GetResourceSchedule(r.Context(), resourceID, from, to)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, schedule)
}

type createBookingRequest struct {
	UserID string `json:"user_id"`
	SlotID string `json:"slot_id"`
}

func (h *Handler) createBooking(w http.ResponseWriter, r *http.Request) {
	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	booking, err := h.svc.CreateBooking(r.Context(), service.CreateBookingInput{
		UserID: req.UserID,
		SlotID: req.SlotID,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, booking)
}

func (h *Handler) cancelBooking(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	booking, err := h.svc.CancelBooking(r.Context(), bookingID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, booking)
}

func (h *Handler) getBooking(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	booking, err := h.svc.GetBooking(r.Context(), bookingID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, booking)
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, domain.ErrInvalidTimeRange):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, domain.ErrSlotAlreadyBooked):
		writeError(w, http.StatusConflict, err)
	case errors.Is(err, domain.ErrBookingAlreadyEnded):
		writeError(w, http.StatusConflict, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeError(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
