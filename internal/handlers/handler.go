package handlers

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"

	"go.uber.org/zap"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var loggInit = zap.NewExample()
var logger = loggInit.Sugar()

const (
	// const invalid err meassage
	InvalidMessage = "namespace missing required team label"
	requiredLabel  = "team"
)

var (
	DenyStatus metav1.Status = metav1.Status{
		Message: InvalidMessage,
		Status:  metav1.StatusFailure,
		Code:    400,
	}
)

type Name struct {
	Name string `json:"name"`
}

type Namespace struct {
	Metadata Metadata `jsons:"metadata"`
}

type Metadata struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

func (m *Metadata) isEmpty() bool {
	return m.Name == ""
}

// function to validate resource creation based on authorized content
func HandleFunc(w http.ResponseWriter, r *http.Request) {

	defer loggInit.Sync()

	arReview := v1beta1.AdmissionReview{}

	if err := json.NewDecoder(r.Body).Decode(&arReview); err != nil {
		logger.Error("client messge: %v, err response: %v", r.Body, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if arReview.Request == nil {
		logger.Error("invlid request body: client messge: %v", r.Body)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// unmarshalled Admission Request Payload
	// https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#webhook-request-and-response
	raw := arReview.Request.Object.Raw
	ns := Namespace{}

	if err := json.Unmarshal(raw, &ns); err != nil {
		logger.Error("client messge: %v, err response: %v", r.Body, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	} else if ns.Metadata.isEmpty() {
		logger.Error("invlid request body: client messge: %v", r.Body)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	arReview.Response = &v1beta1.AdmissionResponse{
		UID:     arReview.Request.UID,
		Allowed: true,
	}

	admitLabel, ok := ns.Metadata.Labels[requiredLabel]
	if len(ns.Metadata.Labels) == 0 || !ok || admitLabel == "" {
		arReview.Response.Allowed = false
		arReview.Response.Result = &DenyStatus
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&arReview)

}

// kubelet health and readiness probez

func Healthz(w http.ResponseWriter, r *http.Request) {
	healthz_path := r.RequestURI
	if healthz_path == "/healthz" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		logger.Infof("ns-Validator healthz : OK")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Internal Error: Error Path: health probe paths is healthz"))
	}

}

func Readyz(w http.ResponseWriter, r *http.Request) {
	readyz_path := r.RequestURI
	var con_check *int

	if readyz_path == "/readyz" {
		kas_endpoint := os.Getenv("KUBERNETES_PORT_443GO c_TCP_ADDR")
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{},
			},
		}
		// add "https://"+kas_endpoint to streamline check
		req, err := http.NewRequest("GET", kas_endpoint, nil)
		if err != nil {
			logger.Error("Readyz error: failed building Kubernete API Server endpoint: %s : err: %v", kas_endpoint, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Readyz error: error connecting to Kubernetes API Server endpoint: err: %v", err)
		}

		con_check = &resp.StatusCode

		if *con_check == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Internal Error: Error Path: health probe paths is healthz"))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		}

	}

}
