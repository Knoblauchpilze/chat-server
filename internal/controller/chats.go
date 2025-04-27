package controller

import (
	"fmt"
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func MessageEndpoints(service service.RoomService) rest.Routes {
	var out rest.Routes

	getHandler := createComponentAwareHttpHandler(subscribeToMessages, service)
	get := rest.NewRoute(http.MethodGet, "/users/:id/subscribe", getHandler)
	out = append(out, get)

	return out
}

func subscribeToMessages(c echo.Context, s service.RoomService) error {
	maybeId := c.Param("id")
	id, err := uuid.Parse(maybeId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid id syntax")
	}

	// TODO: Support this
	// See: // https://echo.labstack.com/docs/cookbook/sse
	msg := fmt.Sprintf("Not implemented for: %v", id)
	return c.JSON(http.StatusInternalServerError, msg)
}
