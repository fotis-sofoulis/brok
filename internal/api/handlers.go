package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fotis-sofoulis/brok/internal/database"
	"github.com/google/uuid"
)

type ApiConfig struct {
	FileServerHits atomic.Int32
	DB             *database.Queries
	Platform       string
}

type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string `json:"email"`
}

func (cfg *ApiConfig) DisplayMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`
		<html>
		  <body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		  </body>
		</html>
		`, cfg.FileServerHits.Load())
	w.Write([]byte(html))
}

func (cfg *ApiConfig) ResetUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		RespondWithError(w, http.StatusForbidden, "Not allowed in this environment")
		return
	}

	if err := cfg.DB.DropUsers(r.Context()); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed reseting the users table")
		return
	}

	RespondWithJSON(w, http.StatusOK, "Users table reset successfully")
}

func (cfg *ApiConfig) ValidateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type parameters struct {
		Body string `json:"body"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	message := strings.Split(params.Body, " ")

	for i, word := range message {
		if _, exists := badWords[strings.ToLower(word)]; exists {
			message[i] = "****"
		}
	}

	RespondWithJSON(w, http.StatusOK, CleanedResp{Cleaned: strings.Join(message, " ")})
}

func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), params.Email)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Failed to create user")
		return
	}

	resp := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	RespondWithJSON(w, http.StatusCreated, resp)
}
