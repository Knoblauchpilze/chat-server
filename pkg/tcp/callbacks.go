package tcp

type OnDisconnect func() error
type OnReadError func(err error) error
type OnReadData func(data []byte) error

type ConnectionCallbacks struct {
	DisconnectCallback OnDisconnect
	ReadErrorCallback  OnReadError
	ReadDataCallback   OnReadData
}

func (c ConnectionCallbacks) OnDisconnect() error {
	if c.DisconnectCallback == nil {
		return nil
	}

	return c.DisconnectCallback()
}

func (c ConnectionCallbacks) OnReadError(err error) error {
	if c.ReadErrorCallback == nil {
		return nil
	}

	return c.ReadErrorCallback(err)
}

func (c ConnectionCallbacks) OnReadData(data []byte) error {
	if c.ReadDataCallback == nil {
		return nil
	}

	return c.ReadDataCallback(data)
}
