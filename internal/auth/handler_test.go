package auth

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"todo-app/pkg/locale"
)

func TestHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	e := echo.New()
	loginRequestData := `{"email": "go", "password": "go"}`
	loginRequest := LoginRequest{
		Email:    "go",
		Password: "go",
	}
	loginResponse := LoginResponse{
		User: UserInfo{
			FirstName: "go",
			LastName:  "go",
			Email:     "go",
		},
	}

	t.Run("success login request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginRequestData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		logger := zap.NewNop().Sugar()
		h := &endpointHandler{logger: logger, service: mockService, e: e}

		mockService.
			EXPECT().
			Login(ctx.Request().Context(), loginRequest).
			Return(loginResponse, nil).
			Times(1)

		if assert.NoError(t, h.login(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// unmarshalling response to check if response data is ok
			var response LoginResponse
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, loginResponse.User.FirstName, response.User.FirstName)
			assert.Equal(t, loginResponse.User.LastName, response.User.LastName)
			assert.Equal(t, loginResponse.User.Email, response.User.Email)
		}

		ctrl.Finish()
	})

	t.Run("invalid login request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`asdkj[=-()&**96`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		logger := zap.NewNop().Sugar()
		h := &endpointHandler{logger: logger, service: mockService, e: e}

		if assert.NoError(t, h.login(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), locale.ErrorInvalidBody)
		}

		ctrl.Finish()
	})
}

func TestHandler_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	e := echo.New()

	t.Run("success logout request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer valid_token")

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		logger := zap.NewNop().Sugar()
		h := &endpointHandler{logger: logger, service: mockService, e: e}

		mockService.
			EXPECT().
			Logout(ctx.Request().Context(), "valid_token").
			Return(nil).
			Times(1)

		if assert.NoError(t, h.logout(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		ctrl.Finish()
	})

	t.Run("not auth header logout request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		logger := zap.NewNop().Sugar()
		h := &endpointHandler{logger: logger, service: mockService, e: e}

		if assert.NoError(t, h.logout(ctx)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Contains(t, rec.Body.String(), locale.ErrorMissingToken)
		}

		ctrl.Finish()
	})

	t.Run("auth header not valid logout request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "invalid_token")

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		logger := zap.NewNop().Sugar()
		h := &endpointHandler{logger: logger, service: mockService, e: e}

		if assert.NoError(t, h.logout(ctx)) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Contains(t, rec.Body.String(), locale.ErrorInvalidToken)
		}

		ctrl.Finish()
	})
}

func TestHandler_RefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := NewMockService(ctrl)
	e := echo.New()

	t.Run("success refresh token request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`{"refresh_token": "valid_token"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		logger := zap.NewNop().Sugar()
		h := &endpointHandler{logger: logger, service: mockService, e: e}

		loginResponse := LoginResponse{
			User: UserInfo{
				FirstName: "ez",
				LastName:  "pz",
				Email:     "go",
			},
		}

		mockService.
			EXPECT().
			RefreshToken(ctx.Request().Context(), "valid_token").
			Return(loginResponse, nil).
			Times(1)

		if assert.NoError(t, h.refresh(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("missing refresh token request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(``))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		logger := zap.NewNop().Sugar()
		h := &endpointHandler{logger: logger, service: mockService, e: e}

		if assert.NoError(t, h.refresh(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), locale.ErrorMissingRefreshToken)
		}
	})

	t.Run("invalid refresh token request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`-ajs87&*`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		logger := zap.NewNop().Sugar()
		h := &endpointHandler{logger: logger, service: mockService, e: e}

		if assert.NoError(t, h.refresh(ctx)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), locale.ErrorInvalidBody)
		}
	})
}
