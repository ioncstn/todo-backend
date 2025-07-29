package todos

import (
	"net/http"
	"strconv"
	"time"
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
			Method:  http.MethodGet,
			Path:    "/",
			Handler: h.hello,
		},
		{
			Method:  http.MethodGet,
			Path:    "/todos",
			Handler: h.getAll,
		},
		{
			Method:  http.MethodPost,
			Path:    "/todos",
			Handler: h.create,
		},
		{
			Method:  http.MethodPut,
			Path:    "/todos/:id",
			Handler: h.updateById,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/todos/:id",
			Handler: h.deleteById,
		},
	}

	for _, endpoint := range endpoints {
		handlers.Method(h.e, endpoint.Method, endpoint.Path, endpoint.Handler)
	}
}

// @Summary Hello endpoint
// @Description This endpoint returns a simple "hello world" message
// @ID hello
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /hello [get]
func (h *endpointHandler) hello(ctx echo.Context) error {
	h.logger.Infow("testing zappy...",
		"attempt", 3,
		"backoff", time.Second,
	)

	return ctx.JSON(http.StatusOK, map[string]string{"hello": "world"})
}

// @Summary Get all todo items
// @Description This endpoint returns all todo items, with pagination
// @Tags todos
// @ID getAll
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param order query string false "Order of items in relation to Done"
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} errors.ResponseError "Internal Server Error"
// @Router /todos [get]
func (h *endpointHandler) getAll(ctx echo.Context) error {
	h.logger.Infow("reading todo item...")
	userId := ctx.Get("user_id").(uint)
	details := PaginationDetails{}

	details.Page, _ = strconv.Atoi(ctx.QueryParam("page"))
	details.Limit, _ = strconv.Atoi(ctx.QueryParam("limit"))
	details.Order = ctx.QueryParam("order")

	items, metadata, err := h.service.GetAllForUser(ctx.Request().Context(), userId, details)
	if err != nil {
		h.logger.Warn("could not read todo items", "error", err.Error())

		return ctx.JSON(http.StatusInternalServerError, e.ResponseError{Message: locale.ErrorCouldNotReadTodoItems})
	}

	return ctx.JSON(http.StatusOK, PaginatedResponse{Data: items, Meta: metadata})
}

// @Summary Create a new todo item
// @Description This endpoint creates a new todo item
// @Tags todos
// @ID create
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param todo body ToDoItem true "ToDo item to create"
// @Success 200 {object} ToDoItem
// @Failure 400 {object} errors.ResponseError "Bad Request"
// @Router /todos [post]
func (h *endpointHandler) create(ctx echo.Context) error {
	h.logger.Infow("creating todo item...")
	userId := ctx.Get("user_id").(uint)

	item := ToDoItem{}
	err := ctx.Bind(&item)
	if err != nil {
		h.logger.Warn("could not bind body to todo-item struct", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorCouldNotReadTodoItem})
	}

	item.UserID = userId

	err = h.service.Create(ctx.Request().Context(), &item)
	if err != nil {
		h.logger.Warn("could not create todo-item", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorInvalidTodoItem, Details: err.Error()})
	}
	h.logger.Infow("created todo item successfully")

	return ctx.JSON(http.StatusOK, item)
}

// @Summary Update a todo item by ID
// @Description This endpoint updates a todo item by its ID
// @Tags todos
// @ID updateById
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ToDo Item ID"
// @Param todo body ToDoItemUpdateInput true "ToDo item update data"
// @Success 200 {object} ToDoItem
// @Failure 400 {object} errors.ResponseError "Bad Request"
// @Router /todos/{id} [put]
func (h *endpointHandler) updateById(ctx echo.Context) error {
	h.logger.Infow("updating todo item...")
	userId := ctx.Get("user_id").(uint)

	id, err := handlers.GetUrlId(ctx, h.logger)
	if err != nil {
		h.logger.Warn("could not get id from url", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorInvalidID})
	}

	item, err := h.service.GetById(ctx.Request().Context(), id)
	if err != nil {
		h.logger.Error("could not get item", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorCouldNotReadTodoItem})
	}
	if item.UserID != userId {
		h.logger.Info("user tried to modify todo of other user")

		return ctx.JSON(http.StatusUnauthorized, e.ResponseError{Message: locale.ErrorNotFoundRecord})
	}

	itemInput := ToDoItemUpdateInput{}
	err = ctx.Bind(&itemInput)
	if err != nil {
		h.logger.Warn("could not bind body to todo-item struct", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorInvalidBody, Details: err.Error()})
	}

	item, err = h.service.UpdateById(ctx.Request().Context(), id, itemInput)
	if err != nil {
		h.logger.Warn("could not update todo-item", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: err.Error()})
	}

	return ctx.JSON(http.StatusOK, item)
}

// @Summary Delete a todo item by ID
// @Description This endpoint deletes a todo item by its ID
// @Tags todos
// @ID deleteById
// @Security BearerAuth
// @Produce json
// @Param id path int true "ToDo Item ID"
// @Success 200 {string} string ""
// @Failure 400 {object} errors.ResponseError "Bad Request"
// @Router /todos/{id} [delete]
func (h *endpointHandler) deleteById(ctx echo.Context) error {
	h.logger.Infow("deleting todo item...")
	userId := ctx.Get("user_id").(uint)

	id, err := handlers.GetUrlId(ctx, h.logger)
	if err != nil {
		h.logger.Warn("could not get id from url", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorInvalidID})
	}

	item, err := h.service.GetById(ctx.Request().Context(), id)
	if err != nil {
		h.logger.Error("could not get item", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorCouldNotReadTodoItem})
	}
	if item.UserID != userId {
		h.logger.Info("user tried to delete todo of other user")

		return ctx.JSON(http.StatusUnauthorized, e.ResponseError{Message: locale.ErrorNotFoundRecord})
	}

	err = h.service.DeleteById(ctx.Request().Context(), id)
	if err != nil {
		h.logger.Warn("could not delete todo-item", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: locale.ErrorCouldNotDelete, Details: err.Error()})
	}

	return ctx.JSON(http.StatusOK, "")
}
