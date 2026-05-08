package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mikelawson03/chirpy/internal/auth"
	"github.com/mikelawson03/chirpy/internal/database"
)

type chirpRequest struct {
	Content string    `json:"body"`
	UserId  uuid.UUID `json:"user_id"`
}

type chirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {

	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Bearer Token not retrieved successfully: ", err)
		return
	}

	id, err := auth.ValidateJWT(tok, cfg.secret)
	if err != nil || id == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or incorrect bearer token: ", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req chirpRequest
	err = decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	cleaned, err := ValidateChirp(req.Content)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error validating chirp: ", err)
	}

	p := database.CreateChirpParams{
		Body:   cleaned,
		UserID: id,
	}

	chirp, err := cfg.dbQueries.CreateChirp(context.Background(), p)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database Error: ", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, chirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	var resp []database.Chirp
	var err error
	author_id := r.URL.Query().Get("author_id")
	if author_id != "" {
		uid, err := uuid.Parse(author_id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Invalid user id: ", err)
			return
		}
		resp, err = cfg.dbQueries.GetChirpsByAuthor(context.Background(), uid)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "User ID not found: ", err)
			return
		}
	} else {
		resp, err = cfg.dbQueries.GetChirps(context.Background())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Database Error: ", err)
			return
		}
	}

	chirps := []chirpResponse{}

	for _, chirp := range resp {
		chirps = append(chirps, chirpResponse{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		})
	}

	sortOrder := r.URL.Query().Get("sort")
	if sortOrder == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
	}

	respondWithJSON(w, http.StatusOK, chirps)

}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	// get access token
	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Bearer Token not retrieved successfully: ", err)
		return
	}

	// validate token
	uid, err := auth.ValidateJWT(tok, cfg.secret)
	if err != nil || uid == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or incorrect bearer token: ", err)
		return
	}

	// verify account id for authorization
	idStr := r.PathValue("chirpID")
	chirpid, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing request: ", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpById(context.Background(), chirpid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found: ", err)
	}

	if uid != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "Chirp may only be deleted by the chirp's author", nil)
		return
	}

	// Delete chirp
	err = cfg.dbQueries.DeleteChirp(context.Background(), chirpid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting chirp: ", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (cfg *apiConfig) handlerGetChirpById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("chirpID")

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing request:  ", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpById(context.Background(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Database Error: ", err)
		return
	}

	respondWithJSON(w, http.StatusOK, chirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	})

}

func ValidateChirp(body string) (string, error) {
	if len(body) > 140 {
		return "", errors.New("Chirp too long")
	}

	cleaned := cleanChirp(body)
	return cleaned, nil

}

func cleanChirp(s string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	split := strings.Split(s, " ")

	for i, word := range split {

		if _, ok := badWords[strings.ToLower(word)]; ok {
			split[i] = "****"
		}
	}

	return strings.Join(split, " ")
}
