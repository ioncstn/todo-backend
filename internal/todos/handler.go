package todos

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
	e "todo-app/internal/errors"
)

type EndpointHandler interface {
	AddEndpoints()
}

type endpointHandler struct {
	logger    *zap.SugaredLogger
	service   Service
	e         *echo.Echo
	validator *validator.Validate
}

type endpoint struct {
	Method  string
	Path    string
	Handler echo.HandlerFunc
}

func GetEndpointHandler(
	logger *zap.SugaredLogger,
	service Service,
	e *echo.Echo,
	validator *validator.Validate,
) EndpointHandler {
	return &endpointHandler{
		logger:    logger,
		service:   service,
		e:         e,
		validator: validator,
	}
}

func (h *endpointHandler) AddEndpoints() {
	var endpoints = []endpoint{
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
		method(h.e, endpoint.Method, endpoint.Path, endpoint.Handler)
	}
}

func method(e *echo.Echo, method string, path string, handler echo.HandlerFunc) {
	switch method {
	case "GET":
		e.GET(path, handler)
	case "POST":
		e.POST(path, handler)
	case "PUT":
		e.PUT(path, handler)
	case "DELETE":
		e.DELETE(path, handler)
	default:
		panic("unsupported method: " + method)
	}
}

func (h *endpointHandler) getUrlId(ctx echo.Context) (uint, error) {
	idString := ctx.Param("id")
	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		h.logger.Warn("could not parse id", "error", err.Error())

		return 0, err
	}

	return uint(id), nil
}

func (h *endpointHandler) hello(ctx echo.Context) error {
	h.logger.Infow("testing zappy...",
		"attempt", 3,
		"backoff", time.Second,
	)

	return ctx.JSON(http.StatusOK, map[string]string{"hello": "world"})
}

func (h *endpointHandler) getAll(ctx echo.Context) error {
	h.logger.Infow("reading todo item...")

	items, err := h.service.GetAll(ctx.Request().Context())
	if err != nil {
		h.logger.Warn("could not read todo items", "error", err.Error())

		return ctx.JSON(http.StatusInternalServerError, e.ResponseError{Message: "could not read todo-items"})
	}

	return ctx.JSON(http.StatusOK, items)
}

func (h *endpointHandler) create(ctx echo.Context) error {
	h.logger.Infow("creating todo item...")

	item := ToDoItem{}
	err := ctx.Bind(&item)
	if err != nil {
		h.logger.Warn("could not bind body to todo-item struct", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: "could not read todo-item"})
	}

	err = h.service.Create(ctx.Request().Context(), &item)
	if err != nil {
		h.logger.Warn("could not create todo-item", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: "could not create todo-item"})
	}
	h.logger.Infow("created todo item successfully")

	return ctx.JSON(http.StatusOK, item)
}

func (h *endpointHandler) updateById(ctx echo.Context) error {
	h.logger.Infow("updating todo item...")

	itemInput := ToDoItemUpdateInput{}
	err := ctx.Bind(&itemInput)
	if err != nil {
		h.logger.Warn("could not bind body to todo-item struct", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: "could not read todo-item"})
	}

	id, err := h.getUrlId(ctx)
	if err != nil {
		h.logger.Warn("could not get id from url", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: "invalid id"})
	}

	item, err := h.service.UpdateById(ctx.Request().Context(), id, itemInput)
	if err != nil {
		h.logger.Warn("could not update todo-item", "error", err.Error())

		if errors.Is(err, e.ResponseErrorNotFound) {
			return ctx.JSON(http.StatusNotFound, e.ResponseErrorNotFound)
		}

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: err.Error()})
	}

	return ctx.JSON(http.StatusOK, item)
}

func (h *endpointHandler) deleteById(ctx echo.Context) error {
	h.logger.Infow("deleting todo item...")

	id, err := h.getUrlId(ctx)
	if err != nil {
		h.logger.Warn("could not get id from url", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: "invalid id"})
	}

	err = h.service.DeleteById(ctx.Request().Context(), id)
	if err != nil {
		h.logger.Warn("could not delete todo-item", "error", err.Error())

		return ctx.JSON(http.StatusBadRequest, e.ResponseError{Message: "could not delete todo-item"})
	}

	return ctx.JSON(http.StatusOK, "")
}
