package handler

import (
	"context"
	authDto "ecomm/ecomm-api/handler/dto/auth"
	"encoding/json"
	"errors"
	"net/http"
)

type AuthHandler struct {
	authService AuthService
}

type AuthService interface {
	Register(ctx context.Context, request *authDto.RegisterRequest) (authDto.RegisterResponse, error)
	Authenticate(ctx context.Context, request *authDto.LoginRequest) (authDto.LoginResponse, error)
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authDto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseWithError(w, r, errors.New("invalid request payload"))
		return
	}
	res, err := h.authService.Register(r.Context(), &req)

	if err != nil {
		responseWithError(w, r, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, res)
}

func (h *AuthHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	var req authDto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseWithError(w, r, errors.New("invalid request payload"))
		return
	}
	res, err := h.authService.Authenticate(r.Context(), &req)

	if err != nil {
		responseWithError(w, r, err)
		return
	}

	respondWithJSON(w, http.StatusOK, res)
}
