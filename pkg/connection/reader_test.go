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

	processed, timeout, err := readFromConnection(sampleUuid, conn, Callbacks{})

	assert.Equal(t, 0, processed)
	assert.False(t, timeout)
	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_ReadFromConnection_ReadTimeout(t *testing.T) {
	client, _ := newTestConnection(t, 1501)
	opts := connectionOptions{
		ReadTimeout: 100 * time.Millisecond,
	}
	conn := WithOptions(client, opts)

	processed, timeout, err := readFromConnection(sampleUuid, conn, Callbacks{})

	assert.Equal(t, 0, processed)
	assert.True(t, timeout)
	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_ReadFromConnection_Disconnect(t *testing.T) {
	client, server := newTestConnection(t, 1502)
	err := server.Close()
	assert.Nil(t, err, "Actual err: %v", err)
	conn := Wrap(client)

	processed, timeout, err := readFromConnection(sampleUuid, conn, Callbacks{})

	assert.Equal(t, 0, processed)
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
		ReadDataCallback: func(id uuid.UUID, data []byte) int {
			actualId = id
			actualData = data
			return 15
		},
	}
	processed, timeout, err := readFromConnection(sampleUuid, conn, callbacks)

	assert.Equal(t, 15, processed)
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
		DisconnectCallback: func(id uuid.UUID) {
			actualId = id
		},
	}
	processed, timeout, err := readFromConnection(sampleUuid, conn, callbacks)

	assert.Equal(t, 0, processed)
	assert.False(t, timeout)
	assert.True(t, errors.IsErrorWithCode(err, ErrClientDisconnected), "Actual err: %v", err)
	assert.Equal(t, sampleUuid, actualId)
}
