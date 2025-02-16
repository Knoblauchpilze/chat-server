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
)

func TestUnit_ConnectionAcceptor_ListenWithContext(t *testing.T) {
	ca := NewConnectionAcceptor(newTestAcceptorConfig(6000), logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	asyncCancelContext(reasonableWaitTimeForAcceptorToBeUp, cancel)

	err := ca.Listen(cancellable)

	assert.Nil(t, err, "Actual err: %v", err)
}

func TestUnit_ConnectionAcceptor_WhenStartedMultipleTimes_ExpectFailure(t *testing.T) {
	ca := NewConnectionAcceptor(newTestAcceptorConfig(6001), logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	// Start the first acceptor: it should run without error.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := ca.Listen(cancellable)
		assert.Nil(t, err)
	}()

	// Wait a bit to be sure the first one is up.
	time.Sleep(100 * time.Millisecond)

	// Start the second acceptor: this should fail.
	err := ca.Listen(context.Background())
	assert.True(t, errors.IsErrorWithCode(err, ErrAlreadyListening))

	cancel()
	wg.Wait()
}

func TestUnit_ConnectionAcceptor_WhenPortIsNotFree_ExpectStartReturnsInitializationFailure(t *testing.T) {
	cancellable, cancel := context.WithCancel(context.Background())
	log := logger.New(os.Stdout)
	ca1 := NewConnectionAcceptor(newTestAcceptorConfig(6002), log)
	ca2 := NewConnectionAcceptor(newTestAcceptorConfig(6002), log)

	var err1, err2 error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		err1 = ca1.Listen(cancellable)
	}()

	// Wait for the first acceptor to be up.
	time.Sleep(100 * time.Millisecond)

	go func() {
		defer wg.Done()
		err2 = ca2.Listen(context.Background())
	}()

	// Wait a bit longer to be sure that the second acceptor
	// has time to fail
	time.Sleep(500 * time.Millisecond)

	cancel()
	wg.Wait()

	assert.Nil(t, err1, "Actual err: %v", err1)
	assert.True(t, errors.IsErrorWithCode(err2, ErrTcpInitialization), "Actual err: %v", err2)
}

func TestUnit_ConnectionAcceptor_WhenAcceptorIsStopped_ExpectConnectionToNotBeClosed(t *testing.T) {
	ca := NewConnectionAcceptor(newTestAcceptorConfig(6008), logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca, cancellable)

	conn, err := net.Dial("tcp", ":6008")
	assert.Nil(t, err, "Actual err: %v", err)

	cancel()
	wg.Wait()

	assertConnectionIsStillOpen(t, conn)
}

func TestUnit_ConnectionAcceptor_ConnectDisconnect(t *testing.T) {
	ca := NewConnectionAcceptor(newTestAcceptorConfig(6003), logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca, cancellable)

	conn, err := net.Dial("tcp", ":6003")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	conn.Close()

	cancel()
	wg.Wait()
}

func TestUnit_ConnectionAcceptor_OnConnect_ExpectCallbackNotified(t *testing.T) {
	config := newTestAcceptorConfig(6004)
	var called int
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		called++
	}

	ca := NewConnectionAcceptor(config, logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca, cancellable)

	conn, err := net.Dial("tcp", ":6004")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	conn.Close()

	cancel()
	wg.Wait()

	assert.Equal(t, 1, called)
}

func TestUnit_ConnectionAcceptor_WhenOnConnectCallbackSucceeds_ExpectConnectionToStayOpen(t *testing.T) {
	config := newTestAcceptorConfig(6005)
	var called int
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		called++
	}

	ca := NewConnectionAcceptor(config, logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca, cancellable)

	conn, err := net.Dial("tcp", ":6005")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	assertConnectionIsStillOpen(t, conn)

	cancel()
	wg.Wait()

	assert.Equal(t, 1, called)
}

func TestUnit_ConnectionAcceptor_WhenOnConnectCallbackFails_ExpectConnectionToBeClosed(t *testing.T) {
	config := newTestAcceptorConfig(6006)
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		panic(errSample)
	}

	ca := NewConnectionAcceptor(config, logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca, cancellable)

	conn, err := net.Dial("tcp", ":6006")
	assert.Nil(t, err, "Unexpected dial error: %v", err)

	// Allow the callback to be processed.
	time.Sleep(50 * time.Millisecond)

	fmt.Printf("checking connection is closed\n")
	assertConnectionIsClosed(t, conn)
	fmt.Printf("after connection check\n")

	cancel()
	wg.Wait()
}

func TestUnit_ConnectionAcceptor_WhenOnConnectCallbackFails_ExpectAcceptorStillAcceptsOtherConnections(t *testing.T) {
	config := newTestAcceptorConfig(6009)
	var called int
	doPanic := true
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		called++
		if doPanic {
			doPanic = false
			panic(errSample)
		}
	}

	ca := NewConnectionAcceptor(config, logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca, cancellable)

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
}

func TestUnit_ConnectionAcceptor_WhenOnConnectTakesLong_ExpectOtherConnectionsCanBeProcessedConcurrently(t *testing.T) {
	config := newTestAcceptorConfig(6007)
	var called atomic.Int32
	config.Callbacks.ConnectCallback = func(conn net.Conn) {
		fmt.Printf("onConnect called\n")
		time.Sleep(200 * time.Millisecond)
		fmt.Printf("onConnect called and done\n")
		called.Add(1)
	}

	ca := NewConnectionAcceptor(config, logger.New(os.Stdout))
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncRunAcceptorAndWaitForItToBeUp(t, ca, cancellable)

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

	assert.Equal(t, int32(2), called.Load())
	// If this would not be concurrent we would expect at least 400ms of
	// processing time for the callbacks.
	assert.LessOrEqual(t, elapsed, 300*time.Millisecond)
}

func newTestAcceptorConfig(port uint16) AcceptorConfig {
	return AcceptorConfig{
		Port: port,
	}
}
