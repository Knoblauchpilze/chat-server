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
	callbacks := Callbacks{
		Message: generateMessageCallback(dispatcher, repos.Message),
	}

	return NewProcessor(messageQueueSize, callbacks)
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
