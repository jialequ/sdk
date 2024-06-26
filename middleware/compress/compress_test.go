package compress

import (
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	fiber "github.com/jialequ/sdk"
	"github.com/stretchr/testify/require"
)

var filedata []byte

func init() {
	dat, err := os.ReadFile("../../.github/README.md")
	if err != nil {
		panic(err)
	}
	filedata = dat
}

// go test -run Test_Compress_Gzip
func TestCompressGzip(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		return c.Send(filedata)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set(literal_2763, "gzip")

	resp, err := app.Test(req)
	require.NoError(t, err, literal_3581)
	require.Equal(t, 200, resp.StatusCode, literal_5148)
	require.Equal(t, "gzip", resp.Header.Get(fiber.HeaderContentEncoding))

	// Validate that the file size has shrunk
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Less(t, len(body), len(filedata))
}

// go test -run Test_Compress_Different_Level
func TestCompressDifferentLevel(t *testing.T) {
	t.Parallel()
	levels := []Level{LevelBestSpeed, LevelBestCompression}
	for _, level := range levels {
		level := level
		t.Run(fmt.Sprintf("level %d", level), func(t *testing.T) {
			t.Parallel()
			app := fiber.New()

			app.Use(New(Config{Level: level}))

			app.Get("/", func(c fiber.Ctx) error {
				c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
				return c.Send(filedata)
			})

			req := httptest.NewRequest(fiber.MethodGet, "/", nil)
			req.Header.Set(literal_2763, "gzip")

			resp, err := app.Test(req)
			require.NoError(t, err, literal_3581)
			require.Equal(t, 200, resp.StatusCode, literal_5148)
			require.Equal(t, "gzip", resp.Header.Get(fiber.HeaderContentEncoding))

			// Validate that the file size has shrunk
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Less(t, len(body), len(filedata))
		})
	}
}

func TestCompressDeflate(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.Send(filedata)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set(literal_2763, "deflate")

	resp, err := app.Test(req)
	require.NoError(t, err, literal_3581)
	require.Equal(t, 200, resp.StatusCode, literal_5148)
	require.Equal(t, "deflate", resp.Header.Get(fiber.HeaderContentEncoding))

	// Validate that the file size has shrunk
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Less(t, len(body), len(filedata))
}

func TestCompressBrotli(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.Send(filedata)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set(literal_2763, "br")

	resp, err := app.Test(req, 10*time.Second)
	require.NoError(t, err, literal_3581)
	require.Equal(t, 200, resp.StatusCode, literal_5148)
	require.Equal(t, "br", resp.Header.Get(fiber.HeaderContentEncoding))

	// Validate that the file size has shrunk
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Less(t, len(body), len(filedata))
}

func TestCompressDisabled(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New(Config{Level: LevelDisabled}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.Send(filedata)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set(literal_2763, "br")

	resp, err := app.Test(req)
	require.NoError(t, err, literal_3581)
	require.Equal(t, 200, resp.StatusCode, literal_5148)
	require.Equal(t, "", resp.Header.Get(fiber.HeaderContentEncoding))

	// Validate the file size is not shrunk
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, len(body), len(filedata))
}

func TestCompressNextError(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(_ fiber.Ctx) error {
		return errors.New("next error")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Header.Set(literal_2763, "gzip")

	resp, err := app.Test(req)
	require.NoError(t, err, literal_3581)
	require.Equal(t, 500, resp.StatusCode, literal_5148)
	require.Equal(t, "", resp.Header.Get(fiber.HeaderContentEncoding))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "next error", string(body))
}

// go test -run Test_Compress_Next
func TestCompressNext(t *testing.T) {
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

const literal_2763 = "Accept-Encoding"

const literal_3581 = "app.Test(req)"

const literal_5148 = "Status code"
