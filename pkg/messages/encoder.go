package messages

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

func Encode(msg Message) ([]byte, error) {

	var out bytes.Buffer
	writer := bufio.NewWriter(&out)

	var err error
	switch msg.Type() {
	case CLIENT_CONNECTED:
		err = encodeClientConnectedMessage(msg, writer)
	case CLIENT_DISCONNECTED:
		err = encodeClientDisconnectedMessage(msg, writer)
	case DIRECT_MESSAGE:
		err = encodeDirectMessageMessage(msg, writer)
	case ROOM_MESSAGE:
		err = encodeRoomMessageMessage(msg, writer)
	}

	if err != nil {
		return nil, errors.WrapCode(err, ErrMessageEncodingFailed)
	}
	if err := finalizeMessageEncoding(writer); err != nil {
		return nil, errors.WrapCode(err, ErrMessageEncodingFailed)
	}

	return out.Bytes(), nil
}

func encodeClientConnectedMessage(msg Message, writer io.Writer) error {
	ccMsg, ok := msg.(*clientConnectedMessage)
	if !ok {
		return errors.NewCode(ErrUnrecognizedMessageImplementation)
	}

	if err := tryEncodeData(writer, CLIENT_CONNECTED); err != nil {
		return err
	}
	if err := tryEncodeData(writer, ccMsg.client); err != nil {
		return err
	}

	return nil
}

func encodeClientDisconnectedMessage(msg Message, writer io.Writer) error {
	cdMsg, ok := msg.(*clientDisconnectedMessage)
	if !ok {
		return errors.NewCode(ErrUnrecognizedMessageImplementation)
	}

	if err := tryEncodeData(writer, CLIENT_DISCONNECTED); err != nil {
		return err
	}
	if err := tryEncodeData(writer, cdMsg.client); err != nil {
		return err
	}

	return nil
}

func encodeDirectMessageMessage(msg Message, writer io.Writer) error {
	return errors.NotImplemented()
}

func encodeRoomMessageMessage(msg Message, writer io.Writer) error {
	return errors.NotImplemented()
}

func tryEncodeData(writer io.Writer, data any) error {
	if err := binary.Write(writer, binary.LittleEndian, data); err != nil {
		return err
	}

	return nil
}

func finalizeMessageEncoding(writer *bufio.Writer) error {
	if err := writer.Flush(); err != nil {
		return errors.WrapCode(err, ErrMessageEncodingFailed)
	}

	return nil
}
