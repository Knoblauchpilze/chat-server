package tcp

import "github.com/google/uuid"

type OnDisconnect func(id uuid.UUID)
type OnReadError func(id uuid.UUID, err error)
type OnPanic func(id uuid.UUID, err error)
type OnReadData func(id uuid.UUID, data []byte)

type ConnectionCallbacks struct {
	DisconnectCallbacks []OnDisconnect
	ReadErrorCallbacks  []OnReadError
	PanicCallbacks      []OnPanic
	ReadDataCallbacks   []OnReadData
}

func (c ConnectionCallbacks) OnDisconnect(id uuid.UUID) {
	for _, callback := range c.DisconnectCallbacks {
		callback(id)
	}
}

func (c ConnectionCallbacks) OnReadError(id uuid.UUID, err error) {
	for _, callback := range c.ReadErrorCallbacks {
		callback(id, err)
	}
}

func (c ConnectionCallbacks) OnPanic(id uuid.UUID, err error) {
	for _, callback := range c.PanicCallbacks {
		callback(id, err)
	}
}

func (c ConnectionCallbacks) OnReadData(id uuid.UUID, data []byte) {
	for _, callback := range c.ReadDataCallbacks {
		callback(id, data)
	}
}
