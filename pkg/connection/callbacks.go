package connection

import "github.com/google/uuid"

type OnDisconnect func(id uuid.UUID)
type OnReadError func(id uuid.UUID, err error)
type OnReadData func(id uuid.UUID, data []byte)

type Callbacks struct {
	DisconnectCallbacks []OnDisconnect
	ReadErrorCallbacks  []OnReadError
	ReadDataCallbacks   []OnReadData
}

func (c Callbacks) OnDisconnect(id uuid.UUID) {
	for _, callback := range c.DisconnectCallbacks {
		callback(id)
	}
}

func (c Callbacks) OnReadError(id uuid.UUID, err error) {
	for _, callback := range c.ReadErrorCallbacks {
		callback(id, err)
	}
}

func (c Callbacks) OnReadData(id uuid.UUID, data []byte) {
	for _, callback := range c.ReadDataCallbacks {
		callback(id, data)
	}
}
