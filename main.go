package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/7minutech/chripy/internal/auth"
	"github.com/7minutech/chripy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const expirationDays = 60

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	secret         string
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type JWT struct {
	Token string `json:"token"`
}

func (apiCfg *apiConfig) handlerMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	defer r.Body.Close()

	body := fmt.Sprintf(
		"<html><body>"+
			"<h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p>"+
			"</body></html>", apiCfg.fileserverHits.Load())
	w.Write([]byte(body))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func (apiCfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	if apiCfg.platform != "dev" {
		msg := "must be in dev platform"
		respondWithError(w, http.StatusForbidden, msg, fmt.Errorf("error: trying reset while not platform is not dev"))
		return
	}

	if err := apiCfg.dbQueries.DeleteUsers(r.Context()); err != nil {
		msg := "could not delete users"
		respondWithError(w, http.StatusInternalServerError, msg, err)
		return
	}
}

func (apiCfg *apiConfig) handerUser(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	var params = parameters{}

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		msg := "could not decode request body"
		respondWithError(w, http.StatusBadRequest, msg, err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		msg := "could not hash password"
		respondWithError(w, http.StatusBadRequest, msg, err)
		return
	}

	userParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	user, err := apiCfg.dbQueries.CreateUser(r.Context(), userParams)

	if err != nil {
		msg := "could not create user"
		respondWithError(w, http.StatusInternalServerError, msg, err)
		return
	}

	resp := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusCreated, resp)
}

func (apiCfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	var params = parameters{}

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		msg := "could not decode request body"
		respondWithError(w, http.StatusBadRequest, msg, err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		msg := "could not hash password"
		respondWithError(w, http.StatusBadRequest, msg, err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		msg := "could not get token"
		respondWithError(w, http.StatusUnauthorized, msg, err)
		return
	}

	userID, err := auth.ValidateJWT(token, apiCfg.secret)
	if err != nil {
		msg := "could not validate token"
		respondWithError(w, http.StatusUnauthorized, msg, err)
		return
	}

	userParams := database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	}

	user, err := apiCfg.dbQueries.UpdateUser(r.Context(), userParams)
	if err != nil {
		msg := "could not update user"
		respondWithError(w, http.StatusInternalServerError, msg, err)
		return
	}

	resp := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (apiCfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {

	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		msg := "token was not given in headers"
		respondWithError(w, http.StatusBadRequest, msg, err)
		return
	}

	userID, err := auth.ValidateJWT(tok, apiCfg.secret)
	if err != nil {
		msg := "token was not valid"
		respondWithError(w, http.StatusUnauthorized, msg, err)
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	const maxChirpLength int = 140

	var params parameters

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleanedBody := cleanProfanity(params.Body)

	chirpyParams := database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	}

	chirp, err := apiCfg.dbQueries.CreateChirp(r.Context(), chirpyParams)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create chirp", err)
		return
	}

	resp := convertChirp(chirp)

	respondWithJSON(w, http.StatusCreated, resp)
}

func (apiCfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {

	dbChirps, err := apiCfg.dbQueries.GetChirps(r.Context())

	if err != nil {
		msg := "could not select all chirps"
		respondWithError(w, http.StatusInternalServerError, msg, err)
	}

	chirps := mapChirp(dbChirps)

	respondWithJSON(w, http.StatusOK, chirps)
}

func (apiCfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {

	chirpIDStr := r.PathValue("chirpID")

	if chirpIDStr == "" {
		msg := "chirp id was not found in url path"
		respondWithError(w, http.StatusBadRequest, msg, nil)
	}

	chirpID, err := uuid.Parse(chirpIDStr)

	if err != nil {
		msg := "could not parse chirpID from string to uuid"
		respondWithError(w, http.StatusBadRequest, msg, err)
		return
	}

	dbChrip, err := apiCfg.dbQueries.GetChirp(r.Context(), chirpID)

	if errors.Is(err, sql.ErrNoRows) {
		msg := "chirp does not exist"
		respondWithError(w, http.StatusNotFound, msg, err)
		return
	}

	if err != nil {
		msg := "db error getting chrip"
		respondWithError(w, http.StatusInternalServerError, msg, err)
		return
	}

	chirp := convertChirp(dbChrip)

	respondWithJSON(w, http.StatusOK, chirp)
}

func (apiCfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	var params = parameters{}

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		msg := "could not decode request body"
		respondWithError(w, http.StatusBadRequest, msg, err)
		return
	}

	user, err := apiCfg.dbQueries.GetUserByEmail(r.Context(), params.Email)

	if errors.Is(err, sql.ErrNoRows) {
		msg := "Incorrect email or password"
		respondWithError(w, http.StatusUnauthorized, msg, err)
		return
	}

	if err != nil {
		msg := "could not get user"
		respondWithError(w, http.StatusInternalServerError, msg, err)
		return
	}

	ok, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if !ok || err != nil {
		msg := "Incorrect email or password"
		respondWithError(w, http.StatusUnauthorized, msg, err)
		return
	}

	tok, err := auth.MakeJWT(user.ID, apiCfg.secret, time.Hour)
	if err != nil {
		msg := "could not create JWT"
		respondWithError(w, http.StatusInternalServerError, msg, err)
	}

	refreshTok, _ := auth.MakeRefreshToken()

	expirationDate := time.Now().AddDate(0, 0, expirationDays)

	refreshTokenParams := database.CreateRefreshTokenParams{Token: refreshTok, UserID: user.ID, ExpiresAt: expirationDate, RevokedAt: sql.NullTime{}}

	dbRefreshToken, err := apiCfg.dbQueries.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		msg := "Could not create refresh token"
		respondWithError(w, http.StatusInternalServerError, msg, err)
	}

	resp := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        tok,
		RefreshToken: dbRefreshToken.Token,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (apiCfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {

	refreshTok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		msg := "Couldn't get refresh token from header"
		respondWithError(w, http.StatusBadRequest, msg, err)
		return
	}

	dbRefreshTok, err := apiCfg.dbQueries.GetRefreshToken(r.Context(), refreshTok)

	if errors.Is(err, sql.ErrNoRows) {
		msg := "refresh token does not exist"
		respondWithError(w, http.StatusUnauthorized, msg, err)
		return
	}

	if err != nil {
		msg := "Couldn't get refresh token"
		respondWithError(w, http.StatusInternalServerError, msg, err)
		return
	}

	if dbRefreshTok.RevokedAt.Valid {
		msg := "refresh token is expired"
		respondWithError(w, http.StatusUnauthorized, msg, err)
		return
	}

	user, err := apiCfg.dbQueries.GetUserByRefreshToken(r.Context(), dbRefreshTok.Token)
	if err != nil {
		msg := "couldn't get user by refresh token"
		respondWithError(w, http.StatusInternalServerError, msg, err)
		return
	}

	tok, err := auth.MakeJWT(user.ID, apiCfg.secret, time.Hour)
	if err != nil {
		msg := "could not create JWT"
		respondWithError(w, http.StatusInternalServerError, msg, err)
		return
	}

	resp := JWT{Token: tok}

	respondWithJSON(w, http.StatusOK, resp)

}

func (apiCfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {

	refreshTok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		msg := "Couldn't get refresh token from header"
		respondWithError(w, http.StatusBadRequest, msg, err)
		return
	}

	dbRefreshTok, err := apiCfg.dbQueries.GetRefreshToken(r.Context(), refreshTok)

	if errors.Is(err, sql.ErrNoRows) {
		msg := "refresh token does not exist"
		respondWithError(w, http.StatusUnauthorized, msg, err)
		return
	}

	if err != nil {
		msg := "Couldn't get refresh token"
		respondWithError(w, http.StatusInternalServerError, msg, err)
		return
	}

	if dbRefreshTok.RevokedAt.Valid {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	err = apiCfg.dbQueries.UpdateRefreshToken(r.Context(), dbRefreshTok.Token)
	if err != nil {
		msg := "Couldn't update refresh token"
		respondWithError(w, http.StatusInternalServerError, msg, err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

}

func main() {
	godotenv.Load(".env")

	platform := os.Getenv("PLATFORM")
	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("Secret")
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal("failed to open data base")
	}

	queries := database.New(db)

	const filepathRoot = "."
	const port = "8080"

	var apiCfg = apiConfig{
		dbQueries: queries,
		platform:  platform,
		secret:    secret,
	}

	mux := http.NewServeMux()

	handlerFile := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handlerFile))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetric)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerValidateChirp)
	mux.HandleFunc("POST /api/users", apiCfg.handerUser)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
