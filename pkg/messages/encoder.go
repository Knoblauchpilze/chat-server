package messages

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

func Encode(msg Message) ([]byte, error) {
	switch msg.Type() {
	case CLIENT_CONNECTED:
		return encodeClientConnectedMessage(msg)
	case CLIENT_DISCONNECTED:
		return encodeClientDisconnectedMessage(msg)
	case DIRECT_MESSAGE:
		return encodeDirectMessageMessage(msg)
	case ROOM_MESSAGE:
		return encodeRoomMessageMessage(msg)
	}
	return nil, errors.NotImplemented()
}

func encodeClientConnectedMessage(msg Message) ([]byte, error) {
	return nil, errors.NotImplemented()
}

func encodeClientDisconnectedMessage(msg Message) ([]byte, error) {
	return nil, errors.NotImplemented()
}

func encodeDirectMessageMessage(msg Message) ([]byte, error) {
	return nil, errors.NotImplemented()
}

func encodeRoomMessageMessage(msg Message) ([]byte, error) {
	return nil, errors.NotImplemented()
}
