package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/communication"
	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_SseEvent_NoContent(t *testing.T) {
	e := sseEvent{}

	rec := httptest.NewRecorder()
	err := e.send(rec)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusOK, rec.Code)
	body, err := io.ReadAll(rec.Body)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, []byte{}, body, "Actual body: %s", string(body))
}

func TestUnit_SseEvent_CommentWithoutData(t *testing.T) {
	e := sseEvent{
		Comment: []byte("my-comment"),
	}

	rec := httptest.NewRecorder()
	err := e.send(rec)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusOK, rec.Code)
	body, err := io.ReadAll(rec.Body)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := []byte(`: my-comment

`)
	assert.Equal(
		t,
		expected,
		body,
		"Expected %s, got: %s",
		string(expected),
		string(body),
	)
}

func TestUnit_SseEvent_DataWithNewLine(t *testing.T) {
	e := sseEvent{
		Data: []byte("line 1\nline 2"),
	}

	rec := httptest.NewRecorder()
	err := e.send(rec)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusOK, rec.Code)
	body, err := io.ReadAll(rec.Body)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := []byte(`id: 
data: line 1
data: line 2

`)
	assert.Equal(
		t,
		expected,
		body,
		"Expected %s, got: %s",
		string(expected),
		string(body),
	)
}

func TestUnit_SseEvent_FromMessage(t *testing.T) {
	msg := persistence.Message{
		Id:        uuid.New(),
		ChatUser:  uuid.New(),
		Room:      uuid.New(),
		Message:   "my-message",
		CreatedAt: time.Date(2025, 5, 4, 17, 54, 40, 0, time.UTC),
	}

	e, err := fromMessage(msg)
	assert.Nil(t, err, "Actual err: %v", err)

	response := communication.ToMessageDtoResponse(msg)
	expected, err := json.Marshal(response)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, expected, e.Data)
}

func TestUnit_SseEvent_WriteMessage(t *testing.T) {
	msg := persistence.Message{
		Id:        uuid.MustParse("3cdfb4ea-d372-443e-92c0-6eea2f7cd2f0"),
		ChatUser:  uuid.MustParse("a21b5378-9020-49b5-8021-8a059b3ecef4"),
		Room:      uuid.MustParse("1948785c-9981-47b8-b280-847f57810964"),
		Message:   "this is a message",
		CreatedAt: time.Date(2025, 5, 4, 17, 56, 45, 0, time.UTC),
	}

	e, err := fromMessage(msg)
	assert.Nil(t, err, "Actual err: %v", err)

	rec := httptest.NewRecorder()
	err = e.send(rec)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusOK, rec.Code)
	body, err := io.ReadAll(rec.Body)
	assert.Nil(t, err, "Actual err: %v", err)

	response := communication.ToMessageDtoResponse(msg)
	msgAsJson, err := json.Marshal(response)
	assert.Nil(t, err, "Actual err: %v", err)

	expected := fmt.Sprintf(
		`id: %s
data: %s

`,
		msg.Id.String(),
		msgAsJson,
	)
	assert.Equal(
		t,
		[]byte(expected),
		body,
		"Expected %s, got: %s",
		string(expected),
		string(body),
	)
}
