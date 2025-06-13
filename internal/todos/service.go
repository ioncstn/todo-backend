package todos

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"
	e "todo-app/internal/errors"
)

type Service interface {
	Create(ctx context.Context, item *ToDoItem) error
	GetAll(ctx context.Context) ([]ToDoItem, error)
	GetById(ctx context.Context, id uint) (ToDoItem, error)
	UpdateById(ctx context.Context, id uint, item ToDoItemUpdateInput) (ToDoItem, error)
	DeleteById(ctx context.Context, id uint) error
}

type service struct {
	logger    *zap.SugaredLogger
	db        *gorm.DB
	validator *validator.Validate
}

func GetService(
	logger *zap.SugaredLogger,
	db *gorm.DB,
	validator *validator.Validate,
) Service {
	return &service{
		logger:    logger,
		db:        db,
		validator: validator,
	}
}

func (s *service) Create(ctx context.Context, item *ToDoItem) error {
	if err := s.validator.Struct(item); err != nil {
		return e.ResponseError{Message: "invalid item", Details: err.Error()}
	}

	result := s.db.WithContext(ctx).Create(item)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *service) GetAll(ctx context.Context) ([]ToDoItem, error) {
	var items []ToDoItem
	result := s.db.WithContext(ctx).Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

func (s *service) GetById(ctx context.Context, id uint) (ToDoItem, error) {
	var item ToDoItem
	result := s.db.WithContext(ctx).First(&item, id)
	if result.Error != nil {
		return ToDoItem{}, result.Error
	}

	return item, nil
}

func (s *service) UpdateById(ctx context.Context, id uint, item ToDoItemUpdateInput) (ToDoItem, error) {
	updates := map[string]interface{}{}

	if item.Text != nil {
		updates["text"] = item.Text
	}
	if item.Done != nil {
		updates["done"] = item.Done
	}

	if len(updates) == 0 {
		return ToDoItem{}, errors.New("no updates found")
	}

	result := s.db.WithContext(ctx).Model(&ToDoItem{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return ToDoItem{}, result.Error
	}

	if result.RowsAffected == 0 {
		return ToDoItem{}, e.ResponseErrorNotFound
	}

	var updatedItem ToDoItem

	s.db.WithContext(ctx).First(&updatedItem, id)

	return updatedItem, nil
}

func (s *service) DeleteById(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&ToDoItem{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
