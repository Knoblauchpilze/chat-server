package internal

import (
	"context"
	"fmt"
	"sync"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/chat-server/internal/service"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
)

func RunServer(ctx context.Context, config Configuration, log logger.Logger) error {
	dbConn, err := db.New(ctx, config.Database)
	if err != nil {
		return err
	}
	defer dbConn.Close(ctx)

	repos := repositories.New(dbConn)
	// TODO: Correctly setup the message processor
	// TODO: Correctly setup the client manager
	services := service.New(config.ConnectTimeout, dbConn, repos, nil, nil, log)

	var tcpErr error
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				tcpErr = errors.New(fmt.Sprintf("TCP server panicked: %v", err))
			}
		}()
		tcpErr = RunTcpServer(ctx, config, services, log)
	}()

	props := HttpServerProps{
		Config:   config,
		DbConn:   dbConn,
		Services: services,
		Log:      log,
	}
	httpErr := RunHttpServer(ctx, props)

	wg.Wait()

	if tcpErr != nil {
		return tcpErr
	}
	return httpErr
}
