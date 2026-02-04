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

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}
