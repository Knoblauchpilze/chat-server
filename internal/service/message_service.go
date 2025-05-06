package service

import (
	"context"
	"fmt"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/process"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MessageService interface {
	PostMessage(ctx context.Context, messageDto communication.MessageDtoRequest) error
	ServeClient(ctx context.Context, user uuid.UUID, response *echo.Response) error
}

type messageServiceImpl struct {
	conn db.Connection

	processor messages.Processor
	manager   clients.Manager
}

func NewMessageService(
	conn db.Connection,
	processor messages.Processor,
	manager clients.Manager,
) MessageService {
	return &messageServiceImpl{
		conn:      conn,
		processor: processor,
		manager:   manager,
	}
}

func (s *messageServiceImpl) PostMessage(
	ctx context.Context, messageDto communication.MessageDtoRequest,
) error {
	message := communication.FromMessageDtoRequest(messageDto)

	if message.Message == "" {
		return errors.NewCode(ErrEmptyMessage)
	}

	s.processor.Enqueue(message)

	return nil
}

func (s *messageServiceImpl) ServeClient(
	ctx context.Context, user uuid.UUID, response *echo.Response,
) error {
	// TODO: We could add some ping/pong mechanism. This could serve as a base
	// for idle checking
	// TODO: Make the message queue's size configurable
	client, err := clients.New(1, user, response)
	if err != nil {
		return err
	}

	fmt.Printf("connecting %s\n", user)
	if err := s.manager.OnConnect(user, client); err != nil {
		return err
	}

	fmt.Printf("starting %s\n", user)
	done := process.SafeRunAsync(client.Start)

	select {
	case <-ctx.Done():
		err = client.Stop()
		fmt.Printf("done, err: %v\n", err)
	case err = <-done:
		fmt.Printf("client failed, err: %v\n", err)
	}

	fmt.Printf("disconnecting: %s (err: %v)\n", user, err)
	s.manager.OnDisconnect(user)

	return err
}
