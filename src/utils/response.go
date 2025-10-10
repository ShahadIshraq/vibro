package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"vibro/src/models"
)

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		log.Printf("Error: %s - %v", message, err)
	}

	response := models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}

	RespondJSON(w, status, response)
}
