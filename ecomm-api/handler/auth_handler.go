package handler

import (
	"context"
	authDto "ecomm/ecomm-api/handler/dto/auth"
	"ecomm/ecomm-api/service"
	"ecomm/util"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

type AuthService interface {
	Register(ctx context.Context, request *authDto.RegisterRequest) (authDto.RegisterResponse, error)
	Authenticate(ctx context.Context, request *authDto.LoginRequest) (authDto.LoginResponse, error)
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req authDto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseWithError(w, r, errors.New("invalid request payload"))
		return
	}

	// TODO: Здесь должна быть реальная логика проверки учетных данных пользователя (из БД, например)
	// и получение user ID.
	// Для примера, используем хардкод:
	if req.Email != "test@example.com" || req.Password != "password" {
		responseWithError(w, r, errors.New("invalid credentials"))
		return
	}
	userID := "some-user-id-from-db" // Замените на реальный userID после аутентификации

	// Генерация Access Token
	accessToken, err := service.GenerateToken(userID, util.GetAccessTokenExpiration())
	if err != nil {
		responseWithError(w, r, fmt.Errorf("failed to generate access token: %w", err))
		return
	}

	resp := authDto.LoginResponse{
		AccessToken: accessToken,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authDto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseWithError(w, r, errors.New("invalid request payload"))
		return
	}
}
