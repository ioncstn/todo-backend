package todos

import (
	"context"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"testing"
	"todo-app/pkg/locale"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)
	v := validator.New()
	logger := zap.NewNop().Sugar()
	service := GetService(logger, mockRepo, v)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		todo := &ToDoItem{Text: "buy milk"}
		mockRepo.
			EXPECT().
			Create(ctx, todo).
			Return(nil).
			Times(1)

		err := service.Create(ctx, todo)
		assert.NoError(t, err)

		ctrl.Finish()
	})

	t.Run("validation error", func(t *testing.T) {
		todo := &ToDoItem{Text: ""} // Empty text should fail validation

		err := service.Create(ctx, todo)
		assert.Error(t, err)

		ctrl.Finish()
	})
}

func TestService_GetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)
	v := validator.New()
	logger := zap.NewNop().Sugar()
	service := GetService(logger, mockRepo, v)
	ctx := context.Background()

	expectedTodos := []ToDoItem{
		{Text: "Todo 1", Done: false},
		{Text: "Todo 2", Done: true},
	}

	mockRepo.
		EXPECT().
		CountAll(ctx).
		Return(len(expectedTodos)).
		Times(2)

	mockRepo.
		EXPECT().
		GetAll(ctx, PaginationDetails{}).
		Return(expectedTodos, nil).
		Times(1)

	todos, metadata, err := service.GetAll(ctx, PaginationDetails{})
	assert.NoError(t, err)
	assert.Equal(t, expectedTodos, todos)
	assert.Equal(t, metadata.ResultCount, metadata.TotalCount)

	mockRepo.
		EXPECT().
		GetAll(ctx, PaginationDetails{Limit: 1, Page: 1}).
		Return([]ToDoItem{expectedTodos[0]}, nil).
		Times(1)

	todos, metadata, err = service.GetAll(ctx, PaginationDetails{Limit: 1, Page: 1})
	assert.NoError(t, err)
	assert.Equal(t, expectedTodos[0], todos[0])
	assert.Equal(t, metadata.ResultCount, 1)

	ctrl.Finish()
}

func TestService_GetById(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)
	v := validator.New()
	logger := zap.NewNop().Sugar()
	service := GetService(logger, mockRepo, v)
	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		expectedTodo := ToDoItem{Text: "found me"}
		expectedTodo.ID = 1
		mockRepo.
			EXPECT().
			GetById(ctx, uint(1)).
			Return(expectedTodo, nil).
			Times(1)

		todo, err := service.GetById(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, expectedTodo, todo)

		ctrl.Finish()
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.
			EXPECT().
			GetById(ctx, uint(99)).
			Return(ToDoItem{}, gorm.ErrRecordNotFound).
			Times(1)

		_, err := service.GetById(ctx, 99)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)

		ctrl.Finish()
	})
}

func TestService_UpdateById(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)
	v := validator.New()
	logger := zap.NewNop().Sugar()
	service := GetService(logger, mockRepo, v)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		updateInput := ToDoItemUpdateInput{Text: stringPtr("updated text")}
		updates := map[string]interface{}{"text": "updated text"}
		updatedTodo := ToDoItem{Text: "updated text"}
		updatedTodo.ID = 1

		mockRepo.
			EXPECT().
			Update(ctx, uint(1), updates).
			Return(nil).
			Times(1)
		mockRepo.
			EXPECT().
			GetById(ctx, uint(1)).
			Return(updatedTodo, nil).
			Times(1)

		todo, err := service.UpdateById(ctx, 1, updateInput)
		assert.NoError(t, err)
		assert.Equal(t, updatedTodo, todo)

		ctrl.Finish()
	})

	t.Run("no updates provided", func(t *testing.T) {
		updateInput := ToDoItemUpdateInput{}

		_, err := service.UpdateById(ctx, 1, updateInput)
		assert.Error(t, err)
		assert.Equal(t, locale.ErrorNotFoundUpdates, err.Error())

		ctrl.Finish()
	})

	t.Run("update fails", func(t *testing.T) {
		updateInput := ToDoItemUpdateInput{Done: boolPtr(true)}
		updates := map[string]interface{}{"done": true}
		mockRepo.
			EXPECT().
			Update(ctx, uint(1), updates).
			Return(gorm.ErrInvalidDB).
			Times(1)

		_, err := service.UpdateById(ctx, 1, updateInput)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidDB, err)

		ctrl.Finish()
	})
}

func TestService_DeleteById(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := NewMockRepository(ctrl)
	v := validator.New()
	logger := zap.NewNop().Sugar()
	service := GetService(logger, mockRepo, v)
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockRepo.
			EXPECT().
			Delete(ctx, uint(1)).
			Return(nil).
			Times(1)

		err := service.DeleteById(ctx, 1)
		assert.NoError(t, err)

		ctrl.Finish()
	})

	t.Run("delete fails", func(t *testing.T) {
		mockRepo.
			EXPECT().
			Delete(ctx, uint(1)).
			Return(gorm.ErrInvalidDB).
			Times(1)

		err := service.DeleteById(ctx, 1)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrInvalidDB, err)

		ctrl.Finish()
	})
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
