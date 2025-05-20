package controller

import (
	"fmt"
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func RegistrationEndpoints(service service.RegistrationService) rest.Routes {
	var out rest.Routes

	postHandler := createComponentAwareHttpHandler(addUserInRoom, service)
	post := rest.NewRoute(http.MethodPost, "/rooms/:id/users", postHandler)
	out = append(out, post)

	deleteHandler := createComponentAwareHttpHandler(deleteUserFromRoom, service)
	delete := rest.NewRoute(http.MethodDelete, "/rooms/:room/users/:user", deleteHandler)
	out = append(out, delete)

	return out
}

func addUserInRoom(c echo.Context, s service.RegistrationService) error {
	maybeId := c.Param("id")
	room, err := uuid.Parse(maybeId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid id syntax")
	}

	var registrationDtoRequest communication.RoomRegistrationDtoRequest
	err = c.Bind(&registrationDtoRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid registration syntax")
	}

	err = s.RegisterUserInRoom(
		c.Request().Context(), registrationDtoRequest.User, room,
	)
	if err != nil {
		if errors.IsErrorWithCode(err, repositories.ErrNoSuchUser) {
			return c.JSON(http.StatusBadRequest, "Invalid user id")
		}
		if errors.IsErrorWithCode(err, repositories.ErrNoSuchRoom) {
			return c.JSON(http.StatusBadRequest, "Invalid room id")
		}
		if errors.IsErrorWithCode(err, repositories.ErrUserAlreadyRegisteredInRoom) {
			return c.JSON(http.StatusConflict, "User already registered in room")
		}

		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func deleteUserFromRoom(c echo.Context, s service.RegistrationService) error {
	maybeId := c.Param("room")
	room, err := uuid.Parse(maybeId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid id syntax")
	}

	maybeId = c.Param("user")
	user, err := uuid.Parse(maybeId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid id syntax")
	}

	fmt.Printf("deleting user %s from room %s\n", user, room)

	err = s.UnregisterUserInRoom(c.Request().Context(), user, room)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusNoContent)
}
