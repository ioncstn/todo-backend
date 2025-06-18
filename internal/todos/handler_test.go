package todos

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"todo-app/pkg/locale"

	err "todo-app/pkg/errors"
)

func TestHandler_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	e := echo.New()
	todoItemData := `{"text":"go for a run", "done": false}`
	todoItem := ToDoItem{
		Text: "go for a run",
		Done: false,
	}

	req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(todoItemData))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	logger := zap.NewNop().Sugar()
	h := &endpointHandler{logger: logger, service: mockService, e: e}

	mockService.
		EXPECT().
		Create(ctx.Request().Context(), &todoItem).
		Return(nil).
		Times(1)

	if assert.NoError(t, h.create(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// unmarshalling response to check if response data is ok
		var responseItem ToDoItem
		err := json.Unmarshal(rec.Body.Bytes(), &responseItem)
		assert.NoError(t, err)

		assert.Equal(t, todoItem.Text, responseItem.Text)
		assert.Equal(t, todoItem.Done, responseItem.Done)
	}
}

func TestHandler_GetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	e := echo.New()
	todoItems := []ToDoItem{
		{
			Text: "go for a run",
			Done: false,
		},
		{
			Text: "go for a walk",
			Done: false,
		},
		{
			Text: "go for a jog",
			Done: false,
		},
	}

	logger := zap.NewNop().Sugar()
	h := &endpointHandler{logger: logger, service: mockService, e: e}

	t.Run("all items", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/todos", nil)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		mockService.
			EXPECT().
			GetAll(ctx.Request().Context(), PaginationDetails{}).
			Return(todoItems, PaginationMetadata{
				ResultCount: len(todoItems),
				TotalCount:  len(todoItems),
			}, nil).
			Times(1)

		if assert.NoError(t, h.getAll(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var responseData PaginatedResponse
			err := json.Unmarshal(rec.Body.Bytes(), &responseData)
			assert.NoError(t, err)

			assert.Equal(t, responseData.Data, todoItems)
			assert.Equal(t, responseData.Meta.ResultCount, len(todoItems))
			assert.Equal(t, responseData.Meta.TotalCount, len(todoItems))
		}
	})

	t.Run("paginated items", func(t *testing.T) {
		paginationDetails := PaginationDetails{
			Page:  1,
			Limit: 2,
		}

		q := make(url.Values)
		q.Set("page", strconv.Itoa(paginationDetails.Page))
		q.Set("limit", strconv.Itoa(paginationDetails.Limit))
		req := httptest.NewRequest(http.MethodGet, "/todos?"+q.Encode(), nil)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		mockService.
			EXPECT().
			GetAll(ctx.Request().Context(), paginationDetails).
			Return(todoItems[:2], PaginationMetadata{
				ResultCount: 2,
				TotalCount:  3,
			}, nil).
			Times(1)

		if assert.NoError(t, h.getAll(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var responseData PaginatedResponse
			err := json.Unmarshal(rec.Body.Bytes(), &responseData)
			assert.NoError(t, err)

			assert.Equal(t, responseData.Data, todoItems[:2])
			assert.Equal(t, responseData.Meta.ResultCount, 2)
			assert.Equal(t, responseData.Meta.TotalCount, 3)
		}
	})
}

func TestHandler_UpdateById(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	e := echo.New()
	logger := zap.NewNop().Sugar()
	h := &endpointHandler{logger: logger, service: mockService, e: e}

	t.Run("update item", func(t *testing.T) {
		item := ToDoItem{
			Model: gorm.Model{ID: 1},
			Text:  "go for a run",
			Done:  false,
		}
		itemBody := `{"text":"go for a walk"}`
		updatedItem := ToDoItem{
			Model: gorm.Model{ID: 1},
			Text:  "go for a walk",
			Done:  false,
		}

		updateInput := ToDoItemUpdateInput{
			Text: stringPtr("go for a walk"),
		}

		req := httptest.NewRequest(http.MethodPut, "/todos/1", strings.NewReader(itemBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		ctx.SetPath("/todos/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")

		mockService.
			EXPECT().
			UpdateById(ctx.Request().Context(), item.Model.ID, updateInput).
			Return(updatedItem, nil).
			Times(1)

		if assert.NoError(t, h.updateById(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var responseItem ToDoItem
			err := json.Unmarshal(rec.Body.Bytes(), &responseItem)
			assert.NoError(t, err)

			assert.Equal(t, updatedItem.Text, responseItem.Text)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		itemBody := `{"text":"go for a walk"}`

		req := httptest.NewRequest(http.MethodPut, "/todos/abc", strings.NewReader(itemBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		ctx.SetPath("/todos/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("abc")

		if assert.NoError(t, h.updateById(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError err.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorInvalidID, responseError.Message)
		}
	})

	t.Run("no updates", func(t *testing.T) {
		itemBody := `{"text": ""}`
		item := ToDoItem{
			Model: gorm.Model{ID: 1},
			Text:  "go for a run",
			Done:  false,
		}
		updateInput := ToDoItemUpdateInput{
			Text: stringPtr(""),
		}

		req := httptest.NewRequest(http.MethodPut, "/todos/1", strings.NewReader(itemBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		ctx.SetPath("/todos/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")

		mockService.
			EXPECT().
			UpdateById(ctx.Request().Context(), item.Model.ID, updateInput).
			Return(ToDoItem{}, errors.New(locale.ErrorNotFoundUpdates)).
			Times(1)

		if assert.NoError(t, h.updateById(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError err.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorNotFoundUpdates, responseError.Message)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		itemBody := `{"text":123, "done": 23}`

		req := httptest.NewRequest(http.MethodPut, "/todos/1", strings.NewReader(itemBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		ctx.SetPath("/todos/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")

		if assert.NoError(t, h.updateById(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError err.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorInvalidBody, responseError.Message)
		}
	})
}

func TestHandler_DeleteById(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	e := echo.New()
	logger := zap.NewNop().Sugar()
	h := &endpointHandler{logger: logger, service: mockService, e: e}

	t.Run("delete item", func(t *testing.T) {
		id := 1
		req := httptest.NewRequest(http.MethodDelete, "/todos/"+strconv.Itoa(id), nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		ctx.SetPath("/todos/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues(strconv.Itoa(id))

		mockService.
			EXPECT().
			DeleteById(ctx.Request().Context(), uint(id)).
			Return(nil).
			Times(1)

		if assert.NoError(t, h.deleteById(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var response string
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "", response)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/todos/abc", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		ctx.SetPath("/todos/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("abc")

		if assert.NoError(t, h.updateById(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError err.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorInvalidID, responseError.Message)
		}
	})
}
