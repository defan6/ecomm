package service

import (
	"context"
	"ecomm/domain"
	authDto "ecomm/ecomm-api/handler/dto/auth"
	"ecomm/mapper"
	"ecomm/util"
	"strconv"
	"time"
)

type UserStorer interface {
	ExistsByEmail(ctx context.Context, email string) bool
	SaveUser(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

type PasswordEncoder interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) bool
}

type TokenGenerator interface {
	GenerateToken(userID string, role string, duration time.Duration) (string, error)
}

type TokenValidator interface {
	ValidateToken(tokenString string) (*util.JWTClaims, error)
}
type AuthService struct {
	userStorer      UserStorer
	passwordEncoder PasswordEncoder
	tokenGenerator  TokenGenerator
}

func NewAuthService(userStorer UserStorer, passwordEncoder PasswordEncoder, tokenGenerator TokenGenerator) *AuthService {
	return &AuthService{userStorer: userStorer, passwordEncoder: passwordEncoder, tokenGenerator: tokenGenerator}
}

func (s *AuthService) Register(ctx context.Context, request *authDto.RegisterRequest) (authDto.RegisterResponse, error) {

	if s.ExistsByEmail(ctx, request.Email) {
		return authDto.RegisterResponse{}, NewErrEmailAlreadyExists("register", "user", request.Email, nil)
	}

	user := mapper.MapToUserFromRegisterReq(request)
	hashedPassword, err := s.passwordEncoder.HashPassword(user.Password)
	if err != nil {
		return authDto.RegisterResponse{}, err
	}
	user.Password = hashedPassword
	user.Role = domain.RoleUser
	err = s.userStorer.SaveUser(ctx, user)
	if err != nil {
		return authDto.RegisterResponse{}, err
	}
	userRes := mapper.MapToUserResFromUser(user)
	return userRes, nil
}

func (s *AuthService) Authenticate(ctx context.Context, request *authDto.LoginRequest) (authDto.LoginResponse, error) {
	u, err := s.userStorer.FindByEmail(ctx, request.Email)
	if err != nil {
		return authDto.LoginResponse{}, err
	}
	if !s.passwordEncoder.CheckPassword(request.Password, u.Password) {
		return authDto.LoginResponse{}, NewErrInvalidEmailOrPassword("login", "user", request.Email, nil)
	}
	tkn, err := s.tokenGenerator.GenerateToken(strconv.FormatInt(u.ID, 10), u.Role, util.GetAccessTokenExpiration())
	if err != nil {
		return authDto.LoginResponse{}, err
	}
	return authDto.LoginResponse{AccessToken: tkn}, nil
}

func (s *AuthService) ExistsByEmail(context context.Context, email string) bool {
	return s.userStorer.ExistsByEmail(context, email)
}
