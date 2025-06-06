package clients

import (
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
	}

	return messages.NewProcessor(messageQueueSize, callbacks), nil
}

func generateStartCallback(rw http.ResponseWriter) messages.StartCallback {
	flusher := rw.(http.Flusher)

	return func() error {
		// https://echo.labstack.com/docs/cookbook/sse
		rw.Header().Set("Content-Type", "text/event-stream")
		rw.Header().Set("Cache-Control", "no-cache")
		rw.Header().Set("Connection", "keep-alive")

		// https://github.com/tmaxmax/go-sse/blob/e429bb3114f36f65a121c25918e1131b8de6affe/session.go#L69
		flusher.Flush()
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
