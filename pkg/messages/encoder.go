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
		err = encodeDirectMessage(msg, writer)
	case ROOM_MESSAGE:
		err = encodeRoomMessage(msg, writer)
	}

	if err != nil {
		return nil, errors.WrapCode(err, ErrMessageEncodingFailed)
	}
	if err := finalizeMessageEncoding(writer); err != nil {
		return nil, errors.WrapCode(err, ErrMessageEncodingFailed)
	}

	return out.Bytes(), nil
}

func encodeClientConnectedMessage(inMsg Message, writer io.Writer) error {
	msg, ok := inMsg.(*ClientConnectedMessage)
	if !ok {
		return errors.NewCode(ErrUnrecognizedMessageImplementation)
	}

	if err := tryEncodeData(writer, CLIENT_CONNECTED); err != nil {
		return err
	}
	if err := tryEncodeData(writer, msg.Client); err != nil {
		return err
	}

	return nil
}

func encodeClientDisconnectedMessage(inMsg Message, writer io.Writer) error {
	msg, ok := inMsg.(*ClientDisconnectedMessage)
	if !ok {
		return errors.NewCode(ErrUnrecognizedMessageImplementation)
	}

	if err := tryEncodeData(writer, CLIENT_DISCONNECTED); err != nil {
		return err
	}
	if err := tryEncodeData(writer, msg.Client); err != nil {
		return err
	}

	return nil
}

func encodeDirectMessage(inMsg Message, writer io.Writer) error {
	msg, ok := inMsg.(*directMessage)
	if !ok {
		return errors.NewCode(ErrUnrecognizedMessageImplementation)
	}

	if err := tryEncodeData(writer, DIRECT_MESSAGE); err != nil {
		return err
	}
	if err := tryEncodeData(writer, msg.Emitter); err != nil {
		return err
	}
	if err := tryEncodeData(writer, msg.Receiver); err != nil {
		return err
	}
	if err := tryEncodeData(writer, int32(len(msg.Content))); err != nil {
		return err
	}
	if err := tryEncodeData(writer, []byte(msg.Content)); err != nil {
		return err
	}

	return nil
}

func encodeRoomMessage(inMsg Message, writer io.Writer) error {
	msg, ok := inMsg.(*roomMessage)
	if !ok {
		return errors.NewCode(ErrUnrecognizedMessageImplementation)
	}

	if err := tryEncodeData(writer, ROOM_MESSAGE); err != nil {
		return err
	}
	if err := tryEncodeData(writer, msg.Emitter); err != nil {
		return err
	}
	if err := tryEncodeData(writer, msg.Room); err != nil {
		return err
	}
	if err := tryEncodeData(writer, int32(len(msg.Content))); err != nil {
		return err
	}
	if err := tryEncodeData(writer, []byte(msg.Content)); err != nil {
		return err
	}

	return nil
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
