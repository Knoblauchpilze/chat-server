package messages

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
)

func Decode(data []byte) (Message, int, error) {
	reader := bytes.NewReader(data)
	initialSize := reader.Len()

	var msgType MessageType
	err := binary.Read(reader, binary.LittleEndian, &msgType)
	if err != nil {
		return nil, 0, errors.WrapCode(err, ErrUnrecognizedMessageFormat)
	}

	var msg Message

	switch msgType {
	case CLIENT_CONNECTED:
		msg, err = decodeClientConnectedMessage(reader)
	case CLIENT_DISCONNECTED:
		msg, err = decodeClientDisconnectedMessage(reader)
	case DIRECT_MESSAGE:
		msg, err = decodeDirectMessageMessage(reader)
	case ROOM_MESSAGE:
		msg, err = decodeRoomMessageMessage(reader)
	default:
		err = errors.NewCode(ErrUnsupportedMessageType)
	}

	var readSize int
	if err == nil {
		readSize = initialSize - reader.Len()
	}
	return msg, readSize, err
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
	var emitter, room uuid.UUID
	if err := tryDecodeDataAndWrapError(reader, &emitter); err != nil {
		return nil, err
	}
	if err := tryDecodeDataAndWrapError(reader, &room); err != nil {
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

	msg := NewRoomMessage(emitter, room, string(content))
	return msg, nil
}

func tryDecodeDataAndWrapError(reader io.Reader, data any) error {
	if err := binary.Read(reader, binary.LittleEndian, data); err != nil {
		return errors.WrapCode(err, ErrMessageDecodingFailed)
	}

	return nil
}
