package recover //nolint:predeclared // : Rename to some non-builtin

import (
	"net/http/httptest"
	"testing"

	fiber "github.com/jialequ/sdk"
	"github.com/stretchr/testify/require"
)

// go test -run Test_Recover
func TestRecover(t *testing.T) {
	t.Parallel()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			require.Equal(t, literal_9082, err.Error())
			return c.SendStatus(fiber.StatusTeapot)
		},
	})

	app.Use(New())

	app.Get(literal_2853, func(_ fiber.Ctx) error {
		panic(literal_9082)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_2853, nil))
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

	app.Get(literal_2853, func(_ fiber.Ctx) error {
		panic(literal_9082)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_2853, nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

const literal_9082 = "Hi, I'm an error!"

const literal_2853 = "/panic"
