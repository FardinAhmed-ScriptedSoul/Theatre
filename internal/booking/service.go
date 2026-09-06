package booking

import (
	"context"
	"errors"
)

type Service struct {
	store BookingStore
}

func NewService(store BookingStore) *Service {
	return &Service{store}
}

func (s *Service) Book(b Booking) error {
	_, err := s.store.Book(b)
	return err
}

func (s *Service) Hold(b Booking) (Booking, error) {
	return s.store.Book(b)
}

func (s *Service) ListBookings(movieID string) []Booking {
	return s.store.ListBookings(movieID)
}

type sessionStore interface {
	Confirm(ctx context.Context, sessionID string, userID string) (Booking, error)
	Release(ctx context.Context, sessionID string, userID string) error
}

func (s *Service) Confirm(ctx context.Context, sessionID string, userID string) (Booking, error) {
	store, ok := s.store.(sessionStore)
	if !ok {
		return Booking{}, errors.New("booking store does not support confirmation")
	}
	return store.Confirm(ctx, sessionID, userID)
}

func (s *Service) Release(ctx context.Context, sessionID string, userID string) error {
	store, ok := s.store.(sessionStore)
	if !ok {
		return errors.New("booking store does not support release")
	}
	return store.Release(ctx, sessionID, userID)
}
