package users

import (
	"context"
	"errors"
	"time"
	"todo-app/pkg/email"
	"todo-app/pkg/locale"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) (User, error)
	VerifyEmail(ctx context.Context, token string) error
}

type service struct {
	logger       *zap.SugaredLogger
	repository   Repository
	validator    *validator.Validate
	emailService email.Service
}

func GetService(logger *zap.SugaredLogger, repo Repository, validator *validator.Validate, emailService email.Service) Service {
	return &service{
		logger:       logger,
		repository:   repo,
		validator:    validator,
		emailService: emailService,
	}
}

func (s *service) Create(ctx context.Context, user *User) error {
	if err := s.validator.Struct(user); err != nil {
		return err
	}

	// Generate verification token
	verificationToken := uuid.New().String()
	expiryTime := time.Now().Add(24 * time.Hour)

	user.EmailVerificationToken = verificationToken
	user.EmailVerificationExpiry = &expiryTime
	user.IsEmailVerified = false

	err := s.repository.Create(ctx, user)
	if err != nil {
		return err
	}

	// Send verification email
	err = s.emailService.SendVerificationEmail(user.Email, user.FirstName, verificationToken)
	if err != nil {
		s.logger.Errorw("failed to send verification email", "error", err, "email", user.Email)
		// Don't return error here - user is created but email failed to send
	}

	return nil
}

func (s *service) VerifyEmail(ctx context.Context, token string) error {
	user, err := s.repository.GetByEmailVerificationToken(ctx, token)
	if err != nil {
		return errors.New("invalid verification token")
	}

	// Check if token is expired
	if user.EmailVerificationExpiry != nil && time.Now().After(*user.EmailVerificationExpiry) {
		return errors.New("verification token has expired")
	}

	// Check if email is already verified
	if user.IsEmailVerified {
		return errors.New("email is already verified")
	}

	// Mark email as verified
	err = s.repository.VerifyEmail(ctx, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) Update(ctx context.Context, user *User) (User, error) {
	actualUser, err := s.repository.GetById(ctx, user.ID)
	if err != nil {
		return User{}, err
	}

	updates := map[string]interface{}{}

	if actualUser.Email != user.Email {
		updates["email"] = user.Email
	}
	if actualUser.FirstName != user.FirstName {
		updates["first_name"] = user.FirstName
	}
	if actualUser.LastName != user.LastName {
		updates["last_name"] = user.LastName
	}
	if bcrypt.CompareHashAndPassword([]byte(actualUser.Password), []byte(user.Password)) != nil {
		updates["password"] = user.Password
	}

	if len(updates) == 0 {
		return User{}, errors.New(locale.ErrorNotFoundUpdates)
	}

	err = s.repository.Update(ctx, user.ID, updates)
	if err != nil {
		return User{}, err
	}

	updatedUser, err := s.repository.GetById(ctx, user.ID)
	if err != nil {
		return User{}, errors.New(locale.ErrorNotFoundRecord)
	}

	updatedUser.Password = ""
	return updatedUser, nil
}
