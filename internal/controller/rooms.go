package controller

import (
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RoomEndpoints(service service.RoomService) Routes {
	var out Routes

	postHandler := createComponentAwareHttpHandler(createRoom, service)
	post := rest.NewRoute(http.MethodPost, "/rooms", postHandler)
	out = append(out, post)

	getHandler := createComponentAwareHttpHandler(getRoom, service)
	get := rest.NewRoute(http.MethodGet, "/rooms/:id", getHandler)
	out = append(out, get)

	listHandler := createComponentAwareHttpHandler(listRoom, service)
	list := rest.NewRoute(http.MethodGet, "/rooms", listHandler)
	out = append(out, list)

	listUserForRoomHandler := createComponentAwareHttpHandler(listUserForRoom, service)
	listUserForRoom := rest.NewRoute(http.MethodGet, "/rooms/:id/users", listUserForRoomHandler)
	out = append(out, listUserForRoom)

	listMessageForRoomHandler := createComponentAwareHttpHandler(listMessageForRoom, service)
	listMessageForRoom := rest.NewRoute(http.MethodGet, "/rooms/:id/messages", listMessageForRoomHandler)
	out = append(out, listMessageForRoom)

	deleteHandler := createComponentAwareHttpHandler(deleteRoom, service)
	delete := rest.NewRoute(http.MethodDelete, "/rooms/:id", deleteHandler)
	out = append(out, delete)

	return out
}

func createRoom(c *gin.Context, s service.RoomService) {
	var roomDtoRequest communication.RoomDtoRequest
	err := c.ShouldBindJSON(&roomDtoRequest)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid room syntax")
		return
	}

	out, err := s.Create(c.Request.Context(), roomDtoRequest)
	if err != nil {
		if codedErr, ok := errors.AsErrorWithCode(err); ok && codedErr.Code == service.ErrInvalidName {
			c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid room name")
			return
		}
		if dbErr, ok := db.AsDatabaseError(err); ok && dbErr.Code == db.ErrUniqueConstraintViolation {
			c.AbortWithStatusJSON(http.StatusConflict, "Room name already in use")
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, out)
}

func getRoom(c *gin.Context, s service.RoomService) {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid id syntax")
		return
	}

	out, err := s.Get(c.Request.Context(), id)
	if err != nil {
		if err == db.ErrNoMatchingRows {
			c.AbortWithStatusJSON(http.StatusNotFound, "No such room")
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, out)
}

func listRoom(c *gin.Context, s service.RoomService) {
	rooms, err := s.List(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	out, err := marshalNilToEmptySlice(rooms)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "application/json", out)
}

func listUserForRoom(c *gin.Context, s service.RoomService) {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid id syntax")
		return
	}

	users, err := s.ListUserForRoom(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	out, err := marshalNilToEmptySlice(users)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "application/json", out)
}

func listMessageForRoom(c *gin.Context, s service.RoomService) {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid id syntax")
		return
	}

	messages, err := s.ListMessageForRoom(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	out, err := marshalNilToEmptySlice(messages)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "application/json", out)
}

func deleteRoom(c *gin.Context, s service.RoomService) {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid id syntax")
		return
	}

	err = s.Delete(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}
