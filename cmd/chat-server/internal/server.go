package internal

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/process"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/server"
	"github.com/Knoblauchpilze/chat-server/internal/controller"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"golang.org/x/sync/errgroup"
)

func RunServer(ctx context.Context, config Configuration, log logger.Logger) error {
	dbConn, err := db.New(ctx, config.Database)
	if err != nil {
		return err
	}
	defer dbConn.Close(ctx)

	repos := repositories.New(dbConn)

	manager := clients.NewManager()
	processor := messages.NewMessageProcessor(config.MessageQueueSize, manager, repos)

	services := service.Services{
		Room:    service.NewRoomService(dbConn, repos),
		User:    service.NewUserService(dbConn, repos),
		Message: service.NewMessageService(dbConn, processor, manager),
	}

	s, err := configureHttpServer(config.Server, dbConn, services, log)
	if err != nil {
		return err
	}

	group, errCtx := errgroup.WithContext(ctx)

	// Voluntarily ignore errors: they can only be produced (as of now, true)
	// when the process is not configured correctly. It is configured correctly
	// here so an error can't happen.
	waitProcessor, _ := process.StartWithSignalHandler(errCtx, processor)
	waitManager, _ := process.StartWithSignalHandler(errCtx, manager)
	waitServer, _ := process.StartWithSignalHandler(errCtx, s)

	group.Go(waitProcessor)
	group.Go(waitManager)
	group.Go(waitServer)

	return group.Wait()
}

func configureHttpServer(
	config server.Config,
	dbConn db.Connection,
	services service.Services,
	log logger.Logger,
) (server.Server, error) {
	s := server.NewWithLogger(config, log)

	for _, route := range controller.HealthCheckEndpoints(dbConn) {
		if err := s.AddRoute(route); err != nil {
			return s, err
		}
	}

	for _, route := range controller.RoomEndpoints(services.Room) {
		if err := s.AddRoute(route); err != nil {
			return s, err
		}
	}

	for _, route := range controller.UserEndpoints(services.User) {
		if err := s.AddRoute(route); err != nil {
			return s, err
		}
	}

	for _, route := range controller.MessageEndpoints(services.Message) {
		if err := s.AddRoute(route); err != nil {
			return s, err
		}
	}

	return s, nil
}
