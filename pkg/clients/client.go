package clients

import (
	"fmt"

	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
)

func New(
	messageQueueSize int,
) messages.Processor {
	return messages.NewProcessor(
		messageQueueSize,
		generateMessageCallback(),
		generateFinishCallback(),
	)
}

func generateMessageCallback() messages.MessageCallback {
	return func(msg persistence.Message) error {
		// TODO: Handle sending through SSE
		fmt.Printf("[warn] message from %v in room %v: \n\"%s\"\n", msg.ChatUser, msg.Room, msg.Message)
		return nil
	}
}

func generateFinishCallback() messages.FinishCallback {
	return func() error {
		// TODO: Handle closing through SSE
		fmt.Printf("[warn] should close SSE\n")
		return nil
	}
}
