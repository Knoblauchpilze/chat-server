package controller

import (
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func MessageEndpoints(service service.MessageService) rest.Routes {
	var out rest.Routes

	getHandler := createComponentAwareHttpHandler(subscribeToMessages, service)
	get := rest.NewRoute(http.MethodGet, "/users/:id/subscribe", getHandler)
	out = append(out, get)

	postHandler := createComponentAwareHttpHandler(postMessage, service)
	post := rest.NewRoute(http.MethodPost, "/rooms/:id/messages", postHandler)
	out = append(out, post)

	return out
}

func postMessage(c echo.Context, s service.MessageService) error {
	var messageDtoRequest communication.MessageDtoRequest
	err := c.Bind(&messageDtoRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid message syntax")
	}

	// TODO: Override the room id with the one from the path

	err = s.PostMessage(c.Request().Context(), messageDtoRequest)
	if err != nil {
		if errors.IsErrorWithCode(err, service.ErrEmptyMessage) {
			return c.JSON(http.StatusBadRequest, "Invalid empty message")
		}

		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusAccepted)
}

func subscribeToMessages(c echo.Context, s service.MessageService) error {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid id syntax")
	}

	err = s.ServeClient(c.Request().Context(), id, c.Response())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusOK)
}
