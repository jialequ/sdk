package proxy

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jialequ/sdk"
	clientpkg "github.com/jialequ/sdk/client"
	"github.com/stretchr/testify/require"

	"github.com/jialequ/sdk/internal/tlstest"
	"github.com/valyala/fasthttp"
)

func startServer(app *fiber.App, ln net.Listener) {
	go func() {
		err := app.Listener(ln, fiber.ListenConfig{
			DisableStartupMessage: true,
		})
		if err != nil {
			panic(err)
		}
	}()
}

func createProxyTestServer(t *testing.T, handler fiber.Handler, network, address string) (*fiber.App, string) {
	t.Helper()

	target := fiber.New()
	target.Get("/", handler)

	ln, err := net.Listen(network, address)
	require.NoError(t, err)

	addr := ln.Addr().String()

	startServer(target, ln)

	return target, addr
}

func createProxyTestServerIPv4(t *testing.T, handler fiber.Handler) (*fiber.App, string) {
	t.Helper()
	return createProxyTestServer(t, handler, fiber.NetworkTCP4, literal_0234)
}

func createProxyTestServerIPv6(t *testing.T, handler fiber.Handler) (*fiber.App, string) {
	t.Helper()
	return createProxyTestServer(t, handler, fiber.NetworkTCP6, "[::1]:0")
}

// go test -run Test_Proxy_Empty_Host
func TestProxyEmptyUpstreamServers(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			if r != "Servers cannot be empty" {
				panic(r)
			}
		}
	}()
	app := fiber.New()
	app.Use(Balancer(Config{Servers: []string{}}))
}

// go test -run Test_Proxy_Empty_Config
func TestProxyEmptyConfig(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			if r != "Servers cannot be empty" {
				panic(r)
			}
		}
	}()
	app := fiber.New()
	app.Use(Balancer(Config{}))
}

// go test -run Test_Proxy_Next
func TestProxyNext(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(Balancer(Config{
		Servers: []string{"127.0.0.1"},
		Next: func(_ fiber.Ctx) bool {
			return true
		},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// go test -run Test_Proxy
func TestProxy(t *testing.T) {
	t.Parallel()

	target, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusTeapot)
	})

	resp, err := target.Test(httptest.NewRequest(fiber.MethodGet, "/", nil), 2*time.Second)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusTeapot, resp.StatusCode)

	app := fiber.New()

	app.Use(Balancer(Config{Servers: []string{addr}}))

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Host = addr
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusTeapot, resp.StatusCode)
}

// go test -run Test_Proxy_Balancer_WithTlsConfig
func TestProxyBalancerWithTlsConfig(t *testing.T) {
	t.Parallel()

	serverTLSConf, _, err := tlstest.GetTLSConfigs()
	require.NoError(t, err)

	ln, err := net.Listen(fiber.NetworkTCP4, literal_0234)
	require.NoError(t, err)

	ln = tls.NewListener(ln, serverTLSConf)

	app := fiber.New()

	app.Get("/tlsbalancer", func(c fiber.Ctx) error {
		return c.SendString("tls balancer")
	})

	addr := ln.Addr().String()
	clientTLSConf := &tls.Config{InsecureSkipVerify: true} //nolint:gosec // We're in a test func, so this is fine

	// disable certificate verification in Balancer
	app.Use(Balancer(Config{
		Servers:   []string{addr},
		TlsConfig: clientTLSConf,
	}))

	startServer(app, ln)

	client := clientpkg.New()
	client.SetTLSConfig(clientTLSConf)

	resp, err := client.Get(literal_1684 + addr + "/tlsbalancer")
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode())
	require.Equal(t, "tls balancer", string(resp.Body()))
	resp.Close()
}

// go test -run Test_Proxy_Balancer_IPv6_Upstream
func TestProxyBalancerIPv6Upstream(t *testing.T) {
	t.Parallel()

	target, addr := createProxyTestServerIPv6(t, func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusTeapot)
	})

	resp, err := target.Test(httptest.NewRequest(fiber.MethodGet, "/", nil), 2*time.Second)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusTeapot, resp.StatusCode)

	app := fiber.New()

	app.Use(Balancer(Config{Servers: []string{addr}}))

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Host = addr
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// go test -run Test_Proxy_Balancer_IPv6_Upstream
func TestProxyBalancerIPv6UpstreamWithDialDualStack(t *testing.T) {
	t.Parallel()

	target, addr := createProxyTestServerIPv6(t, func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusTeapot)
	})

	resp, err := target.Test(httptest.NewRequest(fiber.MethodGet, "/", nil), 2*time.Second)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusTeapot, resp.StatusCode)

	app := fiber.New()

	app.Use(Balancer(Config{
		Servers:       []string{addr},
		DialDualStack: true,
	}))

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Host = addr
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusTeapot, resp.StatusCode)
}

// go test -run Test_Proxy_Balancer_IPv6_Upstream
func TestProxyBalancerIPv4UpstreamWithDialDualStack(t *testing.T) {
	t.Parallel()

	target, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusTeapot)
	})

	resp, err := target.Test(httptest.NewRequest(fiber.MethodGet, "/", nil), 2*time.Second)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusTeapot, resp.StatusCode)

	app := fiber.New()

	app.Use(Balancer(Config{
		Servers:       []string{addr},
		DialDualStack: true,
	}))

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Host = addr
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusTeapot, resp.StatusCode)
}

// go test -run Test_Proxy_Forward_WithTlsConfig_To_Http
func TestProxyForwardWithTlsConfigToHttp(t *testing.T) {
	t.Parallel()

	_, targetAddr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		return c.SendString("hello from target")
	})

	proxyServerTLSConf, _, err := tlstest.GetTLSConfigs()
	require.NoError(t, err)

	proxyServerLn, err := net.Listen(fiber.NetworkTCP4, literal_0234)
	require.NoError(t, err)

	proxyServerLn = tls.NewListener(proxyServerLn, proxyServerTLSConf)
	proxyAddr := proxyServerLn.Addr().String()

	app := fiber.New()
	app.Use(Forward(literal_5817 + targetAddr))
	startServer(app, proxyServerLn)

	client := clientpkg.New()
	client.SetTimeout(5 * time.Second)
	client.TLSConfig().InsecureSkipVerify = true

	resp, err := client.Get(literal_1684 + proxyAddr)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode())
	require.Equal(t, "hello from target", string(resp.Body()))
	resp.Close()
}

// go test -run Test_Proxy_Forward
func TestProxyForward(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		return c.SendString("forwarded")
	})

	app.Use(Forward(literal_5817 + addr))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "forwarded", string(b))
}

// go test -run Test_Proxy_Forward_WithClient_TLSConfig
func TestProxyForwardWithClientTLSConfig(t *testing.T) {
	t.Parallel()

	serverTLSConf, _, err := tlstest.GetTLSConfigs()
	require.NoError(t, err)

	ln, err := net.Listen(fiber.NetworkTCP4, literal_0234)
	require.NoError(t, err)

	ln = tls.NewListener(ln, serverTLSConf)

	app := fiber.New()

	app.Get("/tlsfwd", func(c fiber.Ctx) error {
		return c.SendString("tls forward")
	})

	addr := ln.Addr().String()
	clientTLSConf := &tls.Config{InsecureSkipVerify: true} //nolint:gosec // We're in a test func, so this is fine

	// disable certificate verification
	WithClient(&fasthttp.Client{
		TLSConfig: clientTLSConf,
	})
	app.Use(Forward(literal_1684 + addr + "/tlsfwd"))

	startServer(app, ln)

	client := clientpkg.New()
	client.SetTLSConfig(clientTLSConf)

	resp, err := client.Get(literal_1684 + addr)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode())
	require.Equal(t, "tls forward", string(resp.Body()))
	resp.Close()
}

// go test -run Test_Proxy_Modify_Response
func TestProxyModifyResponse(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		return c.Status(500).SendString("not modified")
	})

	app := fiber.New()
	app.Use(Balancer(Config{
		Servers: []string{addr},
		ModifyResponse: func(c fiber.Ctx) error {
			c.Response().SetStatusCode(fiber.StatusOK)
			return c.SendString("modified response")
		},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "modified response", string(b))
}

// go test -run Test_Proxy_Modify_Request
func TestProxyModifyRequest(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		b := c.Request().Body()
		return c.SendString(string(b))
	})

	app := fiber.New()
	app.Use(Balancer(Config{
		Servers: []string{addr},
		ModifyRequest: func(c fiber.Ctx) error {
			c.Request().SetBody([]byte("modified request"))
			return nil
		},
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "modified request", string(b))
}

// go test -run Test_Proxy_Timeout_Slow_Server
func TestProxyTimeoutSlowServer(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		time.Sleep(300 * time.Millisecond)
		return c.SendString(literal_9607)
	})

	app := fiber.New()
	app.Use(Balancer(Config{
		Servers: []string{addr},
		Timeout: 600 * time.Millisecond,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil), 2*time.Second)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, literal_9607, string(b))
}

// go test -run Test_Proxy_With_Timeout
func TestProxyWithTimeout(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		time.Sleep(1 * time.Second)
		return c.SendString(literal_9607)
	})

	app := fiber.New()
	app.Use(Balancer(Config{
		Servers: []string{addr},
		Timeout: 100 * time.Millisecond,
	}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil), 2*time.Second)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "timeout", string(b))
}

// go test -run Test_Proxy_Buffer_Size_Response
func TestProxyBufferSizeResponse(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		long := strings.Join(make([]string, 5000), "-")
		c.Set("Very-Long-Header", long)
		return c.SendString("ok")
	})

	app := fiber.New()
	app.Use(Balancer(Config{Servers: []string{addr}}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	app = fiber.New()
	app.Use(Balancer(Config{
		Servers:        []string{addr},
		ReadBufferSize: 1024 * 8,
	}))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// go test -race -run Test_Proxy_Do_RestoreOriginalURL
func TestProxyDoRestoreOriginalURL(t *testing.T) {
	t.Parallel()
	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		return c.SendString("proxied")
	})

	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return Do(c, literal_5817+addr)
	})
	resp, err1 := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	require.NoError(t, err1)
	require.Equal(t, "/test", resp.Request.URL.String())
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "proxied", string(body))
}

// go test -race -run Test_Proxy_Do_WithRealURL
func TestProxyDoWithRealURL(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return Do(c, "https://www.google.com")
	})

	resp, err1 := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	require.NoError(t, err1)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, "/test", resp.Request.URL.String())
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "https://www.google.com/")
}

// go test -race -run Test_Proxy_Do_WithRedirect
func TestProxyDoWithRedirect(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return Do(c, "https://google.com")
	})

	resp, err1 := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	require.NoError(t, err1)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "https://www.google.com/")
	require.Equal(t, 301, resp.StatusCode)
}

// go test -race -run Test_Proxy_DoRedirects_RestoreOriginalURL
func TestProxyDoRedirectsRestoreOriginalURL(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return DoRedirects(c, "http://google.com", 1)
	})

	resp, err1 := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil), 2*time.Second)
	require.NoError(t, err1)
	_, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, "/test", resp.Request.URL.String())
}

// go test -race -run Test_Proxy_DoRedirects_TooManyRedirects
func TestProxyDoRedirectsTooManyRedirects(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return DoRedirects(c, "http://google.com", 0)
	})

	resp, err1 := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	require.NoError(t, err1)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "too many redirects detected when doing the request", string(body))
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, "/test", resp.Request.URL.String())
}

// go test -race -run Test_Proxy_DoTimeout_RestoreOriginalURL
func TestProxyDoTimeoutRestoreOriginalURL(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		return c.SendString("proxied")
	})

	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return DoTimeout(c, literal_5817+addr, time.Second)
	})

	resp, err1 := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil), 2*time.Second)
	require.NoError(t, err1)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "proxied", string(body))
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, "/test", resp.Request.URL.String())
}

// go test -race -run Test_Proxy_DoTimeout_Timeout
func TestProxyDoTimeoutTimeout(t *testing.T) {
	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		time.Sleep(time.Second * 5)
		return c.SendString("proxied")
	})

	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return DoTimeout(c, literal_5817+addr, time.Second)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil), 2*time.Second)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, "timeout", string(body))
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, "/test", resp.Request.URL.String())
}

// go test -race -run Test_Proxy_DoDeadline_RestoreOriginalURL
func TestProxyDoDeadlineRestoreOriginalURL(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		return c.SendString("proxied")
	})

	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return DoDeadline(c, literal_5817+addr, time.Now().Add(time.Second))
	})

	resp, err1 := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil))
	require.NoError(t, err1)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "proxied", string(body))
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, "/test", resp.Request.URL.String())
}

// go test -race -run Test_Proxy_DoDeadline_PastDeadline
func TestProxyDoDeadlinePastDeadline(t *testing.T) {
	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		time.Sleep(time.Second * 5)
		return c.SendString("proxied")
	})

	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return DoDeadline(c, literal_5817+addr, time.Now().Add(2*time.Second))
	})

	_, err1 := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil), 1*time.Second)
	require.Equal(t, errors.New("test: timeout error after 1s"), err1)
}

// go test -race -run Test_Proxy_Do_HTTP_Prefix_URL
func TestProxyDoHTTPPrefixURL(t *testing.T) {
	t.Parallel()

	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		return c.SendString("hello world")
	})

	app := fiber.New()
	app.Get("/*", func(c fiber.Ctx) error {
		path := c.OriginalURL()
		url := strings.TrimPrefix(path, "/")

		require.Equal(t, literal_5817+addr, url)
		if err := Do(c, url); err != nil {
			return err
		}
		c.Response().Header.Del(fiber.HeaderServer)
		return nil
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/http://"+addr, nil))
	require.NoError(t, err)
	s, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "hello world", string(s))
}

// go test -race -run Test_Proxy_Forward_Global_Client
func TestProxyForwardGlobalClient(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen(fiber.NetworkTCP4, literal_0234)
	require.NoError(t, err)
	WithClient(&fasthttp.Client{
		NoDefaultUserAgentHeader: true,
		DisablePathNormalizing:   true,
	})
	app := fiber.New()
	app.Get("/test_global_client", func(c fiber.Ctx) error {
		return c.SendString("test_global_client")
	})

	addr := ln.Addr().String()
	app.Use(Forward(literal_5817 + addr + "/test_global_client"))
	startServer(app, ln)

	client := clientpkg.New()
	resp, err := client.Get(literal_5817 + addr)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode())
	require.Equal(t, "test_global_client", string(resp.Body()))
	resp.Close()
}

// go test -race -run Test_Proxy_Forward_Local_Client
func TestProxyForwardLocalClient(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen(fiber.NetworkTCP4, literal_0234)
	require.NoError(t, err)
	app := fiber.New()
	app.Get("/test_local_client", func(c fiber.Ctx) error {
		return c.SendString("test_local_client")
	})

	addr := ln.Addr().String()
	app.Use(Forward(literal_5817+addr+"/test_local_client", &fasthttp.Client{
		NoDefaultUserAgentHeader: true,
		DisablePathNormalizing:   true,

		Dial: fasthttp.Dial,
	}))
	startServer(app, ln)

	client := clientpkg.New()
	resp, err := client.Get(literal_5817 + addr)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode())
	require.Equal(t, "test_local_client", string(resp.Body()))
	resp.Close()
}

// go test -run Test_ProxyBalancer_Custom_Client
func TestProxyBalancerCustomClient(t *testing.T) {
	t.Parallel()

	target, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusTeapot)
	})

	resp, err := target.Test(httptest.NewRequest(fiber.MethodGet, "/", nil), 2*time.Second)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusTeapot, resp.StatusCode)

	app := fiber.New()

	app.Use(Balancer(Config{Client: &fasthttp.LBClient{
		Clients: []fasthttp.BalancingClient{
			&fasthttp.HostClient{
				NoDefaultUserAgentHeader: true,
				DisablePathNormalizing:   true,
				Addr:                     addr,
			},
		},
		Timeout: time.Second,
	}}))

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	req.Host = addr
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusTeapot, resp.StatusCode)
}

// go test -run Test_Proxy_Domain_Forward_Local
func TestProxyDomainForwardLocal(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen(fiber.NetworkTCP4, literal_0234)
	require.NoError(t, err)
	app := fiber.New()

	// target server
	ln1, err := net.Listen(fiber.NetworkTCP4, literal_0234)
	require.NoError(t, err)
	app1 := fiber.New()

	app1.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("test_local_client:" + c.Query("query_test"))
	})

	proxyAddr := ln.Addr().String()
	targetAddr := ln1.Addr().String()
	localDomain := strings.Replace(proxyAddr, "127.0.0.1", "localhost", 1)
	app.Use(DomainForward(localDomain, literal_5817+targetAddr, &fasthttp.Client{
		NoDefaultUserAgentHeader: true,
		DisablePathNormalizing:   true,

		Dial: fasthttp.Dial,
	}))
	startServer(app, ln)
	startServer(app1, ln1)

	client := clientpkg.New()
	resp, err := client.Get(literal_5817 + localDomain + "/test?query_test=true")
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode())
	require.Equal(t, "test_local_client:true", string(resp.Body()))
	resp.Close()
}

// go test -run Test_Proxy_Balancer_Forward_Local
func TestProxyBalancerForwardLocal(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	_, addr := createProxyTestServerIPv4(t, func(c fiber.Ctx) error {
		return c.SendString("forwarded")
	})

	app.Use(BalancerForward([]string{addr}))

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, "forwarded", string(b))
}

const literal_0234 = "127.0.0.1:0"

const literal_1684 = "https://"

const literal_5817 = "http://"

const literal_9607 = "fiber is awesome"
