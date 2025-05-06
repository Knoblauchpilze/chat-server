package messages

import (
	"context"
	"fmt"

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

		fmt.Printf("received message: %+v (err: %v)\n", msg, err)

		fmt.Printf("broadcasting except to %s\n", msg.ChatUser)
		dispatcher.BroadcastExcept(msg.ChatUser, msg)

		return nil
	}
}
