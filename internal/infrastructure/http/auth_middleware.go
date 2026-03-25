package http

import (
	stdhttp "net/http"
	"strings"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
	"github.com/avito-internships/test-backend-1-EdOoO21/internal/infrastructure/http/generated"
)

const bearerHeaderParts = 2

func authMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			writeAPIError(w, shared.ErrUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", bearerHeaderParts)
		if len(parts) != bearerHeaderParts || !strings.EqualFold(parts[0], "Bearer") {
			writeAPIError(w, shared.ErrUnauthorized)
			return
		}

		services, ok := servicesFromContext(r.Context())
		if !ok || services.JWT == nil {
			writeAPIError(w, apiError{Status: stdhttp.StatusInternalServerError, Code: generated.INTERNALERROR, Message: "internal server error"})
			return
		}

		claims, err := services.JWT.ParseToken(parts[1])
		if err != nil {
			writeAPIError(w, apiError{Status: stdhttp.StatusUnauthorized, Code: generated.UNAUTHORIZED, Message: "invalid token", Err: err})
			return
		}

		actor := shared.Actor{UserID: claims.UserID, Role: claims.Role}
		next.ServeHTTP(w, r.WithContext(withActor(r.Context(), actor)))
	})
}

func isPublicPath(path string) bool {
	switch path {
	case "/dummyLogin", "/login", "/register":
		return true
	default:
		return false
	}
}
