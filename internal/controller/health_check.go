package controller

import (
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v4"
)

func HealthCheckEndpoints(pool db.Connection) rest.Routes {
	var out rest.Routes

	getHandler := createComponentAwareHttpHandler(healthcheck, pool)
	get := rest.NewRoute(http.MethodGet, "/healthcheck", getHandler)
	out = append(out, get)

	return out
}

func healthcheck(c echo.Context, pool db.Connection) error {
	err := pool.Ping(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, err)
	}

	return c.JSON(http.StatusOK, "OK")
}
