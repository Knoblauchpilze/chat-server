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
		// We should solve this by verifying in the message_service that
		// the user is allowed to send the message
		// More generally we should probably not allow users to access to
		// rooms that they did not join
		if err != nil {
			return err
		}

		// TODO: We probably don't want to broadcast but have a SendToMultiple
		// with a list of people to handle the room messages.
		dispatcher.BroadcastExcept(msg.ChatUser, msg)

		return nil
	}
}
