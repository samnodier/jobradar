package httpx

import (
	"encoding/json"
	"log"
	"net/http"
)

// Heaper to format all the successful responses
func RespondJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to marshal JSON response: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func RespondError(w http.ResponseWriter, code int, msg string) {
	if code > 500 {
		log.Printf("server error (%d): %s", code, msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	RespondJSON(w, code, errorResponse{
		Error: msg,
	})
}
