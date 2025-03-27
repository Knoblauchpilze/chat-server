package clients

import (
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Handshake_WhenNoDataSent_ExpectFailure(t *testing.T) {
	_, server := newTestConnection(t, 7400)

	timeout := 50 * time.Millisecond
	_, err := Handshake(server, timeout)

	assert.True(
		t,
		errors.IsErrorWithCode(err, HandshakeTimeout),
		"Actual err: %v",
		err,
	)
}

func TestUnit_Handshake_WhenPartialDataSent_ExpectFailure(t *testing.T) {
	client, server := newTestConnection(t, 7401)

	id := uuid.New()
	n, err := client.Write(id[:5])
	assert.Equal(t, 5, n)
	assert.Nil(t, err, "Actual err: %v", err)

	timeout := 200 * time.Millisecond
	_, err = Handshake(server, timeout)

	assert.True(
		t,
		errors.IsErrorWithCode(err, IncompleteHandshake),
		"Actual err: %v",
		err,
	)
}

func TestUnit_Handshake_WhenDataIsNotAnId_ExpectFailure(t *testing.T) {
	client, server := newTestConnection(t, 7402)

	expected := uuid.New()
	n, err := client.Write(expected[:])
	assert.Equal(t, len(expected), n)
	assert.Nil(t, err, "Actual err: %v", err)

	timeout := 200 * time.Millisecond
	actual, err := Handshake(server, timeout)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, expected, actual)
}
