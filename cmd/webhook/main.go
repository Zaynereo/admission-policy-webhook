package main

import (
	"flag"
	"log"
	"os"

	"github.com/zaynereo/admission-policy-webhook/internal/server"
)

func main() {
	addr := envOr("WEBHOOK_LISTEN_ADDR", ":8443")
	cert := envOr("WEBHOOK_TLS_CERT", "certs/tls.crt")
	key := envOr("WEBHOOK_TLS_KEY", "certs/tls.key")

	flag.StringVar(&addr, "addr", addr, "HTTPS listen address")
	flag.StringVar(&cert, "tls-cert", cert, "TLS certificate file")
	flag.StringVar(&key, "tls-key", key, "TLS private key file")
	flag.Parse()

	cfg := server.DefaultConfig()
	cfg.Addr = addr

	log.Printf("admission webhook listening on %s (TLS)", cfg.Addr)
	if err := server.ListenAndServeTLS(cfg, cert, key); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
