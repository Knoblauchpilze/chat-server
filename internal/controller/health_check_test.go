package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestIT_HealthcheckController(t *testing.T) {
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	req := httptest.NewRequest(http.MethodGet, "/healtcheck", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := healthcheck(ctx, dbConn)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "\"OK\"\n", rw.Body.String())
}

func TestIT_HealthcheckController_WhenConnectionClosed_ExpectServiceUnavailable(t *testing.T) {
	dbConn := newTestDbConnection(t)
	dbConn.Close(context.Background())

	req := httptest.NewRequest(http.MethodGet, "/healtcheck", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := healthcheck(ctx, dbConn)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rw.Code)
	expectedResponse := `
	{
		"Code": 100,
		"Message": "An unexpected error occurred"
	}`
	assert.JSONEq(t, expectedResponse, rw.Body.String())
}

func TestUnit_HealthcheckController_WhenConnectionFails_ExpectServiceUnavailable(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healtcheck", nil)
	ctx, rw := generateTestEchoContextFromRequest(req)

	err := healthcheck(ctx, &mockDbConn{})

	assert.Nil(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rw.Code)
	expectedResponse := `
	{
		"Code": 100,
		"Message": "An unexpected error occurred"
	}`
	assert.JSONEq(t, expectedResponse, rw.Body.String())
}

type mockDbConn struct {
	db.Connection
}

func (m *mockDbConn) Ping(ctx context.Context) error {
	return errors.NewCode(db.NotConnected)
}
