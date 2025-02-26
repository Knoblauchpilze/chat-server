package messages

import (
	"bytes"
	"encoding/binary"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

func Decode(data []byte) (Message, error) {
	reader := bytes.NewReader(data)

	var msgType MessageType
	err := binary.Read(reader, binary.LittleEndian, &msgType)
	if err != nil {
		return nil, errors.WrapCode(err, ErrUnrecognizedMessageFormat)
	}

	switch msgType {
	case CLIENT_CONNECTED:
		return decodeClientConnectedMessage(reader)
	case CLIENT_DISCONNECTED:
		return decodeClientDisconnectedMessage(reader)
	case DIRECT_MESSAGE:
		return decodeDirectMessageMessage(reader)
	case ROOM_MESSAGE:
		return decodeRoomMessageMessage(reader)
	}

	return nil, errors.NewCode(ErrUnsupportedMessageType)
}

func decodeClientConnectedMessage(reader *bytes.Reader) (Message, error) {
	return nil, errors.NotImplemented()
}

func decodeClientDisconnectedMessage(reader *bytes.Reader) (Message, error) {
	return nil, errors.NotImplemented()
}

func decodeDirectMessageMessage(reader *bytes.Reader) (Message, error) {
	return nil, errors.NotImplemented()
}

func decodeRoomMessageMessage(reader *bytes.Reader) (Message, error) {
	return nil, errors.NotImplemented()
}
