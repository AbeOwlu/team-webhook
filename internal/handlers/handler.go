package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	loggInit := zap.NewExample()
	defer loggInit.Sync()
	logger := loggInit.Sugar()

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
