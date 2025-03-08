package service

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const reasonableWaitTimeForServerToBeUp = 50 * time.Millisecond

func asyncRunMessageProcessingService(
	t *testing.T, service MessageProcessingService,
) *sync.WaitGroup {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if panicErr := recover(); panicErr != nil {
				assert.Failf(t, "Server panicked", "Panic details: %v", panicErr)
			}
		}()
		service.Start()
	}()

	time.Sleep(reasonableWaitTimeForServerToBeUp)

	return &wg
}
