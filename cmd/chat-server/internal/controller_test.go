package internal

import (
	"context"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_ListenAndServe_StartAndStopWithContext(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	asyncCancelContext(100*time.Millisecond, cancel)

	err := listenAndServe(cancellable, newTestConfig(7000), logger.New(os.Stdout))

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_ListenAndServe_WhenServerIsStopped_ExpectClientConnectionToBeClosed(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTestConfig(7001)
	wg := asyncListenAndServe(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7001")
	assert.Nil(t, err, "Actual err: %v", err)

	cancel()
	wg.Wait()

	assertConnectionIsClosed(t, conn)
}

func TestUnit_ListenAndServe_WhenClientConnects_ExpectCallbackNotified(t *testing.T) {
	config := newTestConfig(7002)
	var called int
	config.Callbacks.ConnectCallback = func(uuid.UUID, net.Conn) bool {
		called++
		return true
	}
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncListenAndServe(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7002")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	conn.Close()

	cancel()
	wg.Wait()

	assert.Equal(t, 1, called)
}

func TestUnit_ListenAndServe_WhenClientSendsData_ExpectCallbackNotified(t *testing.T) {
	config := newTestConfig(7003)
	var called int
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) bool {
		called++
		return true
	}
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncListenAndServe(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7003")
	assert.Nil(t, err, "Actual err: %v", err)

	n, err := conn.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err, "Actual err: %v", err)

	conn.Close()

	cancel()
	wg.Wait()

	assert.Equal(t, 1, called)
}

func TestUnit_ListenAndServe_WhenClientDisconnects_ExpectCallbackNotified(t *testing.T) {
	config := newTestConfig(7004)
	var called int
	config.Callbacks.DisconnectCallback = func(uuid.UUID) {
		called++
	}
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncListenAndServe(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7004")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	conn.Close()

	cancel()
	wg.Wait()

	assert.Equal(t, 1, called)
}

func TestUnit_ListenAndServe_WhenClientConnectsAndIsDenied_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTestConfig(7005)
	var called int
	config.Callbacks.ConnectCallback = func(uuid.UUID, net.Conn) bool {
		called++
		return false
	}
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncListenAndServe(t, config, cancellable)

	conn, dialErr := net.Dial("tcp", ":7005")
	assert.Nil(t, dialErr, "Actual err: %v", dialErr)

	// Wait for connection to be processed.
	time.Sleep(100 * time.Millisecond)

	assertConnectionIsClosed(t, conn)

	cancel()
	wg.Wait()

	assert.Equal(t, 1, called)
}

func TestUnit_ListenAndServe_WhenReadDataCallbackIndicatesToCloseTheConnection_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTestConfig(7006)
	var called int
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) bool {
		called++
		return false
	}
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncListenAndServe(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7006")
	assert.Nil(t, err, "Actual err: %v", err)

	n, err := conn.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err, "Actual err: %v", err)

	// Wait long enough for the read timeout to expire and connection
	// to be effectively closed.
	time.Sleep(1100 * time.Millisecond)
	assertConnectionIsClosed(t, conn)

	cancel()
	wg.Wait()

	assert.Equal(t, 1, called)
}

func TestUnit_ListenAndServe_WhenDataReadCallbackPanics_ExpectServerDoesNotCrash(t *testing.T) {
	config := newTestConfig(7007)
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
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncListenAndServe(t, config, cancellable)

	// First attempt panics, the connection should be closed.
	conn, err := net.Dial("tcp", ":7007")
	assert.Nil(t, err, "Actual err: %v", err)

	n, err := conn.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err, "Actual err: %v", err)

	// Wait long enough for the read timeout to expire and connection
	// to be effectively closed.
	time.Sleep(1100 * time.Millisecond)
	assertConnectionIsClosed(t, conn)

	// The second attempt does not, expect to be able to write data.
	conn, err = net.Dial("tcp", ":7007")
	assert.Nil(t, err, "Actual err: %v", err)

	n, err = conn.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err, "Actual err: %v", err)

	time.Sleep(100 * time.Millisecond)
	assertConnectionIsStillOpen(t, conn)

	conn.Close()

	cancel()
	wg.Wait()

	assert.Equal(t, sampleData, actual)
}

func newTestConfig(port uint16) Configuration {
	conf := DefaultConfig()
	conf.Port = port
	return conf
}

func asyncListenAndServe(
	t *testing.T,
	config Configuration,
	ctx context.Context,
) *sync.WaitGroup {
	var err error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if panicErr := recover(); panicErr != nil {
				assert.Failf(t, "Server panicked", "Panic details: %v", panicErr)
			}
		}()
		err = listenAndServe(ctx, config, logger.New(os.Stdout))
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(reasonableWaitTimeForServerToBeUp)

	return &wg
}
