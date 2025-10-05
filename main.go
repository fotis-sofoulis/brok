package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/fotis-sofoulis/brok/internal/api"
)

const (
	filepathRoot = "."
	port = "8080"
)

func main() {
	mux := http.NewServeMux()
	cfg := &api.ApiConfig{}
	appHandler :=  http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", cfg.MiddlewareMetricIncrease(appHandler))

	server := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}
	
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()

		type parameters struct {
			Body string `json:"body"`
		}

		type errorResp struct {
			Error string `json:"error"`
		}

		type validResp struct {
			Valid bool `json:"valid"`
		}
		
		w.Header().Set("Content-Type", "application/json")

		params := parameters{}
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			msg, mErr := json.Marshal(errorResp{Error: "Something went wrong"})
			if mErr != nil {
				w.Header().Del("Content-Type")
				io.WriteString(w, "Something went really wrong")
				return
			}
			w.Write(msg)
			return
		}

		if len(params.Body) > 140 {
			w.WriteHeader(http.StatusBadRequest)
			msg, mErr := json.Marshal(errorResp{Error: "Chirp is too long"})
			if mErr != nil {
				w.Header().Del("Content-Type")
				io.WriteString(w, "Something went really wrong")
				return
			}
			w.Write(msg)
			return
		}

		resp, err := json.Marshal(validResp{Valid: true})
		if err != nil {
			w.Header().Del("Content-Type")
			io.WriteString(w, "Something went really wrong")
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	})

	mux.HandleFunc("GET /admin/metrics", cfg.DisplayMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.ResetMetrics)


	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())


}
