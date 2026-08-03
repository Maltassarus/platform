package handler

import (
	"encoding/json"
	"net/http"

	"platform/internal/model"
	"platform/internal/service"
	"platform/pkg/auth"
)

type AuthHandler struct {
	userService *service.UserService
	jwtManager  *auth.JWTManager
}

func NewAuthHandler(userService *service.UserService, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	sendSuccess(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.userService.Login(&req)
	if err != nil {
		sendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := h.jwtManager.GenerateToken(user.ID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	sendSuccess(w, http.StatusOK, model.AuthResponse{
		Token: token,
		User:  user,
	})
}

func sendSuccess(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(model.ErrorResponse{Error: message})
}
