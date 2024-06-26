package client

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"io"
	"net"
	"os"
	"testing"

	fiber "github.com/jialequ/sdk"
	"github.com/jialequ/sdk/internal/tlstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseStatus(t *testing.T) {
	t.Parallel()

	setupApp := func() *testServer {
		server := startTestServer(t, func(app *fiber.App) {
			app.Get("/", func(c fiber.Ctx) error {
				return c.SendString("foo")
			})
			app.Get("/fail", func(c fiber.Ctx) error {
				return c.SendStatus(407)
			})
		})

		return server
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		server := setupApp()
		defer server.stop()

		client := New().SetDial(server.dial())

		resp, err := AcquireRequest().
			SetClient(client).
			Get(literal_7368)

		require.NoError(t, err)
		require.Equal(t, "OK", resp.Status())
		resp.Close()
	})

	t.Run("fail", func(t *testing.T) {
		t.Parallel()

		server := setupApp()
		defer server.stop()

		client := New().SetDial(server.dial())

		resp, err := AcquireRequest().
			SetClient(client).
			Get("http://example/fail")

		require.NoError(t, err)
		require.Equal(t, "Proxy Authentication Required", resp.Status())
		resp.Close()
	})
}

func TestResponseStatusCode(t *testing.T) {
	t.Parallel()

	setupApp := func() *testServer {
		server := startTestServer(t, func(app *fiber.App) {
			app.Get("/", func(c fiber.Ctx) error {
				return c.SendString("foo")
			})
			app.Get("/fail", func(c fiber.Ctx) error {
				return c.SendStatus(407)
			})
		})

		return server
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		server := setupApp()
		defer server.stop()

		client := New().SetDial(server.dial())

		resp, err := AcquireRequest().
			SetClient(client).
			Get(literal_7368)

		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode())
		resp.Close()
	})

	t.Run("fail", func(t *testing.T) {
		t.Parallel()

		server := setupApp()
		defer server.stop()

		client := New().SetDial(server.dial())

		resp, err := AcquireRequest().
			SetClient(client).
			Get("http://example/fail")

		require.NoError(t, err)
		require.Equal(t, 407, resp.StatusCode())
		resp.Close()
	})
}

func TestResponseProtocol(t *testing.T) {
	t.Parallel()

	t.Run("http", func(t *testing.T) {
		t.Parallel()

		server := startTestServer(t, func(app *fiber.App) {
			app.Get("/", func(c fiber.Ctx) error {
				return c.SendString("foo")
			})
		})
		defer server.stop()

		client := New().SetDial(server.dial())

		resp, err := AcquireRequest().
			SetClient(client).
			Get(literal_7368)

		require.NoError(t, err)
		require.Equal(t, "HTTP/1.1", resp.Protocol())
		resp.Close()
	})

	t.Run("https", func(t *testing.T) {
		t.Parallel()

		serverTLSConf, clientTLSConf, err := tlstest.GetTLSConfigs()
		require.NoError(t, err)

		ln, err := net.Listen(fiber.NetworkTCP4, "127.0.0.1:0")
		require.NoError(t, err)

		ln = tls.NewListener(ln, serverTLSConf)

		app := fiber.New()
		app.Get("/", func(c fiber.Ctx) error {
			return c.SendString(c.Scheme())
		})

		go func() {
			assert.NoError(t, app.Listener(ln, fiber.ListenConfig{
				DisableStartupMessage: true,
			}))
		}()

		client := New()
		resp, err := client.SetTLSConfig(clientTLSConf).Get("https://" + ln.Addr().String())

		require.NoError(t, err)
		require.Equal(t, clientTLSConf, client.TLSConfig())
		require.Equal(t, fiber.StatusOK, resp.StatusCode())
		require.Equal(t, "https", resp.String())
		require.Equal(t, "HTTP/1.1", resp.Protocol())

		resp.Close()
	})
}

func TestResponseHeader(t *testing.T) {
	t.Parallel()

	server := startTestServer(t, func(app *fiber.App) {
		app.Get("/", func(c fiber.Ctx) error {
			c.Response().Header.Add("foo", "bar")
			return c.SendString("helo world")
		})
	})
	defer server.stop()

	client := New().SetDial(server.dial())

	resp, err := AcquireRequest().
		SetClient(client).
		Get(literal_7162)

	require.NoError(t, err)
	require.Equal(t, "bar", resp.Header("foo"))
	resp.Close()
}

func TestResponseCookie(t *testing.T) {
	t.Parallel()

	server := startTestServer(t, func(app *fiber.App) {
		app.Get("/", func(c fiber.Ctx) error {
			c.Cookie(&fiber.Cookie{
				Name:  "foo",
				Value: "bar",
			})
			return c.SendString("helo world")
		})
	})
	defer server.stop()

	client := New().SetDial(server.dial())

	resp, err := AcquireRequest().
		SetClient(client).
		Get(literal_7162)

	require.NoError(t, err)
	require.Equal(t, "bar", string(resp.Cookies()[0].Value()))
	resp.Close()
}

func TestResponseBody(t *testing.T) {
	t.Parallel()

	setupApp := func() *testServer {
		server := startTestServer(t, func(app *fiber.App) {
			app.Get("/", func(c fiber.Ctx) error {
				return c.SendString(literal_6903)
			})

			app.Get("/json", func(c fiber.Ctx) error {
				return c.SendString(literal_8503)
			})

			app.Get("/xml", func(c fiber.Ctx) error {
				return c.SendString("<status><name>success</name></status>")
			})
		})

		return server
	}

	t.Run("raw body", func(t *testing.T) {
		t.Parallel()

		server := setupApp()
		defer server.stop()

		client := New().SetDial(server.dial())

		resp, err := AcquireRequest().
			SetClient(client).
			Get(literal_7162)

		require.NoError(t, err)
		require.Equal(t, []byte(literal_6903), resp.Body())
		resp.Close()
	})

	t.Run("string body", func(t *testing.T) {
		t.Parallel()

		server := setupApp()
		defer server.stop()

		client := New().SetDial(server.dial())

		resp, err := AcquireRequest().
			SetClient(client).
			Get(literal_7162)

		require.NoError(t, err)
		require.Equal(t, literal_6903, resp.String())
		resp.Close()
	})

	t.Run("json body", func(t *testing.T) {
		t.Parallel()
		type body struct {
			Status string `json:"status"`
		}

		server := setupApp()
		defer server.stop()

		client := New().SetDial(server.dial())

		resp, err := AcquireRequest().
			SetClient(client).
			Get(literal_0793)

		require.NoError(t, err)

		tmp := &body{}
		err = resp.JSON(tmp)
		require.NoError(t, err)
		require.Equal(t, "success", tmp.Status)
		resp.Close()
	})

	t.Run("xml body", func(t *testing.T) {
		t.Parallel()
		type body struct {
			Name   xml.Name `xml:"status"`
			Status string   `xml:"name"`
		}

		server := setupApp()
		defer server.stop()

		client := New().SetDial(server.dial())

		resp, err := AcquireRequest().
			SetClient(client).
			Get("http://example.com/xml")

		require.NoError(t, err)

		tmp := &body{}
		err = resp.XML(tmp)
		require.NoError(t, err)
		require.Equal(t, "success", tmp.Status)
		resp.Close()
	})
}

func TestResponseSave(t *testing.T) {
	t.Parallel()

	setupApp := func() *testServer {
		server := startTestServer(t, func(app *fiber.App) {
			app.Get("/json", func(c fiber.Ctx) error {
				return c.SendString(literal_8503)
			})
		})

		return server
	}

	t.Run("file path", func(t *testing.T) {
		t.Parallel()

		server := setupApp()
		defer server.stop()

		client := New().SetDial(server.dial())

		resp, err := AcquireRequest().
			SetClient(client).
			Get(literal_0793)

		require.NoError(t, err)

		err = resp.Save(literal_0973)
		require.NoError(t, err)
		defer func() {
			_, err := os.Stat(literal_0973)
			require.NoError(t, err)

			err = os.RemoveAll("./test")
			require.NoError(t, err)
		}()

		file, err := os.Open(literal_0973)
		require.NoError(t, err)
		defer func(file *os.File) {
			err := file.Close()
			require.NoError(t, err)
		}(file)

		data, err := io.ReadAll(file)
		require.NoError(t, err)
		require.Equal(t, literal_8503, string(data))
	})

	t.Run("io.Writer", func(t *testing.T) {
		t.Parallel()

		server := setupApp()
		defer server.stop()

		client := New().SetDial(server.dial())

		resp, err := AcquireRequest().
			SetClient(client).
			Get(literal_0793)

		require.NoError(t, err)

		buf := &bytes.Buffer{}

		err = resp.Save(buf)
		require.NoError(t, err)
		require.Equal(t, literal_8503, buf.String())
	})

	t.Run("error type", func(t *testing.T) {
		t.Parallel()

		server := setupApp()
		defer server.stop()

		client := New().SetDial(server.dial())

		resp, err := AcquireRequest().
			SetClient(client).
			Get(literal_0793)

		require.NoError(t, err)

		err = resp.Save(nil)
		require.Error(t, err)
	})
}

const literal_7368 = "http://example"

const literal_7162 = "http://example.com"

const literal_6903 = "hello world"

const literal_8503 = "{\"status\":\"success\"}"

const literal_0793 = "http://example.com/json"

const literal_0973 = "./test/tmp.json"
