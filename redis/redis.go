package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"paperboy-back"
	"time"

	"github.com/go-redis/redis/v8"
)

// Redis wraps a SummaryService and Redis to provide a cache.
type Redis struct {
	rdb *redis.Client
	ss  paperboy.SummaryService
}

// NewSummaryCache returns a new read-through cache for service.
func NewSummaryCache(addr, port, pass string, db int, ss paperboy.SummaryService) paperboy.SummaryService {
	// Initialize new redis client.
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", addr, port),
		Password: pass,
		DB:       db,
	})

	// Construct and return cache.
	return &Redis{rdb: rdb, ss: ss}
}

// Summary returns a pointer to a summary for a given objectID.
func (r *Redis) Summary(objectID string) (*paperboy.Summary, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
	defer cancel()

	// Checks the local cache before querying service.
	s, _ := r.rdb.Get(ctx, objectID).Result()
	if len(s) > 0 {
		var ret paperboy.Summary
		err := json.Unmarshal([]byte(s), &ret)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", "unable to unmarshal json", err)
		}
		return &ret, nil
	}

	// Otherwise, fetch from the underlying service.
	sum, err := r.ss.Summary(objectID)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", "unable to fetch from service", err)
	} else if sum != nil {
		json, err := json.Marshal(sum)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", "unable to unmarshal json", err)
		}

		err = r.rdb.Set(ctx, objectID, json, 1*time.Hour).Err()
	}
	return sum, nil
}

// Summaries returns a slice of the most recent summaries with a given sectionID such as
// 'world' or 'tech'. A limit must be set for the maximum number of documents fetched.
func (r *Redis) Summaries(sectionID, startID string, size int) ([]*paperboy.Summary, string, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 500*time.Millisecond)
	defer cancel()

	// Checks the local cache before querying service.
	sstr, err := r.rdb.Get(ctx, fmt.Sprintf("%s:%s:%v", sectionID, startID, size)).Result()
	if err == nil {
		var ret paperboy.SummariesResponse
		err := json.Unmarshal([]byte(sstr), &ret)
		if err != nil {
			return nil, "", fmt.Errorf("%q: %w", "unable to unmarshal json", err)
		}
		return ret.Summaries, ret.LastID, nil
	}

	// Otherwise, fetch from the underlying service.
	sum, last, err := r.ss.Summaries(sectionID, startID, size)
	if err != nil {
		return nil, "", fmt.Errorf("%q: %w", "unable to retrieve summaries", err)
	} else if sum != nil && len(startID) > 0 {
		json, err := json.Marshal(paperboy.SummariesResponse{LastID: last, Summaries: sum})
		if err != nil {
			return nil, "", fmt.Errorf("%q: %w", "unable to unmarshal json", err)
		}
		r.rdb.Set(ctx, fmt.Sprintf("%s:%s:%v", sectionID, startID, size), json, 1*time.Hour).Err()
	}
	return sum, last, nil
}

// Create inserts a summary into the database if possible, otherwise,
// it will update the existing entry.
func (r *Redis) Create(s *paperboy.Summary) error {
	return r.ss.Create(s)
}
