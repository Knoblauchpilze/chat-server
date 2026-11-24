package internal

import (
	"context"
	"log/slog"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/process"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/server"
	"github.com/Knoblauchpilze/chat-server/internal/controller"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"golang.org/x/sync/errgroup"
)

func RunServer(ctx context.Context, config Configuration, log *slog.Logger) error {
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
		Registration: service.NewRegistrationService(dbConn, repos),
		Room:         service.NewRoomService(dbConn, repos),
		User:         service.NewUserService(dbConn, repos),
		Message:      service.NewMessageService(opts),
	}

	s, err := configureHttpServer(config.Server, dbConn, services, log)
	if err != nil {
		return err
	}

	listener, err := s.Bind(config.Port)
	if err != nil {
		return err
	}
	defer listener.Close()

	group, errCtx := errgroup.WithContext(ctx)

	// Voluntarily ignore errors: they can only be produced (as of now, true)
	// when the process is not configured correctly. These processes all
	// define both run and interrupt functions.
	waitProcessor, _ := process.AsyncStartWithSignalHandler(errCtx, process.Process{
		Run: processor.Start, Interrupt: processor.Stop,
	})
	waitManager, _ := process.AsyncStartWithSignalHandler(errCtx, process.Process{
		Run: manager.Start, Interrupt: manager.Stop,
	})
	serverProcess := process.NewContextProcess(func(ctx context.Context) error {
		return s.Serve(ctx, listener)
	})
	waitServer, _ := process.AsyncStartWithSignalHandler(errCtx, serverProcess)

	group.Go(waitProcessor)
	group.Go(waitManager)
	group.Go(waitServer)

	return group.Wait()
}

func configureHttpServer(
	config server.Config,
	dbConn *db.Connection,
	services service.Services,
	log *slog.Logger,
) (*server.HttpServer, error) {
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

	for _, route := range controller.RegistrationEndpoints(services.Registration) {
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
