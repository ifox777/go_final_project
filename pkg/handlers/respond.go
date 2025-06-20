package handlers

import (
	"encoding/json"
	"net/http"
)

// Вспомогательные функции для ответов
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func respondWithSuccess(w http.ResponseWriter, code int, id int64) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(SuccessResponse{ID: id})
}
