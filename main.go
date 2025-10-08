package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/fotis-sofoulis/brok/internal/api"
	"github.com/fotis-sofoulis/brok/internal/database"
)

const (
	filepathRoot = "."
	port         = "8080"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return
	}
	dbQueries := database.New(db)
	platform := os.Getenv("PLATFORM")

	mux := http.NewServeMux()
	cfg := &api.ApiConfig{
		DB:       dbQueries,
		Platform: platform,
	}
	appHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", cfg.MiddlewareMetricIncrease(appHandler))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /api/chirps", cfg.CreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.GetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.GetChirpById)
	mux.HandleFunc("POST /api/users", cfg.CreateUser)
	mux.HandleFunc("GET /admin/metrics", cfg.DisplayMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.ResetUsers)

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())

}
