package users

import (
	"fmt"
	"net/http"
	e "todo-app/pkg/errors"
	"todo-app/pkg/handlers"
	"todo-app/pkg/locale"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type endpointHandler struct {
	logger  *zap.SugaredLogger
	service Service
	e       *echo.Echo
}

func GetEndpointHandler(
	logger *zap.SugaredLogger,
	service Service,
	e *echo.Echo,
) handlers.EndpointHandler {
	return &endpointHandler{
		logger:  logger,
		service: service,
		e:       e,
	}
}

func (h *endpointHandler) AddEndpoints() {
	var endpoints = []handlers.Endpoint{
		{
			Method:  http.MethodPost,
			Path:    "/user",
			Handler: h.create,
		},
		{
			Method:  http.MethodPut,
			Path:    "/user/:id",
			Handler: h.update,
		},
		{
			Method:  http.MethodGet,
			Path:    "/verify-email",
			Handler: h.verifyEmail,
		},
	}

	for _, endpoint := range endpoints {
		handlers.Method(h.e, endpoint.Method, endpoint.Path, endpoint.Handler)
	}
}

// @Summary Create a new user
// @Description This endpoint creates a new user and sends an email verification link
// @Tags users
// @ID create-user
// @Accept json
// @Produce json
// @Param user body User true "User details"
// @Success 200 {object} User
// @Failure 400 {object} errors.ResponseError "Invalid user data"
// @Failure 500 {object} errors.ResponseError "Internal server error"
// @Router /user [post]
func (h *endpointHandler) create(ctx echo.Context) error {
	h.logger.Infow("creating user...")

	user := User{}
	err := ctx.Bind(&user)
	if err != nil {
		h.logger.Warn("could not bind body to user struct", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorCouldNotReadUser})
	}

	err = h.service.Create(ctx.Request().Context(), &user)
	if err != nil {
		h.logger.Warn("could not create user", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorInvalidUser, Details: err.Error()})
	}
	h.logger.Infow("created user successfully")

	user.Password = ""
	return ctx.JSON(http.StatusOK, user)
}

// @Summary Verify email address
// @Description This endpoint verifies a user's email address using the verification token
// @Tags users
// @ID verify-email
// @Produce html
// @Param token query string true "Email verification token"
// @Success 200 {string} string "Email verified successfully"
// @Failure 400 {string} string "Invalid or expired token"
// @Failure 500 {string} string "Internal server error"
// @Router /verify-email [get]
func (h *endpointHandler) verifyEmail(ctx echo.Context) error {
	h.logger.Infow("verifying email...")

	token := ctx.QueryParam("token")
	if token == "" {
		h.logger.Warn("verification token not provided")
		return ctx.HTML(http.StatusBadRequest, getVerificationHTML("Error", "Verification token is required", false))
	}

	err := h.service.VerifyEmail(ctx.Request().Context(), token)
	if err != nil {
		h.logger.Warn("could not verify email", "error", err.Error())
		return ctx.HTML(http.StatusBadRequest, getVerificationHTML("Verification Failed", err.Error(), false))
	}

	h.logger.Infow("email verified successfully")
	return ctx.HTML(http.StatusOK, getVerificationHTML("Email Verified!", "Email verified successfully. You can now use your account.", true))
}

func getVerificationHTML(title, message string, success bool) string {
	color := "#dc3545" // Red for error
	if success {
		color = "#28a745" // Green for success
	}

	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>%s</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
				.container { max-width: 600px; margin: 0 auto; background-color: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center; }
				.icon { font-size: 48px; margin-bottom: 20px; color: %s; }
				h1 { color: %s; margin-bottom: 10px; }
				p { color: #666; line-height: 1.6; }
				.btn { display: inline-block; margin-top: 20px; padding: 10px 20px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="icon">%s</div>
				<h1>%s</h1>
				<p>%s</p>
				<a href="/" class="btn">Go to Home</a>
			</div>
		</body>
		</html>
	`, title, color, color, func() string {
		if success {
			return "✅"
		}
		return "❌"
	}(), title, message)
}

// @Summary Update an existing user
// @Description This endpoint updates an existing user based on the provided ID and returns the updated user details
// @Tags users
// @ID update-user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body User true "Updated user details"
// @Success 200 {object} User
// @Failure 400 {object} errors.ResponseError "Invalid user data or ID"
// @Failure 500 {object} errors.ResponseError "Internal server error"
// @Router /update/{id} [put]
func (h *endpointHandler) update(ctx echo.Context) error {
	h.logger.Infow("updating user...")

	user := User{}
	err := ctx.Bind(&user)
	if err != nil {
		h.logger.Warn("could not bind body to user struct", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorCouldNotReadUser})
	}

	id, err := handlers.GetUrlId(ctx, h.logger)
	if err != nil {
		h.logger.Warn("could not get id from url", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorInvalidID})
	}
	user.ID = id

	user, err = h.service.Update(ctx.Request().Context(), &user)
	if err != nil {
		h.logger.Warn("could not update user", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: err.Error()})
	}
	h.logger.Infow("updated user successfully")

	return ctx.JSON(http.StatusOK, user)
}
