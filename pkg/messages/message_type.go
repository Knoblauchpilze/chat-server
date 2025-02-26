package messages

type MessageType int32

const (
	CLIENT_CONNECTED MessageType = iota
	CLIENT_DISCONNECTED
	DIRECT_MESSAGE
	ROOM_MESSAGE
)

func (m MessageType) String() string {
	switch m {
	case CLIENT_CONNECTED:
		return "CLIENT_CONNECTED"
	case CLIENT_DISCONNECTED:
		return "CLIENT_DISCONNECTED"
	case DIRECT_MESSAGE:
		return "DIRECT_MESSAGE"
	case ROOM_MESSAGE:
		return "ROOM_MESSAGE"
	}

	return "UNKNOWN"
}
