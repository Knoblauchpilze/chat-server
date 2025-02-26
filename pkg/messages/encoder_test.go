package messages

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var sampleUuid = uuid.MustParse("2dbf2622-2a95-4bd1-9b38-2f7b4ce65ffe")

func TestUnit_Encode_ClientConnectedMessage(t *testing.T) {
	msg := NewClientConnectedMessage(sampleUuid)

	actual, err := Encode(msg)

	assert.Nil(t, err, "Actual err: %v", err)
	expected := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// UUID
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}
	assert.Equal(t, expected, actual)
}

func TestUnit_Encode_ClientDisconnectedMessage(t *testing.T) {
	msg := NewClientDisconnectedMessage(sampleUuid)

	actual, err := Encode(msg)

	assert.Nil(t, err, "Actual err: %v", err)
	expected := []byte{
		// CLIENT_DISCONNECTED
		0x01, 0x0, 0x0, 0x0,
		// UUID
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}
	assert.Equal(t, expected, actual)
}
