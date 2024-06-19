package expvar

import (
	"bytes"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func TestNonExpvarPath(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("escaped")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "escaped", string(b))
}

func TestExpvarIndex(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("escaped")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/debug/vars", nil))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, fiber.MIMEApplicationJSONCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.True(t, bytes.Contains(b, []byte("cmdline")))
	require.True(t, bytes.Contains(b, []byte("memstat")))
}

func TestExpvarFilter(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("escaped")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/debug/vars?r=cmd", nil))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, fiber.MIMEApplicationJSONCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.True(t, bytes.Contains(b, []byte("cmdline")))
	require.False(t, bytes.Contains(b, []byte("memstat")))
}

func TestExpvarOtherPath(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("escaped")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/debug/vars/302", nil))
	require.NoError(t, err)
	require.Equal(t, 302, resp.StatusCode)
}

// go test -run Test_Expvar_Next
func TestExpvarNext(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New(Config{
		Next: func(_ fiber.Ctx) bool {
			return true
		},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/debug/vars", nil))
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
}
