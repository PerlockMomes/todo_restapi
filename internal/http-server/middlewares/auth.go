package middlewares

import (
	"net/http"
)

func Auth(authService *AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(write http.ResponseWriter, request *http.Request) {

			if err := authService.ValidateJWT(request); err != nil {
				http.Error(write, "authentication required", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(write, request)
		})
	}
}
