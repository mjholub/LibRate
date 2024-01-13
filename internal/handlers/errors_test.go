package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"codeberg.org/mjh/LibRate/tests"
)

type testController struct {
	log *zerolog.Logger
}

func newTestController(log *zerolog.Logger) *testController {
	return &testController{log: log}
}

// test especially whether the log message is written as expected
func TestInternalError(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()
	tc := newTestController(&logger)

	app := tests.NewAppWithLogger(&logger)

	port, err := tests.TryFindFreePort()
	require.Nil(t, err)

	app.Get("/internal-error", tc.internalErrorTestHandler)

	go func() {
		err = app.Listen(fmt.Sprintf(":%d", port))
		require.Nil(t, err)
	}()

	req := httptest.NewRequest("GET", fmt.Sprintf("http://localhost:%d/internal-error", port), nil)
	assert.NotEmpty(t, req)

	resp, err := app.Test(req)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Contains(t, buf.String(), "test message")
	assert.Contains(t, buf.String(), "test error")
}

func (tc *testController) internalErrorTestHandler(c *fiber.Ctx) error {
	message := "test message"
	err := errors.New("test error")
	return InternalError(tc.log, c, message, err)
}
