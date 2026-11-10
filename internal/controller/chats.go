package controller

import (
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func MessageEndpoints(service service.MessageService) Routes {
	var out Routes

	getHandler := createComponentAwareHttpHandler(subscribeToMessages, service)
	get := rest.NewRawRoute(http.MethodGet, "/users/:id/subscribe", getHandler)
	out = append(out, get)

	postHandler := createComponentAwareHttpHandler(postMessage, service)
	post := rest.NewRoute(http.MethodPost, "/rooms/:id/messages", postHandler)
	out = append(out, post)

	return out
}

func postMessage(c *gin.Context, s service.MessageService) {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid id syntax")
		return
	}

	var messageDtoRequest communication.MessageDtoRequest
	err = c.ShouldBindJSON(&messageDtoRequest)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid message syntax")
		return
	}

	messageDtoRequest.Room = id

	err = s.PostMessage(c.Request.Context(), messageDtoRequest)
	if err != nil {
		if codedErr, ok := errors.AsErrorWithCode(err); ok && codedErr.Code == service.ErrEmptyMessage {
			c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid empty message")
			return
		} else if codedErr, ok := errors.AsErrorWithCode(err); ok && codedErr.Code == service.ErrUserNotInRoom {
			c.AbortWithStatusJSON(http.StatusBadRequest, "User is not registered in the room")
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusAccepted)
}

func subscribeToMessages(c *gin.Context, s service.MessageService) {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid id syntax")
		return
	}

	// TODO: We could pass on the logger taken from the context
	err = s.ServeClient(c.Request.Context(), id, c.Writer)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
}
