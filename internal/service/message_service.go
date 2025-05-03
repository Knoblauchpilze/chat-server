package service

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
)

type MessageService interface {
	PostMessage(ctx context.Context, messageDto communication.MessageDtoRequest) error
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
