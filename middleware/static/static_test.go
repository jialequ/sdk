package static

import (
	"embed"
	"io"
	"io/fs"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

// go test -run Test_Static_Index_Default
func TestStaticIndexDefault(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/prefix", New("../../.github/workflows"))

	app.Get("", New("../../.github/"))

	app.Get("test", New("", Config{
		IndexNames: []string{"index.html"},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Hello, World!")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/not-found", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 404, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextPlainCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Cannot GET /not-found", string(body))
}

// go test -run Test_Static_Index
func TestStaticDirect(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/*", New("../../.github"))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/index.html", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Hello, World!")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodPost, "/index.html", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 405, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextPlainCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/testdata/testRoutes.json", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMEApplicationJSON, resp.Header.Get("Content-Type"))
	require.Equal(t, "", resp.Header.Get(fiber.HeaderCacheControl), "CacheControl Control")

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "test_routes")
}

// go test -run Test_Static_MaxAge
func TestStaticMaxAge(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/*", New("../../.github", Config{
		MaxAge: 100,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/index.html", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, "text/html; charset=utf-8", resp.Header.Get(fiber.HeaderContentType))
	require.Equal(t, "public, max-age=100", resp.Header.Get(fiber.HeaderCacheControl), "CacheControl Control")
}

// go test -run Test_Static_Custom_CacheControl
func TestStaticCustomCacheControl(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/*", New("../../.github", Config{
		ModifyResponse: func(c fiber.Ctx) error {
			if strings.Contains(c.GetRespHeader("Content-Type"), "text/html") {
				c.Response().Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
			}
			return nil
		},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/index.html", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get(fiber.HeaderCacheControl), "CacheControl Control")

	normalResp, normalErr := app.Test(httptest.NewRequest(fiber.MethodGet, "/config.yml", nil))
	require.NoError(t, normalErr, "app.Test(req)")
	require.Equal(t, "", normalResp.Header.Get(fiber.HeaderCacheControl), "CacheControl Control")
}

func TestStaticDisableCache(t *testing.T) {
	// Skip on Windows. It's not possible to delete a file that is in use.
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}

	t.Parallel()

	app := fiber.New()

	file, err := os.Create("../../.github/test.txt")
	require.NoError(t, err)
	_, err = file.WriteString("Hello, World!")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	// Remove the file even if the test fails
	defer func() {
		_ = os.Remove("../../.github/test.txt") //nolint:errcheck // not needed
	}()

	app.Get("/*", New("../../.github/", Config{
		CacheDuration: -1,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/test.txt", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, "", resp.Header.Get(fiber.HeaderCacheControl), "CacheControl Control")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Hello, World!")

	require.NoError(t, os.Remove("../../.github/test.txt"))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/test.txt", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, "", resp.Header.Get(fiber.HeaderCacheControl), "CacheControl Control")

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Cannot GET /test.txt", string(body))
}

func TestStaticNotFoundHandler(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/*", New("../../.github", Config{
		NotFoundHandler: func(c fiber.Ctx) error {
			return c.SendString("Custom 404")
		},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/not-found", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 404, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextPlainCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Custom 404", string(body))
}

// go test -run Test_Static_Download
func TestStaticDownload(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/fiber.png", New("../../.github/testdata/fs/img/fiber.png", Config{
		Download: true,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/fiber.png", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, "image/png", resp.Header.Get(fiber.HeaderContentType))
	require.Equal(t, `attachment`, resp.Header.Get(fiber.HeaderContentDisposition))
}

// go test -run Test_Static_Group
func TestStaticGroup(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	grp := app.Group("/v1", func(c fiber.Ctx) error {
		c.Set("Test-Header", "123")
		return c.Next()
	})

	grp.Get("/v2*", New("../../.github/index.html"))

	req := httptest.NewRequest(fiber.MethodGet, "/v1/v2", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))
	require.Equal(t, "123", resp.Header.Get("Test-Header"))

	grp = app.Group("/v2")
	grp.Get("/v3*", New("../../.github/index.html"))

	req = httptest.NewRequest(fiber.MethodGet, "/v2/v3/john/doe", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))
}

func TestStaticWildcard(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("*", New("../../.github/index.html"))

	req := httptest.NewRequest(fiber.MethodGet, "/yesyes/john/doe", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Test file")
}

func TestStaticPrefixWildcard(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test*", New("../../.github/index.html"))

	req := httptest.NewRequest(fiber.MethodGet, "/test/john/doe", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	app.Get("/my/nameisjohn*", New("../../.github/index.html"))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/my/nameisjohn/no/its/not", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Test file")
}

func TestStaticPrefix(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/john*", New("../../.github"))

	req := httptest.NewRequest(fiber.MethodGet, "/john/index.html", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	app.Get("/prefix*", New("../../.github/testdata"))

	req = httptest.NewRequest(fiber.MethodGet, "/prefix/index.html", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	app.Get("/single*", New("../../.github/testdata/testRoutes.json"))

	req = httptest.NewRequest(fiber.MethodGet, "/single", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMEApplicationJSON, resp.Header.Get(fiber.HeaderContentType))
}

func TestStaticTrailingSlash(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/john*", New("../../.github"))

	req := httptest.NewRequest(fiber.MethodGet, "/john/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	app.Get("/john_without_index*", New("../../.github/testdata/fs/css"))

	req = httptest.NewRequest(fiber.MethodGet, "/john_without_index/", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 404, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextPlainCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	app.Use("/john", New("../../.github"))

	req = httptest.NewRequest(fiber.MethodGet, "/john/", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	req = httptest.NewRequest(fiber.MethodGet, "/john", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	app.Use("/john_without_index/", New("../../.github/testdata/fs/css"))

	req = httptest.NewRequest(fiber.MethodGet, "/john_without_index/", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 404, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextPlainCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))
}

func TestStaticNext(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/*", New("../../.github", Config{
		Next: func(c fiber.Ctx) bool {
			return c.Get("X-Custom-Header") == "skip"
		},
	}))

	app.Get("/*", func(c fiber.Ctx) error {
		return c.SendString("You've skipped app.Static")
	})

	t.Run("app.Static is skipped: invoking Get handler", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(fiber.MethodGet, "/", nil)
		req.Header.Set("X-Custom-Header", "skip")
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
		require.Equal(t, fiber.MIMETextPlainCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "You've skipped app.Static")
	})

	t.Run("app.Static is not skipped: serving index.html", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(fiber.MethodGet, "/", nil)
		req.Header.Set("X-Custom-Header", "don't skip")
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
		require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "Hello, World!")
	})
}

func TestRouteStaticRoot(t *testing.T) {
	t.Parallel()

	dir := "../../.github/testdata/fs/css"
	app := fiber.New()
	app.Get("/*", New(dir, Config{
		Browse: true,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/style.css", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "app.Test(req)")
	require.Contains(t, string(body), "color")

	app = fiber.New()
	app.Get("/*", New(dir))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 404, resp.StatusCode, "Status code")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/style.css", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "app.Test(req)")
	require.Contains(t, string(body), "color")
}

func TestRouteStaticHasPrefix(t *testing.T) {
	t.Parallel()

	dir := "../../.github/testdata/fs/css"
	app := fiber.New()
	app.Get("/static*", New(dir, Config{
		Browse: true,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/static", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/static/", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/static/style.css", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "app.Test(req)")
	require.Contains(t, string(body), "color")

	app = fiber.New()
	app.Get("/static/*", New(dir, Config{
		Browse: true,
	}))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/static", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/static/", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/static/style.css", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "app.Test(req)")
	require.Contains(t, string(body), "color")

	app = fiber.New()
	app.Get("/static*", New(dir))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/static", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 404, resp.StatusCode, "Status code")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/static/", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 404, resp.StatusCode, "Status code")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/static/style.css", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "app.Test(req)")
	require.Contains(t, string(body), "color")

	app = fiber.New()
	app.Get("/static*", New(dir))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/static", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 404, resp.StatusCode, "Status code")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/static/", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 404, resp.StatusCode, "Status code")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/static/style.css", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "app.Test(req)")
	require.Contains(t, string(body), "color")
}

func TestStaticFS(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/*", New("", Config{
		FS:     os.DirFS("../../.github/testdata/fs"),
		Browse: true,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/css/style.css", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.Equal(t, fiber.MIMETextCSSCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "app.Test(req)")
	require.Contains(t, string(body), "color")
}

/*func Test_Static_FS_DifferentRoot(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/*", New("fs", Config{
		FS:         os.DirFS("../../.github/testdata"),
		IndexNames: []string{"index2.html"},
		Browse:     true,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "app.Test(req)")
	require.Contains(t, string(body), "<h1>Hello, World!</h1>")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/css/style.css", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.Equal(t, fiber.MIMETextCSSCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "app.Test(req)")
	require.Contains(t, string(body), "color")
}*/

//go:embed static.go config.go
var fsTestFilesystem embed.FS

func TestStaticFSBrowse(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	app.Get("/embed*", New("", Config{
		FS:     fsTestFilesystem,
		Browse: true,
	}))

	app.Get("/dirfs*", New("", Config{
		FS:     os.DirFS("../../.github/testdata/fs/css"),
		Browse: true,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/dirfs", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "app.Test(req)")
	require.Contains(t, string(body), "style.css")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/dirfs/style.css", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.Equal(t, fiber.MIMETextCSSCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "app.Test(req)")
	require.Contains(t, string(body), "color")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/embed", nil))
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "app.Test(req)")
	require.Contains(t, string(body), "static.go")
}

func TestStaticFSPrefixWildcard(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test*", New("index.html", Config{
		FS:         os.DirFS("../../.github"),
		IndexNames: []string{"not_index.html"},
	}))

	req := httptest.NewRequest(fiber.MethodGet, "/test/john/doe", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "app.Test(req)")
	require.Equal(t, 200, resp.StatusCode, "Status code")
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Test file")
}

func TestisFile(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		path       string
		filesystem fs.FS
		expected   bool
		gotError   error
	}{
		{
			name:       "file",
			path:       "index.html",
			filesystem: os.DirFS("../../.github"),
			expected:   true,
		},
		{
			name:       "file",
			path:       "index2.html",
			filesystem: os.DirFS("../../.github"),
			expected:   false,
			gotError:   fs.ErrNotExist,
		},
		{
			name:       "directory",
			path:       ".",
			filesystem: os.DirFS("../../.github"),
			expected:   false,
		},
		{
			name:       "directory",
			path:       "not_exists",
			filesystem: os.DirFS("../../.github"),
			expected:   false,
			gotError:   fs.ErrNotExist,
		},
		{
			name:       "directory",
			path:       ".",
			filesystem: os.DirFS("../../.github/testdata/fs/css"),
			expected:   false,
		},
		{
			name:       "file",
			path:       "../../.github/testdata/fs/css/style.css",
			filesystem: nil,
			expected:   true,
		},
		{
			name:       "file",
			path:       "../../.github/testdata/fs/css/style2.css",
			filesystem: nil,
			expected:   false,
			gotError:   fs.ErrNotExist,
		},
		{
			name:       "directory",
			path:       "../../.github/testdata/fs/css",
			filesystem: nil,
			expected:   false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c := c
			t.Parallel()

			actual, err := isFile(c.path, c.filesystem)
			require.ErrorIs(t, err, c.gotError)
			require.Equal(t, c.expected, actual)
		})
	}
}