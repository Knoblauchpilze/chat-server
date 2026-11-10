package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestIT_HealthcheckController(t *testing.T) {
	dbConn := newTestDbConnection(t)
	defer dbConn.Close(context.Background())

	req := httptest.NewRequest(http.MethodGet, "/healtcheck", nil)
	ctx, rw := generateTestGinContextFromRequest(req)

	healthcheck(ctx, dbConn)
	assert.Equal(t, http.StatusOK, ctx.Writer.Status())
	assert.Equal(t, "\"OK\"", rw.Body.String())
}

func TestIT_HealthcheckController_WhenConnectionClosed_ExpectServiceUnavailable(t *testing.T) {
	dbConn := newTestDbConnection(t)
	dbConn.Close(context.Background())

	req := httptest.NewRequest(http.MethodGet, "/healtcheck", nil)
	ctx, rw := generateTestGinContextFromRequest(req)

	healthcheck(ctx, dbConn)
	assert.Equal(t, http.StatusServiceUnavailable, ctx.Writer.Status())
	expectedResponse := `
	{
		"code": 100,
		"message": "an unexpected error occurred"
	}`
	assert.JSONEq(t, expectedResponse, rw.Body.String())
}

func TestUnit_HealthcheckController_WhenConnectionFails_ExpectServiceUnavailable(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healtcheck", nil)
	ctx, rw := generateTestGinContextFromRequest(req)

	healthcheck(ctx, &mockDbConn{})
	assert.Equal(t, http.StatusServiceUnavailable, ctx.Writer.Status())
	expectedResponse := `
	{
		"code": 100,
		"message": "an unexpected error occurred"
	}`
	assert.JSONEq(t, expectedResponse, rw.Body.String())
}

type mockDbConn struct {
	*db.Connection
}

func (m *mockDbConn) Ping(ctx context.Context) error {
	return db.ErrNotConnected
}
