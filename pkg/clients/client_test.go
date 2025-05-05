package clients

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Client_CorrectlySetsSseHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	client, err := New(1, uuid.New(), rec)
	assert.Nil(t, err, "Actual err: %v", err)

	wg := asyncStartClientAndAssertNoError(t, client)

	err = client.Stop()
	wg.Wait()
	assert.Nil(t, err, "Actual err: %v", err)

	headers := rec.Header()
	header, ok := headers["Content-Type"]
	assert.Truef(t, ok, "Expected Content-Type header to be set: %v", headers)
	assert.Equal(t, []string{"text/event-stream"}, header)
	header, ok = headers["Cache-Control"]
	assert.Truef(t, ok, "Expected Cache-Control header to be set: %v", headers)
	assert.Equal(t, []string{"no-cache"}, header)
	header, ok = headers["Connection"]
	assert.Truef(t, ok, "Expected Connection to be set: %v", headers)
	assert.Equal(t, []string{"keep-alive"}, header)
}

func TestUnit_Client_SendsMessageAsSse(t *testing.T) {
	rec := httptest.NewRecorder()
	client, err := New(1, uuid.New(), rec)
	assert.Nil(t, err, "Actual err: %v", err)

	wg := asyncStartClientAndAssertNoError(t, client)

	msg := persistence.Message{
		Id:        uuid.MustParse("8f102c70-8eba-4094-bd4d-7f70d71b21f2"),
		ChatUser:  uuid.MustParse("f2d9ce22-179d-431c-b63d-43d5a8ab5e18"),
		Room:      uuid.MustParse("111838db-a871-47be-9149-c974fd356316"),
		Message:   "Hello",
		CreatedAt: time.Date(2025, 5, 4, 20, 56, 16, 0, time.UTC),
	}
	client.Enqueue(msg)

	// Wait for the message to be sent
	time.Sleep(50 * time.Millisecond)

	err = client.Stop()
	wg.Wait()
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusOK, rec.Code)
	actual := rec.Body.String()
	expected := `id: 8f102c70-8eba-4094-bd4d-7f70d71b21f2
data: {"id":"8f102c70-8eba-4094-bd4d-7f70d71b21f2","user":"f2d9ce22-179d-431c-b63d-43d5a8ab5e18","room":"111838db-a871-47be-9149-c974fd356316","message":"Hello","created_at":"2025-05-04T20:56:16Z"}

`
	assert.Equal(t, expected, actual)
}

func asyncStartClientAndAssertNoError(
	t *testing.T, client Client,
) *sync.WaitGroup {
	return asyncStartClientAndAssertError(t, client, nil)
}

func asyncStartClientAndAssertError(
	t *testing.T, client Client, expectedErr error,
) *sync.WaitGroup {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Processor panicked: %v", r)
			}
		}()

		err := client.Start()
		assert.Equal(t, expectedErr, err, "Actual err: %v", err)
	}()

	// Wait a bit for the client to start
	time.Sleep(50 * time.Millisecond)

	return &wg
}
