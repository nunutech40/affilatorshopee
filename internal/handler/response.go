package handler

import (
	"encoding/json"
	"net/http"
)

type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Success: true, Data: data, Error: nil})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeErrorFields(w, status, code, message, nil)
}

func writeErrorFields(w http.ResponseWriter, status int, code, message string, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{
		Success: false,
		Data:    nil,
		Error:   ErrorDetail{Code: code, Message: message, Fields: fields},
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target interface{}) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
