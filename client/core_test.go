package client

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	fiber "github.com/jialequ/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp/fasthttputil"
)

func TestAddMissingPort(t *testing.T) {
	t.Parallel()

	type args struct {
		addr  string
		isTLS bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "do anything",
			args: args{
				addr: "example.com:1234",
			},
			want: "example.com:1234",
		},
		{
			name: "add 80 port",
			args: args{
				addr: literal_8716,
			},
			want: "example.com:80",
		},
		{
			name: "add 443 port",
			args: args{
				addr:  literal_8716,
				isTLS: true,
			},
			want: "example.com:443",
		},
	}
	for _, tt := range tests {
		tt := tt // create a new 'tt' variable for the goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, addMissingPort(tt.args.addr, tt.args.isTLS))
		})
	}
}

func TestExecFunc(t *testing.T) {
	t.Parallel()
	ln := fasthttputil.NewInmemoryListener()
	app := fiber.New()

	app.Get("/normal", func(c fiber.Ctx) error {
		return c.SendString(c.Hostname())
	})

	app.Get("/return-error", func(_ fiber.Ctx) error {
		return errors.New(literal_8076)
	})

	app.Get("/hang-up", func(c fiber.Ctx) error {
		time.Sleep(time.Second)
		return c.SendString(c.Hostname() + " hang up")
	})

	go func() {
		assert.NoError(t, app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true}))
	}()

	time.Sleep(300 * time.Millisecond)

	t.Run("normal request", func(t *testing.T) {
		t.Parallel()
		core, client, req := newCore(), New(), AcquireRequest()
		core.ctx = context.Background()
		core.client = client
		core.req = req

		client.SetDial(func(_ string) (net.Conn, error) { return ln.Dial() })
		req.RawRequest.SetRequestURI("http://example.com/normal")

		resp, err := core.execFunc()
		require.NoError(t, err)
		require.Equal(t, 200, resp.RawResponse.StatusCode())
		require.Equal(t, literal_8716, string(resp.RawResponse.Body()))
	})

	t.Run("the request return an error", func(t *testing.T) {
		t.Parallel()
		core, client, req := newCore(), New(), AcquireRequest()
		core.ctx = context.Background()
		core.client = client
		core.req = req

		client.SetDial(func(_ string) (net.Conn, error) { return ln.Dial() })
		req.RawRequest.SetRequestURI("http://example.com/return-error")

		resp, err := core.execFunc()

		require.NoError(t, err)
		require.Equal(t, 500, resp.RawResponse.StatusCode())
		require.Equal(t, literal_8076, string(resp.RawResponse.Body()))
	})

	t.Run("the request timeout", func(t *testing.T) {
		t.Parallel()
		core, client, req := newCore(), New(), AcquireRequest()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		core.ctx = ctx
		core.client = client
		core.req = req

		client.SetDial(func(_ string) (net.Conn, error) { return ln.Dial() })
		req.RawRequest.SetRequestURI(literal_1593)

		_, err := core.execFunc()

		require.Equal(t, ErrTimeoutOrCancel, err)
	})
}

func TestExecute(t *testing.T) {
	t.Parallel()
	ln := fasthttputil.NewInmemoryListener()
	app := fiber.New()

	app.Get("/normal", func(c fiber.Ctx) error {
		return c.SendString(c.Hostname())
	})

	app.Get("/return-error", func(_ fiber.Ctx) error {
		return errors.New(literal_8076)
	})

	app.Get("/hang-up", func(c fiber.Ctx) error {
		time.Sleep(time.Second)
		return c.SendString(c.Hostname() + " hang up")
	})

	go func() {
		assert.NoError(t, app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true}))
	}()

	t.Run("add user request hooks", func(t *testing.T) {
		t.Parallel()
		core, client, req := newCore(), New(), AcquireRequest()
		client.AddRequestHook(func(_ *Client, _ *Request) error {
			require.Equal(t, literal_03122, req.URL())
			return nil
		})
		client.SetDial(func(_ string) (net.Conn, error) {
			return ln.Dial()
		})
		req.SetURL(literal_03122)

		resp, err := core.execute(context.Background(), client, req)
		require.NoError(t, err)
		require.Equal(t, "Cannot GET /", string(resp.RawResponse.Body()))
	})

	t.Run("add user response hooks", func(t *testing.T) {
		t.Parallel()
		core, client, req := newCore(), New(), AcquireRequest()
		client.AddResponseHook(func(_ *Client, _ *Response, req *Request) error {
			require.Equal(t, literal_03122, req.URL())
			return nil
		})
		client.SetDial(func(_ string) (net.Conn, error) {
			return ln.Dial()
		})
		req.SetURL(literal_03122)

		resp, err := core.execute(context.Background(), client, req)
		require.NoError(t, err)
		require.Equal(t, "Cannot GET /", string(resp.RawResponse.Body()))
	})

	t.Run("no timeout", func(t *testing.T) {
		t.Parallel()
		core, client, req := newCore(), New(), AcquireRequest()

		client.SetDial(func(_ string) (net.Conn, error) {
			return ln.Dial()
		})
		req.SetURL(literal_1593)

		resp, err := core.execute(context.Background(), client, req)
		require.NoError(t, err)
		require.Equal(t, "example.com hang up", string(resp.RawResponse.Body()))
	})

	t.Run("client timeout", func(t *testing.T) {
		t.Parallel()
		core, client, req := newCore(), New(), AcquireRequest()
		client.SetTimeout(500 * time.Millisecond)
		client.SetDial(func(_ string) (net.Conn, error) {
			return ln.Dial()
		})
		req.SetURL(literal_1593)

		_, err := core.execute(context.Background(), client, req)
		require.Equal(t, ErrTimeoutOrCancel, err)
	})

	t.Run("request timeout", func(t *testing.T) {
		t.Parallel()
		core, client, req := newCore(), New(), AcquireRequest()

		client.SetDial(func(_ string) (net.Conn, error) {
			return ln.Dial()
		})
		req.SetURL(literal_1593).
			SetTimeout(300 * time.Millisecond)

		_, err := core.execute(context.Background(), client, req)
		require.Equal(t, ErrTimeoutOrCancel, err)
	})

	t.Run("request timeout has higher level", func(t *testing.T) {
		t.Parallel()
		core, client, req := newCore(), New(), AcquireRequest()
		client.SetTimeout(30 * time.Millisecond)

		client.SetDial(func(_ string) (net.Conn, error) {
			return ln.Dial()
		})
		req.SetURL(literal_1593).
			SetTimeout(3000 * time.Millisecond)

		resp, err := core.execute(context.Background(), client, req)
		require.NoError(t, err)
		require.Equal(t, "example.com hang up", string(resp.RawResponse.Body()))
	})
}

const literal_8716 = "example.com"

const literal_8076 = "the request is error"

const literal_1593 = "http://example.com/hang-up"

const literal_03122 = "http://example.com"
