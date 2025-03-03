package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"time"
)

type SigninRequest struct {
	Password string `json:"password"`
}

type SigninResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"token,omitempty"`
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Разрешаем CORS для фронтенда
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:7540")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	envPassword := os.Getenv("TODO_PASSWORD")
	if envPassword == "" {
		respondWithError(w, http.StatusInternalServerError, "Auth disabled")
		return
	}

	if req.Password != envPassword {
		respondWithError(w, http.StatusUnauthorized, "Неверный пароль")
		return
	}

	// Генерация токена с хешем пароля
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"passHash": fmt.Sprintf("%x", sha256.Sum256([]byte(envPassword))),
		"exp":      time.Now().Add(8 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(envPassword))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Token error")
		return
	}

	// Установка куки
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: time.Now().Add(8 * time.Hour),
		Path:    "/",
	})

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
