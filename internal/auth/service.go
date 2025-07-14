package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"time"
	"todo-app/internal/users"
	"todo-app/pkg/locale"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
	Logout(ctx context.Context, token string) error
	ValidateToken(tokenString string) (*JWTClaims, error)
	RefreshToken(ctx context.Context, refreshToken string) (LoginResponse, error)
}

type service struct {
	logger          *zap.SugaredLogger
	userRepository  users.Repository
	authRepository  Repository
	validator       *validator.Validate
	jwtSecret       []byte
	tokenExpiration time.Duration
}

func GetService(
	logger *zap.SugaredLogger,
	userRepo users.Repository,
	authRepo Repository,
	validator *validator.Validate,
) Service {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "password"
	}

	return &service{
		logger:          logger,
		userRepository:  userRepo,
		authRepository:  authRepo,
		validator:       validator,
		jwtSecret:       []byte(jwtSecret),
		tokenExpiration: 20 * time.Minute,
	}
}

func (s *service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	if err := s.validator.Struct(req); err != nil {
		return LoginResponse{}, err
	}

	user, err := s.userRepository.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warnw("user not found", "email", req.Email)
		return LoginResponse{}, errors.New(locale.ErrorInvalidCredentials)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warnw("invalid password", "email", req.Email)
		return LoginResponse{}, errors.New(locale.ErrorInvalidCredentials)
	}

	if user.IsEmailVerified == false {
		s.logger.Warnw("email not verified", "email", req.Email)
		return LoginResponse{}, errors.New(locale.ErrorEmailUnverified)
	}

	expiresAt := time.Now().Add(s.tokenExpiration)
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "todo-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.logger.Errorw("failed to sign token", "error", err)
		return LoginResponse{}, errors.New(locale.ErrorInternalServer)
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		s.logger.Errorw("failed to generate refresh token", "error", err)
		return LoginResponse{}, errors.New(locale.ErrorInternalServer)
	}

	refreshTokenRecord := RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		IsRevoked: false,
	}

	if err := s.authRepository.SaveRefreshToken(ctx, &refreshTokenRecord); err != nil {
		s.logger.Errorw("failed to store refresh token", "error", err)
		return LoginResponse{}, errors.New(locale.ErrorInternalServer)
	}

	return LoginResponse{
		Token:     tokenString,
		Refresh:   refreshToken,
		ExpiresAt: expiresAt.Unix(),
		User: UserInfo{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
	}, nil
}

func (s *service) Logout(ctx context.Context, token string) error {
	claims, err := s.ValidateToken(token)
	if err != nil {
		return err
	}

	return s.authRepository.RevokeRefreshTokensByUserID(ctx, claims.UserID)
}

func (s *service) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (LoginResponse, error) {
	// Validate refresh token
	tokenRecord, err := s.authRepository.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return LoginResponse{}, errors.New(locale.ErrorInvalidToken)
	}

	if tokenRecord.IsRevoked || time.Now().After(tokenRecord.ExpiresAt) {
		return LoginResponse{}, errors.New(locale.ErrorInvalidToken)
	}

	user, err := s.userRepository.GetById(ctx, tokenRecord.UserID)
	if err != nil {
		return LoginResponse{}, errors.New(locale.ErrorUserNotFound)
	}

	// Generate new JWT token
	expiresAt := time.Now().Add(s.tokenExpiration)
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "todo-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.logger.Errorw("failed to sign token", "error", err)
		return LoginResponse{}, errors.New(locale.ErrorInternalServer)
	}

	return LoginResponse{
		Token:     tokenString,
		Refresh:   refreshToken,
		ExpiresAt: expiresAt.Unix(),
		User: UserInfo{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
	}, nil
}

func (s *service) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
