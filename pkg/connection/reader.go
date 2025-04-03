package connection

import (
	bterr "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
)

type connectionReadResult struct {
	processed int
	available int
	timeout   bool
}

func readFromConnection(
	id uuid.UUID, conn connection, callbacks Callbacks,
) (out connectionReadResult, err error) {
	var data []byte

	out = connectionReadResult{
		processed: 0,
		available: 0,
		timeout:   false,
	}

	data, err = conn.Read()

	if bterr.IsErrorWithCode(err, ErrClientDisconnected) {
		callbacks.OnDisconnect(id)
	} else if bterr.IsErrorWithCode(err, ErrReadTimeout) {
		out.timeout = true
		err = nil
	} else if err != nil {
		callbacks.OnReadError(id, err)
	}

	out.available = len(data)

	// This block needs to be after the error has potentially been reset
	// in case of a timeout. We might still have data to read despite the
	// timeout and we want to process it.
	if err == nil && len(data) > 0 {
		out.processed = callbacks.OnReadData(id, data)
	}

	return
}
