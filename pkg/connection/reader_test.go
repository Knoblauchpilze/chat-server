package connection

import (
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_ReadFromConnection_NoError(t *testing.T) {
	client, server := newTestConnection(t, 1500)
	conn := Wrap(client)
	asyncWriteSampleDataToConnection(t, server)

	timeout, err := readFromConnection(sampleUuid, conn, Callbacks{})

	assert.False(t, timeout)
	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_ReadFromConnection_ReadTimeout(t *testing.T) {
	client, _ := newTestConnection(t, 1501)
	opts := connectionOptions{
		ReadTimeout: 100 * time.Millisecond,
	}
	conn := WithOptions(client, opts)

	timeout, err := readFromConnection(sampleUuid, conn, Callbacks{})

	assert.True(t, timeout)
	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_ReadFromConnection_Disconnect(t *testing.T) {
	client, server := newTestConnection(t, 1502)
	err := server.Close()
	assert.Nil(t, err, "Actual err: %v", err)
	conn := Wrap(client)

	timeout, err := readFromConnection(sampleUuid, conn, Callbacks{})

	assert.False(t, timeout)
	assert.True(t, errors.IsErrorWithCode(err, ErrClientDisconnected), "Actual err: %v", err)
}

func TestUnit_ReadFromConnection_ReadWithCallback(t *testing.T) {
	client, server := newTestConnection(t, 1503)
	conn := Wrap(client)
	asyncWriteSampleDataToConnection(t, server)

	var actualId uuid.UUID
	var actualData []byte
	callbacks := Callbacks{
		ReadDataCallbacks: []OnReadData{
			func(id uuid.UUID, data []byte) {
				actualId = id
				actualData = data
			},
		},
	}
	timeout, err := readFromConnection(sampleUuid, conn, callbacks)

	assert.False(t, timeout)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, sampleUuid, actualId)
	assert.Equal(t, sampleData, actualData)
}

func TestUnit_ReadFromConnection_DisconnectWithCallback(t *testing.T) {
	client, server := newTestConnection(t, 1504)
	err := server.Close()
	assert.Nil(t, err, "Actual err: %v", err)
	conn := Wrap(client)

	var actualId uuid.UUID
	callbacks := Callbacks{
		DisconnectCallbacks: []OnDisconnect{
			func(id uuid.UUID) {
				actualId = id
			},
		},
	}
	timeout, err := readFromConnection(sampleUuid, conn, callbacks)

	assert.False(t, timeout)
	assert.True(t, errors.IsErrorWithCode(err, ErrClientDisconnected), "Actual err: %v", err)
	assert.Equal(t, sampleUuid, actualId)
}
