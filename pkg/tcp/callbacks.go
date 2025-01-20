package tcp

type OnDisconnect func()
type OnReadError func(err error)
type OnPanic func(err error)
type OnReadData func(data []byte)

type ConnectionCallbacks struct {
	DisconnectCallback OnDisconnect
	ReadErrorCallback  OnReadError
	PanicCallback      OnPanic
	ReadDataCallback   OnReadData
}

func (c ConnectionCallbacks) OnDisconnect() {
	if c.DisconnectCallback == nil {
		return
	}

	c.DisconnectCallback()
}

func (c ConnectionCallbacks) OnReadError(err error) {
	if c.ReadErrorCallback == nil {
		return
	}

	c.ReadErrorCallback(err)
}

func (c ConnectionCallbacks) OnPanic(err error) {
	if c.PanicCallback == nil {
		return
	}

	c.PanicCallback(err)
}

func (c ConnectionCallbacks) OnReadData(data []byte) {
	if c.ReadDataCallback == nil {
		return
	}

	c.ReadDataCallback(data)
}
