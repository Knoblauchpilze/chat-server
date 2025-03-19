package tcp

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const reasonableReadTimeout = 100 * time.Millisecond

func TestUnit_ConnectionManager_WhenCloseIsCalled_ExpectClientConnectionToBeClosed(t *testing.T) {
	cm := newConnectionManager(newTestManagerConfig(), logger.New(os.Stdout))

	client, server := newTestConnection(t, 5100)

	cm.OnClientConnected(server)
	// Wait for the connection to be processed
	time.Sleep(50 * time.Millisecond)

	cm.Close()

	time.Sleep(50 * time.Millisecond)

	assertConnectionIsClosed(t, client)
}

func TestUnit_ConnectionManager_WhenCloseIsCalled_ExpectOnConnectDeniesConnection(t *testing.T) {
	cm := newConnectionManager(newTestManagerConfig(), logger.New(os.Stdout))
	cm.Close()
	client, server := newTestConnection(t, 5101)

	cm.OnClientConnected(server)

	// Wait for the connection to be processed
	time.Sleep(100 * time.Millisecond)

	assertConnectionIsClosed(t, client)
}

func TestUnit_ConnectionManager_WhenCloseIsCalled_ExpectOnDisconnectToBeCalledOnlyOnce(t *testing.T) {
	config := newTestManagerConfig()
	var called int
	config.Callbacks.DisconnectCallback = func(id uuid.UUID) {
		called++
	}
	cm := newConnectionManager(config, logger.New(os.Stdout))

	_, server := newTestConnection(t, 5100)

	cm.OnClientConnected(server)
	// Wait for the connection to be processed
	time.Sleep(50 * time.Millisecond)

	cm.Close()

	assert.Equal(t, 1, called)
}

func TestUnit_ConnectionManager_WhenClientConnects_ExpectCallbackNotified(t *testing.T) {
	config := newTestManagerConfig()
	called := make(chan struct{}, 1)
	config.Callbacks.ConnectCallback = func(id uuid.UUID, conn net.Conn) bool {
		defer func() {
			called <- struct{}{}
		}()
		return true
	}
	cm := newConnectionManager(config, logger.New(os.Stdout))

	_, server := newTestConnection(t, 5102)

	cm.OnClientConnected(server)

	<-called
}

func TestUnit_ConnectionManager_WhenClientSendsData_ExpectCallbackNotified(t *testing.T) {
	config := newTestManagerConfig()
	called := make(chan struct{}, 1)
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) (int, bool) {
		defer func() {
			called <- struct{}{}
		}()
		return 0, true
	}

	cm := newConnectionManager(config, logger.New(os.Stdout))

	client, server := newTestConnection(t, 5103)

	cm.OnClientConnected(server)

	n, err := client.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err, "Actual err: %v", err)

	<-called
}

func TestUnit_ConnectionManager_WhenClientConnectsAndIsDenied_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTestManagerConfig()
	called := make(chan struct{}, 1)
	config.Callbacks.ConnectCallback = func(id uuid.UUID, conn net.Conn) bool {
		defer func() {
			called <- struct{}{}
		}()
		return false
	}

	cm := newConnectionManager(config, logger.New(os.Stdout))

	client, server := newTestConnection(t, 5104)

	cm.OnClientConnected(server)

	<-called

	assertConnectionIsClosed(t, client)
}

func TestUnit_ConnectionManager_WhenReadDataCallbackIndicatesToCloseTheConnection_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTestManagerConfig()
	called := make(chan struct{}, 1)
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) (int, bool) {
		defer func() {
			called <- struct{}{}
		}()
		return len(data), false
	}

	cm := newConnectionManager(config, logger.New(os.Stdout))

	client, server := newTestConnection(t, 5105)

	cm.OnClientConnected(server)

	n, err := client.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err, "Actual err: %v", err)

	<-called
	// Wait long enough for the read timeout to expire and connection
	// to be effectively closed.
	time.Sleep(200 * time.Millisecond)

	assertConnectionIsClosed(t, client)
}

func TestUnit_ConnectionManager_WhenReadDataCallbackIndicatesToCloseTheConnection_ExpectDisconnectCallbackIsCalled(t *testing.T) {
	config := newTestManagerConfig()
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) (int, bool) {
		return len(data), false
	}
	called := make(chan struct{}, 1)
	config.Callbacks.DisconnectCallback = func(id uuid.UUID) {
		called <- struct{}{}
	}

	cm := newConnectionManager(config, logger.New(os.Stdout))

	client, server := newTestConnection(t, 5105)

	cm.OnClientConnected(server)

	n, err := client.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err, "Actual err: %v", err)

	<-called
}

func TestUnit_ConnectionManager_WhenClientDisconnects_ExpectCallbackNotified(t *testing.T) {
	config := newTestManagerConfig()
	called := make(chan struct{}, 1)
	config.Callbacks.DisconnectCallback = func(id uuid.UUID) {
		called <- struct{}{}
	}

	cm := newConnectionManager(config, logger.New(os.Stdout))

	client, server := newTestConnection(t, 5106)

	cm.OnClientConnected(server)
	client.Close()

	<-called
}

func TestUnit_ConnectionManager_WhenDataReadCallbackPanics_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTestManagerConfig()
	called := make(chan struct{}, 1)
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) (int, bool) {
		defer func() {
			called <- struct{}{}
		}()
		panic(errSample)
	}

	cm := newConnectionManager(config, logger.New(os.Stdout))

	client, server := newTestConnection(t, 5107)

	cm.OnClientConnected(server)

	n, err := client.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err, "Actual err: %v", err)

	<-called
	// Wait long enough for the read timeout to expire and connection
	// to be effectively closed.
	time.Sleep(200 * time.Millisecond)

	assertConnectionIsClosed(t, client)
}

func newTestManagerConfig() managerConfig {
	return managerConfig{
		ReadTimeout: reasonableReadTimeout,
	}
}
