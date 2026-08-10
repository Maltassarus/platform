package response

import (
	"encoding/json"
	"net/http"

	"platform/internal/model"
)

// Success отправляет успешный ответ
func Success(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// Error отправляет ошибку
func Error(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(model.ErrorResponse{Error: message})
}
