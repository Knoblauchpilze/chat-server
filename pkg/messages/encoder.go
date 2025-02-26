package messages

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

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
	ccMsg, ok := msg.(*clientConnectedMessage)
	if !ok {
		return nil, errors.NewCode(ErrUnrecognizedMessageImplementation)
	}

	var out bytes.Buffer
	writer := bufio.NewWriter(&out)

	if err := tryEncodeDataAndWrapError(writer, CLIENT_CONNECTED); err != nil {
		return nil, err
	}
	if err := tryEncodeDataAndWrapError(writer, ccMsg.client); err != nil {
		return nil, errors.WrapCode(err, ErrMessageEncodingFailed)
	}

	if err := finalizeMessageEncoding(writer); err != nil {
		return nil, errors.WrapCode(err, ErrMessageEncodingFailed)
	}

	return out.Bytes(), nil
}

func encodeClientDisconnectedMessage(msg Message) ([]byte, error) {
	cdMsg, ok := msg.(*clientDisconnectedMessage)
	if !ok {
		return nil, errors.NewCode(ErrUnrecognizedMessageImplementation)
	}

	var out bytes.Buffer
	writer := bufio.NewWriter(&out)

	if err := tryEncodeDataAndWrapError(writer, CLIENT_DISCONNECTED); err != nil {
		return nil, err
	}
	if err := tryEncodeDataAndWrapError(writer, cdMsg.client); err != nil {
		return nil, errors.WrapCode(err, ErrMessageEncodingFailed)
	}

	if err := finalizeMessageEncoding(writer); err != nil {
		return nil, errors.WrapCode(err, ErrMessageEncodingFailed)
	}

	return out.Bytes(), nil
}

func encodeDirectMessageMessage(msg Message) ([]byte, error) {
	return nil, errors.NotImplemented()
}

func encodeRoomMessageMessage(msg Message) ([]byte, error) {
	return nil, errors.NotImplemented()
}

func tryEncodeDataAndWrapError(writer io.Writer, data any) error {
	if err := binary.Write(writer, binary.LittleEndian, data); err != nil {
		return errors.WrapCode(err, ErrMessageEncodingFailed)
	}

	return nil
}

func finalizeMessageEncoding(writer *bufio.Writer) error {
	if err := writer.Flush(); err != nil {
		return errors.WrapCode(err, ErrMessageEncodingFailed)
	}

	return nil
}
