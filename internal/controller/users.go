package controller

import (
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/db/pgx"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func UserEndpoints(service service.UserService) rest.Routes {
	var out rest.Routes

	postHandler := createComponentAwareHttpHandler(createUser, service)
	post := rest.NewRoute(http.MethodPost, "/users", postHandler)
	out = append(out, post)

	getHandler := createComponentAwareHttpHandler(getUser, service)
	get := rest.NewRoute(http.MethodGet, "/users/:id", getHandler)
	out = append(out, get)

	listHandler := createComponentAwareHttpHandler(listUsers, service)
	list := rest.NewRoute(http.MethodGet, "/users", listHandler)
	out = append(out, list)

	listForUserHandler := createComponentAwareHttpHandler(listForUser, service)
	listForUser := rest.NewRoute(http.MethodGet, "/users/:id/rooms", listForUserHandler)
	out = append(out, listForUser)

	deleteHandler := createComponentAwareHttpHandler(deleteUser, service)
	delete := rest.NewRoute(http.MethodDelete, "/users/:id", deleteHandler)
	out = append(out, delete)

	return out
}

func createUser(c echo.Context, s service.UserService) error {
	var userDtoRequest communication.UserDtoRequest
	err := c.Bind(&userDtoRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user syntax")
	}

	out, err := s.Create(c.Request().Context(), userDtoRequest)
	if err != nil {
		if errors.IsErrorWithCode(err, service.ErrInvalidName) {
			return c.JSON(http.StatusBadRequest, "Invalid user name")
		}
		if errors.IsErrorWithCode(err, pgx.UniqueConstraintViolation) {
			return c.JSON(http.StatusConflict, "User name already in use")
		}

		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, out)
}

func getUser(c echo.Context, s service.UserService) error {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid id syntax")
	}

	out, err := s.Get(c.Request().Context(), id)
	if err != nil {
		if errors.IsErrorWithCode(err, db.NoMatchingRows) {
			return c.JSON(http.StatusNotFound, "No such user")
		}

		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, out)
}

func listUsers(c echo.Context, s service.UserService) error {
	const userNameKey = "name"
	maybeName := c.QueryParam(userNameKey)
	exists := (maybeName != "")

	var users []communication.UserDtoResponse
	var err error

	if exists {
		var user communication.UserDtoResponse
		user, err = s.GetByName(c.Request().Context(), maybeName)
		if err == nil {
			users = append(users, user)
		}
	} else {
		return c.JSON(
			http.StatusBadRequest,
			"Please provide a user name as filtering parameter",
		)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	out, err := marshalNilToEmptySlice(users)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSONBlob(http.StatusOK, out)
}

func listForUser(c echo.Context, s service.UserService) error {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid id syntax")
	}

	rooms, err := s.ListForUser(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	out, err := marshalNilToEmptySlice(rooms)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSONBlob(http.StatusOK, out)
}

func deleteUser(c echo.Context, s service.UserService) error {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid id syntax")
	}

	err = s.Delete(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusNoContent)
}
