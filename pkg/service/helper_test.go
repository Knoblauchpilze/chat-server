package service

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const reasonableWaitTimeForServerToBeUp = 400 * time.Millisecond

func asyncRunService(
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
