package http

import (
	stdhttp "net/http"
)

type SystemHandler struct{}

func NewSystemHandler() SystemHandler {
	return SystemHandler{}
}

// Info godoc
// @Summary Service healthcheck
// @Description Возвращает простой статус готовности сервиса.
// @Tags system
// @Produce json
// @Success 200 {object} http.InfoResponse
// @Router /_info [get]
func (h SystemHandler) Info(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	response := map[string]string{
		"status": "ok",
	}

	writeJSON(w, stdhttp.StatusOK, response)
}
