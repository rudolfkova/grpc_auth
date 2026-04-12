// Package handler ...
package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// writeProblem — стабильные code + message (для публичных API вроде /api/me/characters).
func writeProblem(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"code": code, "message": message})
}

func decodeJSON(r *http.Request, v any) error {
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return json.NewDecoder(r.Body).Decode(v)
}
