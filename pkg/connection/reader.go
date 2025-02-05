package connection

import (
	bterr "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
)

func readFromConnection(id uuid.UUID, conn connection, callbacks Callbacks) (timeout bool, err error) {
	var data []byte

	data, err = conn.Read()

	if err == nil {
		callbacks.OnReadData(id, data)
	} else if bterr.IsErrorWithCode(err, ErrClientDisconnected) {
		callbacks.OnDisconnect(id)
	} else if bterr.IsErrorWithCode(err, ErrReadTimeout) {
		timeout = true
		err = nil
	} else {
		callbacks.OnReadError(id, err)
	}

	return
}
