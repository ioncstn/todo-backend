package todos

import (
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"todo-app/pkg/locale"

	localErr "todo-app/pkg/errors"
)

func TestHandler_Create(t *testing.T) {
	os.Setenv("APP_ENV", "test")
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
	ctx.Set("user_id", uint(0))

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
	os.Setenv("APP_ENV", "test")
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
		ctx.Set("user_id", uint(0))

		mockService.
			EXPECT().
			GetAllForUser(ctx.Request().Context(), uint(0), PaginationDetails{}).
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
		ctx.Set("user_id", uint(0))

		mockService.
			EXPECT().
			GetAllForUser(ctx.Request().Context(), uint(0), paginationDetails).
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
	os.Setenv("APP_ENV", "test")
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	e := echo.New()
	logger := zap.NewNop().Sugar()
	h := &endpointHandler{logger: logger, service: mockService, e: e}

	t.Run("update item", func(t *testing.T) {
		item := ToDoItem{
			ID:   1,
			Text: "go for a run",
			Done: false,
		}
		itemBody := `{"text":"go for a walk"}`
		updatedItem := ToDoItem{
			ID:   1,
			Text: "go for a walk",
			Done: false,
		}

		updateInput := ToDoItemUpdateInput{
			Text: stringPtr("go for a walk"),
		}

		req := httptest.NewRequest(http.MethodPut, "/todos/1", strings.NewReader(itemBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.Set("user_id", uint(0))

		ctx.SetPath("/todos/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")

		mockService.
			EXPECT().
			GetById(ctx.Request().Context(), item.ID).
			Return(item, nil).
			Times(1)

		mockService.
			EXPECT().
			UpdateById(ctx.Request().Context(), item.ID, updateInput).
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
		ctx.Set("user_id", uint(0))

		ctx.SetPath("/todos/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("abc")

		if assert.NoError(t, h.updateById(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorInvalidID, responseError.Message)
		}
	})

	t.Run("no updates", func(t *testing.T) {
		itemBody := `{"text": ""}`
		item := ToDoItem{
			ID:   1,
			Text: "go for a run",
			Done: false,
		}
		updateInput := ToDoItemUpdateInput{
			Text: stringPtr(""),
		}

		req := httptest.NewRequest(http.MethodPut, "/todos/1", strings.NewReader(itemBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.Set("user_id", uint(0))

		ctx.SetPath("/todos/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")

		mockService.
			EXPECT().
			GetById(ctx.Request().Context(), item.ID).
			Return(item, nil).
			Times(1)

		mockService.
			EXPECT().
			UpdateById(ctx.Request().Context(), item.ID, updateInput).
			Return(ToDoItem{}, errors.New(locale.ErrorNotFoundUpdates)).
			Times(1)

		if assert.NoError(t, h.updateById(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorNotFoundUpdates, responseError.Message)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		item := ToDoItem{
			ID:   1,
			Text: "123",
		}
		itemBody := `{"text":123, "done": 23}`

		req := httptest.NewRequest(http.MethodPut, "/todos/1", strings.NewReader(itemBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.Set("user_id", uint(0))

		ctx.SetPath("/todos/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("1")

		mockService.
			EXPECT().
			GetById(ctx.Request().Context(), item.ID).
			Return(item, nil).
			Times(1)

		if assert.NoError(t, h.updateById(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorInvalidBody, responseError.Message)
		}
	})
}

func TestHandler_DeleteById(t *testing.T) {
	os.Setenv("APP_ENV", "test")
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	e := echo.New()
	logger := zap.NewNop().Sugar()
	h := &endpointHandler{logger: logger, service: mockService, e: e}

	t.Run("delete item", func(t *testing.T) {
		userId := uint(0)
		id := 1
		item := ToDoItem{
			UserID: userId,
		}
		req := httptest.NewRequest(http.MethodDelete, "/todos/"+strconv.Itoa(id), nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.Set("user_id", userId)

		ctx.SetPath("/todos/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues(strconv.Itoa(id))

		mockService.
			EXPECT().
			GetById(ctx.Request().Context(), uint(id)).
			Return(item, nil).
			Times(1)

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
		userId := uint(0)
		req := httptest.NewRequest(http.MethodPut, "/todos/abc", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.Set("user_id", userId)

		ctx.SetPath("/todos/:id")
		ctx.SetParamNames("id")
		ctx.SetParamValues("abc")

		if assert.NoError(t, h.updateById(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var responseError localErr.ResponseError
			err := json.Unmarshal(rec.Body.Bytes(), &responseError)
			assert.NoError(t, err)

			assert.Equal(t, locale.ErrorInvalidID, responseError.Message)
		}
	})
}
