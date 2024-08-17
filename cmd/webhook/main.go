package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/AbeOwlu/team-webhook/internal/handlers"

	"go.uber.org/zap"
)

var (
	TLScert, TLSkey string
	port            string = ":8443"
)

func main() {
	loggInit := zap.NewExample()
	defer loggInit.Sync()
	logger := loggInit.Sugar()

	flag.StringVar(&TLScert, "tlscert", "/etc/certs/tls.crt", "Generated Webhook server cert")
	flag.StringVar(&TLSkey, "tlskey", "/etc/certs/tls.key", "Generated Webhook server key")
	flag.Parse()

	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/readyz", readyz)
	http.HandleFunc("/", handlers.HandleFunc)

	err := http.ListenAndServeTLS(port, TLScert, TLSkey, nil)
	if err != nil {
		logger.Fatal("Webhook error: %v",
			err)
	}

}

func healthz(w http.ResponseWriter, r *http.Request) {
	healthz_path := r.RequestURI
	if healthz_path == "/healthz" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Internal Error: Error Path: health probe paths is healthz"))
	}

}

func readyz(w http.ResponseWriter, r *http.Request) {
	readyz_path := r.RequestURI
	var con_check *int

	if readyz_path == "/readyz" {
		kas_endpoint := os.Getenv("KUBERNETES_PORT_$$#_TCP_ADDR")
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{},
			},
		}
		// add "https://"+kas_endpoint to streamline check
		req, err := http.NewRequest("GET", kas_endpoint, nil)
		if err != nil {
			log.Printf("Readyz error: failed building Kubernete API Server endpoint: %s : err: %v", kas_endpoint, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Readyz error: error connecting to Kubernetes API Server endpoint: err: %v", err)
		}

		con_check = &resp.StatusCode

		if &con_check != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Internal Error: Error Path: health probe paths is healthz"))
		}

	}

}
