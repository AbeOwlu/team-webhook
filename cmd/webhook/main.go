package main

import (
	"flag"
	"net/http"

	"example-webhook.com/internal/handlers"

	"go.uber.org/zap"
)

const (
// default message
)

func main() {
	loggInit := zap.NewExample()
	defer loggInit.Sync()
	logger := loggInit.Sugar()

	flag.StringVar(&handlers.TLScert, "tlscert", "/etc/certs/tls.crt", "Generated Webhook server cert")
	flag.StringVar(&handlers.TLSkey, "tlskey", "/etc/certs/tls.key", "Generated Webhook server key")
	flag.Parse()

	http.HandleFunc("/", handlers.HandleFunc)

	err := http.ListenAndServeTLS(":8443", handlers.TLScert, handlers.TLSkey, nil)
	if err != nil {
		logger.Fatal("Webhook error: %v",
			err)
	}

}
