package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/fotis-sofoulis/brok/internal/database"
)

type ApiConfig struct {
	FileServerHits atomic.Int32
	DB			   *database.Queries
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

func (cfg *ApiConfig) ResetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	cfg.FileServerHits.Swap(0)
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

	badWords := map[string]struct{} {
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
