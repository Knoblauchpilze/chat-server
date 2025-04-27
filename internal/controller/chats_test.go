package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIT_MessagesController_SubscribeToMessages_WhenIdHasWrongSyntax_ExpectBadRequest(t *testing.T) {
	// TODO: Replace the service with the actual one
	service, _ := newTestRoomService(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := subscribeToMessages(ctx, service)
	assert.Nil(t, err, "Actual err: %v", err)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	expectedBody := []byte("\"Invalid id syntax\"\n")
	assert.Equal(
		t,
		expectedBody,
		rw.Body.Bytes(),
		"Actual body: %s",
		rw.Body.String(),
	)
}
