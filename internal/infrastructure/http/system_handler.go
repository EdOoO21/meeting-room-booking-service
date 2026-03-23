package http

import (
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
