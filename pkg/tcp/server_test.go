package tcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const reasonableWaitTimeForServerToBeUp = 500 * time.Millisecond

func TestUnit_Server_Start_StopWithContext(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	s := NewServer(newTestServerConfig(6000), logger.New(os.Stdout))

	asyncCancelContext(reasonableWaitTimeForServerToBeUp, cancel)

	err := s.Start(cancellable)

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_Server_Start_WhenPortIsNotFree_ExpectInitializationFailure(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	log := logger.New(os.Stdout)
	s1 := NewServer(newTestServerConfig(6001), log)
	s2 := NewServer(newTestServerConfig(6001), log)

	var err1, err2 error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		err1 = s1.Start(cancellable)
	}()
	go func() {
		defer wg.Done()
		// Wait for the first server to be up
		time.Sleep(reasonableWaitTimeForServerToBeUp)
		err2 = s2.Start(context.Background())
	}()

	// Wait a bit longer to be sure that the second server
	// has the time to fail
	time.Sleep(3 * reasonableWaitTimeForServerToBeUp)

	cancel()
	wg.Wait()

	assert.Nil(t, err1, "Actual err: %v", err1)
	assert.True(t, errors.IsErrorWithCode(err2, ErrTcpInitialization), "Actual err: %v", err2)
}

func TestUnit_Server_ConnectDisconnect(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	s := NewServer(newTestServerConfig(6002), logger.New(os.Stdout))

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	openConnectionAndCloseIt(t, 6002)

	cancel()
	wg.Wait()

	require.Nil(t, *serverErr, "Actual err: %v", *serverErr)
}

func TestUnit_Server_WhenServerStopped_ExpectClosesBeforeConnectionCloses(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	s := NewServer(newTestServerConfig(6003), logger.New(os.Stdout))

	wgServer, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, dialErr := net.Dial("tcp", ":6003")
	assert.Nil(t, dialErr, "Actual err: %v", dialErr)

	var connClosed atomic.Bool

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer connClosed.Store(true)

		// Wait long enough for cancel to be called in the main thread
		// and shutdown timeout of the server to be reached
		time.Sleep(5 * time.Second)

		err := conn.Close()
		assert.Nil(t, err)
	}()

	cancel()
	wgServer.Wait()
	assert.False(t, connClosed.Load())
	wg.Wait()

	require.Nil(t, *serverErr, "Actual err: %v", *serverErr)
}

func TestUnit_Server_OnConnect_ExpectCallbackToBeCalled(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	config := newTestServerConfig(6004)
	var called int
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		called++
	}
	s := NewServer(config, logger.New(os.Stdout))

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	openConnectionAndCloseIt(t, 6004)

	cancel()
	wg.Wait()

	require.Equal(t, 1, called)
	require.Nil(t, *serverErr, "Actual err: %v", *serverErr)
}

func newTestServerConfig(port uint16) Config {
	return Config{
		Port: port,
	}
}

func asyncCancelContext(delay time.Duration, cancel context.CancelFunc) {
	go func() {
		time.Sleep(delay)
		cancel()
	}()
}

func asyncRunServerAndWaitForItToBeUp(t *testing.T, s Server, ctx context.Context) (*sync.WaitGroup, *error) {
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
		err = s.Start(ctx)
	}()

	time.Sleep(reasonableWaitTimeForServerToBeUp)

	return &wg, &err
}

func openConnectionAndCloseIt(t *testing.T, port uint16) {
	openConnectionAndSendData(t, port, nil)
}

func openConnectionAndSendData(t *testing.T, port uint16, data []byte) {
	address := fmt.Sprintf(":%d", port)
	conn, dialErr := net.Dial("tcp", address)
	assert.Nil(t, dialErr, "Unexpected dial error: %v", dialErr)

	if data != nil {
		_, err := conn.Write(sampleData)
		assert.Nil(t, err, "Unexpected error while sending data: %v", err)
	}

	closeErr := conn.Close()
	assert.Nil(t, closeErr, "Unexpected close error: %v", closeErr)

	const reasonableWaitTimeForConnectionToBeProcessed = 50 * time.Millisecond
	time.Sleep(reasonableWaitTimeForConnectionToBeProcessed)
}
