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
)

func main() {
	loggInit := zap.NewExample()
	defer loggInit.Sync()
	logger := loggInit.Sugar()

	flag.StringVar(&TLScert, "tlscert", "/etc/certs/tls.crt", "Generated Webhook server cert")
	flag.StringVar(&TLSkey, "tlskey", "/etc/certs/tls.key", "Generated Webhook server key")
	flag.Parse()

	http.HandleFunc("/", handlers.HandleFunc)

	err := http.ListenAndServeTLS(port, TLScert, TLSkey, nil)
	if err != nil {
		logger.Fatal("Webhook error: %v",
			err)
	}

}
