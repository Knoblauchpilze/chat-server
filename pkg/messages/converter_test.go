package messages

import (
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_ToMessageStruct_ClientConnectedMessage(t *testing.T) {
	client := uuid.New()
	in := NewClientConnectedMessage(client)

	actual, err := ToMessageStruct[ClientConnectedMessage](in)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, CLIENT_CONNECTED, actual.Type())
	assert.Equal(t, in, actual)
}

func TestUnit_ToMessageStruct_WhenWrongType_ExpectError(t *testing.T) {
	client := uuid.New()
	in := NewClientConnectedMessage(client)

	_, err := ToMessageStruct[ClientDisconnectedMessage](in)

	assert.True(t, errors.IsErrorWithCode(err, ErrUnrecognizedMessageImplementation))
}

func TestUnit_ToMessageStruct_ClientDisonnectedMessage(t *testing.T) {
	client := uuid.New()
	in := NewClientDisconnectedMessage(client)

	actual, err := ToMessageStruct[ClientDisconnectedMessage](in)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, CLIENT_DISCONNECTED, actual.Type())
	assert.Equal(t, in, actual)
}

func TestUnit_ToMessageStruct_DirectMessage(t *testing.T) {
	emitter := uuid.New()
	receiver := uuid.New()
	in := NewDirectMessage(emitter, receiver, "some content")

	actual, err := ToMessageStruct[DirectMessage](in)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, DIRECT_MESSAGE, actual.Type())
	assert.Equal(t, in, actual)
}

func TestUnit_ToMessageStruct_RoomMessage(t *testing.T) {
	emitter := uuid.New()
	room := uuid.New()
	in := NewRoomMessage(emitter, room, "some content")

	actual, err := ToMessageStruct[RoomMessage](in)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, ROOM_MESSAGE, actual.Type())
	assert.Equal(t, in, actual)
}
