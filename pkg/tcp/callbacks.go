package tcp

import "github.com/google/uuid"

type OnDisconnect func(id uuid.UUID)
type OnReadError func(id uuid.UUID, err error)
type OnPanic func(id uuid.UUID, err error)
type OnReadData func(id uuid.UUID, data []byte)

type ConnectionCallbacks struct {
	DisconnectCallback OnDisconnect
	ReadErrorCallback  OnReadError
	PanicCallback      OnPanic
	ReadDataCallback   OnReadData
}

func (c ConnectionCallbacks) OnDisconnect(id uuid.UUID) {
	if c.DisconnectCallback == nil {
		return
	}

	c.DisconnectCallback(id)
}

func (c ConnectionCallbacks) OnReadError(id uuid.UUID, err error) {
	if c.ReadErrorCallback == nil {
		return
	}

	c.ReadErrorCallback(id, err)
}

func (c ConnectionCallbacks) OnPanic(id uuid.UUID, err error) {
	if c.PanicCallback == nil {
		return
	}

	c.PanicCallback(id, err)
}

func (c ConnectionCallbacks) OnReadData(id uuid.UUID, data []byte) {
	if c.ReadDataCallback == nil {
		return
	}

	c.ReadDataCallback(id, data)
}
