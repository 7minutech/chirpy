package main

import (
	"testing"
	"time"

	"github.com/7minutech/chirpy/internal/database"
	"github.com/google/uuid"
)

func TestMapChirp(t *testing.T) {
	var dbChripTime1 = time.Date(2025, time.April, 10, 20, 0, 0, 0, time.UTC)
	var dbChripTime2 = time.Date(2025, time.November, 10, 20, 0, 0, 0, time.UTC)
	var dbChripId1 = uuid.New()
	var dbChirpId2 = uuid.New()
	var dbChripUserId1 = uuid.New()
	var dbChirpUserId2 = uuid.New()
	cases := []struct {
		input    []database.Chirp
		expected []Chirp
	}{
		{
			input: []database.Chirp{
				{
					ID:        dbChripId1,
					CreatedAt: dbChripTime1,
					UpdatedAt: dbChripTime1,
					Body:      "hello there",
					UserID:    dbChripUserId1,
				},
				{
					ID:        dbChirpId2,
					CreatedAt: dbChripTime2,
					UpdatedAt: dbChripTime2,
					Body:      "bye there",
					UserID:    dbChirpUserId2,
				},
			},
			expected: []Chirp{
				{
					ID:        dbChripId1,
					CreatedAt: dbChripTime1,
					UpdatedAt: dbChripTime1,
					Body:      "hello there",
					UserID:    dbChripUserId1,
				},
				{
					ID:        dbChirpId2,
					CreatedAt: dbChripTime2,
					UpdatedAt: dbChripTime2,
					Body:      "bye there",
					UserID:    dbChirpUserId2,
				},
			},
		},
	}

	for _, c := range cases {
		actual := mapChirp(c.input)

		if len(c.expected) != len(actual) {
			t.Errorf("len(actual) == %d, expected: %d", len(c.expected), len(actual))
		}

		for i := range c.expected {
			if c.expected[i] != actual[i] {
				t.Errorf("actual[%d] == %v, expected: %v", i, actual[i], c.expected[i])
			}
		}
	}
}
