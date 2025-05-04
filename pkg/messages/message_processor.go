package messages

import (
	"context"

	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/Knoblauchpilze/chat-server/pkg/repositories"
)

func NewMessageProcessor(
	messageQueueSize int,
	dispatcher Dispatcher,
	repos repositories.Repositories,
) Processor {
	return newProcessor(
		messageQueueSize,
		generateMessageCallback(dispatcher, repos.Message),
		nil,
	)
}

func generateMessageCallback(
	dispatcher Dispatcher,
	messageRepo repositories.MessageRepository,
) MessageCallback {
	return func(msg persistence.Message) error {
		_, err := messageRepo.Create(context.Background(), msg)
		if err != nil {
			return err
		}

		out := NewRoomMessage(msg.ChatUser, msg.Room, msg.Message)
		dispatcher.BroadcastExcept(msg.ChatUser, out)

		return nil
	}
}
