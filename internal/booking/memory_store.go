package booking

type MemoryStore struct {
	bookings map[string]Booking
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		bookings:map[string]Booking{},
	}
}
//book a specific movie
func (s *MemoryStore)	Book(b Booking) (Booking, error){
	//look if the seat is taken if not then populate the map
	if _, exists := s.bookings[b.SeatID]; exists {
		return b, ErrSeatAlreadyBooked
	}
	s.bookings[b.SeatID] = b
	return b, nil
}
//find booking for a specific movie
func (s *MemoryStore)	ListBookings(movieID string) []Booking{
	var result []Booking
	for _,b := range s.bookings{
		if b.MovieID == movieID{
			result = append(result,b)
		}
	}
	return result
}

