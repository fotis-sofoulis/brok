package main

import (
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

	mux.HandleFunc("POST /api/validate_chirp", cfg.ValidateChirp)
	mux.HandleFunc("GET /admin/metrics", cfg.DisplayMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.ResetMetrics)


	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())


}
