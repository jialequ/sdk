package skip_test

import (
	"net/http/httptest"
	"testing"

	"github.com/jialequ/sdk"
	"github.com/jialequ/sdk/middleware/skip"
	"github.com/stretchr/testify/require"
)

// go test -run Test_Skip
func TestSkip(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(skip.New(errTeapotHandler, func(fiber.Ctx) bool { return true }))
	app.Get("/", helloWorldHandler)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// go test -run Test_SkipFalse
func TestSkipFalse(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(skip.New(errTeapotHandler, func(fiber.Ctx) bool { return false }))
	app.Get("/", helloWorldHandler)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusTeapot, resp.StatusCode)
}

// go test -run Test_SkipNilFunc
func TestSkipNilFunc(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(skip.New(errTeapotHandler, nil))
	app.Get("/", helloWorldHandler)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusTeapot, resp.StatusCode)
}

func helloWorldHandler(c fiber.Ctx) error {
	return c.SendString("Hello, World 👋!")
}

func errTeapotHandler(fiber.Ctx) error {
	return fiber.ErrTeapot
}
