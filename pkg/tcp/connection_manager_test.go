package tcp

import (
	"os"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const reasonableReadTimeout = 100 * time.Millisecond

func TestUnit_ConnectionManager_WhenCloseIsCalled_ExpectClientConnectionToBeClosed(t *testing.T) {
	cm := NewConnectionManager(newTestManagerConfig(), logger.New(os.Stdout))

	client, server := newTestConnection(t, 7000)

	cm.OnClientConnected(server)
	cm.Close()

	time.Sleep(50 * time.Millisecond)

	assertConnectionIsClosed(t, client)
}

func TestUnit_ConnectionManager_WhenClientConnects_ExpectCallbackNotified(t *testing.T) {
	config := newTestManagerConfig()
	var called int
	config.Callbacks.ConnectCallback = func(id uuid.UUID, address string) bool {
		called++
		return true
	}
	cm := NewConnectionManager(config, logger.New(os.Stdout))

	_, server := newTestConnection(t, 7001)

	cm.OnClientConnected(server)

	assert.Equal(t, 1, called)
}

func TestUnit_ConnectionManager_WhenClientSendsData_ExpectCallbackNotified(t *testing.T) {
	config := newTestManagerConfig()
	var called int
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) bool {
		called++
		return true
	}

	cm := NewConnectionManager(config, logger.New(os.Stdout))

	client, server := newTestConnection(t, 7002)

	cm.OnClientConnected(server)

	n, err := client.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err)

	// Wait for the processing to happen.
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, called)
}

func TestUnit_ConnectionManager_WhenClientConnectsAndIsDenied_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTestManagerConfig()
	var called int
	config.Callbacks.ConnectCallback = func(id uuid.UUID, address string) bool {
		called++
		return false
	}

	cm := NewConnectionManager(config, logger.New(os.Stdout))

	client, server := newTestConnection(t, 7003)

	cm.OnClientConnected(server)

	// Wait for connection to be processed.
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, called)
	assertConnectionIsClosed(t, client)
}

func TestUnit_ConnectionManager_WhenReadDataCallbackIndicatesToCloseTheConnection_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTestManagerConfig()
	var called int
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) bool {
		called++
		return false
	}

	cm := NewConnectionManager(config, logger.New(os.Stdout))

	client, server := newTestConnection(t, 7004)

	cm.OnClientConnected(server)

	n, err := client.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err)

	// Wait long enough for the read timeout to expire.
	time.Sleep(100 * time.Millisecond)

	assertConnectionIsClosed(t, client)
}

func TestUnit_ConnectionManager_WhenClientDisconnects_ExpectCallbackNotified(t *testing.T) {
	config := newTestManagerConfig()
	var called int
	config.Callbacks.DisconnectCallback = func(id uuid.UUID) {
		called++
	}

	cm := NewConnectionManager(config, logger.New(os.Stdout))

	client, server := newTestConnection(t, 7005)

	cm.OnClientConnected(server)
	client.Close()

	// Wait a bit for the processing to happen.
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, called)
}

func TestUnit_ConnectionManager_WhenDataReadCallbackPanics_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTestManagerConfig()
	var called int
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) bool {
		called++
		panic(errSample)
	}

	cm := NewConnectionManager(config, logger.New(os.Stdout))

	client, server := newTestConnection(t, 7006)

	cm.OnClientConnected(server)

	n, err := client.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err)

	// Wait long enough for the read timeout to expire.
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 1, called)
	assertConnectionIsClosed(t, client)
}

func newTestManagerConfig() ManagerConfig {
	return ManagerConfig{
		ReadTimeout: reasonableReadTimeout,
	}
}
