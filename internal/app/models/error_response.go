package models

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents the structure of an error response.
type ErrorResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func HandleErrorResponse(rw http.ResponseWriter, response ErrorResponse) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(response.Status)
	json.NewEncoder(rw).Encode(response)
}
