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

	manager := clients.NewManager(repos)
	processor := messages.NewMessageProcessor(config.MessageQueueSize, manager, repos)

	opts := service.MessageServiceOpts{
		DbConn:                 dbConn,
		Repos:                  repos,
		Processor:              processor,
		Manager:                manager,
		ClientMessageQueueSize: config.ClientMessageQueueSize,
	}

	services := service.Services{
		Room:    service.NewRoomService(dbConn, repos),
		User:    service.NewUserService(dbConn, repos),
		Message: service.NewMessageService(opts),
	}

	s, err := configureHttpServer(config.Server, dbConn, services, log)
	if err != nil {
		return err
	}

	group, errCtx := errgroup.WithContext(ctx)

	// Voluntarily ignore errors: they can only be produced (as of now, true)
	// when the process is not configured correctly. However when using the
	// `StartWithSignalHandler`, the process is build directly from the
	// runnable given in input so there's no risk of having it misconfigured.
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
