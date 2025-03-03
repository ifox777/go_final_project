package middleware

import (
	"crypto/sha256"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		envPassword := os.Getenv("TODO_PASSWORD")
		if envPassword == "" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Auth required", http.StatusUnauthorized)
			return
		}

		// Валидация токена
		token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
			return []byte(envPassword), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Проверка хеша пароля
		claims := token.Claims.(jwt.MapClaims)
		expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(envPassword)))
		if claims["passHash"] != expectedHash {
			http.Error(w, "Password changed", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
