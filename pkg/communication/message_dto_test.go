package communication

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Knoblauchpilze/chat-server/pkg/persistence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_MessageDtoRequest_MarshalsToCamelCase(t *testing.T) {
	dto := MessageDtoRequest{
		User:    uuid.MustParse("e1c24e7d-b7b2-4118-85ac-d724b8e889dd"),
		Room:    uuid.MustParse("aaa3fc52-3ad8-4679-ae92-6c1bddb8d7f9"),
		Message: "my-message",
	}

	out, err := json.Marshal(dto)

	assert.Nil(t, err)
	expectedJson := `
	{
		"user": "e1c24e7d-b7b2-4118-85ac-d724b8e889dd",
		"room": "aaa3fc52-3ad8-4679-ae92-6c1bddb8d7f9",
		"message": "my-message"
	}`
	assert.JSONEq(t, expectedJson, string(out))
}

func TestUnit_FromMessageDtoRequest(t *testing.T) {
	beforeConversion := time.Now()

	dto := MessageDtoRequest{
		User:    uuid.New(),
		Room:    uuid.New(),
		Message: "my-message",
	}

	actual := FromMessageDtoRequest(dto)

	assert.Equal(t, dto.User, actual.ChatUser)
	assert.Equal(t, dto.Room, actual.Room)
	assert.Equal(t, dto.Message, actual.Message)
	assert.True(t, actual.CreatedAt.After(beforeConversion))
}

func TestUnit_MessageDtoResponse_MarshalsToCamelCase(t *testing.T) {
	dto := MessageDtoResponse{
		Id:        uuid.MustParse("a590b448-d3cd-4dbc-a9e3-8d642b1a5814"),
		User:      uuid.MustParse("e1c24e7d-b7b2-4118-85ac-d724b8e889dd"),
		Room:      uuid.MustParse("aaa3fc52-3ad8-4679-ae92-6c1bddb8d7f9"),
		Message:   "my-message",
		CreatedAt: someTime,
	}

	out, err := json.Marshal(dto)

	assert.Nil(t, err)
	expectedJson := `
	{
		"id": "a590b448-d3cd-4dbc-a9e3-8d642b1a5814",
		"user": "e1c24e7d-b7b2-4118-85ac-d724b8e889dd",
		"room": "aaa3fc52-3ad8-4679-ae92-6c1bddb8d7f9",
		"message": "my-message",
		"created_at": "2024-11-12T19:09:36Z"
	}`
	assert.JSONEq(t, expectedJson, string(out))
}

func TestUnit_ToMessageDtoResponse(t *testing.T) {
	entity := persistence.Message{
		Id:       uuid.New(),
		ChatUser: uuid.New(),
		Room:     uuid.New(),
		Message:  "my-message",

		CreatedAt: someTime,
	}

	actual := ToMessageDtoResponse(entity)

	assert.Equal(t, entity.Id, actual.Id)
	assert.Equal(t, entity.ChatUser, actual.User)
	assert.Equal(t, entity.Room, actual.Room)
	assert.Equal(t, entity.Message, actual.Message)
	assert.Equal(t, someTime, actual.CreatedAt)
}
