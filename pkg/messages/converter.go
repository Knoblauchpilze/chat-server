package messages

import "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"

func ToMessageStruct[ConcreteMessageType Message](msg Message) (ConcreteMessageType, error) {
	concrete, ok := msg.(ConcreteMessageType)
	if !ok {
		var out ConcreteMessageType
		return out, errors.NewCode(ErrUnrecognizedMessageImplementation)
	}

	return concrete, nil
}
