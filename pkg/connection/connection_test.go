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

func TestUnit_Connection_ReadWithTimeout_WhenDataReceived_ReturnsData(t *testing.T) {
	client, server := newTestConnection(t, 1602)
	opts := WithReadTimeout(150 * time.Millisecond)
	conn := WithOptions(server, opts)

	wg := asyncWriteDataToConnection(t, client, []byte("hello"))
	data, err := conn.Read()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()

	assert.Equal(t, []byte("hello"), data)
}

func TestUnit_Connection_ReadWithTimeout_WhenDataReceivedAndNoDiscard_ExpectDataToStillBeAvailable(t *testing.T) {
	client, server := newTestConnection(t, 1603)
	opts := WithReadTimeout(150 * time.Millisecond)
	conn := WithOptions(server, opts)

	wg := asyncWriteDataToConnection(t, client, []byte("hello"))
	wg.Wait()

	data, err := conn.Read()
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, []byte("hello"), data)

	data, err = conn.Read()
	assert.True(t, errors.IsErrorWithCode(err, ErrReadTimeout), "Actual err: %v", err)
	assert.Equal(t, []byte("hello"), data)
}

func TestUnit_Connection_ReadWithTimeout_WhenDataReceivedIsTooBig_ExpectError(t *testing.T) {
	client, server := newTestConnection(t, 1604)
	opts := connectionOptions{
		ReadTimeout:               150 * time.Millisecond,
		MaximumMessageSizeInBytes: 2,
	}
	conn := WithOptions(server, opts)

	wg := asyncWriteDataToConnection(t, client, []byte("hello"))
	data, err := conn.Read()
	assert.True(t, errors.IsErrorWithCode(err, ErrTooLargeIncompleteData), "Actual err: %v", err)
	wg.Wait()

	assert.Equal(t, []byte{}, data)
}

func TestUnit_Connection_ReadWithTimeout_WhenDataReceivedAndMoreDataAfterwards_ExpectReturnsAllData(t *testing.T) {
	client, server := newTestConnection(t, 1605)
	opts := WithReadTimeout(150 * time.Millisecond)
	conn := WithOptions(server, opts)

	wg := asyncWriteDataToConnection(t, client, []byte("hel"))
	data, err := conn.Read()
	assert.Nil(t, err, "Actual err: %v\n", err)
	assert.Equal(t, []byte("hel"), data)
	wg.Wait()

	wg = asyncWriteDataToConnection(t, client, []byte("lo"))
	data, err = conn.Read()
	assert.Nil(t, err, "Actual err: %v", err)
	wg.Wait()

	assert.Equal(t, []byte("hello"), data)
}

func TestUnit_Connection_ReadWithTimeout(t *testing.T) {
	client, server := newTestConnection(t, 1606)
	opts := connectionOptions{
		// 2 reads will be over the delay we set for the client connection.
		ReadTimeout:               150 * time.Millisecond,
		MaximumMessageSizeInBytes: 100,
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
	client, server := newTestConnection(t, 1607)
	conn := Wrap(server)

	err := client.Close()
	assert.Nil(t, err, "Actual err: %v", err)
	actual, err := conn.Read()

	assert.True(t, errors.IsErrorWithCode(err, ErrClientDisconnected), "Actual err: %v", err)
	assert.Nil(t, actual)
}

func TestUnit_Connection_DiscardBytes_WhenNoBytes_ExpectNoError(t *testing.T) {
	_, server := newTestConnection(t, 1608)
	conn := Wrap(server)

	assert.NotPanics(
		t,
		func() {
			conn.DiscardBytes(10)
		},
	)
}

func TestUnit_Connection_DiscardBytes_WhenSomeBytes_ExpectDiscardAsRequested(t *testing.T) {
	client, server := newTestConnection(t, 1609)
	opts := WithReadTimeout(150 * time.Millisecond)
	conn := WithOptions(server, opts)

	wg := asyncWriteDataToConnection(t, client, []byte("hello"))
	wg.Wait()

	data, err := conn.Read()
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, []byte("hello"), data)

	conn.DiscardBytes(2)

	data, err = conn.Read()
	assert.True(t, errors.IsErrorWithCode(err, ErrReadTimeout), "Actual err: %v", err)
	assert.Equal(t, []byte("llo"), data)
}

func TestUnit_Connection_DiscardBytes_WhenAllBytesDiscarded_ExpectNoDataLeft(t *testing.T) {
	client, server := newTestConnection(t, 1609)
	opts := WithReadTimeout(150 * time.Millisecond)
	conn := WithOptions(server, opts)

	wg := asyncWriteDataToConnection(t, client, []byte("hello"))
	wg.Wait()

	data, err := conn.Read()
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, []byte("hello"), data)

	conn.DiscardBytes(5)

	data, err = conn.Read()
	assert.True(t, errors.IsErrorWithCode(err, ErrReadTimeout), "Actual err: %v", err)
	assert.Equal(t, []byte{}, data)
}

func TestUnit_Connection_Write(t *testing.T) {
	client, server := newTestConnection(t, 1610)
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
	client, server := newTestConnection(t, 1611)
	conn := Wrap(server)

	err := client.Close()
	assert.Nil(t, err, "Actual err: %v", err)
	actual, err := conn.Write(sampleData)

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(sampleData), actual)
}
