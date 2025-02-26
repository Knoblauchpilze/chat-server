package messages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_Decode_ClientConnectedMessage(t *testing.T) {
	encoded := []byte{
		// CLIENT_CONNECTED
		0x0, 0x0, 0x0, 0x0,
		// UUID
		0x2d, 0xbf, 0x26, 0x22, 0x2a, 0x95, 0x4b, 0xd1, 0x9b, 0x38, 0x2f, 0x7b, 0x4c, 0xe6, 0x5f, 0xfe,
	}

	msg, err := Decode(encoded)

	assert.Nil(t, err, "Actual err: %v", err)
	actual, ok := msg.(*clientConnectedMessage)
	assert.True(t, ok)
	assert.Equal(t, sampleUuid, actual.client)
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
