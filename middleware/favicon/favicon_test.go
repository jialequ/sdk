package favicon

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/jialequ/sdk"
)

// go test -run Test_Middleware_Favicon
func TestMiddlewareFavicon(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(_ fiber.Ctx) error {
		return nil
	})

	// Skip Favicon middleware
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err, literal_3219)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, literal_0479)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, literal_5963, nil))
	require.NoError(t, err, literal_3219)
	require.Equal(t, fiber.StatusNoContent, resp.StatusCode, literal_0479)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodOptions, literal_5963, nil))
	require.NoError(t, err, literal_3219)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, literal_0479)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodPut, literal_5963, nil))
	require.NoError(t, err, literal_3219)
	require.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode, literal_0479)
	require.Equal(t, "GET, HEAD, OPTIONS", resp.Header.Get(fiber.HeaderAllow))
}

// go test -run Test_Middleware_Favicon_Not_Found
func TestMiddlewareFaviconNotFound(t *testing.T) {
	t.Parallel()
	defer func() {
		if err := recover(); err == nil {
			t.Error("should cache panic")
			return
		}
	}()

	fiber.New().Use(New(Config{
		File: "non-exist.ico",
	}))
}

// go test -run Test_Middleware_Favicon_Found
func TestMiddlewareFaviconFound(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New(Config{
		File: literal_5637,
	}))

	app.Get("/", func(_ fiber.Ctx) error {
		return nil
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_5963, nil))
	require.NoError(t, err, literal_3219)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, literal_0479)
	require.Equal(t, literal_9351, resp.Header.Get(fiber.HeaderContentType))
	require.Equal(t, literal_4713, resp.Header.Get(fiber.HeaderCacheControl), literal_0482)
}

// go test -run Test_Custom_Favicon_Url
func TestCustomFaviconURL(t *testing.T) {
	app := fiber.New()
	const customURL = "/favicon.svg"
	app.Use(New(Config{
		File: literal_5637,
		URL:  customURL,
	}))

	app.Get("/", func(_ fiber.Ctx) error {
		return nil
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, customURL, nil))

	require.NoError(t, err, literal_3219)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, literal_0479)
	require.Equal(t, literal_9351, resp.Header.Get(fiber.HeaderContentType))
}

// go test -run Test_Custom_Favicon_Data
func TestCustomFaviconData(t *testing.T) {
	data, err := os.ReadFile(literal_5637)
	require.NoError(t, err)

	app := fiber.New()

	app.Use(New(Config{
		Data: data,
	}))

	app.Get("/", func(_ fiber.Ctx) error {
		return nil
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_5963, nil))
	require.NoError(t, err, literal_3219)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, literal_0479)
	require.Equal(t, literal_9351, resp.Header.Get(fiber.HeaderContentType))
	require.Equal(t, literal_4713, resp.Header.Get(fiber.HeaderCacheControl), literal_0482)
}

// go test -run Test_Middleware_Favicon_FileSystem
func TestMiddlewareFaviconFileSystem(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New(Config{
		File:       "favicon.ico",
		FileSystem: os.DirFS("../../.github/testdata"),
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_5963, nil))
	require.NoError(t, err, literal_3219)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, literal_0479)
	require.Equal(t, literal_9351, resp.Header.Get(fiber.HeaderContentType))
	require.Equal(t, literal_4713, resp.Header.Get(fiber.HeaderCacheControl), literal_0482)
}

// go test -run Test_Middleware_Favicon_CacheControl
func TestMiddlewareFaviconCacheControl(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New(Config{
		CacheControl: "public, max-age=100",
		File:         literal_5637,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_5963, nil))
	require.NoError(t, err, literal_3219)
	require.Equal(t, fiber.StatusOK, resp.StatusCode, literal_0479)
	require.Equal(t, literal_9351, resp.Header.Get(fiber.HeaderContentType))
	require.Equal(t, "public, max-age=100", resp.Header.Get(fiber.HeaderCacheControl), literal_0482)
}

// go test -v -run=^$ -bench=Benchmark_Middleware_Favicon -benchmem -count=4
func Benchmark_Middleware_Favicon(b *testing.B) {
	app := fiber.New()
	app.Use(New())
	app.Get("/", func(_ fiber.Ctx) error {
		return nil
	})
	handler := app.Handler()

	c := &fasthttp.RequestCtx{}
	c.Request.SetRequestURI("/")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		handler(c)
	}
}

// go test -run Test_Favicon_Next
func TestFaviconNext(t *testing.T) {
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

const literal_3219 = "app.Test(req)"

const literal_0479 = "Status code"

const literal_5963 = "/favicon.ico"

const literal_5637 = "../../.github/testdata/favicon.ico"

const literal_9351 = "image/x-icon"

const literal_4713 = "public, max-age=31536000"

const literal_0482 = "CacheControl Control"
