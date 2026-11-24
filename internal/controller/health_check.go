package controller

import (
	"context"
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-gonic/gin"
)

type healthCheckConnection interface {
	Ping(context.Context) error
}

func HealthCheckEndpoints(pool healthCheckConnection) Routes {
	var out Routes

	getHandler := createComponentAwareHttpHandler(healthcheck, pool)
	get := rest.NewRoute(http.MethodGet, "/healthcheck", getHandler)
	out = append(out, get)

	return out
}

func healthcheck(c *gin.Context, pool healthCheckConnection) {
	err := pool.Ping(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, err)
		return
	}

	c.JSON(http.StatusOK, "OK")
}
