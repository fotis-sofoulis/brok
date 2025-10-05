package api

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type ApiConfig struct {
	FileServerHits atomic.Int32
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
