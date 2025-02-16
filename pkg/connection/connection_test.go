package connection

import (
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// https://stackoverflow.com/questions/30688685/how-does-one-test-net-conn-in-unit-tests-in-golang

func TestUnit_Connection_Read(t *testing.T) {
	client, server := newTestConnection()
	conn := Wrap(server)

	wg := asyncWriteSampleDataToConnection(t, client)
	actual, err := conn.Read()
	wg.Wait()

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, sampleData, actual)
}

func TestUnit_Connection_ReadWithTimeout_WhenNoDataWritten_ReturnsNoData(t *testing.T) {
	_, server := newTestConnection()
	opts := connectionOptions{
		ReadTimeout: 100 * time.Millisecond,
	}
	conn := WithOptions(server, opts)

	actual, err := conn.Read()

	assert.True(t, errors.IsErrorWithCode(err, ErrReadTimeout))
	assert.Equal(t, []byte{}, actual)
}

func TestUnit_Connection_ReadWithTimeout(t *testing.T) {
	client, server := newTestConnection()
	opts := connectionOptions{
		// 2 reads will be over the delay we set for the client connection.
		ReadTimeout: 150 * time.Millisecond,
	}
	conn := WithOptions(server, opts)

	wg := asyncWriteSampleDataToConnectionWithDelay(t, client, 200*time.Millisecond)
	firstRead, err := conn.Read()
	assert.True(t, errors.IsErrorWithCode(err, ErrReadTimeout))

	secondRead, err := conn.Read()
	assert.Nil(t, err, "Actual err: %v", err)

	wg.Wait()

	assert.Equal(t, []byte{}, firstRead)
	assert.Equal(t, sampleData, secondRead)
}

func TestUnit_Connection_Read_WhenDisconnect_ReturnsExplicitError(t *testing.T) {
	client, server := newTestConnection()
	conn := Wrap(server)

	err := client.Close()
	assert.Nil(t, err, "Actual err: %v", err)
	actual, err := conn.Read()

	assert.True(t, errors.IsErrorWithCode(err, ErrClientDisconnected), "Actual err: %v", err)
	assert.Nil(t, actual)
}

func TestUnit_Connection_Write(t *testing.T) {
	client, server := newTestConnection()
	conn := Wrap(server)

	wg, actual := asyncReadDataFromConnection(t, client)
	sizeWritten, err := conn.Write(sampleData)
	wg.Wait()

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, sizeWritten, actual.size)
	assert.Equal(t, sampleData, actual.data[:actual.size])
}

func TestUnit_Connection_Write_WhenDisconnect_ReturnsExplicitError(t *testing.T) {
	client, server := newTestConnection()
	conn := Wrap(server)

	err := client.Close()
	assert.Nil(t, err, "Actual err: %v", err)
	actual, err := conn.Write(sampleData)

	assert.True(t, errors.IsErrorWithCode(err, ErrClientDisconnected), "Actual err: %v", err)
	assert.Equal(t, 0, actual)
}
