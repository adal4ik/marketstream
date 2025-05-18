package utils

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type APIError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	Resource string `json:"resource"`
}

func (e *APIError) Send(w http.ResponseWriter) {
	j, err := json.MarshalIndent(e, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	w.Write(j)
}

func (r *APIResponse) Send(w http.ResponseWriter) {
	j, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)
	w.Write(j)
}
