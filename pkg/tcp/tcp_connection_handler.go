package tcp

import (
	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
	"github.com/labstack/echo/v4"
)

func NewHandler(log logger.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Infof("Received connection")
		return nil
	}
}
