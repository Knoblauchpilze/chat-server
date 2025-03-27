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
	"github.com/Knoblauchpilze/chat-server/pkg/clients"
	"github.com/Knoblauchpilze/chat-server/pkg/messages"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_TcpListenAndServe_StartAndStopWithContext(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	asyncCancelContext(100*time.Millisecond, cancel)

	err := tcpListenAndServe(cancellable, newTcpTestConfig(7000), logger.New(os.Stdout))

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_TcpListenAndServe_WhenServerIsStopped_ExpectClientConnectionToBeClosed(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTcpTestConfig(7001)
	wg := asyncTcpListenAndServe(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7001")
	assert.Nil(t, err, "Actual err: %v", err)

	cancel()
	wg.Wait()

	assertConnectionIsClosed(t, conn)
}

func TestUnit_TcpListenAndServe_WhenClientConnects_ExpectCallbackNotified(t *testing.T) {
	config := newTcpTestConfig(7002)
	called := make(chan struct{}, 1)
	config.Callbacks.ConnectCallback = func(net.Conn) (bool, uuid.UUID) {
		called <- struct{}{}
		return true, uuid.Nil
	}
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncTcpListenAndServe(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7002")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	conn.Close()

	cancel()
	wg.Wait()

	<-called
}

func TestUnit_TcpListenAndServe_WhenClientSendsData_ExpectCallbackNotified(t *testing.T) {
	config := newTcpTestConfig(7003)
	called := make(chan struct{}, 1)
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) (int, bool) {
		called <- struct{}{}
		return len(data), true
	}
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncTcpListenAndServe(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7003")
	assert.Nil(t, err, "Actual err: %v", err)

	n, err := conn.Write(sampleData)
	assert.Equal(t, n, len(sampleData))
	assert.Nil(t, err, "Actual err: %v", err)

	conn.Close()

	cancel()
	wg.Wait()

	<-called
}

func TestUnit_TcpListenAndServe_WhenClientDisconnects_ExpectCallbackNotified(t *testing.T) {
	config := newTcpTestConfig(7004)
	called := make(chan struct{}, 1)
	config.Callbacks.DisconnectCallback = func(uuid.UUID) {
		called <- struct{}{}
	}
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncTcpListenAndServe(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7004")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	conn.Close()

	cancel()
	wg.Wait()

	<-called
}

func TestUnit_TcpListenAndServe_WhenClientConnectsAndIsDenied_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTcpTestConfig(7005)
	config.Callbacks.ConnectCallback = func(net.Conn) (bool, uuid.UUID) {
		return false, uuid.Nil
	}
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncTcpListenAndServe(t, config, cancellable)

	conn, dialErr := net.Dial("tcp", ":7005")
	assert.Nil(t, dialErr, "Actual err: %v", dialErr)

	// Wait for connection to be processed.
	time.Sleep(100 * time.Millisecond)

	assertConnectionIsClosed(t, conn)

	cancel()
	wg.Wait()
}

func TestUnit_TcpListenAndServe_WhenReadDataCallbackIndicatesToCloseTheConnection_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTcpTestConfig(7006)
	var called int
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) (int, bool) {
		called++
		return len(data), false
	}
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncTcpListenAndServe(t, config, cancellable)

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

func TestUnit_TcpListenAndServe_WhenDataReadCallbackPanics_ExpectServerDoesNotCrash(t *testing.T) {
	config := newTcpTestConfig(7007)
	var called atomic.Int32
	doPanic := true
	var actual []byte
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) (int, bool) {
		called.Add(1)
		if doPanic {
			doPanic = !doPanic
			panic(errSample)
		}
		actual = data
		return len(data), true
	}
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncTcpListenAndServe(t, config, cancellable)

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

func TestUnit_TcpListenAndServe_WhenMessageIsSentInMultiplePieces_ExpectItToCorrectlyBeReceived(t *testing.T) {
	config := newTcpTestConfig(7008)
	var called atomic.Int32
	var receivedMessage messages.Message
	config.Callbacks.ReadDataCallback = func(id uuid.UUID, data []byte) (int, bool) {
		called.Add(1)

		msg, n, err := messages.Decode(data)
		if err != nil {
			return 0, true
		}

		receivedMessage = msg
		return n, true
	}
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncTcpListenAndServe(t, config, cancellable)

	conn, err := net.Dial("tcp", ":7008")
	assert.Nil(t, err, "Actual err: %v", err)

	clientId1 := uuid.New()
	clientId2 := uuid.New()
	msg := messages.NewDirectMessage(clientId1, clientId2, "hello")
	data, err := messages.Encode(msg)
	assert.Nil(t, err, "Actual err: %v", err)

	// Send first 10 bytes of the message
	assert.Greater(t, len(data), 10)
	n, err := conn.Write(data[:10])
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, 10, n)

	// Wait for the callback to be processed
	time.Sleep(100 * time.Millisecond)

	// Send the rest of the data
	n, err = conn.Write(data[10:])
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, len(data)-10, n)

	// Wait for the callback to be processed
	time.Sleep(100 * time.Millisecond)

	assertConnectionIsStillOpen(t, conn)

	conn.Close()

	cancel()
	wg.Wait()

	assert.GreaterOrEqual(t, called.Load(), int32(2))
	actual, err := messages.ToMessageStruct[messages.DirectMessage](receivedMessage)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, clientId1, actual.Emitter)
	assert.Equal(t, clientId2, actual.Receiver)
	assert.Equal(t, "hello", actual.Content)
}

func newTcpTestConfig(port uint16) Configuration {
	conf := DefaultConfig()
	conf.TcpPort = port
	conf.Callbacks = clients.Callbacks{
		ConnectCallback: func(net.Conn) (bool, uuid.UUID) {
			return true, uuid.Nil
		},
	}
	return conf
}

func asyncTcpListenAndServe(
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
		err = tcpListenAndServe(ctx, config, logger.New(os.Stdout))
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(reasonableWaitTimeForServerToBeUp)

	return &wg
}
