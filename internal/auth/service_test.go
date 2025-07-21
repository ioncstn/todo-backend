package auth

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
	"todo-app/internal/users"
	"todo-app/pkg/locale"
)

func TestService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAuthRepo := NewMockRepository(ctrl)
	mockUserRepo := users.NewMockRepository(ctrl)
	v := validator.New()
	logger := zap.NewNop().Sugar()
	service := GetService(logger, mockUserRepo, mockAuthRepo, v)
	ctx := context.Background()

	password := "test"
	userPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := users.User{
		ID:        1,
		Email:     "test@test.com",
		FirstName: "test",
		LastName:  "test",
		Password:  string(userPassword),
	}

	t.Run("successful login", func(t *testing.T) {
		mockUserRepo.
			EXPECT().
			GetByEmail(ctx, user.Email).
			Return(user, nil).
			Times(1)

		mockAuthRepo.
			EXPECT().
			SaveRefreshToken(ctx, gomock.Any()).
			Return(nil).
			Times(1)

		loginRequest := LoginRequest{
			Email:    user.Email,
			Password: password,
		}

		response, err := service.Login(ctx, loginRequest)
		assert.NoError(t, err)
		assert.Equal(t, user.FirstName, response.User.FirstName)
		assert.Equal(t, user.LastName, response.User.LastName)
		assert.Equal(t, user.Email, response.User.Email)

		ctrl.Finish()
	})

	t.Run("failed login", func(t *testing.T) {
		mockUserRepo.
			EXPECT().
			GetByEmail(ctx, user.Email).
			Return(user, nil).
			Times(1)

		loginRequest := LoginRequest{
			Email:    user.Email,
			Password: "wrong_password",
		}

		response, err := service.Login(ctx, loginRequest)
		assert.Error(t, err)
		assert.Equal(t, response, LoginResponse{})
		assert.Equal(t, err, errors.New(locale.ErrorInvalidCredentials))

		ctrl.Finish()
	})
}

func TestService_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAuthRepo := NewMockRepository(ctrl)
	mockUserRepo := users.NewMockRepository(ctrl)
	v := validator.New()
	logger := zap.NewNop().Sugar()
	service := GetService(logger, mockUserRepo, mockAuthRepo, v)
	ctx := context.Background()

	password := "test"
	userPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := users.User{
		ID:        1,
		Email:     "test@test.com",
		FirstName: "test",
		LastName:  "test",
		Password:  string(userPassword),
	}

	t.Run("successful logout", func(t *testing.T) {
		mockUserRepo.
			EXPECT().
			GetByEmail(ctx, user.Email).
			Return(user, nil).
			Times(1)

		mockAuthRepo.
			EXPECT().
			SaveRefreshToken(ctx, gomock.Any()).
			Return(nil).
			Times(1)

		loginRequest := LoginRequest{
			Email:    user.Email,
			Password: password,
		}

		response, err := service.Login(ctx, loginRequest)
		assert.NoError(t, err)
		assert.Equal(t, user.FirstName, response.User.FirstName)
		assert.Equal(t, user.LastName, response.User.LastName)
		assert.Equal(t, user.Email, response.User.Email)

		mockAuthRepo.
			EXPECT().
			RevokeRefreshTokensByUserID(ctx, user.ID).
			Return(nil).
			Times(1)

		err = service.Logout(ctx, response.Token)

		assert.NoError(t, err)

		ctrl.Finish()
	})

	t.Run("failed logout", func(t *testing.T) {
		err := service.Logout(ctx, "invalid_token")

		assert.Error(t, err)

		ctrl.Finish()
	})
}

func TestService_RefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockAuthRepo := NewMockRepository(ctrl)
	mockUserRepo := users.NewMockRepository(ctrl)
	v := validator.New()
	logger := zap.NewNop().Sugar()
	service := GetService(logger, mockUserRepo, mockAuthRepo, v)
	ctx := context.Background()

	t.Run("successful refresh token", func(t *testing.T) {
		token := "valid_token"
		tokenRecord := RefreshToken{
			UserID:    1,
			IsRevoked: false,
			ExpiresAt: time.Now().Add(time.Hour * 24),
		}

		user := users.User{
			ID:        1,
			Email:     "test@test.com",
			FirstName: "first test",
			LastName:  "last test",
		}

		mockAuthRepo.
			EXPECT().
			GetRefreshToken(ctx, token).
			Return(tokenRecord, nil).
			Times(1)

		mockUserRepo.
			EXPECT().
			GetById(ctx, user.ID).
			Return(user, nil).
			Times(1)

		response, err := service.RefreshToken(ctx, token)

		assert.NoError(t, err)
		assert.Equal(t, user.FirstName, response.User.FirstName)
		assert.Equal(t, user.LastName, response.User.LastName)
		assert.Equal(t, user.Email, response.User.Email)

		ctrl.Finish()
	})

	t.Run("expired refresh token", func(t *testing.T) {
		token := "expired_token"
		tokenRecord := RefreshToken{
			UserID:    1,
			IsRevoked: false,
			ExpiresAt: time.Now().Add(-time.Hour * 24),
		}

		mockAuthRepo.
			EXPECT().
			GetRefreshToken(ctx, token).
			Return(tokenRecord, nil).
			Times(1)

		response, err := service.RefreshToken(ctx, token)

		assert.Error(t, err)
		assert.Equal(t, response, LoginResponse{})
		assert.Equal(t, err, errors.New(locale.ErrorInvalidToken))

		ctrl.Finish()
	})

	t.Run("revoked refresh token", func(t *testing.T) {
		token := "revoked_token"
		tokenRecord := RefreshToken{
			UserID:    1,
			IsRevoked: true,
			ExpiresAt: time.Now().Add(time.Hour * 24),
		}

		mockAuthRepo.
			EXPECT().
			GetRefreshToken(ctx, token).
			Return(tokenRecord, nil).
			Times(1)

		response, err := service.RefreshToken(ctx, token)

		assert.Error(t, err)
		assert.Equal(t, response, LoginResponse{})
		assert.Equal(t, err, errors.New(locale.ErrorInvalidToken))

		ctrl.Finish()
	})
}
