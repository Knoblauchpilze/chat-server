package messages

import "github.com/google/uuid"

type Message interface {
	Type() MessageType
}

type clientConnectedMessage struct {
	client uuid.UUID
}

func NewClientConnectedMessage(client uuid.UUID) Message {
	return &clientConnectedMessage{
		client: client,
	}
}

func (m *clientConnectedMessage) Type() MessageType {
	return CLIENT_CONNECTED
}

type clientDisconnectedMessage struct {
	client uuid.UUID
}

func NewClientDisconnectedMessage(client uuid.UUID) Message {
	return &clientDisconnectedMessage{
		client: client,
	}
}

func (m *clientDisconnectedMessage) Type() MessageType {
	return CLIENT_DISCONNECTED
}

type directMessage struct {
	emitter  uuid.UUID
	receiver uuid.UUID
	content  string
}

func NewDirectMessage(emitter uuid.UUID, receiver uuid.UUID, content string) Message {
	return &directMessage{
		emitter:  emitter,
		receiver: receiver,
		content:  content,
	}
}

func (m *directMessage) Type() MessageType {
	return DIRECT_MESSAGE
}
