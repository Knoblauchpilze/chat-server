package connection

import "github.com/google/uuid"

type OnDisconnect func(id uuid.UUID)
type OnReadError func(id uuid.UUID, err error)

// OnReadData is a callback that is called when data is read from a connection.
// The return value indicates how many bytes were processed.
type OnReadData func(id uuid.UUID, data []byte) int

type Callbacks struct {
	DisconnectCallback OnDisconnect
	ReadErrorCallback  OnReadError
	ReadDataCallback   OnReadData
}

func (c Callbacks) OnDisconnect(id uuid.UUID) {
	if c.DisconnectCallback != nil {
		c.DisconnectCallback(id)
	}
}

func (c Callbacks) OnReadError(id uuid.UUID, err error) {
	if c.ReadErrorCallback != nil {
		c.ReadErrorCallback(id, err)
	}
}

func (c Callbacks) OnReadData(id uuid.UUID, data []byte) int {
	if c.ReadDataCallback == nil {
		return 0
	}
	return c.ReadDataCallback(id, data)
}
