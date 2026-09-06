package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const defaultHoldTTL = 7 * time.Minute

// RedisStore implements session-based seat booking backed by Redis.
//
// Key design:
//
//	seat:{movieID}:{seatID}   → session JSON (TTL = held, no TTL = confirmed)
//	session:{sessionID}       → seat key     (reverse lookup)
type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

func sessionKey(id string) string {
	return fmt.Sprintf("session:%s", id)
}

func (s *RedisStore) Book(b Booking) (Booking, error) {
	session, err := s.hold(b)
	if err != nil {
		return Booking{}, err
	}

	return session, nil
}

func (s *RedisStore) ListBookings(movieID string) []Booking {
	pattern := fmt.Sprintf("seat:%s:*", movieID)
	var sessions []Booking

	ctx := context.Background()

	iter := s.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		val, err := s.rdb.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}
		session, err := parseSession(val)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions
}

func (s *RedisStore) hold(b Booking) (Booking, error) {
	id := uuid.New().String()
	ctx := context.Background()
	key := fmt.Sprintf("seat:%s:%s", b.MovieID, b.SeatID)
	b.ID = id
	b.Status = "held"
	val, err := json.Marshal(b)
	if err != nil {
		return Booking{}, err
	}
	res := s.rdb.SetArgs(ctx, key, val, redis.SetArgs{
		Mode: "NX",
		TTL:  defaultHoldTTL,
	})
	if err := res.Err(); err != nil {
		if err == redis.Nil {
			return Booking{}, ErrSeatAlreadyBooked
		}
		return Booking{}, err
	}

	if err := s.rdb.Set(ctx, sessionKey(id), key, defaultHoldTTL).Err(); err != nil {
		s.rdb.Del(ctx, key)
		return Booking{}, err
	}

	return Booking{
		ID:      id,
		MovieID: b.MovieID,
		SeatID:  b.SeatID,
		UserID:  b.UserID,
		Status:  "held",
	}, nil
}

func parseSession(val string) (Booking, error) {
	var data Booking
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return Booking{}, err
	}
	return Booking{
		ID:      data.ID,
		MovieID: data.MovieID,
		SeatID:  data.SeatID,
		UserID:  data.UserID,
		Status:  data.Status,
	}, nil
}

// Confirm converts a held session into a permanent booking.
// Removes the TTL (PERSIST) so the key never expires.
func (s *RedisStore) Confirm(ctx context.Context, sessionID string, userID string) (Booking, error) {
	session, sk, err := s.getSession(ctx, sessionID, userID)
	if err != nil {
		return Booking{}, err
	}

	session.Status = "confirmed"
	val, err := json.Marshal(session)
	if err != nil {
		return Booking{}, err
	}
	if err := s.rdb.Set(ctx, sk, val, 0).Err(); err != nil {
		return Booking{}, err
	}
	if err := s.rdb.Persist(ctx, sessionKey(sessionID)).Err(); err != nil {
		return Booking{}, err
	}
	return session, nil
}

func (s *RedisStore) getSession(ctx context.Context, sessionID string, userID string) (Booking, string, error) {
	sk, err := s.rdb.Get(ctx, sessionKey(sessionID)).Result()
	if err != nil {
		return Booking{}, "", err
	}

	val, err := s.rdb.Get(ctx, sk).Result()
	if err != nil {
		return Booking{}, "", err
	}

	session, err := parseSession(val)
	if err != nil {
		return Booking{}, "", err
	}

	if session.UserID != userID {
		return Booking{}, "", fmt.Errorf("session does not belong to user")
	}

	return session, sk, nil
}

func (s *RedisStore) Release(ctx context.Context, sessionID string, userID string) error {
	_, sk, err := s.getSession(ctx, sessionID, userID)
	if err != nil {
		return err
	}

	s.rdb.Del(ctx, sk, sessionKey(sessionID))
	return nil
}
