package tcp

import (
	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
	"github.com/labstack/echo/v4"
)

type tcpConnectionHandler struct {
	log logger.Logger
}

func NewHandler(log logger.Logger) echo.HandlerFunc {
	h := tcpConnectionHandler{
		log: log,
	}

	return func(c echo.Context) error {
		return h.HandleRequest(c)
	}
}

func (h *tcpConnectionHandler) HandleRequest(c echo.Context) error {
	// https://github.com/venilnoronha/tcp-echo-server/blob/master/main.go#L43
	h.log.Infof("Received connection")
	return nil
}
