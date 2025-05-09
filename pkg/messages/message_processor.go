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
		// TODO: Returning an error here means the processing of messages
		// will stop. Probably we should not do that and just go on
		// At this point we can't return an error to the client anyway
		if err != nil {
			return err
		}

		// TODO: Also here, we probably don't want to return the error
		err = dispatcher.BroadcastExcept(msg.ChatUser, msg)
		if err != nil {
			return err
		}

		return nil
	}
}
