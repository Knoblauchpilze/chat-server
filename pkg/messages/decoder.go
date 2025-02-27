package messages

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
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
	var clientId uuid.UUID
	if err := tryDecodeDataAndWrapError(reader, &clientId); err != nil {
		return nil, err
	}

	msg := NewClientConnectedMessage(clientId)
	return msg, nil
}

func decodeClientDisconnectedMessage(reader *bytes.Reader) (Message, error) {
	var clientId uuid.UUID
	if err := tryDecodeDataAndWrapError(reader, &clientId); err != nil {
		return nil, err
	}

	msg := NewClientDisconnectedMessage(clientId)
	return msg, nil
}

func decodeDirectMessageMessage(reader *bytes.Reader) (Message, error) {
	var emitter, receiver uuid.UUID
	if err := tryDecodeDataAndWrapError(reader, &emitter); err != nil {
		return nil, err
	}
	if err := tryDecodeDataAndWrapError(reader, &receiver); err != nil {
		return nil, err
	}

	var contentLength int32
	if err := tryDecodeDataAndWrapError(reader, &contentLength); err != nil {
		return nil, err
	}

	content := make([]byte, contentLength)
	if err := tryDecodeDataAndWrapError(reader, content); err != nil {
		return nil, err
	}

	msg := NewDirectMessage(emitter, receiver, string(content))
	return msg, nil
}

func decodeRoomMessageMessage(reader *bytes.Reader) (Message, error) {
	return nil, errors.NotImplemented()
}

func tryDecodeDataAndWrapError(reader io.Reader, data any) error {
	if err := binary.Read(reader, binary.LittleEndian, data); err != nil {
		return errors.WrapCode(err, ErrMessageDecodingFailed)
	}

	return nil
}
