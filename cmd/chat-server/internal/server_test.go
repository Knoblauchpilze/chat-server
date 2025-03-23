package internal

import (
	"context"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestIT_RunServer_StartAndStopWithContext(t *testing.T) {
	conf := newTestServerConfig(7300, 7301)
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncListenAndServe(t, conf, cancellable)

	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()
}

func TestIT_RunServer_CanConnectOnHttpPort(t *testing.T) {
	conf := newTestServerConfig(7302, 7303)
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncListenAndServe(t, conf, cancellable)

	url := "http://localhost:7303/v1/chats/healthcheck"
	rw := doRequest(t, http.MethodGet, url)
	assert.Equal(t, http.StatusOK, rw.StatusCode)
	assertResponseContainsDetails(t, rw, success, `"OK"`)

	cancel()
	wg.Wait()
}

func TestIT_RunServer_CanConnectOnTcpPort(t *testing.T) {
	conf := newTestServerConfig(7304, 7305)
	cancellable, cancel := context.WithCancel(context.Background())

	wg := asyncListenAndServe(t, conf, cancellable)

	conn, err := net.Dial("tcp", ":7304")
	assert.Nil(t, err, "Actual err: %v", err)

	time.Sleep(100 * time.Millisecond)
	assertConnectionIsStillOpen(t, conn)

	cancel()
	wg.Wait()

	assertConnectionIsClosed(t, conn)
}

func newTestServerConfig(tcpPort uint16, httpPort uint16) Configuration {
	baseConfig := DefaultConfig()
	baseConfig.Server.Port = httpPort

	return Configuration{
		Server:   baseConfig.Server,
		TcpPort:  tcpPort,
		Database: dbTestConfig,
	}
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
		err = RunServer(ctx, config, logger.New(os.Stdout))
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	time.Sleep(reasonableWaitTimeForServerToBeUp)

	return &wg
}
