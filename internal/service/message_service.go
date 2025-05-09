package service

import (
	"context"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/process"
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MessageService interface {
	PostMessage(ctx context.Context, messageDto communication.MessageDtoRequest) error
	ServeClient(ctx context.Context, user uuid.UUID, response *echo.Response) error
}

type MessageServiceOpts struct {
	DbConn                 db.Connection
	Repos                  repositories.Repositories
	Processor              messages.Processor
	Manager                clients.Manager
	ClientMessageQueueSize int
}

type messageServiceImpl struct {
	conn     db.Connection
	roomRepo repositories.RoomRepository

	processor              messages.Processor
	manager                clients.Manager
	clientMessageQueueSize int
}

func NewMessageService(opts MessageServiceOpts) MessageService {
	return &messageServiceImpl{
		conn:                   opts.DbConn,
		roomRepo:               opts.Repos.Room,
		processor:              opts.Processor,
		manager:                opts.Manager,
		clientMessageQueueSize: opts.ClientMessageQueueSize,
	}
}

func (s *messageServiceImpl) PostMessage(
	ctx context.Context, messageDto communication.MessageDtoRequest,
) error {
	message := communication.FromMessageDtoRequest(messageDto)

	if message.Message == "" {
		return errors.NewCode(ErrEmptyMessage)
	}

	registered, err := s.roomRepo.UserInRoom(ctx, message.ChatUser, message.Room)
	if err != nil {
		return err
	}
	if !registered {
		return errors.NewCode(ErrUserNotInRoom)
	}

	s.processor.Enqueue(message)

	return nil
}

func (s *messageServiceImpl) ServeClient(
	ctx context.Context, user uuid.UUID, response *echo.Response,
) error {
	// TODO: We could add some ping/pong mechanism. This could serve as a base
	// for idle checking
	client, err := clients.New(s.clientMessageQueueSize, user, response)
	if err != nil {
		return err
	}

	if err := s.manager.OnConnect(user, client); err != nil {
		return err
	}

	done := process.SafeRunAsync(client.Start)

	select {
	case <-ctx.Done():
		err = client.Stop()
	case err = <-done:
	}

	s.manager.OnDisconnect(user)

	return err
}
