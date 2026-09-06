package booking

import "sync"

type ConcurrentStore struct {
	mu       sync.RWMutex
	bookings map[string]Booking
}

func NewConcurrentStore() *ConcurrentStore {
	return &ConcurrentStore{
		bookings: map[string]Booking{},
	}
}

//book a specific movie
func (s *ConcurrentStore) Book(b Booking) (Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	//look if the seat is taken if not then populate the map
	if _, exists := s.bookings[b.SeatID]; exists {
		return b, ErrSeatAlreadyBooked
	}
	s.bookings[b.SeatID] = b
	return b, nil
}

//find booking for a specific movie
func (s *ConcurrentStore) ListBookings(movieID string) []Booking {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Booking
	for _, b := range s.bookings {
		if b.MovieID == movieID {
			result = append(result, b)
		}
	}
	return result
}
