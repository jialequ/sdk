package pprof

import (
	"bytes"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jialequ/sdk"
	"github.com/stretchr/testify/require"
)

func TestNonPprofPath(t *testing.T) {
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

func TestNonPprofPathWithPrefix(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{Prefix: literal_0874}))

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

func TestPprofIndex(t *testing.T) {
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("escaped")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_3162, nil))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.True(t, bytes.Contains(b, []byte("<title>/debug/pprof/</title>")))
}

func TestPprofIndexWithPrefix(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{Prefix: literal_0874}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("escaped")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_6407, nil))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(b), "<title>/debug/pprof/</title>")
}

func TestPprofSubs(t *testing.T) {
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("escaped")
	})

	subs := []string{
		"cmdline", "profile", "symbol", "trace", "allocs", "block",
		"goroutine", "heap", "mutex", "threadcreate",
	}

	for _, sub := range subs {
		sub := sub
		t.Run(sub, func(t *testing.T) {
			target := literal_3162 + sub
			if sub == "profile" {
				target += "?seconds=1"
			}
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, target, nil), 5*time.Second)
			require.NoError(t, err)
			require.Equal(t, 200, resp.StatusCode)
		})
	}
}

func TestPprofSubsWithPrefix(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{Prefix: literal_0874}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("escaped")
	})

	subs := []string{
		"cmdline", "profile", "symbol", "trace", "allocs", "block",
		"goroutine", "heap", "mutex", "threadcreate",
	}

	for _, sub := range subs {
		sub := sub
		t.Run(sub, func(t *testing.T) {
			target := literal_6407 + sub
			if sub == "profile" {
				target += "?seconds=1"
			}
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, target, nil), 5*time.Second)
			require.NoError(t, err)
			require.Equal(t, 200, resp.StatusCode)
		})
	}
}

func TestPprofOther(t *testing.T) {
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("escaped")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/debug/pprof/302", nil))
	require.NoError(t, err)
	require.Equal(t, 302, resp.StatusCode)
}

func TestPprofOtherWithPrefix(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{Prefix: literal_0874}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("escaped")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/federated-fiber/debug/pprof/302", nil))
	require.NoError(t, err)
	require.Equal(t, 302, resp.StatusCode)
}

// go test -run Test_Pprof_Next
func TestPprofNext(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		Next: func(_ fiber.Ctx) bool {
			return true
		},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_3162, nil))
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
}

// go test -run Test_Pprof_Next_WithPrefix
func TestPprofNextWithPrefix(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		Next: func(_ fiber.Ctx) bool {
			return true
		},
		Prefix: literal_0874,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_6407, nil))
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
}

const literal_0874 = "/federated-fiber"

const literal_3162 = "/debug/pprof/"

const literal_6407 = "/federated-fiber/debug/pprof/"
