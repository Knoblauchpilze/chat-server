package connection

import (
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// https://stackoverflow.com/questions/30688685/how-does-one-test-net-conn-in-unit-tests-in-golang

func TestUnit_Connection_Read(t *testing.T) {
	client, server := newTestConnection(t, 1600)
	conn := Wrap(server)

	wg := asyncWriteSampleDataToConnection(t, client)
	actual, err := conn.Read()
	wg.Wait()

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, sampleData, actual)
}

func TestUnit_Connection_ReadWithTimeout_WhenNoDataWritten_ReturnsNoData(t *testing.T) {
	_, server := newTestConnection(t, 1601)
	opts := connectionOptions{
		ReadTimeout: 100 * time.Millisecond,
	}
	conn := WithOptions(server, opts)

	actual, err := conn.Read()

	assert.True(t, errors.IsErrorWithCode(err, ErrReadTimeout))
	assert.Equal(t, []byte{}, actual)
}

func TestUnit_Connection_ReadWithTimeout_WhenPartialDataReceived_ReturnsNoData(t *testing.T) {
	client, server := newTestConnection(t, 1602)
	opts := WithReadTimeout(150 * time.Millisecond)
	conn := WithOptions(server, opts)

	wg := asyncWriteDataToConnection(t, client, []byte("hello"))
	data, err := conn.Read()
	assert.True(t, errors.IsErrorWithCode(err, ErrReadTimeout), "Actual err: %v", err)
	wg.Wait()

	assert.Equal(t, []byte{}, data)
}

func TestUnit_Connection_ReadWithTimeout_WhenPartialDataReceivedIsTooBig_ExpectError(t *testing.T) {
	client, server := newTestConnection(t, 1602)
	opts := connectionOptions{
		ReadTimeout:                  150 * time.Millisecond,
		IncompleteMessageSizeInBytes: 2,
	}
	conn := WithOptions(server, opts)

	wg := asyncWriteDataToConnection(t, client, []byte("hello"))
	data, err := conn.Read()
	assert.True(t, errors.IsErrorWithCode(err, ErrTooLargeIncompleteData), "Actual err: %v", err)
	wg.Wait()

	assert.Equal(t, []byte{}, data)
}

func TestUnit_Connection_ReadWithTimeout_WhenPartialDataReceivedAndMoreDataAfterwards_ExpectReturnsCompleteData(t *testing.T) {
	client, server := newTestConnection(t, 1602)
	opts := WithReadTimeout(150 * time.Millisecond)
	conn := WithOptions(server, opts)

	wg := asyncWriteDataToConnection(t, client, []byte("hel"))
	data, err := conn.Read()
	assert.True(t, errors.IsErrorWithCode(err, ErrReadTimeout), "Actual err: %v", err)
	assert.Equal(t, []byte{}, data)
	wg.Wait()

	wg = asyncWriteDataToConnection(t, client, []byte("lo\n"))
	data, err = conn.Read()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()

	assert.Equal(t, []byte("hello\n"), data)
}

func TestUnit_Connection_ReadWithTimeout(t *testing.T) {
	client, server := newTestConnection(t, 1602)
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
	client, server := newTestConnection(t, 1603)
	conn := Wrap(server)

	err := client.Close()
	assert.Nil(t, err, "Actual err: %v", err)
	actual, err := conn.Read()

	assert.True(t, errors.IsErrorWithCode(err, ErrClientDisconnected), "Actual err: %v", err)
	assert.Nil(t, actual)
}

func TestUnit_Connection_Write(t *testing.T) {
	client, server := newTestConnection(t, 1604)
	conn := Wrap(server)

	wg, actual := asyncReadDataFromConnection(t, client)
	sizeWritten, err := conn.Write(sampleData)
	wg.Wait()

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, sizeWritten, actual.size)
	assert.Equal(t, sampleData, actual.data[:actual.size])
}

func TestUnit_Connection_Write_WhenDisconnect_ReturnsNoError(t *testing.T) {
	// This topic indicates that no error is returned when writing to a closed connection
	// https://groups.google.com/g/golang-nuts/c/MRIOQ-82ofM
	client, server := newTestConnection(t, 1605)
	conn := Wrap(server)

	err := client.Close()
	assert.Nil(t, err, "Actual err: %v", err)
	actual, err := conn.Write(sampleData)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(sampleData), actual)
}
