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

func TestUnit_Server_StartAndStopWithContext(t *testing.T) {
	s := NewServer(newTestServerConfig(6000), logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	asyncCancelContext(reasonableWaitTimeForServerToBeUp, cancel)

	err := s.Start(cancellable)

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_Server_WhenStartedMultipleTimes_ExpectFailure(t *testing.T) {
	s := NewServer(newTestServerConfig(6001), logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	// Start the first server: it should run without error.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.Start(cancellable)
		assert.Nil(t, err)
	}()

	// Wait a bit to be sure the first one is up.
	time.Sleep(100 * time.Millisecond)

	// Start the second server: this should fail.
	err := s.Start(context.Background())
	assert.True(t, errors.IsErrorWithCode(err, ErrAlreadyListening))

	cancel()
	wg.Wait()
}

func TestUnit_Server_WhenPortIsNotFree_ExpectStartReturnsInitializationFailure(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	log := logger.New(os.Stdout)
	s1 := NewServer(newTestServerConfig(6002), log)
	s2 := NewServer(newTestServerConfig(6002), log)

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
	// has time to fail
	time.Sleep(500 * time.Millisecond)

	cancel()
	wg.Wait()

	assert.Nil(t, err1, "Actual err: %v", err1)
	assert.True(t, errors.IsErrorWithCode(err2, ErrTcpInitialization), "Actual err: %v", err2)
}

func TestUnit_Server_WhenServerIsStopped_ExpectConnectionToNotBeClosed(t *testing.T) {
	s := NewServer(newTestServerConfig(6008), logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, err := net.Dial("tcp", ":6008")
	assert.Nil(t, err, "Actual err: %v", err)

	cancel()
	wg.Wait()

	require.Nil(t, *serverErr, "Actual err: %v", *serverErr)
	assertConnectionIsStillOpen(t, conn)
}

func TestUnit_Server_ConnectDisconnect(t *testing.T) {
	s := NewServer(newTestServerConfig(6003), logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, dialErr := net.Dial("tcp", ":6003")
	assert.Nil(t, dialErr, "Unexpected dial error: %v", dialErr)

	conn.Close()

	cancel()
	wg.Wait()

	require.Nil(t, *serverErr, "Actual err: %v", *serverErr)
}

func TestUnit_Server_OnConnect_ExpectCallbackNotified(t *testing.T) {
	config := newTestServerConfig(6004)
	var called int
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		called++
	}

	s := NewServer(config, logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, dialErr := net.Dial("tcp", ":6004")
	assert.Nil(t, dialErr, "Unexpected dial error: %v", dialErr)

	conn.Close()

	cancel()
	wg.Wait()

	assert.Nil(t, *serverErr, "Actual err: %v", *serverErr)
	assert.Equal(t, 1, called)
}

func TestUnit_Server_WhenOnConnectCallbackSucceeds_ExpectConnectionToStayOpen(t *testing.T) {
	config := newTestServerConfig(6005)
	var called int
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		called++
	}

	s := NewServer(config, logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, dialErr := net.Dial("tcp", ":6005")
	assert.Nil(t, dialErr, "Unexpected dial error: %v", dialErr)

	assertConnectionIsStillOpen(t, conn)

	cancel()
	wg.Wait()

	assert.Nil(t, *serverErr, "Actual err: %v", *serverErr)
	assert.Equal(t, 1, called)
}

func TestUnit_Server_WhenOnConnectCallbackFails_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTestServerConfig(6006)
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		panic(errSample)
	}

	s := NewServer(config, logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	conn, dialErr := net.Dial("tcp", ":6006")
	assert.Nil(t, dialErr, "Unexpected dial error: %v", dialErr)

	assertConnectionIsClosed(t, conn)

	cancel()
	wg.Wait()

	assert.Nil(t, *serverErr, "Actual err: %v", *serverErr)
}

func TestUnit_Server_WhenOnConnectCallbackFails_ExpectServerStillAcceptsOtherConnections(t *testing.T) {
	config := newTestServerConfig(6009)
	var called int
	doPanic := true
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		called++
		if doPanic {
			doPanic = false
			panic(errSample)
		}
	}

	s := NewServer(config, logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	// First connection should panic and be closed
	conn, err := net.Dial("tcp", ":6009")
	assert.Nil(t, err, "Unexpected dial error: %v", err)
	time.Sleep(50 * time.Millisecond)
	assertConnectionIsClosed(t, conn)

	// Second connection should work
	conn, err = net.Dial("tcp", ":6009")
	assert.Nil(t, err, "Unexpected dial error: %v", err)
	time.Sleep(50 * time.Millisecond)
	assertConnectionIsStillOpen(t, conn)

	cancel()
	wg.Wait()

	assert.Nil(t, *serverErr, "Actual err: %v", *serverErr)
}

func TestUnit_Server_WhenOnConnectTakesLong_ExpectOtherConnectionsCanBeProcessedConcurrently(t *testing.T) {
	config := newTestServerConfig(6007)
	var called atomic.Int32
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		fmt.Printf("onConnect called\n")
		time.Sleep(200 * time.Millisecond)
		fmt.Printf("onConnect called and done\n")
		called.Add(1)
	}

	s := NewServer(config, logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg, serverErr := asyncRunServerAndWaitForItToBeUp(t, s, cancellable)

	wgConn1 := asyncOpenConnectionAndCloseIt(t, 6007)
	wgConn2 := asyncOpenConnectionAndCloseIt(t, 6007)

	start := time.Now()

	wgConn2.Wait()
	wgConn1.Wait()

	// Wait for connections to be processed before closing
	time.Sleep(100 * time.Millisecond)

	cancel()
	wg.Wait()

	elapsed := time.Since(start)

	require.Nil(t, *serverErr, "Actual err: %v", *serverErr)
	assert.Equal(t, int32(2), called.Load())
	// If this would not be concurrent we would expect at least 400ms of
	// processing time for the callbacks.
	assert.LessOrEqual(t, elapsed, 300*time.Millisecond)
}

func newTestServerConfig(port uint16) Config {
	return Config{
		Port: port,
	}
}
