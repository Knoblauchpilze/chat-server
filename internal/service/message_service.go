package service

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
)

type MessageService interface {
	PostMessage(ctx context.Context, messageDto communication.MessageDtoRequest) (communication.MessageDtoResponse, error)
}

type messageServiceImpl struct {
	conn db.Connection

	messageRepo repositories.MessageRepository
}

func NewMessageService(conn db.Connection, repos repositories.Repositories) MessageService {
	return &messageServiceImpl{
		conn:        conn,
		messageRepo: repos.Message,
	}
}

func (s *messageServiceImpl) PostMessage(
	ctx context.Context, messageDto communication.MessageDtoRequest,
) (communication.MessageDtoResponse, error) {
	message := communication.FromMessageDtoRequest(messageDto)

	if message.Message == "" {
		return communication.MessageDtoResponse{}, errors.NewCode(ErrEmptyMessage)
	}

	createdMessage, err := s.messageRepo.Create(ctx, message)
	if err != nil {
		return communication.MessageDtoResponse{}, err
	}

	out := communication.ToMessageDtoResponse(createdMessage)
	return out, nil
}
