package messages

type MessageType int

const (
	CLIENT_CONNECTED MessageType = iota
	CLIENT_DISCONNECTED
	DATA_RECEIVED
	DIRECT_MESSAGE
	ROOM_MESSAGE
)

func (m MessageType) String() string {
	switch m {
	case CLIENT_CONNECTED:
		return "CLIENT_CONNECTED"
	case CLIENT_DISCONNECTED:
		return "CLIENT_DISCONNECTED"
	case DATA_RECEIVED:
		return "DATA_RECEIVED"
	case DIRECT_MESSAGE:
		return "DIRECT_MESSAGE"
	case ROOM_MESSAGE:
		return "ROOM_MESSAGE"
	}

	return "UNKNOWN"
}
