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
