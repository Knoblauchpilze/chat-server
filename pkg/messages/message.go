package messages

import "github.com/google/uuid"

type Message interface {
	Type() MessageType
}

type ClientConnectedMessage struct {
	Client uuid.UUID
}

func NewClientConnectedMessage(client uuid.UUID) Message {
	return &ClientConnectedMessage{
		Client: client,
	}
}

func (m *ClientConnectedMessage) Type() MessageType {
	return CLIENT_CONNECTED
}

type ClientDisconnectedMessage struct {
	Client uuid.UUID
}

func NewClientDisconnectedMessage(client uuid.UUID) Message {
	return &ClientDisconnectedMessage{
		Client: client,
	}
}

func (m *ClientDisconnectedMessage) Type() MessageType {
	return CLIENT_DISCONNECTED
}

type directMessage struct {
	Emitter  uuid.UUID
	Receiver uuid.UUID
	Content  string
}

func NewDirectMessage(emitter uuid.UUID, receiver uuid.UUID, content string) Message {
	return &directMessage{
		Emitter:  emitter,
		Receiver: receiver,
		Content:  content,
	}
}

func (m *directMessage) Type() MessageType {
	return DIRECT_MESSAGE
}

type roomMessage struct {
	Emitter uuid.UUID
	Room    uuid.UUID
	Content string
}

func NewRoomMessage(emitter uuid.UUID, room uuid.UUID, content string) Message {
	return &roomMessage{
		Emitter: emitter,
		Room:    room,
		Content: content,
	}
}

func (m *roomMessage) Type() MessageType {
	return ROOM_MESSAGE
}
