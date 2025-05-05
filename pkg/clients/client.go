package clients

import (
	"fmt"
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
)

type Client messages.Processor

func New(
	messageQueueSize int,
	_ uuid.UUID,
	rw http.ResponseWriter,
) (Client, error) {
	_, ok := rw.(http.Flusher)
	if !ok {
		return nil, errors.NewCode(ErrUnsupportedConnection)
	}

	callbacks := messages.Callbacks{
		Start:   generateStartCallback(rw),
		Message: generateMessageCallback(rw),
		Finish:  generateFinishCallback(),
	}

	return messages.NewProcessor(messageQueueSize, callbacks), nil
}

func generateStartCallback(rw http.ResponseWriter) messages.StartCallback {
	return func() error {
		// https://echo.labstack.com/docs/cookbook/sse
		rw.Header().Set("Content-Type", "text/event-stream")
		rw.Header().Set("Cache-Control", "no-cache")
		rw.Header().Set("Connection", "keep-alive")
		return nil
	}
}

func generateMessageCallback(rw http.ResponseWriter) messages.MessageCallback {
	// We verify in New that this conversion will succeed
	flusher := rw.(http.Flusher)

	return func(msg persistence.Message) error {
		e, err := fromMessage(msg)
		if err != nil {
			return errors.WrapCode(err, ErrSseStreamFailed)
		}

		// TODO: We should probably have some synchronization mechanism here.
		// Or at least check if this is already handled.
		err = e.send(rw)
		if err != nil {
			return errors.WrapCode(err, ErrSseStreamFailed)
		}

		flusher.Flush()

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
