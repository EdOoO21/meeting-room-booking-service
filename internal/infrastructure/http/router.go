package http

import (
	stdhttp "net/http"
	"time"

	// Register generated Swagger docs.
	_ "github.com/avito-internships/test-backend-1-EdOoO21/docs"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/http/generated"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

const routerTimeout = 30 * time.Second

func NewRouter(services Services) stdhttp.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(requestLogger(services.Logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(routerTimeout))
	router.Use(func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			next.ServeHTTP(w, r.WithContext(withServices(r.Context(), services)))
		})
	})

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

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
