package main

import (
	"flag"
	"net/http"

	"github.com/AbeOwlu/team-webhook/internal/handlers"

	"go.uber.org/zap"
)

var (
	TLScert, TLSkey string
	port            string = ":8443"
	probezPort      string = ":8080"
)

func main() {
	loggInit := zap.NewExample()
	defer loggInit.Sync()
	logger := loggInit.Sugar()

	flag.StringVar(&TLScert, "tlscert", "/etc/certs/tls.crt", "Generated Webhook server cert")
	flag.StringVar(&TLSkey, "tlskey", "/etc/certs/tls.key", "Generated Webhook server key")
	flag.Parse()

	http.HandleFunc("/healthz", handlers.Healthz)
	http.HandleFunc("/readyz", handlers.Readyz)
	http.HandleFunc("/", handlers.HandleFunc)

	go func() {
		err := http.ListenAndServe(probezPort, nil)
		if err != nil {
			logger.Fatal("Webhook Probe error: %v",
				err)
		}
	}()

	err := http.ListenAndServeTLS(port, TLScert, TLSkey, nil)
	if err != nil {
		logger.Fatal("Webhook error: %v",
			err)
	}

}
