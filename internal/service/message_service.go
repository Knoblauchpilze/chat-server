package service

import (
	"context"

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
}

func NewMessageService(conn db.Connection, processor messages.Processor) MessageService {
	return &messageServiceImpl{
		conn:      conn,
		processor: processor,
	}
}

func (s *messageServiceImpl) PostMessage(
	ctx context.Context, messageDto communication.MessageDtoRequest,
) error {
	message := communication.FromMessageDtoRequest(messageDto)

	if message.Message == "" {
		return errors.NewCode(ErrEmptyMessage)
	}

	err := s.processor.Enqueue(message)
	if err != nil {
		return err
	}

	return nil
}

func (s *messageServiceImpl) ServeClient(
	ctx context.Context, user uuid.UUID, response *echo.Response,
) error {
	// TODO: We could add some ping/pong mechanism. This could serve as a base
	// for idle checking
	// TODO: Make the message queue's size configurable
	c, err := clients.New(1, user, response)
	if err != nil {
		return err
	}

	done := process.SafeRunAsync(c.Start)

	select {
	case <-ctx.Done():
	case err = <-done:
	}

	return err
}
