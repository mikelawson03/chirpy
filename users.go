package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mikelawson03/chirpy/internal/auth"
	"github.com/mikelawson03/chirpy/internal/database"
)

type userInbound struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type userResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	// Get user data from http request
	user, err := decodeUser(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding JSON: ", err)
		return
	}

	// Check that email address and password exist
	if user.Email == "" || user.Password == "" {
		respondWithError(w, http.StatusUnauthorized, "Must provide both username and password", nil)
		return
	}

	// Get hashed password from database
	res, err := cfg.dbQueries.GetUserByEmail(context.Background(), user.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database Error: ", err)
		return
	}

	// check hashed password against entered password
	valid, err := auth.CheckPasswordHash(user.Password, res.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error validating login information: ", err)
		return
	}

	// respond with error if not valid
	if !valid {
		respondWithError(w, http.StatusUnauthorized, "Incorrect login information", nil)
		return
	}

	// create access token
	authTok, err := auth.MakeJWT(res.ID, cfg.secret, time.Duration(1)*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating access token: ", err)
		return
	}

	// create reference token and store in db
	p := database.CreateRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		UserID:    res.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	}
	refToken, err := cfg.dbQueries.CreateRefreshToken(context.Background(), p)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating refresh token: ", err)
		return
	}

	respondWithJSON(w, http.StatusOK, userResponse{
		ID:           res.ID,
		CreatedAt:    res.CreatedAt,
		UpdatedAt:    res.UpdatedAt,
		Email:        res.Email,
		Token:        authTok,
		RefreshToken: refToken.Token,
		IsChirpyRed:  res.IsChirpyRed,
	})
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refTok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error retrieving Bearer Token: ", err)
		return
	}

	// look up refresh token
	res, err := cfg.dbQueries.GetUserByRefToken(context.Background(), refTok)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error retrieving user: ", err)
		return
	}

	//check if token is revoked or expired
	if res.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Token Revoked: ", err)
		return
	}

	if res.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Token Expired: ", err)
		return
	}

	accTok, err := auth.MakeJWT(res.UserID, cfg.secret, time.Duration(1)*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error generating new access token: ", err)
		return
	}

	type token struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, token{
		Token: accTok,
	})

}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refTok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error retrieving Bearer Token: ", err)
		return
	}

	err = cfg.dbQueries.RevokeToken(context.Background(), refTok)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error revoking token: ", err)
	}

	respondWithJSON(w, http.StatusNoContent, "")
}

func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	user, err := decodeUser(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding JSON: ", err)
		return
	}

	if user.Email == "" || user.Password == "" {
		respondWithError(w, http.StatusUnauthorized, "Must provide both username and password", nil)
		return
	}

	hashedPW, err := auth.HashPassword(user.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password: ", err)
		return
	}

	res, err := cfg.dbQueries.CreateUser(context.Background(), database.CreateUserParams{
		Email:          user.Email,
		HashedPassword: hashedPW,
	})
	if err != nil {
		fmt.Println(err)
		respondWithError(w, http.StatusInternalServerError, "Error creating user: ", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, userResponse{
		ID:          res.ID,
		CreatedAt:   res.CreatedAt,
		UpdatedAt:   res.UpdatedAt,
		Email:       res.Email,
		IsChirpyRed: res.IsChirpyRed,
	})

}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get Access Token
	accToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error retrieving access token: ", err)
		return
	}

	// Validate Token and return user ID
	uid, err := auth.ValidateJWT(accToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error validating token: ", err)
		return
	}

	// Decode user request into userInbound struct
	userInfo, err := decodeUser(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error decoding user: ", err)
		return
	}

	// Hash provided PW
	hashPW, err := auth.HashPassword(userInfo.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password: ", err)
		return
	}

	p := database.UpdateUserInfoParams{
		Email:          userInfo.Email,
		HashedPassword: hashPW,
		ID:             uid,
	}

	res, err := cfg.dbQueries.UpdateUserInfo(context.Background(), p)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating user information: ", err)
		return
	}

	respondWithJSON(w, http.StatusOK, userResponse{
		ID:        res.ID,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
		Email:     res.Email,
	})
}

func (cfg *apiConfig) handlerResetUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Reset only allowed in dev environment", nil)
	}
	err := cfg.dbQueries.ResetUsers(context.Background())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database Error: ", err)
		return
	}

	respondWithJSON(w, http.StatusOK, nil)
}

func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	type polkaEvent struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetApiKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing API key: ", err)
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Incorrect API key: %v, Expected: %v", apiKey, cfg.polkaKey), nil)
		return
	}

	decoder := json.NewDecoder(r.Body)
	event := polkaEvent{}
	err = decoder.Decode(&event)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding event: ", err)
		return
	}

	if event.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	uid, err := uuid.Parse(event.Data.UserId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing user ID: ", err)
		return
	}

	_, err = cfg.dbQueries.UpgradeUser(context.Background(), uid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found: ", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

}

func decodeUser(r *http.Request) (*userInbound, error) {
	decoder := json.NewDecoder(r.Body)
	user := userInbound{}
	err := decoder.Decode(&user)
	if err != nil {
		return &userInbound{}, err
	}

	return &user, err
}
