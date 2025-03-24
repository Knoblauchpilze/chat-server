package internal

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/server"
	"github.com/Knoblauchpilze/chat-server/internal/controller"
	"github.com/Knoblauchpilze/chat-server/internal/service"
)

type HttpServerProps struct {
	Config   Configuration
	DbConn   db.Connection
	Services service.Services
	Log      logger.Logger
}

func RunHttpServer(
	ctx context.Context, props HttpServerProps) error {
	s := server.NewWithLogger(props.Config.Server, props.Log)

	for _, route := range controller.HealthCheckEndpoints(props.DbConn) {
		if err := s.AddRoute(route); err != nil {
			return err
		}
	}

	for _, route := range controller.RoomEndpoints(props.Services.Room) {
		if err := s.AddRoute(route); err != nil {
			return err
		}
	}

	for _, route := range controller.UserEndpoints(props.Services.User) {
		if err := s.AddRoute(route); err != nil {
			return err
		}
	}

	return s.Start(ctx)
}
