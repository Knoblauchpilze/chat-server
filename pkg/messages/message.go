package messages

type MessageType int

const (
	CLIENT_CONNECTED MessageType = iota
	CLIENT_DISCONNECTED
	DATA_RECEIVED
)

type Message interface {
	Type() MessageType
}

type messageImpl struct {
	kind MessageType
}

func NewMessage(kind MessageType) Message {
	return messageImpl{
		kind: kind,
	}
}

func (m messageImpl) Type() MessageType {
	return m.kind
}
