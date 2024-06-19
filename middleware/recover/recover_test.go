package recover //nolint:predeclared // TODO: Rename to some non-builtin

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

// go test -run Test_Recover
func TestRecover(t *testing.T) {
	t.Parallel()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			require.Equal(t, "Hi, I'm an error!", err.Error())
			return c.SendStatus(fiber.StatusTeapot)
		},
	})

	app.Use(New())

	app.Get("/panic", func(_ fiber.Ctx) error {
		panic("Hi, I'm an error!")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/panic", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusTeapot, resp.StatusCode)
}

// go test -run Test_Recover_Next
func TestRecoverNext(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		Next: func(_ fiber.Ctx) bool {
			return true
		},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestRecoverEnableStackTrace(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		EnableStackTrace: true,
	}))

	app.Get("/panic", func(_ fiber.Ctx) error {
		panic("Hi, I'm an error!")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/panic", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
