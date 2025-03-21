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

func RoomEndpoints(service service.RoomService) rest.Routes {
	var out rest.Routes

	postHandler := createComponentAwareHttpHandler(createRoom, service)
	post := rest.NewRoute(http.MethodPost, "", postHandler)
	out = append(out, post)

	getHandler := createComponentAwareHttpHandler(getRoom, service)
	get := rest.NewRoute(http.MethodGet, ":id", getHandler)
	out = append(out, get)

	deleteHandler := createComponentAwareHttpHandler(deleteRoom, service)
	delete := rest.NewRoute(http.MethodDelete, ":id", deleteHandler)
	out = append(out, delete)

	return out
}

func createRoom(c echo.Context, s service.RoomService) error {
	var roomDtoRequest communication.RoomDtoRequest
	err := c.Bind(&roomDtoRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid room syntax")
	}

	out, err := s.Create(c.Request().Context(), roomDtoRequest)
	if err != nil {
		if errors.IsErrorWithCode(err, service.InvalidName) {
			return c.JSON(http.StatusBadRequest, "Invalid room name")
		}
		if errors.IsErrorWithCode(err, pgx.UniqueConstraintViolation) {
			return c.JSON(http.StatusConflict, "Room name already in use")
		}

		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, out)
}

func getRoom(c echo.Context, s service.RoomService) error {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid id syntax")
	}

	out, err := s.Get(c.Request().Context(), id)
	if err != nil {
		if errors.IsErrorWithCode(err, db.NoMatchingRows) {
			return c.JSON(http.StatusNotFound, "No such room")
		}

		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, out)
}

func deleteRoom(c echo.Context, s service.RoomService) error {
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
