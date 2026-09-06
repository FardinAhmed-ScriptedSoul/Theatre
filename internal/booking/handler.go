package booking

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListSeats(w http.ResponseWriter, r *http.Request) {
	bookings := h.service.ListBookings(r.PathValue("movieID"))
	seats := make([]seatResponse, 0, len(bookings))
	for _, booking := range bookings {
		seats = append(seats, seatResponse{
			SeatID:    booking.SeatID,
			UserID:    booking.UserID,
			Booked:    true,
			Confirmed: booking.Status == "confirmed",
		})
	}
	writeJSON(w, http.StatusOK, seats)
}

func (h *Handler) HoldSeat(w http.ResponseWriter, r *http.Request) {
	current := Booking{
		MovieID: r.PathValue("movieID"),
		SeatID:  r.PathValue("seatID"),
		UserID:  r.Header.Get("X-User-ID"),
	}
	if current.UserID == "" {
		writeError(w, http.StatusBadRequest, errors.New("X-User-ID header is required"))
		return
	}

	result, err := h.service.Hold(current)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrSeatAlreadyBooked) {
			status = http.StatusConflict
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusCreated, holdResponse{
		SessionID: result.ID,
		MovieID:   result.MovieID,
		SeatID:    result.SeatID,
		ExpiresAt: time.Now().Add(defaultHoldTTL),
	})
}

func (h *Handler) ConfirmSession(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Confirm(r.Context(), r.PathValue("sessionID"), r.Header.Get("X-User-ID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) ReleaseSession(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Release(r.Context(), r.PathValue("sessionID"), r.Header.Get("X-User-ID")); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

type seatResponse struct {
	SeatID    string `json:"seat_id"`
	UserID    string `json:"user_id"`
	Booked    bool   `json:"booked"`
	Confirmed bool   `json:"confirmed"`
}

type holdResponse struct {
	SessionID string    `json:"session_id"`
	MovieID   string    `json:"movie_id"`
	SeatID    string    `json:"seat_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
