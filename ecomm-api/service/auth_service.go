package service

import (
	"context"
	"ecomm/domain"
	authDto "ecomm/ecomm-api/handler/dto/auth"
	"ecomm/ecomm-api/storer"
	"ecomm/mapper"
	"ecomm/util"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type UserStorer interface {
	ExistsByEmail(ctx context.Context, email string) bool
	RegisterUser(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

type PasswordEncoder interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) (bool, error)
}

type TokenGenerator interface {
	GenerateToken(user *domain.User) (string, error)
}
type AuthService struct {
	userStorer      UserStorer
	passwordEncoder PasswordEncoder
	tokenGenerator  TokenGenerator
}

func NewAuthService(userStorer UserStorer, passwordEncoder PasswordEncoder, tokenGenerator TokenGenerator) *AuthService {
	return &AuthService{userStorer: userStorer, passwordEncoder: passwordEncoder, tokenGenerator: tokenGenerator}
}

func GenerateToken(userID string, duration time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(util.GetJWTSecret())
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *AuthService) Register(ctx context.Context, request *authDto.RegisterRequest) (authDto.RegisterResponse, error) {

	if s.ExistsByEmail(ctx, request.Email) {
		return authDto.RegisterResponse{}, NewErrUserWithThatEmailAlreadyExists("register", "user", request.Email, nil)
	}

	user := mapper.MapToUserFromRegisterReq(request)
	hashedPassword, err := util.HashPassword(user.Password)
	if err != nil {
		return authDto.RegisterResponse{}, err
	}
	user.Password = hashedPassword
	err = s.userStorer.Register(ctx, user)
	if err != nil {
		return authDto.RegisterResponse{}, err
	}
	userRes := mapper.MapToUserResFromUser(user)
	return userRes, nil
}

func (s *AuthService) Authenticate(ctx context.Context, request *authDto.LoginRequest) (authDto.LoginResponse, error) {

}

func (s *AuthService) ExistsByEmail(context context.Context, email string) bool {
	return s.userStorer.ExistsByEmail(context, email)
}
