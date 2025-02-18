package internal

import (
	"context"
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_Server_StartAndStopWithContext(t *testing.T) {
	s, err := NewServer(newTestServerConfig(7000), logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)
	cancellable, cancel := context.WithCancel(context.Background())

	asyncCancelContext(reasonableWaitTimeForServerToBeUp, cancel)

	err = s.Start(cancellable)

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_Server_WhenServerIsStopped_ExpectClientConnectionToBeClosed(t *testing.T) {
	s, err := NewServer(newTestServerConfig(7001), logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)
	cancellable, cancel := context.WithCancel(context.Background())

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, err := net.Dial("tcp", ":7001")
	assert.Nil(t, err, "Actual err: %v", err)

	cancel()
	wg.Wait()

	require.Nil(t, *serverErr, "Actual err: %v", *serverErr)
	assertConnectionIsClosed(t, conn)
}

func TestUnit_Server_WhenClientConnects_ExpectCallbackNotified(t *testing.T) {
	config := newTestServerConfig(7002)
	var called int
	config.Callbacks.ConnectCallback = func(uuid.UUID, string) bool {
		called++
		return true
	}

	s, err := NewServer(config, logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)
	cancellable, cancel := context.WithCancel(context.Background())

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, err := net.Dial("tcp", ":7002")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	conn.Close()

	cancel()
	wg.Wait()

	assert.Nil(t, *serverErr, "Actual err: %v", *serverErr)
	assert.Equal(t, 1, called)
}

func TestUnit_Server_WhenClientSendsData_ExpectCallbackNotified(t *testing.T) {
	config := newTestServerConfig(7003)
	var called int
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) bool {
		called++
		return true
	}

	s, err := NewServer(config, logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)
	cancellable, cancel := context.WithCancel(context.Background())

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, err := net.Dial("tcp", ":7003")
	assert.Nil(t, err, "Actual err: %v", err)

	n, err := conn.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err)

	conn.Close()

	cancel()
	wg.Wait()

	require.Nil(t, *serverErr, "Actual err: %v", *serverErr)
	assert.Equal(t, 1, called)
}

func TestUnit_Server_WhenClientDisconnects_ExpectCallbackNotified(t *testing.T) {
	config := newTestServerConfig(7004)
	var called int
	config.Callbacks.DisconnectCallback = func(uuid.UUID) {
		called++
	}

	s, err := NewServer(config, logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)
	cancellable, cancel := context.WithCancel(context.Background())

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, err := net.Dial("tcp", ":7004")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	conn.Close()

	cancel()
	wg.Wait()

	assert.Nil(t, *serverErr, "Actual err: %v", *serverErr)
	assert.Equal(t, 1, called)
}

func TestUnit_Server_WhenClientConnectsAndIsDenied_ExpectConnectionToBeClosed(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTestServerConfig(7003)
	var called int
	config.Callbacks.ConnectCallback = func(uuid.UUID, string) bool {
		called++
		return false
	}
	s, err := NewServer(config, logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)

	wgServer, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, dialErr := net.Dial("tcp", ":7003")
	assert.Nil(t, dialErr, "Actual err: %v", dialErr)

	// Wait for connection to be processed.
	time.Sleep(100 * time.Millisecond)

	assertConnectionIsClosed(t, conn)

	cancel()
	wgServer.Wait()

	require.Nil(t, *serverErr, "Actual err: %v", *serverErr)
	assert.Equal(t, 1, called)
}

func TestUnit_Server_WhenReadDataCallbackIndicatesToCloseTheConnection_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTestServerConfig(7004)
	var called int
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) bool {
		called++
		return false
	}

	s, err := NewServer(config, logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)
	cancellable, cancel := context.WithCancel(context.Background())

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, err := net.Dial("tcp", ":7004")
	assert.Nil(t, err, "Actual err: %v", err)

	n, err := conn.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err)

	// Wait long enough for the read timeout to expire.
	time.Sleep(1100 * time.Millisecond)
	assertConnectionIsClosed(t, conn)

	cancel()
	wg.Wait()

	require.Nil(t, *serverErr, "Actual err: %v", *serverErr)
	assert.Equal(t, 1, called)
}

func TestUnit_Server_WhenDataReadCallbackPanics_ExpectServerDoesNotCrash(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTestServerConfig(7006)

	var called atomic.Int32
	doPanic := true
	var actual []byte
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) bool {
		called.Add(1)
		if doPanic {
			doPanic = !doPanic
			panic(errSample)
		}
		actual = data
		return true
	}
	s, err := NewServer(config, logger.New(os.Stdout))
	assert.Nil(t, err, "Actual err: %v", err)

	wgServer, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	// First attempt panics, the connection should be closed.
	conn, err := net.Dial("tcp", ":7006")
	assert.Nil(t, err, "Actual err: %v", err)

	n, err := conn.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err)

	// Wait long enough for the read timeout to expire.
	time.Sleep(1100 * time.Millisecond)
	assertConnectionIsClosed(t, conn)

	// The second attempt does not, expect to be able to write data.
	conn, err = net.Dial("tcp", ":7006")
	assert.Nil(t, err, "Actual err: %v", err)

	n, err = conn.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err)

	time.Sleep(100 * time.Millisecond)
	assertConnectionIsStillOpen(t, conn)

	conn.Close()

	cancel()
	wgServer.Wait()

	require.Nil(t, *serverErr, "Actual err: %v", *serverErr)
	assert.Equal(t, sampleData, actual)
}

func newTestServerConfig(port uint16) Configuration {
	conf := DefaultConfig()
	conf.Port = port
	return conf
}
