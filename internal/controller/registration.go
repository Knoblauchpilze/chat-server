package controller

import (
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegistrationEndpoints(service service.RegistrationService) Routes {
	var out Routes

	postHandler := createComponentAwareHttpHandler(addUserInRoom, service)
	post := rest.NewRoute(http.MethodPost, "/rooms/:id/users", postHandler)
	out = append(out, post)

	deleteHandler := createComponentAwareHttpHandler(deleteUserFromRoom, service)
	delete := rest.NewRoute(http.MethodDelete, "/rooms/:id/users/:user", deleteHandler)
	out = append(out, delete)

	return out
}

func addUserInRoom(c *gin.Context, s service.RegistrationService) {
	maybeId := c.Param("id")
	room, err := uuid.Parse(maybeId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid id syntax")
		return
	}

	var registrationDtoRequest communication.RoomRegistrationDtoRequest
	err = c.ShouldBindJSON(&registrationDtoRequest)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid registration syntax")
		return
	}

	err = s.RegisterUserInRoom(
		c.Request.Context(), registrationDtoRequest.User, room,
	)
	if err != nil {
		if codedErr, ok := errors.AsErrorWithCode(err); ok && codedErr.Code == repositories.ErrNoSuchUser {
			c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid user id")
			return
		}
		if codedErr, ok := errors.AsErrorWithCode(err); ok && codedErr.Code == repositories.ErrNoSuchRoom {
			c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid room id")
			return
		}
		if codedErr, ok := errors.AsErrorWithCode(err); ok && codedErr.Code == repositories.ErrUserAlreadyRegisteredInRoom {
			c.AbortWithStatusJSON(http.StatusConflict, "User already registered in room")
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func deleteUserFromRoom(c *gin.Context, s service.RegistrationService) {
	maybeId := c.Param("id")
	room, err := uuid.Parse(maybeId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid id syntax")
		return
	}

	maybeId = c.Param("user")
	user, err := uuid.Parse(maybeId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid id syntax")
		return
	}

	err = s.UnregisterUserInRoom(c.Request.Context(), user, room)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}
