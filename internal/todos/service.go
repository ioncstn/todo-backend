package todos

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"todo-app/pkg/locale"
)

type Service interface {
	Create(ctx context.Context, item *ToDoItem) error
	GetAllForUser(ctx context.Context, userId uint, details PaginationDetails) ([]ToDoItem, PaginationMetadata, error)
	GetById(ctx context.Context, id uint) (ToDoItem, error)
	UpdateById(ctx context.Context, id uint, item ToDoItemUpdateInput) (ToDoItem, error)
	DeleteById(ctx context.Context, id uint) error
}

type service struct {
	logger     *zap.SugaredLogger
	repository Repository
	validator  *validator.Validate
}

func GetService(logger *zap.SugaredLogger, repo Repository, validator *validator.Validate) Service {
	return &service{
		logger:     logger,
		repository: repo,
		validator:  validator,
	}
}

func (s *service) Create(ctx context.Context, item *ToDoItem) error {
	if err := s.validator.Struct(item); err != nil {
		return err
	}

	return s.repository.Create(ctx, item)
}

func (s *service) GetAllForUser(ctx context.Context, userId uint, details PaginationDetails) ([]ToDoItem, PaginationMetadata, error) {
	s.logger.Infow("get all todos", "details", details)

	items, err := s.repository.GetAllForUser(ctx, userId, details)

	if err != nil {
		return nil, PaginationMetadata{}, err
	}

	totalCount := s.repository.CountAll(ctx)

	return items, PaginationMetadata{TotalCount: totalCount, ResultCount: len(items)}, nil
}

func (s *service) GetById(ctx context.Context, id uint) (ToDoItem, error) {
	return s.repository.GetById(ctx, id)
}

func (s *service) UpdateById(ctx context.Context, id uint, item ToDoItemUpdateInput) (ToDoItem, error) {
	updates := map[string]interface{}{}

	if item.Text != nil && *item.Text != "" {
		updates["text"] = *item.Text
	}
	if item.Done != nil {
		updates["done"] = *item.Done
	}

	if len(updates) == 0 {
		return ToDoItem{}, errors.New(locale.ErrorNotFoundUpdates)
	}

	err := s.repository.Update(ctx, id, updates)
	if err != nil {
		return ToDoItem{}, err
	}

	updatedItem, err := s.repository.GetById(ctx, id)
	if err != nil {
		return ToDoItem{}, errors.New(locale.ErrorNotFoundRecord)
	}

	return updatedItem, nil
}

func (s *service) DeleteById(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}
