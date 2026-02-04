package auth

// LoginRequest определяет структуру для запроса входа
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse определяет структуру для ответа входа
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	// RefreshToken string `json:"refresh_token,omitempty"` // Для будущих refresh-токенов
}
