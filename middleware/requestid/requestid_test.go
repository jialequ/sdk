package requestid

import (
	"net/http/httptest"
	"testing"

	"github.com/jialequ/sdk"
	"github.com/stretchr/testify/require"
)

// go test -run Test_RequestID
func TestRequestID(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	reqid := resp.Header.Get(fiber.HeaderXRequestID)
	require.Len(t, reqid, 36)

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Add(fiber.HeaderXRequestID, reqid)

	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, reqid, resp.Header.Get(fiber.HeaderXRequestID))
}

// go test -run Test_RequestID_Next
func TestRequestIDNext(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		Next: func(_ fiber.Ctx) bool {
			return true
		},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Empty(t, resp.Header.Get(fiber.HeaderXRequestID))
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// go test -run Test_RequestID_Locals
func TestRequestIDFromContext(t *testing.T) {
	t.Parallel()
	reqID := "ThisIsARequestId"

	app := fiber.New()
	app.Use(New(Config{
		Generator: func() string {
			return reqID
		},
	}))

	var ctxVal string

	app.Use(func(c fiber.Ctx) error {
		ctxVal = FromContext(c)
		return c.Next()
	})

	_, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, reqID, ctxVal)
}
