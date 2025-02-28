package messages

import (
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Decode_WhenMessageTypeIsTruncated_ExpectError(t *testing.T) {
	encoded := []byte{
		// Partial CLIENT_CONNECTED
		0x0, 0x0, 0x0,
	}

	_, err := Decode(encoded)

	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrUnrecognizedMessageFormat),
		"Actual err: %v",
		err,
	)
}

func TestUnit_Decode_ClientConnectedMessage(t *testing.T) {
	encoded := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// Partial UUID
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}

	msg, err := Decode(encoded)

	assert.Nil(t, err, "Actual err: %v", err)
	actual, ok := msg.(*clientConnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, sampleUuid, actual.client)
}

func TestUnit_Decode_WhenMessageIsIncomplete_ExpectError(t *testing.T) {
	encoded := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// UUID
		0x2d, 0xbf, 0x26, 0x22,
	}

	_, err := Decode(encoded)

	assert.True(
		t,
		errors.IsErrorWithCode(err, ErrMessageDecodingFailed),
		"Actual err: %v",
		err,
	)
}

func TestUnit_Decode_ClientDisconnectedMessage(t *testing.T) {
	encoded := []byte{
		// CLIENT_DISCONNECTED
		0x01, 0x0, 0x0, 0x0,
		// UUID
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}

	msg, err := Decode(encoded)

	assert.Nil(t, err, "Actual err: %v", err)
	actual, ok := msg.(*clientDisconnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, sampleUuid, actual.client)
}

func TestUnit_Decode_DirectMessage(t *testing.T) {
	encoded := []byte{
		// DIRECT_MESSAGE
		0x02, 0x0, 0x0, 0x0,
		// Emitter
		0x9c, 0xd1, 0x3f, 0x28, 0xc5, 0x60, 0x4a, 0xdd, 0x83, 0xde, 0xeb, 0x6c, 0x47, 0x3d, 0xea, 0x5,
		// Receiver
		0xb2, 0x95, 0x4, 0xfe, 0x27, 0x9e, 0x43, 0x12, 0x8a, 0x7e, 0x3f, 0xc9, 0x5c, 0xe0, 0xaf, 0xa5,
		// Content length
		0x1b, 0x0, 0x0, 0x0,
		// Content
		0x74, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x61, 0x6e, 0x20, 0x61, 0x77, 0x65, 0x73, 0x6f, 0x6d, 0x65, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x21,
	}

	msg, err := Decode(encoded)

	assert.Nil(t, err, "Actual err: %v", err)
	actual, ok := msg.(*directMessage)
	assert.True(t, ok)
	expectedEmitter := uuid.MustParse("9cd13f28-c560-4add-83de-eb6c473dea05")
	assert.Equal(t, expectedEmitter, actual.emitter)
	expectedReceiver := uuid.MustParse("b29504fe-279e-4312-8a7e-3fc95ce0afa5")
	assert.Equal(t, expectedReceiver, actual.receiver)
	assert.Equal(t, "this is an awesome message!", actual.content)
}

func TestUnit_Decode_RoomMessage(t *testing.T) {
	encoded := []byte{
		// ROOM_MESSAGE
		0x03, 0x0, 0x0, 0x0,
		// Emitter
		0x9c, 0xd1, 0x3f, 0x28, 0xc5, 0x60, 0x4a, 0xdd, 0x83, 0xde, 0xeb, 0x6c, 0x47, 0x3d, 0xea, 0x5,
		// Room
		0xb2, 0x95, 0x4, 0xfe, 0x27, 0x9e, 0x43, 0x12, 0x8a, 0x7e, 0x3f, 0xc9, 0x5c, 0xe0, 0xaf, 0xa5,
		// Content length
		0x13, 0x0, 0x0, 0x0,
		// Content
		0x31, 0x38, 0x20, 0x61, 0x6e, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x3f,
	}

	msg, err := Decode(encoded)

	assert.Nil(t, err, "Actual err: %v", err)
	actual, ok := msg.(*roomMessage)
	assert.True(t, ok)
	expectedEmitter := uuid.MustParse("9cd13f28-c560-4add-83de-eb6c473dea05")
	assert.Equal(t, expectedEmitter, actual.emitter)
	expectedRoom := uuid.MustParse("b29504fe-279e-4312-8a7e-3fc95ce0afa5")
	assert.Equal(t, expectedRoom, actual.room)
	assert.Equal(t, "18 another message?", actual.content)
}
