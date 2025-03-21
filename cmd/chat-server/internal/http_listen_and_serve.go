package internal

import (
	"context"
	"os"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/server"
	"github.com/Knoblauchpilze/chat-server/internal/controller"
)

func httpListenAndServe(
	ctx context.Context, config Configuration, conn db.Connection, log logger.Logger) error {
	s := server.NewWithLogger(config.Server, log)

	for _, route := range controller.HealthCheckEndpoints(conn) {
		if err := s.AddRoute(route); err != nil {
			log.Errorf("Failed to register route %v: %v", route.Path(), err)
			os.Exit(1)
		}
	}

	return s.Start(ctx)
}
