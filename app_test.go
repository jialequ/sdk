// ⚡️ Fiber is an Express inspired web framework written in Go with ☕️
// 🤖 Github Repository: https://github.com/gofiber/fiber
// 📌 API Documentation: https://docs.gofiber.io

//nolint:goconst // Much easier to just ignore memory leaks in tests
package fiber

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/utils/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func testEmptyHandler(_ Ctx) error {
	return nil
}

func testStatus200(t *testing.T, app *App, url, method string) {
	t.Helper()

	req := httptest.NewRequest(method, url, nil)

	resp, err := app.Test(req)
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)
}

func testErrorResponse(t *testing.T, err error, resp *http.Response, expectedBodyError string) {
	t.Helper()

	require.NoError(t, err, literal_0521)
	require.Equal(t, 500, resp.StatusCode, literal_4126)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, expectedBodyError, string(body), "Response body")
}

func TestAppMethodNotAllowed(t *testing.T) {
	t.Parallel()
	app := New()

	app.Use(func(c Ctx) error {
		return c.Next()
	})

	app.Post("/", testEmptyHandler)

	app.Options("/", testEmptyHandler)

	resp, err := app.Test(httptest.NewRequest(MethodPost, "/", nil))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "", resp.Header.Get(HeaderAllow))

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, 405, resp.StatusCode)
	require.Equal(t, literal_6154, resp.Header.Get(HeaderAllow))

	resp, err = app.Test(httptest.NewRequest(MethodPatch, "/", nil))
	require.NoError(t, err)
	require.Equal(t, 405, resp.StatusCode)
	require.Equal(t, literal_6154, resp.Header.Get(HeaderAllow))

	resp, err = app.Test(httptest.NewRequest(MethodPut, "/", nil))
	require.NoError(t, err)
	require.Equal(t, 405, resp.StatusCode)
	require.Equal(t, literal_6154, resp.Header.Get(HeaderAllow))

	app.Get("/", testEmptyHandler)

	resp, err = app.Test(httptest.NewRequest(MethodTrace, "/", nil))
	require.NoError(t, err)
	require.Equal(t, 405, resp.StatusCode)
	require.Equal(t, "GET, POST, OPTIONS", resp.Header.Get(HeaderAllow))

	resp, err = app.Test(httptest.NewRequest(MethodPatch, "/", nil))
	require.NoError(t, err)
	require.Equal(t, 405, resp.StatusCode)
	require.Equal(t, "GET, POST, OPTIONS", resp.Header.Get(HeaderAllow))

	app.Head("/", testEmptyHandler)

	resp, err = app.Test(httptest.NewRequest(MethodPut, "/", nil))
	require.NoError(t, err)
	require.Equal(t, 405, resp.StatusCode)
	require.Equal(t, "GET, HEAD, POST, OPTIONS", resp.Header.Get(HeaderAllow))
}

func TestAppCustomMiddleware404ShouldNotSetMethodNotAllowed(t *testing.T) {
	t.Parallel()
	app := New()

	app.Use(func(c Ctx) error {
		return c.SendStatus(404)
	})

	app.Post("/", testEmptyHandler)

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)

	g := app.Group("/with-next", func(c Ctx) error {
		return c.Status(404).Next()
	})

	g.Post("/", testEmptyHandler)

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/with-next", nil))
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
}

func TestAppServerErrorHandlerSmallReadBuffer(t *testing.T) {
	t.Parallel()
	expectedError := regexp.MustCompile(
		`error when reading request headers: small read buffer\. Increase ReadBufferSize\. Buffer size=4096, contents: "GET / HTTP/1.1\\r\\nHost: example\.com\\r\\nVery-Long-Header: -+`,
	)
	app := New()

	app.Get("/", func(_ Ctx) error {
		panic(errors.New("should never called"))
	})

	request := httptest.NewRequest(MethodGet, "/", nil)
	logHeaderSlice := make([]string, 5000)
	request.Header.Set("Very-Long-Header", strings.Join(logHeaderSlice, "-"))
	_, err := app.Test(request)
	if err == nil {
		t.Error("Expect an error at app.Test(request)")
	}

	require.Regexp(t, expectedError, err.Error())
}

func TestAppErrors(t *testing.T) {
	t.Parallel()
	app := New(Config{
		BodyLimit: 4,
	})

	app.Get("/", func(_ Ctx) error {
		return errors.New(literal_2753)
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 500, resp.StatusCode, literal_4126)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, literal_2753, string(body))

	_, err = app.Test(httptest.NewRequest(MethodGet, "/", strings.NewReader("big body")))
	if err != nil {
		require.Equal(t, "body size exceeds the given limit", err.Error(), literal_0521)
	}
}

type customConstraint struct{}

func (*customConstraint) Name() string {
	return "test"
}

func (*customConstraint) Execute(param string, args ...string) bool {
	if param == "test" && len(args) == 1 && args[0] == "test" {
		return true
	}

	if len(args) == 0 && param == "c" {
		return true
	}

	return false
}

func TestAppCustomConstraint(t *testing.T) {
	t.Parallel()
	app := New()
	app.RegisterCustomConstraint(&customConstraint{})

	app.Get("/test/:param<test(test)>", func(c Ctx) error {
		return c.SendString("test")
	})

	app.Get("/test2/:param<test>", func(c Ctx) error {
		return c.SendString("test")
	})

	app.Get("/test3/:param<test()>", func(c Ctx) error {
		return c.SendString("test")
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/test/test", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/test/test2", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 404, resp.StatusCode, literal_4126)

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/test2/c", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/test2/cc", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 404, resp.StatusCode, literal_4126)

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/test3/cc", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 404, resp.StatusCode, literal_4126)
}

func TestAppErrorHandlerCustom(t *testing.T) {
	t.Parallel()
	app := New(Config{
		ErrorHandler: func(c Ctx, _ error) error {
			return c.Status(200).SendString("hi, i'm an custom error")
		},
	})

	app.Get("/", func(_ Ctx) error {
		return errors.New(literal_2753)
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "hi, i'm an custom error", string(body))
}

func TestAppErrorHandlerHandlerStack(t *testing.T) {
	t.Parallel()
	app := New(Config{
		ErrorHandler: func(c Ctx, err error) error {
			require.Equal(t, literal_8913, err.Error())
			return DefaultErrorHandler(c, err)
		},
	})
	app.Use("/", func(c Ctx) error {
		err := c.Next() // call next USE
		require.Equal(t, "2: USE error", err.Error())
		return errors.New(literal_8913)
	}, func(c Ctx) error {
		err := c.Next() // call [0] GET
		require.Equal(t, literal_1437, err.Error())
		return errors.New("2: USE error")
	})
	app.Get("/", func(_ Ctx) error {
		return errors.New(literal_1437)
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 500, resp.StatusCode, literal_4126)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, literal_8913, string(body))
}

func TestAppErrorHandlerRouteStack(t *testing.T) {
	t.Parallel()
	app := New(Config{
		ErrorHandler: func(c Ctx, err error) error {
			require.Equal(t, literal_8913, err.Error())
			return DefaultErrorHandler(c, err)
		},
	})
	app.Use("/", func(c Ctx) error {
		err := c.Next()
		require.Equal(t, literal_1437, err.Error())
		return errors.New(literal_8913) // [2] call ErrorHandler
	})
	app.Get("/test", func(_ Ctx) error {
		return errors.New(literal_1437) // [1] return to USE
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/test", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 500, resp.StatusCode, literal_4126)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, literal_8913, string(body))
}

func TestAppserverErrorHandlerInternalError(t *testing.T) {
	t.Parallel()
	app := New()
	msg := "test err"
	c := app.AcquireCtx(&fasthttp.RequestCtx{}).(*DefaultCtx) //nolint:errcheck, forcetypeassert // not needed

	app.serverErrorHandler(c.fasthttp, errors.New(msg))
	require.Equal(t, string(c.fasthttp.Response.Body()), msg)
	require.Equal(t, StatusBadRequest, c.fasthttp.Response.StatusCode())
}

func TestAppserverErrorHandlerNetworkError(t *testing.T) {
	t.Parallel()
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{}).(*DefaultCtx) //nolint:errcheck, forcetypeassert // not needed

	app.serverErrorHandler(c.fasthttp, &net.DNSError{
		Err:       "test error",
		Name:      "test host",
		IsTimeout: false,
	})
	require.Equal(t, string(c.fasthttp.Response.Body()), utils.StatusMessage(StatusBadGateway))
	require.Equal(t, StatusBadGateway, c.fasthttp.Response.StatusCode())
}

func TestAppNestedParams(t *testing.T) {
	t.Parallel()
	app := New()

	app.Get("/test", func(c Ctx) error {
		return c.Status(400).Send([]byte(literal_4026))
	})
	app.Get("/test/:param", func(c Ctx) error {
		return c.Status(400).Send([]byte(literal_4026))
	})
	app.Get("/test/:param/test", func(c Ctx) error {
		return c.Status(400).Send([]byte(literal_4026))
	})
	app.Get("/test/:param/test/:param2", func(c Ctx) error {
		return c.Status(200).Send([]byte("Good job"))
	})

	req := httptest.NewRequest(MethodGet, "/test/john/test/doe", nil)
	resp, err := app.Test(req)

	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)
}

func TestAppUseParams(t *testing.T) {
	t.Parallel()
	app := New()

	app.Use("/prefix/:param", func(c Ctx) error {
		require.Equal(t, "john", c.Params("param"))
		return nil
	})

	app.Use("/foo/:bar?", func(c Ctx) error {
		require.Equal(t, "foobar", c.Params("bar", "foobar"))
		return nil
	})

	app.Use("/:param/*", func(c Ctx) error {
		require.Equal(t, "john", c.Params("param"))
		require.Equal(t, "doe", c.Params("*"))
		return nil
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/prefix/john", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)

	resp, err = app.Test(httptest.NewRequest(MethodGet, literal_5096, nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/foo", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)

	defer func() {
		if err := recover(); err != nil {
			require.Equal(t, "use: invalid handler func()\n", fmt.Sprintf("%v", err))
		}
	}()

	app.Use("/:param/*", func() {
		// this should panic
	})
}

func TestAppUseUnescapedPath(t *testing.T) {
	t.Parallel()
	app := New(Config{UnescapePath: true, CaseSensitive: true})

	app.Use("/cRéeR/:param", func(c Ctx) error {
		require.Equal(t, "/cRéeR/اختبار", c.Path())
		return c.SendString(c.Params("param"))
	})

	app.Use("/abc", func(c Ctx) error {
		require.Equal(t, "/AbC", c.Path())
		return nil
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/cR%C3%A9eR/%D8%A7%D8%AE%D8%AA%D8%A8%D8%A7%D8%B1", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_0521)
	// check the param result
	require.Equal(t, "اختبار", app.getString(body))

	// with lowercase letters
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/cr%C3%A9er/%D8%A7%D8%AE%D8%AA%D8%A8%D8%A7%D8%B1", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusNotFound, resp.StatusCode, literal_4126)
}

func TestAppUseCaseSensitive(t *testing.T) {
	t.Parallel()
	app := New(Config{CaseSensitive: true})

	app.Use("/abc", func(c Ctx) error {
		return c.SendString(c.Path())
	})

	// wrong letters in the requested route -> 404
	resp, err := app.Test(httptest.NewRequest(MethodGet, "/AbC", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusNotFound, resp.StatusCode, literal_4126)

	// right letters in the requrested route -> 200
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/abc", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)

	// check the detected path when the case insensitive recognition is activated
	app.config.CaseSensitive = false
	// check the case sensitive feature
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/AbC", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_0521)
	// check the detected path result
	require.Equal(t, "/AbC", app.getString(body))
}

func TestAppNotUseStrictRouting(t *testing.T) {
	t.Parallel()
	app := New()

	app.Use("/abc", func(c Ctx) error {
		return c.SendString(c.Path())
	})

	g := app.Group("/foo")
	g.Use("/", func(c Ctx) error {
		return c.SendString(c.Path())
	})

	// wrong path in the requested route -> 404
	resp, err := app.Test(httptest.NewRequest(MethodGet, "/abc/", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)

	// right path in the requrested route -> 200
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/abc", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)

	// wrong path with group in the requested route -> 404
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/foo", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)

	// right path with group in the requrested route -> 200
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/foo/", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)
}

func TestAppUseMultiplePrefix(t *testing.T) {
	t.Parallel()
	app := New()

	app.Use([]string{"/john", "/doe"}, func(c Ctx) error {
		return c.SendString(c.Path())
	})

	g := app.Group("/test")
	g.Use([]string{"/john", "/doe"}, func(c Ctx) error {
		return c.SendString(c.Path())
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/john", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "/john", string(body))

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/doe", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "/doe", string(body))

	resp, err = app.Test(httptest.NewRequest(MethodGet, literal_3546, nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, literal_3546, string(body))

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/test/doe", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "/test/doe", string(body))
}

func TestAppUseStrictRouting(t *testing.T) {
	t.Parallel()
	app := New(Config{StrictRouting: true})

	app.Get("/abc", func(c Ctx) error {
		return c.SendString(c.Path())
	})

	g := app.Group("/foo")
	g.Get("/", func(c Ctx) error {
		return c.SendString(c.Path())
	})

	// wrong path in the requested route -> 404
	resp, err := app.Test(httptest.NewRequest(MethodGet, "/abc/", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusNotFound, resp.StatusCode, literal_4126)

	// right path in the requrested route -> 200
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/abc", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)

	// wrong path with group in the requested route -> 404
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/foo", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusNotFound, resp.StatusCode, literal_4126)

	// right path with group in the requrested route -> 200
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/foo/", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)
}

func TestAppAddMethodTest(t *testing.T) {
	t.Parallel()
	defer func() {
		if err := recover(); err != nil {
			require.Equal(t, "add: invalid http method JANE\n", fmt.Sprintf("%v", err))
		}
	}()

	methods := append(DefaultMethods, "JOHN") //nolint:gocritic // We want a new slice here
	app := New(Config{
		RequestMethods: methods,
	})

	app.Add([]string{"JOHN"}, "/doe", testEmptyHandler)

	resp, err := app.Test(httptest.NewRequest("JOHN", "/doe", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusOK, resp.StatusCode, literal_4126)

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/doe", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusMethodNotAllowed, resp.StatusCode, literal_4126)

	resp, err = app.Test(httptest.NewRequest("UNKNOWN", "/doe", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusNotImplemented, resp.StatusCode, literal_4126)

	app.Add([]string{"JANE"}, "/doe", testEmptyHandler)
}

// go test -run Test_App_GETOnly
func TestAppGETOnly(t *testing.T) {
	t.Parallel()
	app := New(Config{
		GETOnly: true,
	})

	app.Post("/", func(c Ctx) error {
		return c.SendString("Hello 👋!")
	})

	req := httptest.NewRequest(MethodPost, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, literal_0521)
	require.Equal(t, StatusMethodNotAllowed, resp.StatusCode, literal_4126)
}

func TestAppUseParamsGroup(t *testing.T) {
	t.Parallel()
	app := New()

	group := app.Group("/prefix/:param/*")
	group.Use("/", func(c Ctx) error {
		return c.Next()
	})
	group.Get("/test", func(c Ctx) error {
		require.Equal(t, "john", c.Params("param"))
		require.Equal(t, "doe", c.Params("*"))
		return nil
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/prefix/john/doe/test", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)
}

func TestAppChaining(t *testing.T) {
	t.Parallel()
	n := func(c Ctx) error {
		return c.Next()
	}
	app := New()
	app.Use("/john", n, n, n, n, func(c Ctx) error {
		return c.SendStatus(202)
	})
	// check handler count for registered HEAD route
	require.Len(t, app.stack[app.methodInt(MethodHead)][0].Handlers, 5, literal_0521)

	req := httptest.NewRequest(MethodPost, "/john", nil)

	resp, err := app.Test(req)
	require.NoError(t, err, literal_0521)
	require.Equal(t, 202, resp.StatusCode, literal_4126)

	app.Get("/test", n, n, n, n, func(c Ctx) error {
		return c.SendStatus(203)
	})

	req = httptest.NewRequest(MethodGet, "/test", nil)

	resp, err = app.Test(req)
	require.NoError(t, err, literal_0521)
	require.Equal(t, 203, resp.StatusCode, literal_4126)
}

func TestAppOrder(t *testing.T) {
	t.Parallel()
	app := New()

	app.Get("/test", func(c Ctx) error {
		_, err := c.Write([]byte("1"))
		require.NoError(t, err)

		return c.Next()
	})

	app.All("/test", func(c Ctx) error {
		_, err := c.Write([]byte("2"))
		require.NoError(t, err)

		return c.Next()
	})

	app.Use(func(c Ctx) error {
		_, err := c.Write([]byte("3"))
		require.NoError(t, err)

		return nil
	})

	req := httptest.NewRequest(MethodGet, "/test", nil)

	resp, err := app.Test(req)
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "123", string(body))
}

func TestAppMethods(t *testing.T) {
	t.Parallel()
	dummyHandler := testEmptyHandler

	app := New()

	app.Connect(literal_3091, dummyHandler)
	testStatus200(t, app, literal_5096, "CONNECT")

	app.Put(literal_3091, dummyHandler)
	testStatus200(t, app, literal_5096, MethodPut)

	app.Post(literal_3091, dummyHandler)
	testStatus200(t, app, literal_5096, MethodPost)

	app.Delete(literal_3091, dummyHandler)
	testStatus200(t, app, literal_5096, MethodDelete)

	app.Head(literal_3091, dummyHandler)
	testStatus200(t, app, literal_5096, MethodHead)

	app.Patch(literal_3091, dummyHandler)
	testStatus200(t, app, literal_5096, MethodPatch)

	app.Options(literal_3091, dummyHandler)
	testStatus200(t, app, literal_5096, MethodOptions)

	app.Trace(literal_3091, dummyHandler)
	testStatus200(t, app, literal_5096, MethodTrace)

	app.Get(literal_3091, dummyHandler)
	testStatus200(t, app, literal_5096, MethodGet)

	app.All(literal_3091, dummyHandler)
	testStatus200(t, app, literal_5096, MethodPost)

	app.Use(literal_3091, dummyHandler)
	testStatus200(t, app, literal_5096, MethodGet)
}

func TestAppRouteNaming(t *testing.T) {
	t.Parallel()
	app := New()
	handler := func(c Ctx) error {
		return c.SendStatus(StatusOK)
	}
	app.Get("/john", handler).Name("john")
	app.Delete("/doe", handler)
	app.Name("doe")

	jane := app.Group("/jane").Name("jane.")
	group := app.Group("/group")
	subGroup := jane.Group("/sub-group").Name("sub.")

	jane.Get("/test", handler).Name("test")
	jane.Trace("/trace", handler).Name("trace")

	group.Get("/test", handler).Name("test")

	app.Post("/post", handler).Name("post")

	subGroup.Get("/done", handler).Name("done")

	require.Equal(t, "post", app.GetRoute("post").Name)
	require.Equal(t, "john", app.GetRoute("john").Name)
	require.Equal(t, "jane.test", app.GetRoute("jane.test").Name)
	require.Equal(t, "jane.trace", app.GetRoute("jane.trace").Name)
	require.Equal(t, "jane.sub.done", app.GetRoute("jane.sub.done").Name)
	require.Equal(t, "test", app.GetRoute("test").Name)
}

func TestAppNew(t *testing.T) {
	t.Parallel()
	app := New()
	app.Get("/", testEmptyHandler)

	appConfig := New(Config{
		Immutable: true,
	})
	appConfig.Get("/", testEmptyHandler)
}

func TestAppConfig(t *testing.T) {
	t.Parallel()
	app := New(Config{
		StrictRouting: true,
	})
	require.True(t, app.Config().StrictRouting)
}

func TestAppShutdown(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		app := New()
		require.NoError(t, app.Shutdown())
	})

	t.Run("no server", func(t *testing.T) {
		t.Parallel()
		app := &App{}
		if err := app.Shutdown(); err != nil {
			require.ErrorContains(t, err, "shutdown: server is not running")
		}
	})
}

func TestAppShutdownWithTimeout(t *testing.T) {
	t.Parallel()
	app := New()
	app.Get("/", func(c Ctx) error {
		time.Sleep(5 * time.Second)
		return c.SendString("body")
	})

	ln := fasthttputil.NewInmemoryListener()
	go func() {
		err := app.Listener(ln)
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)
	go func() {
		conn, err := ln.Dial()
		assert.NoError(t, err)

		_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: google.com\r\n\r\n"))
		assert.NoError(t, err)
	}()
	time.Sleep(1 * time.Second)

	shutdownErr := make(chan error)
	go func() {
		shutdownErr <- app.ShutdownWithTimeout(1 * time.Second)
	}()

	timer := time.NewTimer(time.Second * 5)
	select {
	case <-timer.C:
		t.Fatal("idle connections not closed on shutdown")
	case err := <-shutdownErr:
		if err == nil || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("unexpected err %v. Expecting %v", err, context.DeadlineExceeded)
		}
	}
}

func TestAppShutdownWithContext(t *testing.T) {
	t.Parallel()

	app := New()
	app.Get("/", func(ctx Ctx) error {
		time.Sleep(5 * time.Second)
		return ctx.SendString("body")
	})

	ln := fasthttputil.NewInmemoryListener()

	go func() {
		err := app.Listener(ln)
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	go func() {
		conn, err := ln.Dial()
		assert.NoError(t, err)

		_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: google.com\r\n\r\n"))
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	shutdownErr := make(chan error)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		shutdownErr <- app.ShutdownWithContext(ctx)
	}()

	select {
	case <-time.After(5 * time.Second):
		t.Fatal("idle connections not closed on shutdown")
	case err := <-shutdownErr:
		if err == nil || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("unexpected err %v. Expecting %v", err, context.DeadlineExceeded)
		}
	}
}

// go test -run Test_App_Mixed_Routes_WithSameLen
func TestAppMixedRoutesWithSameLen(t *testing.T) {
	t.Parallel()
	app := New()

	// middleware
	app.Use(func(c Ctx) error {
		c.Set("TestHeader", "TestValue")
		return c.Next()
	})
	// routes with the same length
	app.Get("/tesbar", func(c Ctx) error {
		c.Type("html")
		return c.Send([]byte("TEST_BAR"))
	})
	app.Get("/foobar", func(c Ctx) error {
		c.Type("html")
		return c.Send([]byte("FOO_BAR"))
	})

	// match get route
	req := httptest.NewRequest(MethodGet, "/foobar", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)
	require.NotEmpty(t, resp.Header.Get(HeaderContentLength))
	require.Equal(t, "TestValue", resp.Header.Get("TestHeader"))
	require.Equal(t, "text/html", resp.Header.Get(HeaderContentType))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "FOO_BAR", string(body))

	// match static route
	req = httptest.NewRequest(MethodGet, "/tesbar", nil)
	resp, err = app.Test(req)
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)
	require.NotEmpty(t, resp.Header.Get(HeaderContentLength))
	require.Equal(t, "TestValue", resp.Header.Get("TestHeader"))
	require.Equal(t, "text/html", resp.Header.Get(HeaderContentType))

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "TEST_BAR")
}

func TestAppGroupInvalid(t *testing.T) {
	t.Parallel()
	defer func() {
		if err := recover(); err != nil {
			require.Equal(t, "use: invalid handler int\n", fmt.Sprintf("%v", err))
		}
	}()
	New().Group("/").Use(1)
}

func TestAppGroup(t *testing.T) {
	t.Parallel()
	dummyHandler := testEmptyHandler

	app := New()

	grp := app.Group("/test")
	grp.Get("/", dummyHandler)
	testStatus200(t, app, "/test", MethodGet)

	grp.Get("/:demo?", dummyHandler)
	testStatus200(t, app, literal_3546, MethodGet)

	grp.Connect("/CONNECT", dummyHandler)
	testStatus200(t, app, "/test/CONNECT", MethodConnect)

	grp.Put("/PUT", dummyHandler)
	testStatus200(t, app, "/test/PUT", MethodPut)

	grp.Post("/POST", dummyHandler)
	testStatus200(t, app, "/test/POST", MethodPost)

	grp.Delete("/DELETE", dummyHandler)
	testStatus200(t, app, "/test/DELETE", MethodDelete)

	grp.Head("/HEAD", dummyHandler)
	testStatus200(t, app, "/test/HEAD", MethodHead)

	grp.Patch("/PATCH", dummyHandler)
	testStatus200(t, app, "/test/PATCH", MethodPatch)

	grp.Options("/OPTIONS", dummyHandler)
	testStatus200(t, app, "/test/OPTIONS", MethodOptions)

	grp.Trace("/TRACE", dummyHandler)
	testStatus200(t, app, "/test/TRACE", MethodTrace)

	grp.All("/ALL", dummyHandler)
	testStatus200(t, app, "/test/ALL", MethodPost)

	grp.Use(dummyHandler)
	testStatus200(t, app, "/test/oke", MethodGet)

	grp.Use("/USE", dummyHandler)
	testStatus200(t, app, "/test/USE/oke", MethodGet)

	api := grp.Group("/v1")
	api.Post("/", dummyHandler)

	resp, err := app.Test(httptest.NewRequest(MethodPost, "/test/v1/", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)

	api.Get(literal_8759, dummyHandler)
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/test/v1/UsErS", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)
}

func TestAppRoute(t *testing.T) {
	t.Parallel()
	dummyHandler := testEmptyHandler

	app := New()

	register := app.Route("/test").
		Get(dummyHandler).
		Head(dummyHandler).
		Post(dummyHandler).
		Put(dummyHandler).
		Delete(dummyHandler).
		Connect(dummyHandler).
		Options(dummyHandler).
		Trace(dummyHandler).
		Patch(dummyHandler)

	testStatus200(t, app, "/test", MethodGet)
	testStatus200(t, app, "/test", MethodHead)
	testStatus200(t, app, "/test", MethodPost)
	testStatus200(t, app, "/test", MethodPut)
	testStatus200(t, app, "/test", MethodDelete)
	testStatus200(t, app, "/test", MethodConnect)
	testStatus200(t, app, "/test", MethodOptions)
	testStatus200(t, app, "/test", MethodTrace)
	testStatus200(t, app, "/test", MethodPatch)

	register.Route("/v1").Get(dummyHandler).Post(dummyHandler)

	resp, err := app.Test(httptest.NewRequest(MethodPost, "/test/v1", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/test/v1", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)

	register.Route("/v1").Route("/v2").Route("/v3").Get(dummyHandler).Trace(dummyHandler)

	resp, err = app.Test(httptest.NewRequest(MethodTrace, "/test/v1/v2/v3", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/test/v1/v2/v3", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)
}

func TestAppDeepGroup(t *testing.T) {
	t.Parallel()
	runThroughCount := 0
	dummyHandler := func(c Ctx) error {
		runThroughCount++
		return c.Next()
	}

	app := New()
	gAPI := app.Group("/api", dummyHandler)
	gV1 := gAPI.Group("/v1", dummyHandler)
	gUser := gV1.Group("/user", dummyHandler)
	gUser.Get("/authenticate", func(c Ctx) error {
		runThroughCount++
		return c.SendStatus(200)
	})
	testStatus200(t, app, "/api/v1/user/authenticate", MethodGet)
	require.Equal(t, 4, runThroughCount, "Loop count")
}

// go test -run Test_App_Next_Method
func TestAppNextMethod(t *testing.T) {
	t.Parallel()
	app := New()

	app.Use(func(c Ctx) error {
		require.Equal(t, MethodGet, c.Method())
		err := c.Next()
		require.Equal(t, MethodGet, c.Method())
		return err
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/", nil))
	require.NoError(t, err, literal_0521)
	require.Equal(t, 404, resp.StatusCode, literal_4126)
}

// go test -v -run=^$ -bench=Benchmark_NewError -benchmem -count=4
func BenchmarkNewError(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewError(200, "test") //nolint:errcheck // not needed
	}
}

// go test -run Test_NewError
func TestNewError(t *testing.T) {
	t.Parallel()
	e := NewError(StatusForbidden, "permission denied")
	require.Equal(t, StatusForbidden, e.Code)
	require.Equal(t, "permission denied", e.Message)
}

// go test -run Test_Test_Timeout
func TestTestTimeout(t *testing.T) {
	t.Parallel()
	app := New()

	app.Get("/", testEmptyHandler)

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/", nil), -1)
	require.NoError(t, err, literal_0521)
	require.Equal(t, 200, resp.StatusCode, literal_4126)

	app.Get("timeout", func(_ Ctx) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	_, err = app.Test(httptest.NewRequest(MethodGet, "/timeout", nil), 20*time.Millisecond)
	require.Error(t, err, literal_0521)
}

type errorReader int

var errErrorReader = errors.New("errorReader")

func (errorReader) Read([]byte) (int, error) {
	return 0, errErrorReader
}

// go test -run Test_Test_DumpError
func TestTestDumpError(t *testing.T) {
	t.Parallel()
	app := New()

	app.Get("/", testEmptyHandler)

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/", errorReader(0)))
	require.Nil(t, resp)
	require.ErrorIs(t, err, errErrorReader)
}

// go test -run Test_App_Handler
func TestAppHandler(t *testing.T) {
	t.Parallel()
	h := New().Handler()
	require.Equal(t, "fasthttp.RequestHandler", reflect.TypeOf(h).String())
}

type invalidView struct{}

func (invalidView) Load() error { return errors.New("invalid view") }

func (invalidView) Render(io.Writer, string, any, ...string) error { panic("implement me") }

// go test -run Test_App_Init_Error_View
func TestAppInitErrorView(t *testing.T) {
	t.Parallel()
	app := New(Config{Views: invalidView{}})

	defer func() {
		if err := recover(); err != nil {
			require.Equal(t, "implement me", fmt.Sprintf("%v", err))
		}
	}()

	err := app.config.Views.Render(nil, "", nil)
	require.NoError(t, err)
}

// go test -run Test_App_Stack
func TestAppStack(t *testing.T) {
	t.Parallel()
	app := New()

	app.Use("/path0", testEmptyHandler)
	app.Get("/path1", testEmptyHandler)
	app.Get("/path2", testEmptyHandler)
	app.Post("/path3", testEmptyHandler)

	stack := app.Stack()
	methodList := app.config.RequestMethods
	require.Equal(t, len(methodList), len(stack))
	require.Len(t, stack[app.methodInt(MethodGet)], 3)
	require.Len(t, stack[app.methodInt(MethodHead)], 1)
	require.Len(t, stack[app.methodInt(MethodPost)], 2)
	require.Len(t, stack[app.methodInt(MethodPut)], 1)
	require.Len(t, stack[app.methodInt(MethodPatch)], 1)
	require.Len(t, stack[app.methodInt(MethodDelete)], 1)
	require.Len(t, stack[app.methodInt(MethodConnect)], 1)
	require.Len(t, stack[app.methodInt(MethodOptions)], 1)
	require.Len(t, stack[app.methodInt(MethodTrace)], 1)
}

// go test -run Test_App_HandlersCount
func TestAppHandlersCount(t *testing.T) {
	t.Parallel()
	app := New()

	app.Use("/path0", testEmptyHandler)
	app.Get("/path2", testEmptyHandler)
	app.Post("/path3", testEmptyHandler)

	count := app.HandlersCount()
	require.Equal(t, uint32(3), count)
}

// go test -run Test_App_ReadTimeout
func TestAppReadTimeout(t *testing.T) {
	t.Parallel()
	app := New(Config{
		ReadTimeout:      time.Nanosecond,
		IdleTimeout:      time.Minute,
		DisableKeepalive: true,
	})

	app.Get("/read-timeout", func(c Ctx) error {
		return c.SendString(literal_4193)
	})

	go func() {
		time.Sleep(500 * time.Millisecond)

		conn, err := net.Dial(NetworkTCP4, "127.0.0.1:4004")
		assert.NoError(t, err)
		defer func(conn net.Conn) {
			err := conn.Close()
			assert.NoError(t, err)
		}(conn)

		_, err = conn.Write([]byte("HEAD /read-timeout HTTP/1.1\r\n"))
		assert.NoError(t, err)

		buf := make([]byte, 1024)
		var n int
		n, err = conn.Read(buf)

		assert.NoError(t, err)
		assert.True(t, bytes.Contains(buf[:n], []byte("408 Request Timeout")))

		assert.NoError(t, app.Shutdown())
	}()

	require.NoError(t, app.Listen(":4004", ListenConfig{DisableStartupMessage: true}))
}

// go test -run Test_App_BadRequest
func TestAppBadRequest(t *testing.T) {
	t.Parallel()
	app := New()

	app.Get("/bad-request", func(c Ctx) error {
		return c.SendString(literal_4193)
	})

	go func() {
		time.Sleep(500 * time.Millisecond)
		conn, err := net.Dial(NetworkTCP4, "127.0.0.1:4005")
		assert.NoError(t, err)
		defer func(conn net.Conn) {
			err := conn.Close()
			assert.NoError(t, err)
		}(conn)

		_, err = conn.Write([]byte("BadRequest\r\n"))
		assert.NoError(t, err)

		buf := make([]byte, 1024)
		var n int
		n, err = conn.Read(buf)

		assert.NoError(t, err)
		assert.True(t, bytes.Contains(buf[:n], []byte("400 Bad Request")))
		assert.NoError(t, app.Shutdown())
	}()

	require.NoError(t, app.Listen(":4005", ListenConfig{DisableStartupMessage: true}))
}

// go test -run Test_App_SmallReadBuffer
func TestAppSmallReadBuffer(t *testing.T) {
	t.Parallel()
	app := New(Config{
		ReadBufferSize: 1,
	})

	app.Get("/small-read-buffer", func(c Ctx) error {
		return c.SendString(literal_4193)
	})

	go func() {
		time.Sleep(500 * time.Millisecond)
		req, err := http.NewRequestWithContext(context.Background(), MethodGet, "http://127.0.0.1:4006/small-read-buffer", nil)
		assert.NoError(t, err)
		var client http.Client
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, 431, resp.StatusCode)
		assert.NoError(t, app.Shutdown())
	}()

	require.NoError(t, app.Listen(":4006", ListenConfig{DisableStartupMessage: true}))
}

func TestAppServer(t *testing.T) {
	t.Parallel()
	app := New()

	require.NotNil(t, app.Server())
}

func TestAppErrorInFasthttpServer(t *testing.T) {
	app := New()
	app.config.ErrorHandler = func(_ Ctx, _ error) error {
		return errors.New("fake error")
	}
	app.server.GetOnly = true

	resp, err := app.Test(httptest.NewRequest(MethodPost, "/", nil))
	require.NoError(t, err)
	require.Equal(t, 500, resp.StatusCode)
}

// go test -race -run Test_App_New_Test_Parallel
func TestAppNewTestParallel(t *testing.T) {
	t.Parallel()
	t.Run("Test_App_New_Test_Parallel_1", func(t *testing.T) {
		t.Parallel()
		app := New(Config{Immutable: true})
		_, err := app.Test(httptest.NewRequest(MethodGet, "/", nil))
		require.NoError(t, err)
	})
	t.Run("Test_App_New_Test_Parallel_2", func(t *testing.T) {
		t.Parallel()
		app := New(Config{Immutable: true})
		_, err := app.Test(httptest.NewRequest(MethodGet, "/", nil))
		require.NoError(t, err)
	})
}

func TestAppReadBodyStream(t *testing.T) {
	t.Parallel()
	app := New(Config{StreamRequestBody: true})
	app.Post("/", func(c Ctx) error {
		// Calling c.Body() automatically reads the entire stream.
		return c.SendString(fmt.Sprintf("%v %s", c.Request().IsBodyStream(), c.Body()))
	})
	testString := "this is a test"
	resp, err := app.Test(httptest.NewRequest(MethodPost, "/", bytes.NewBufferString(testString)))
	require.NoError(t, err, literal_0521)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "io.ReadAll(resp.Body)")
	require.Equal(t, "true "+testString, string(body))
}

func TestAppDisablePreParseMultipartForm(t *testing.T) {
	t.Parallel()
	// Must be used with both otherwise there is no point.
	testString := "this is a test"

	app := New(Config{DisablePreParseMultipartForm: true, StreamRequestBody: true})
	app.Post("/", func(c Ctx) error {
		req := c.Request()
		mpf, err := req.MultipartForm()
		if err != nil {
			return err
		}
		if !req.IsBodyStream() {
			return errors.New("not a body stream")
		}
		file, err := mpf.File["test"][0].Open()
		if err != nil {
			return fmt.Errorf("failed to open: %w", err)
		}
		buffer := make([]byte, len(testString))
		n, err := file.Read(buffer)
		if err != nil {
			return fmt.Errorf("failed to read: %w", err)
		}
		if n != len(testString) {
			return errors.New("bad read length")
		}
		return c.Send(buffer)
	})
	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)
	writer, err := w.CreateFormFile("test", "test")
	require.NoError(t, err, "w.CreateFormFile")
	n, err := writer.Write([]byte(testString))
	require.NoError(t, err, "writer.Write")
	require.Len(t, testString, n, "writer n")
	require.NoError(t, w.Close(), "w.Close()")

	req := httptest.NewRequest(MethodPost, "/", b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := app.Test(req)
	require.NoError(t, err, literal_0521)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "io.ReadAll(resp.Body)")

	require.Equal(t, testString, string(body))
}

func TestAppTestnotimeoutinfinitely(t *testing.T) {
	t.Parallel()
	var err error
	c := make(chan int)

	go func() {
		defer func() { c <- 0 }()
		app := New()
		app.Get("/", func(_ Ctx) error {
			runtime.Goexit()
			return nil
		})

		req := httptest.NewRequest(MethodGet, "/", nil)
		_, err = app.Test(req, -1)
	}()

	tk := time.NewTimer(5 * time.Second)
	defer tk.Stop()

	select {
	case <-tk.C:
		t.Error("hanging test")
		t.FailNow()
	case <-c:
	}

	if err == nil {
		t.Error("unexpected success request")
		t.FailNow()
	}
}

func TestAppTesttimeout(t *testing.T) {
	t.Parallel()

	app := New()
	app.Get("/", func(_ Ctx) error {
		time.Sleep(1 * time.Second)
		return nil
	})

	_, err := app.Test(httptest.NewRequest(MethodGet, "/", nil), 100*time.Millisecond)
	require.Equal(t, errors.New("test: timeout error after 100ms"), err)
}

func TestAppSetTLSHandler(t *testing.T) {
	t.Parallel()
	tlsHandler := &TLSHandler{clientHelloInfo: &tls.ClientHelloInfo{
		ServerName: "example.golang",
	}}

	app := New()
	app.SetTLSHandler(tlsHandler)

	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	require.Equal(t, "example.golang", c.ClientHelloInfo().ServerName)
}

func TestAppAddCustomRequestMethod(t *testing.T) {
	t.Parallel()
	methods := append(DefaultMethods, "TEST") //nolint:gocritic // We want a new slice here
	app := New(Config{
		RequestMethods: methods,
	})
	appMethods := app.config.RequestMethods

	// method name is always uppercase - https://datatracker.ietf.org/doc/html/rfc7231#section-4.1
	require.Equal(t, len(app.stack), len(appMethods))
	require.Equal(t, len(app.stack), len(appMethods))
	require.Equal(t, "TEST", appMethods[len(appMethods)-1])
}

func TestAppGetRoutes(t *testing.T) {
	t.Parallel()
	app := New()
	app.Use(func(c Ctx) error {
		return c.Next()
	})
	handler := func(c Ctx) error {
		return c.SendStatus(StatusOK)
	}
	app.Delete("/delete", handler).Name("delete")
	app.Post("/post", handler).Name("post")
	routes := app.GetRoutes(false)
	require.Len(t, routes, 2+len(app.config.RequestMethods))
	methodMap := map[string]string{"/delete": "delete", "/post": "post"}
	for _, route := range routes {
		name, ok := methodMap[route.Path]
		if ok {
			require.Equal(t, name, route.Name)
		}
	}

	routes = app.GetRoutes(true)
	require.Len(t, routes, 2)
	for _, route := range routes {
		name, ok := methodMap[route.Path]
		require.True(t, ok)
		require.Equal(t, name, route.Name)
	}
}

func TestMiddlewareRouteNamingWithUse(t *testing.T) {
	t.Parallel()
	named := "named"
	app := New()

	app.Get(literal_0891, func(c Ctx) error {
		return c.Next()
	})

	app.Post("/named", func(c Ctx) error {
		return c.Next()
	}).Name(named)

	app.Use(func(c Ctx) error {
		return c.Next()
	}) // no name - logging MW

	app.Use(func(c Ctx) error {
		return c.Next()
	}).Name("corsMW")

	app.Use(func(c Ctx) error {
		return c.Next()
	}).Name("compressMW")

	app.Use(func(c Ctx) error {
		return c.Next()
	}) // no name - cache MW

	grp := app.Group("/pages").Name("pages.")
	grp.Use(func(c Ctx) error {
		return c.Next()
	}).Name("csrfMW")

	grp.Get("/home", func(c Ctx) error {
		return c.Next()
	}).Name("home")

	grp.Get(literal_0891, func(c Ctx) error {
		return c.Next()
	})

	for _, route := range app.GetRoutes() {
		switch route.Path {
		case "/":
			require.Equal(t, "compressMW", route.Name)
		case literal_0891:
			require.Equal(t, "", route.Name)
		case "named":
			require.Equal(t, named, route.Name)
		case "/pages":
			require.Equal(t, "pages.csrfMW", route.Name)
		case "/pages/home":
			require.Equal(t, "pages.home", route.Name)
		case "/pages/unnamed":
			require.Equal(t, "", route.Name)
		}
	}
}

func TestRouteNamingIssue26712685(t *testing.T) {
	t.Parallel()
	app := New()

	app.Get("/", emptyHandler).Name("index")
	require.Equal(t, "/", app.GetRoute("index").Path)

	app.Get("/a/:a_id", emptyHandler).Name("a")
	require.Equal(t, "/a/:a_id", app.GetRoute("a").Path)

	app.Post("/b/:bId", emptyHandler).Name("b")
	require.Equal(t, "/b/:bId", app.GetRoute("b").Path)

	c := app.Group("/c")
	c.Get("", emptyHandler).Name("c.get")
	require.Equal(t, "/c", app.GetRoute("c.get").Path)

	c.Post("", emptyHandler).Name("c.post")
	require.Equal(t, "/c", app.GetRoute("c.post").Path)

	c.Get("/d", emptyHandler).Name("c.get.d")
	require.Equal(t, "/c/d", app.GetRoute("c.get.d").Path)

	d := app.Group(literal_1580)
	d.Get("", emptyHandler).Name("d.get")
	require.Equal(t, literal_1580, app.GetRoute("d.get").Path)

	d.Post("", emptyHandler).Name("d.post")
	require.Equal(t, literal_1580, app.GetRoute("d.post").Path)

	e := app.Group(literal_8490)
	e.Get("", emptyHandler).Name("e.get")
	require.Equal(t, literal_8490, app.GetRoute("e.get").Path)

	e.Post("", emptyHandler).Name("e.post")
	require.Equal(t, literal_8490, app.GetRoute("e.post").Path)

	e.Get("f", emptyHandler).Name("e.get.f")
	require.Equal(t, "/e/:eId/f", app.GetRoute("e.get.f").Path)

	postGroup := app.Group(literal_5726)
	postGroup.Get("", emptyHandler).Name("post.get")
	require.Equal(t, literal_5726, app.GetRoute("post.get").Path)

	postGroup.Post("", emptyHandler).Name("post.update")
	require.Equal(t, literal_5726, app.GetRoute("post.update").Path)

	// Add testcase for routes use the same PATH on different methods
	app.Get(literal_8759, emptyHandler).Name("get-users")
	app.Post(literal_8759, emptyHandler).Name("add-user")
	getUsers := app.GetRoute("get-users")
	require.Equal(t, literal_8759, getUsers.Path)

	addUser := app.GetRoute("add-user")
	require.Equal(t, literal_8759, addUser.Path)

	// Add testcase for routes use the same PATH on different methods (for groups)
	newGrp := app.Group("/name-test")
	newGrp.Get(literal_8759, emptyHandler).Name("grp-get-users")
	newGrp.Post(literal_8759, emptyHandler).Name("grp-add-user")
	getUsers = app.GetRoute("grp-get-users")
	require.Equal(t, "/name-test/users", getUsers.Path)

	addUser = app.GetRoute("grp-add-user")
	require.Equal(t, "/name-test/users", addUser.Path)

	// Add testcase for HEAD route naming
	app.Get(literal_0824, emptyHandler).Name("simple-route")
	app.Head(literal_0824, emptyHandler).Name("simple-route2")

	sRoute := app.GetRoute("simple-route")
	require.Equal(t, literal_0824, sRoute.Path)

	sRoute2 := app.GetRoute("simple-route2")
	require.Equal(t, literal_0824, sRoute2.Path)
}

// go test -v -run=^$ -bench=Benchmark_Communication_Flow -benchmem -count=4
func BenchmarkCommunicationFlow(b *testing.B) {
	app := New()

	app.Get("/", func(c Ctx) error {
		return c.SendString("Hello, World!")
	})

	h := app.Handler()

	fctx := &fasthttp.RequestCtx{}
	fctx.Request.Header.SetMethod(MethodGet)
	fctx.Request.SetRequestURI("/")

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		h(fctx)
	}

	require.Equal(b, 200, fctx.Response.Header.StatusCode())
	require.Equal(b, "Hello, World!", string(fctx.Response.Body()))
}

// go test -v -run=^$ -bench=Benchmark_Ctx_AcquireReleaseFlow -benchmem -count=4
func BenchmarkCtxAcquireReleaseFlow(b *testing.B) {
	app := New()

	fctx := &fasthttp.RequestCtx{}

	b.Run("withoutRequestCtx", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for n := 0; n < b.N; n++ {
			c, _ := app.AcquireCtx(fctx).(*DefaultCtx) //nolint:errcheck // not needed
			app.ReleaseCtx(c)
		}
	})

	b.Run("withRequestCtx", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for n := 0; n < b.N; n++ {
			c, _ := app.AcquireCtx(&fasthttp.RequestCtx{}).(*DefaultCtx) //nolint:errcheck // not needed
			app.ReleaseCtx(c)
		}
	})
}

const literal_0521 = "app.Test(req)"

const literal_4126 = "Status code"

const literal_6154 = "POST, OPTIONS"

const literal_2753 = "hi, i'm an error"

const literal_8913 = "1: USE error"

const literal_1437 = "0: GET error"

const literal_4026 = "Should move on"

const literal_5096 = "/john/doe"

const literal_3546 = "/test/john"

const literal_3091 = "/:john?/:doe?"

const literal_8759 = "/users"

const literal_4193 = "I should not be sent"

const literal_0891 = "/unnamed"

const literal_1580 = "/d/:d_id"

const literal_8490 = "/e/:eId"

const literal_5726 = "/post/:postId"

const literal_0824 = "/simple-route"
