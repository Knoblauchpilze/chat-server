package service

import (
	"context"
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/process"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/google/uuid"
)

type MessageService interface {
	PostMessage(ctx context.Context, messageDto communication.MessageDtoRequest) error
	ServeClient(ctx context.Context, user uuid.UUID, response http.ResponseWriter) error
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
	ctx context.Context, user uuid.UUID, response http.ResponseWriter,
) error {
	// TODO: Make the message queue's size configurable
	c := clients.New(1, user, response)
	return process.SafeRunSync(c.Start)
}
