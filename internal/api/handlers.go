package api

import (
	"database/sql"
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

type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string	`json:"body"`
		UserID    uuid.UUID `json:"user_id"`
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

func (cfg *ApiConfig) CreateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type parameters struct {
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if params.UserID == uuid.Nil {
		RespondWithError(w, http.StatusBadRequest, "User ID is required")
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
	
	args := database.CreateChirpParams {
		Body: strings.Join(message, " "),
		UserID: params.UserID,
	}

	chirp, err := cfg.DB.CreateChirp(r.Context(), args)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create the chirp")
		return
	}

	resp := Chirp {
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}

	RespondWithJSON(w, http.StatusCreated, resp)
}


func (cfg *ApiConfig) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetChirps(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get chirps")
		return
	}

	resp := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		resp[i] = Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		}
	}

	RespondWithJSON(w, http.StatusOK, resp)
}

func (cfg *ApiConfig) GetChirpById(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}
	
	chirp, err := cfg.DB.GetChirpById(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}

		RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirp")
		return
	}

	resp := Chirp {
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}

	RespondWithJSON(w, http.StatusOK, resp)
}
