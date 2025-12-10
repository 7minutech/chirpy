package main

import "github.com/7minutech/chirpy/internal/database"

func mapChirp(dbChirps []database.Chirp) []Chirp {
	var chirps = make([]Chirp, len(dbChirps))
	for i, dbChirp := range dbChirps {
		chirp := convertChirp(dbChirp)
		chirps[i] = chirp
	}
	return chirps
}

func convertChirp(dbChirp database.Chirp) Chirp {
	return Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
}
