package http

import (
	stdhttp "net/http"
	"time"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/http/generated"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(services Services) stdhttp.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(requestLogger(services.Logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))
	router.Use(func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			next.ServeHTTP(w, r.WithContext(withServices(r.Context(), services)))
		})
	})

	systemHandler := NewSystemHandler()
	router.Get("/_info", systemHandler.Info)

	server := NewServer(services)
	generated.HandlerWithOptions(server, generated.ChiServerOptions{
		BaseRouter:       router,
		Middlewares:      []generated.MiddlewareFunc{authMiddleware},
		ErrorHandlerFunc: writeGeneratedParamError,
	})

	return router
}
