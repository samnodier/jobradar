package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) handlerHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "online",
		"message": "JobRadar API is pulsing",
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}
