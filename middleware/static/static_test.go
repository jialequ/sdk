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

	fiber "github.com/jialequ/sdk"
	"github.com/stretchr/testify/require"
)

// go test -run Test_Static_Index_Default
func TestStaticIndexDefault(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/prefix", New("../../.github/workflows"))

	app.Get("", New("../../.github/"))

	app.Get("test", New("", Config{
		IndexNames: []string{literal_4531},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), literal_0385)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/not-found", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 404, resp.StatusCode, literal_1247)
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

	app.Get("/*", New(literal_8694))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_6304, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), literal_0385)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodPost, literal_6304, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 405, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextPlainCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/testdata/testRoutes.json", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMEApplicationJSON, resp.Header.Get("Content-Type"))
	require.Equal(t, "", resp.Header.Get(fiber.HeaderCacheControl), literal_2451)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "test_routes")
}

// go test -run Test_Static_MaxAge
func TestStaticMaxAge(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/*", New(literal_8694, Config{
		MaxAge: 100,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_6304, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, "text/html; charset=utf-8", resp.Header.Get(fiber.HeaderContentType))
	require.Equal(t, "public, max-age=100", resp.Header.Get(fiber.HeaderCacheControl), literal_2451)
}

// go test -run Test_Static_Custom_CacheControl
func TestStaticCustomCacheControl(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/*", New(literal_8694, Config{
		ModifyResponse: func(c fiber.Ctx) error {
			if strings.Contains(c.GetRespHeader("Content-Type"), "text/html") {
				c.Response().Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
			}
			return nil
		},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_6304, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get(fiber.HeaderCacheControl), literal_2451)

	normalResp, normalErr := app.Test(httptest.NewRequest(fiber.MethodGet, "/config.yml", nil))
	require.NoError(t, normalErr, literal_1679)
	require.Equal(t, "", normalResp.Header.Get(fiber.HeaderCacheControl), literal_2451)
}

func TestStaticDisableCache(t *testing.T) {
	// Skip on Windows. It's not possible to delete a file that is in use.
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}

	t.Parallel()

	app := fiber.New()

	file, err := os.Create(literal_5374)
	require.NoError(t, err)
	_, err = file.WriteString(literal_0385)
	require.NoError(t, err)
	require.NoError(t, file.Close())

	// Remove the file even if the test fails
	defer func() {
		_ = os.Remove(literal_5374) //nolint:errcheck // not needed
	}()

	app.Get("/*", New("../../.github/", Config{
		CacheDuration: -1,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/test.txt", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, "", resp.Header.Get(fiber.HeaderCacheControl), literal_2451)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), literal_0385)

	require.NoError(t, os.Remove(literal_5374))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/test.txt", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, "", resp.Header.Get(fiber.HeaderCacheControl), literal_2451)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Cannot GET /test.txt", string(body))
}

func TestStaticNotFoundHandler(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/*", New(literal_8694, Config{
		NotFoundHandler: func(c fiber.Ctx) error {
			return c.SendString("Custom 404")
		},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/not-found", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 404, resp.StatusCode, literal_1247)
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
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
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

	grp.Get("/v2*", New(literal_4912))

	req := httptest.NewRequest(fiber.MethodGet, "/v1/v2", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))
	require.Equal(t, "123", resp.Header.Get("Test-Header"))

	grp = app.Group("/v2")
	grp.Get("/v3*", New(literal_4912))

	req = httptest.NewRequest(fiber.MethodGet, "/v2/v3/john/doe", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))
}

func TestStaticWildcard(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("*", New(literal_4912))

	req := httptest.NewRequest(fiber.MethodGet, "/yesyes/john/doe", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), literal_1925)
}

func TestStaticPrefixWildcard(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test*", New(literal_4912))

	req := httptest.NewRequest(fiber.MethodGet, "/test/john/doe", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	app.Get("/my/nameisjohn*", New(literal_4912))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/my/nameisjohn/no/its/not", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), literal_1925)
}

func TestStaticPrefix(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/john*", New(literal_8694))

	req := httptest.NewRequest(fiber.MethodGet, "/john/index.html", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	app.Get("/prefix*", New("../../.github/testdata"))

	req = httptest.NewRequest(fiber.MethodGet, "/prefix/index.html", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	app.Get("/single*", New("../../.github/testdata/testRoutes.json"))

	req = httptest.NewRequest(fiber.MethodGet, "/single", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMEApplicationJSON, resp.Header.Get(fiber.HeaderContentType))
}

func TestStaticTrailingSlash(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/john*", New(literal_8694))

	req := httptest.NewRequest(fiber.MethodGet, "/john/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	app.Get("/john_without_index*", New(literal_4265))

	req = httptest.NewRequest(fiber.MethodGet, literal_7493, nil)
	resp, err = app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 404, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextPlainCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	app.Use("/john", New(literal_8694))

	req = httptest.NewRequest(fiber.MethodGet, "/john/", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	req = httptest.NewRequest(fiber.MethodGet, "/john", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	app.Use(literal_7493, New(literal_4265))

	req = httptest.NewRequest(fiber.MethodGet, literal_7493, nil)
	resp, err = app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 404, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextPlainCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))
}

func TestStaticNext(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/*", New(literal_8694, Config{
		Next: func(c fiber.Ctx) bool {
			return c.Get(literal_2395) == "skip"
		},
	}))

	app.Get("/*", func(c fiber.Ctx) error {
		return c.SendString("You've skipped app.Static")
	})

	t.Run("app.Static is skipped: invoking Get handler", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(fiber.MethodGet, "/", nil)
		req.Header.Set(literal_2395, "skip")
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
		req.Header.Set(literal_2395, "don't skip")
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
		require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), literal_0385)
	})
}

func TestRouteStaticRoot(t *testing.T) {
	t.Parallel()

	dir := literal_4265
	app := fiber.New()
	app.Get("/*", New(dir, Config{
		Browse: true,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/style.css", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_1679)
	require.Contains(t, string(body), "color")

	app = fiber.New()
	app.Get("/*", New(dir))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 404, resp.StatusCode, literal_1247)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/style.css", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_1679)
	require.Contains(t, string(body), "color")
}

func TestRouteStaticHasPrefix(t *testing.T) {
	t.Parallel()

	dir := literal_4265
	app := fiber.New()
	app.Get(literal_9034, New(dir, Config{
		Browse: true,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_03125, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, literal_4273, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, literal_6243, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_1679)
	require.Contains(t, string(body), "color")

	app = fiber.New()
	app.Get("/static/*", New(dir, Config{
		Browse: true,
	}))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, literal_03125, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, literal_4273, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, literal_6243, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_1679)
	require.Contains(t, string(body), "color")

	app = fiber.New()
	app.Get(literal_9034, New(dir))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, literal_03125, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 404, resp.StatusCode, literal_1247)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, literal_4273, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 404, resp.StatusCode, literal_1247)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, literal_6243, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_1679)
	require.Contains(t, string(body), "color")

	app = fiber.New()
	app.Get(literal_9034, New(dir))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, literal_03125, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 404, resp.StatusCode, literal_1247)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, literal_4273, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 404, resp.StatusCode, literal_1247)

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, literal_6243, nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_1679)
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
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/css/style.css", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.Equal(t, fiber.MIMETextCSSCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_1679)
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
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_1679)
	require.Contains(t, string(body), "<h1>Hello, World!</h1>")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/css/style.css", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.Equal(t, fiber.MIMETextCSSCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_1679)
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
		FS:     os.DirFS(literal_4265),
		Browse: true,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/dirfs", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_1679)
	require.Contains(t, string(body), "style.css")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/dirfs/style.css", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.Equal(t, fiber.MIMETextCSSCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_1679)
	require.Contains(t, string(body), "color")

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/embed", nil))
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_1679)
	require.Contains(t, string(body), "static.go")
}

func TestStaticFSPrefixWildcard(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/test*", New(literal_4531, Config{
		FS:         os.DirFS(literal_8694),
		IndexNames: []string{"not_index.html"},
	}))

	req := httptest.NewRequest(fiber.MethodGet, "/test/john/doe", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, literal_1679)
	require.Equal(t, 200, resp.StatusCode, literal_1247)
	require.NotEmpty(t, resp.Header.Get(fiber.HeaderContentLength))
	require.Equal(t, fiber.MIMETextHTMLCharsetUTF8, resp.Header.Get(fiber.HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), literal_1925)
}

func TestIsFile(t *testing.T) {
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
			path:       literal_4531,
			filesystem: os.DirFS(literal_8694),
			expected:   true,
		},
		{
			name:       "file",
			path:       "index2.html",
			filesystem: os.DirFS(literal_8694),
			expected:   false,
			gotError:   fs.ErrNotExist,
		},
		{
			name:       "directory",
			path:       ".",
			filesystem: os.DirFS(literal_8694),
			expected:   false,
		},
		{
			name:       "directory",
			path:       "not_exists",
			filesystem: os.DirFS(literal_8694),
			expected:   false,
			gotError:   fs.ErrNotExist,
		},
		{
			name:       "directory",
			path:       ".",
			filesystem: os.DirFS(literal_4265),
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
			path:       literal_4265,
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

const literal_4531 = "index.html"

const literal_1679 = "app.Test(req)"

const literal_1247 = "Status code"

const literal_0385 = "Hello, World!"

const literal_8694 = "../../.github"

const literal_6304 = "/index.html"

const literal_2451 = "CacheControl Control"

const literal_5374 = "../../.github/test.txt"

const literal_4912 = "../../.github/index.html"

const literal_1925 = "Test file"

const literal_4265 = "../../.github/testdata/fs/css"

const literal_7493 = "/john_without_index/"

const literal_2395 = "X-Custom-Header"

const literal_9034 = "/static*"

const literal_4273 = "/static/"

const literal_6243 = "/static/style.css"

const literal_03125 = "/static"
