package api

import (
	"encoding/json"
	"net/http"
)

type ErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": code, "error": ErrorBody{code, message}})
}
func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, 405, "method not allowed")
}
func badRequest(w http.ResponseWriter, msg string) { writeError(w, http.StatusBadRequest, 400, msg) }
func notFound(w http.ResponseWriter)               { writeError(w, http.StatusNotFound, 404, "resource not found") }
