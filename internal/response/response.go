package response

import (
	"encoding/json"
	"net/http"
)

func JsonResponse(w http.ResponseWriter, status int, data any) {
	b, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	b, err := json.Marshal(map[string]any{"status": status, "message": message})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}
