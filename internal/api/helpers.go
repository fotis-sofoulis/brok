package api

import (
	"encoding/json"
	"net/http"
)

type ErrorResp struct {
	Error string `json:"error"`
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, err := json.Marshal(ErrorResp{Error: msg})
	if err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
