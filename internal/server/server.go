package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/zaynereo/admission-policy-webhook/internal/admission"
)

// Config holds HTTP server settings.
type Config struct {
	Addr          string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
	MaxHeaderBytes int
}

// DefaultConfig returns conservative defaults for a validating webhook.
func DefaultConfig() Config {
	return Config{
		Addr:           ":8443",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

// NewMux builds routes for health and validation.
func NewMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/readyz", handleHealthz)
	mux.HandleFunc("/validate", handleValidate)
	return mux
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if ct := r.Header.Get("Content-Type"); ct != "" && ct != "application/json" {
		log.Printf("validate: unexpected content-type %q", ct)
	}

	body, err := admission.ReadBoundedBody(r.Body)
	if err != nil {
		out, code, rerr := admission.MalformedResponse(err.Error())
		if rerr != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_, _ = w.Write(out)
		return
	}

	out, code, err := admission.Review(body)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(out)
}

// ListenAndServeTLS serves HTTPS with the provided certificate.
func ListenAndServeTLS(cfg Config, certFile, keyFile string) error {
	srv := &http.Server{
		Addr:           cfg.Addr,
		Handler:        NewMux(),
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		IdleTimeout:    cfg.IdleTimeout,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}
	return srv.ListenAndServeTLS(certFile, keyFile)
}
