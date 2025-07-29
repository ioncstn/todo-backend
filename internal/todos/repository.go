package todos

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, item *ToDoItem) error
	GetAll(ctx context.Context, details PaginationDetails) ([]ToDoItem, error)
	GetAllForUser(ctx context.Context, userId uint, details PaginationDetails) ([]ToDoItem, error)
	GetById(ctx context.Context, id uint) (ToDoItem, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	CountAll(ctx context.Context) int
}

type repository struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func GetRepository(logger *zap.SugaredLogger, db *gorm.DB) Repository {
	return &repository{
		logger: logger,
		db:     db,
	}
}

func (r *repository) Create(ctx context.Context, item *ToDoItem) error {
	result := r.db.WithContext(ctx).Create(item)
	if result.Error != nil {
		r.logger.Errorw("failed to create todo item", "error", result.Error)

		return result.Error
	}

	return nil
}

func (r *repository) GetAll(ctx context.Context, details PaginationDetails) ([]ToDoItem, error) {
	var items []ToDoItem
	db := r.db.WithContext(ctx).Model(&ToDoItem{})

	if details.Page > 0 && details.Limit > 0 {
		db = db.Offset((details.Page - 1) * details.Limit).Limit(details.Limit)
	}

	if details.Order == "asc" || details.Order == "desc" {
		db = db.Order("done " + details.Order)
	}

	result := db.Find(&items)
	if result.Error != nil {
		r.logger.Errorw("failed to find all todo items", "error", result.Error)

		return nil, result.Error
	}

	return items, nil
}

func (r *repository) GetAllForUser(ctx context.Context, userId uint, details PaginationDetails) ([]ToDoItem, error) {
	var items []ToDoItem
	db := r.db.WithContext(ctx).Model(&ToDoItem{})

	if details.Page > 0 && details.Limit > 0 {
		db = db.Offset((details.Page - 1) * details.Limit).Limit(details.Limit)
	}

	if details.Order == "asc" || details.Order == "desc" {
		db = db.Order("done " + details.Order)
	}

	result := db.Where("user_id = ?", userId).Find(&items)
	if result.Error != nil {
		r.logger.Errorw("failed to find all todo items", "error", result.Error)

		return nil, result.Error
	}

	return items, nil
}

func (r *repository) GetById(ctx context.Context, id uint) (ToDoItem, error) {
	var item ToDoItem
	result := r.db.WithContext(ctx).First(&item, id)
	if result.Error != nil {
		r.logger.Errorw("failed to find todo item by id", "id", id, "error", result.Error)

		return ToDoItem{}, result.Error
	}

	return item, nil
}

func (r *repository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&ToDoItem{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		r.logger.Errorw("failed to update todo item", "id", id, "error", result.Error)

		return result.Error
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&ToDoItem{}, id)
	if result.Error != nil {
		r.logger.Errorw("failed to delete todo item", "id", id, "error", result.Error)

		return result.Error
	}

	return nil
}

func (r *repository) CountAll(ctx context.Context) int {
	var count int64
	r.db.WithContext(ctx).Model(&ToDoItem{}).Count(&count)

	return int(count)
}
