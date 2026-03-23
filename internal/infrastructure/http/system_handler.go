package http

import (
	"encoding/json"
	stdhttp "net/http"
)

type SystemHandler struct{}

func NewSystemHandler() SystemHandler {
	return SystemHandler{}
}

func (h SystemHandler) Info(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	response := map[string]string{
		"status": "ok",
	}

	writeJSON(w, stdhttp.StatusOK, response)
}

func writeJSON(w stdhttp.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		stdhttp.Error(w, stdhttp.StatusText(stdhttp.StatusInternalServerError), stdhttp.StatusInternalServerError)
	}
}
