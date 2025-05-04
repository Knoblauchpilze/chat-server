package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
)

// https://echo.labstack.com/docs/cookbook/sse#event-structure-and-marshal-method
type sseEvent struct {
	Id      []byte
	Data    []byte
	Event   []byte
	Retry   []byte
	Comment []byte
}

func fromMessage(msg persistence.Message) (sseEvent, error) {
	out := communication.ToMessageDtoResponse(msg)

	data, err := json.Marshal(out)
	if err != nil {
		return sseEvent{}, err
	}

	e := sseEvent{
		Id: []byte(msg.Id.String()),
		// TODO: Verify the SSE configuration and see if it make sense
		Data: data,
	}

	return e, nil
}

func (e sseEvent) send(rw http.ResponseWriter) error {
	if len(e.Data) == 0 && len(e.Comment) == 0 {
		return nil
	}

	if len(e.Data) > 0 {
		if err := e.writeData(rw); err != nil {
			return err
		}
	}

	return e.writeComment(rw)
}

func (e sseEvent) writeData(rw http.ResponseWriter) error {
	out := fmt.Sprintf("id: %s\n", e.Id)
	if err := writeToResponseWriter([]byte(out), rw); err != nil {
		return errors.WrapCode(err, ErrSseStreamFailed)
	}

	for _, line := range bytes.Split(e.Data, []byte("\n")) {
		out := fmt.Sprintf("data: %s\n", line)
		if err := writeToResponseWriter([]byte(out), rw); err != nil {
			return errors.WrapCode(err, ErrSseStreamFailed)
		}
	}

	if len(e.Event) > 0 {
		out := fmt.Sprintf("event: %s\n", e.Event)
		if err := writeToResponseWriter([]byte(out), rw); err != nil {
			return errors.WrapCode(err, ErrSseStreamFailed)
		}
	}

	if len(e.Retry) > 0 {
		out := fmt.Sprintf("retry: %s\n", e.Retry)
		if err := writeToResponseWriter([]byte(out), rw); err != nil {
			return errors.WrapCode(err, ErrSseStreamFailed)
		}
	}

	return nil
}

func (e sseEvent) writeComment(rw http.ResponseWriter) error {
	if len(e.Comment) == 0 {
		return nil
	}

	out := fmt.Sprintf(": %s\n", e.Comment)
	return writeToResponseWriter([]byte(out), rw)
}

func writeToResponseWriter(data []byte, rw http.ResponseWriter) error {
	n, err := rw.Write(data)

	if n != len(data) {
		return errors.NewCode(ErrPartialSseWrite)
	}

	return err
}
