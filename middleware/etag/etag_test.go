package etag

import (
	"bytes"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/jialequ/sdk"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

// go test -run Test_ETag_Next
func TestETagNext(t *testing.T) {
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

// go test -run Test_ETag_SkipError
func TestETagSkipError(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(_ fiber.Ctx) error {
		return fiber.ErrForbidden
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// go test -run Test_ETag_NotStatusOK
func TestETagNotStatusOK(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusCreated)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

// go test -run Test_ETag_NoBody
func TestETagNoBody(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(_ fiber.Ctx) error {
		return nil
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// go test -run Test_ETag_NewEtag
func TestETagNewEtag(t *testing.T) {
	t.Parallel()
	t.Run(literal_4859, func(t *testing.T) {
		t.Parallel()
		testETagNewEtag(t, false, false)
	})
	t.Run(literal_5967, func(t *testing.T) {
		t.Parallel()
		testETagNewEtag(t, true, false)
	})
	t.Run(literal_0962, func(t *testing.T) {
		t.Parallel()
		testETagNewEtag(t, true, true)
	})
}

func testETagNewEtag(t *testing.T, headerIfNoneMatch, matched bool) { //nolint:revive // We're in a test, so using bools as a flow-control is fine
	t.Helper()

	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(literal_4806)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	if headerIfNoneMatch {
		etag := `"non-match"`
		if matched {
			etag = `"13-1831710635"`
		}
		req.Header.Set(fiber.HeaderIfNoneMatch, etag)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	if !headerIfNoneMatch || !matched {
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		require.Equal(t, `"13-1831710635"`, resp.Header.Get(fiber.HeaderETag))
		return
	}

	if matched {
		require.Equal(t, fiber.StatusNotModified, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, b)
	}
}

// go test -run Test_ETag_WeakEtag
func TestETagWeakEtag(t *testing.T) {
	t.Parallel()
	t.Run(literal_4859, func(t *testing.T) {
		t.Parallel()
		testETagWeakEtag(t, false, false)
	})
	t.Run(literal_5967, func(t *testing.T) {
		t.Parallel()
		testETagWeakEtag(t, true, false)
	})
	t.Run(literal_0962, func(t *testing.T) {
		t.Parallel()
		testETagWeakEtag(t, true, true)
	})
}

func testETagWeakEtag(t *testing.T, headerIfNoneMatch, matched bool) { //nolint:revive // We're in a test, so using bools as a flow-control is fine
	t.Helper()

	app := fiber.New()

	app.Use(New(Config{Weak: true}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(literal_4806)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	if headerIfNoneMatch {
		etag := `W/"non-match"`
		if matched {
			etag = `W/"13-1831710635"`
		}
		req.Header.Set(fiber.HeaderIfNoneMatch, etag)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	if !headerIfNoneMatch || !matched {
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		require.Equal(t, `W/"13-1831710635"`, resp.Header.Get(fiber.HeaderETag))
		return
	}

	if matched {
		require.Equal(t, fiber.StatusNotModified, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, b)
	}
}

// go test -run Test_ETag_CustomEtag
func TestETagCustomEtag(t *testing.T) {
	t.Parallel()
	t.Run(literal_4859, func(t *testing.T) {
		t.Parallel()
		testETagCustomEtag(t, false, false)
	})
	t.Run(literal_5967, func(t *testing.T) {
		t.Parallel()
		testETagCustomEtag(t, true, false)
	})
	t.Run(literal_0962, func(t *testing.T) {
		t.Parallel()
		testETagCustomEtag(t, true, true)
	})
}

func testETagCustomEtag(t *testing.T, headerIfNoneMatch, matched bool) { //nolint:revive // We're in a test, so using bools as a flow-control is fine
	t.Helper()

	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderETag, `"custom"`)
		if bytes.Equal(c.Request().Header.Peek(fiber.HeaderIfNoneMatch), []byte(`"custom"`)) {
			return c.SendStatus(fiber.StatusNotModified)
		}
		return c.SendString(literal_4806)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	if headerIfNoneMatch {
		etag := `"non-match"`
		if matched {
			etag = `"custom"`
		}
		req.Header.Set(fiber.HeaderIfNoneMatch, etag)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	if !headerIfNoneMatch || !matched {
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		require.Equal(t, `"custom"`, resp.Header.Get(fiber.HeaderETag))
		return
	}

	if matched {
		require.Equal(t, fiber.StatusNotModified, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, b)
	}
}

// go test -run Test_ETag_CustomEtagPut
func TestETagCustomEtagPut(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Put("/", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderETag, `"custom"`)
		if !bytes.Equal(c.Request().Header.Peek(fiber.HeaderIfMatch), []byte(`"custom"`)) {
			return c.SendStatus(fiber.StatusPreconditionFailed)
		}
		return c.SendString(literal_4806)
	})

	req := httptest.NewRequest(fiber.MethodPut, "/", nil)
	req.Header.Set(fiber.HeaderIfMatch, `"non-match"`)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusPreconditionFailed, resp.StatusCode)
}

// go test -v -run=^$ -bench=Benchmark_Etag -benchmem -count=4
func Benchmark_Etag(b *testing.B) {
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(literal_4806)
	})

	h := app.Handler()

	fctx := &fasthttp.RequestCtx{}
	fctx.Request.Header.SetMethod(fiber.MethodGet)
	fctx.Request.SetRequestURI("/")

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		h(fctx)
	}

	require.Equal(b, 200, fctx.Response.Header.StatusCode())
	require.Equal(b, `"13-1831710635"`, string(fctx.Response.Header.Peek(fiber.HeaderETag)))
}

const literal_4859 = "without HeaderIfNoneMatch"

const literal_5967 = "with HeaderIfNoneMatch and not matched"

const literal_0962 = "with HeaderIfNoneMatch and matched"

const literal_4806 = "Hello, World!"
