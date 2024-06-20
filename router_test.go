// ⚡️ Fiber is an Express inspired web framework written in Go with ☕️
// 📃 Github Repository: https://github.com/gofiber/fiber
// 📌 API Documentation: https://docs.gofiber.io

package fiber

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/utils/v2"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

var routesFixture routeJSON

func init() {
	dat, err := os.ReadFile("./.github/testdata/testRoutes.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(dat, &routesFixture); err != nil {
		panic(err)
	}
}

func TestRouteMatchSameLength(t *testing.T) {
	t.Parallel()

	app := New()

	app.Get("/:param", func(c Ctx) error {
		return c.SendString(c.Params("param"))
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/:param", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, 200, resp.StatusCode, literal_6917)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, ":param", app.getString(body))

	// with param
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/test", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, 200, resp.StatusCode, literal_6917)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "test", app.getString(body))
}

func TestRouteMatchStar(t *testing.T) {
	t.Parallel()

	app := New()

	app.Get("/*", func(c Ctx) error {
		return c.SendString(c.Params("*"))
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/*", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, 200, resp.StatusCode, literal_6917)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "*", app.getString(body))

	// with param
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/test", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, 200, resp.StatusCode, literal_6917)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "test", app.getString(body))

	// without parameter
	route := Route{
		star:        true,
		path:        "/*",
		routeParser: routeParser{},
	}
	params := [maxParams]string{}
	match := route.match("", "", &params)
	require.True(t, match)
	require.Equal(t, [maxParams]string{}, params)

	// with parameter
	match = route.match("/favicon.ico", "/favicon.ico", &params)
	require.True(t, match)
	require.Equal(t, [maxParams]string{"favicon.ico"}, params)

	// without parameter again
	match = route.match("", "", &params)
	require.True(t, match)
	require.Equal(t, [maxParams]string{}, params)
}

func TestRouteMatchRoot(t *testing.T) {
	t.Parallel()

	app := New()

	app.Get("/", func(c Ctx) error {
		return c.SendString("root")
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, 200, resp.StatusCode, literal_6917)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "root", app.getString(body))
}

func TestRouteMatchParser(t *testing.T) {
	t.Parallel()

	app := New()

	app.Get("/foo/:ParamName", func(c Ctx) error {
		return c.SendString(c.Params("ParamName"))
	})
	app.Get("/Foobar/*", func(c Ctx) error {
		return c.SendString(c.Params("*"))
	})
	resp, err := app.Test(httptest.NewRequest(MethodGet, "/foo/bar", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, 200, resp.StatusCode, literal_6917)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "bar", app.getString(body))

	// with star
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/Foobar/test", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, 200, resp.StatusCode, literal_6917)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "test", app.getString(body))
}

func TestRouteMatchMiddleware(t *testing.T) {
	t.Parallel()

	app := New()

	app.Use("/foo/*", func(c Ctx) error {
		return c.SendString(c.Params("*"))
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/foo/*", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, 200, resp.StatusCode, literal_6917)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "*", app.getString(body))

	// with param
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/foo/bar/fasel", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, 200, resp.StatusCode, literal_6917)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "bar/fasel", app.getString(body))
}

func TestRouteMatchUnescapedPath(t *testing.T) {
	t.Parallel()

	app := New(Config{UnescapePath: true})

	app.Use(literal_0493, func(c Ctx) error {
		return c.SendString("test")
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, literal_2379, nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, StatusOK, resp.StatusCode, literal_6917)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "test", app.getString(body))
	// without special chars
	resp, err = app.Test(httptest.NewRequest(MethodGet, literal_0493, nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, StatusOK, resp.StatusCode, literal_6917)

	// check deactivated behavior
	app.config.UnescapePath = false
	resp, err = app.Test(httptest.NewRequest(MethodGet, literal_2379, nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, StatusNotFound, resp.StatusCode, literal_6917)
}

func TestRouteMatchWithEscapeChar(t *testing.T) {
	t.Parallel()

	app := New()
	// static route and escaped part
	app.Get("/v1/some/resource/name\\:customVerb", func(c Ctx) error {
		return c.SendString("static")
	})
	// group route
	group := app.Group("/v2/\\:firstVerb")
	group.Get("/\\:customVerb", func(c Ctx) error {
		return c.SendString("group")
	})
	// route with resource param and escaped part
	app.Get("/v3/:resource/name\\:customVerb", func(c Ctx) error {
		return c.SendString(c.Params("resource"))
	})

	// check static route
	resp, err := app.Test(httptest.NewRequest(MethodGet, "/v1/some/resource/name:customVerb", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, StatusOK, resp.StatusCode, literal_6917)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "static", app.getString(body))

	// check group route
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/v2/:firstVerb/:customVerb", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, StatusOK, resp.StatusCode, literal_6917)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "group", app.getString(body))

	// check param route
	resp, err = app.Test(httptest.NewRequest(MethodGet, "/v3/awesome/name:customVerb", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, StatusOK, resp.StatusCode, literal_6917)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "awesome", app.getString(body))
}

func TestRouteMatchMiddlewareHasPrefix(t *testing.T) {
	t.Parallel()

	app := New()

	app.Use("/foo", func(c Ctx) error {
		return c.SendString("middleware")
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/foo/bar", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, 200, resp.StatusCode, literal_6917)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "middleware", app.getString(body))
}

func TestRouteMatchMiddlewareRoot(t *testing.T) {
	t.Parallel()

	app := New()

	app.Use("/", func(c Ctx) error {
		return c.SendString("middleware")
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/everything", nil))
	require.NoError(t, err, literal_3769)
	require.Equal(t, 200, resp.StatusCode, literal_6917)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, literal_3769)
	require.Equal(t, "middleware", app.getString(body))
}

func TestRouterRegisterMissingHandler(t *testing.T) {
	t.Parallel()

	app := New()
	defer func() {
		if err := recover(); err != nil {
			require.Equal(t, "missing handler/middleware in route: /doe\n", fmt.Sprintf("%v", err))
		}
	}()
	app.register([]string{"USE"}, "/doe", nil, nil)
}

func TestEnsureRouterInterfaceImplementation(t *testing.T) {
	t.Parallel()

	var app any = (*App)(nil)
	_, ok := app.(Router)
	require.True(t, ok)

	var group any = (*Group)(nil)
	_, ok = group.(Router)
	require.True(t, ok)
}

func TestRouterHandlerCatchError(t *testing.T) {
	t.Parallel()

	app := New()
	app.config.ErrorHandler = func(_ Ctx, _ error) error {
		return errors.New("fake error")
	}

	app.Get("/", func(_ Ctx) error {
		return ErrForbidden
	})

	c := &fasthttp.RequestCtx{}

	app.Handler()(c)

	require.Equal(t, StatusInternalServerError, c.Response.Header.StatusCode())
}

func TestRouterNotFound(t *testing.T) {
	t.Parallel()
	app := New()
	app.Use(func(c Ctx) error {
		return c.Next()
	})
	appHandler := app.Handler()
	c := &fasthttp.RequestCtx{}

	c.Request.Header.SetMethod("DELETE")
	c.URI().SetPath("/this/route/does/not/exist")

	appHandler(c)

	require.Equal(t, 404, c.Response.StatusCode())
	require.Equal(t, "Cannot DELETE /this/route/does/not/exist", string(c.Response.Body()))
}

func TestRouterNotFoundHTMLInject(t *testing.T) {
	t.Parallel()
	app := New()
	app.Use(func(c Ctx) error {
		return c.Next()
	})
	appHandler := app.Handler()
	c := &fasthttp.RequestCtx{}

	c.Request.Header.SetMethod("DELETE")
	c.URI().SetPath("/does/not/exist<script>alert('foo');</script>")

	appHandler(c)

	require.Equal(t, 404, c.Response.StatusCode())
	require.Equal(t, "Cannot DELETE /does/not/exist&lt;script&gt;alert(&#39;foo&#39;);&lt;/script&gt;", string(c.Response.Body()))
}

//////////////////////////////////////////////
///////////////// BENCHMARKS /////////////////
//////////////////////////////////////////////

func registerDummyRoutes(app *App) {
	h := func(_ Ctx) error {
		return nil
	}
	for _, r := range routesFixture.GithubAPI {
		app.Add([]string{r.Method}, r.Path, h)
	}
}

// go test -v -run=^$ -bench=Benchmark_App_MethodNotAllowed -benchmem -count=4
func BenchmarkAppMethodNotAllowed(b *testing.B) {
	app := New()
	h := func(c Ctx) error {
		return c.SendString("Hello World!")
	}
	app.All("/this/is/a/", h)
	app.Get("/this/is/a/dummy/route/oke", h)
	appHandler := app.Handler()
	c := &fasthttp.RequestCtx{}

	c.Request.Header.SetMethod("DELETE")
	c.URI().SetPath("/this/is/a/dummy/route/oke")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		appHandler(c)
	}
	b.StopTimer()
	require.Equal(b, 405, c.Response.StatusCode())
	require.Equal(b, MethodGet, string(c.Response.Header.Peek("Allow")))
	require.Equal(b, utils.StatusMessage(StatusMethodNotAllowed), string(c.Response.Body()))
}

// go test -v ./... -run=^$ -bench=Benchmark_Router_NotFound -benchmem -count=4
func BenchmarkRouterNotFound(b *testing.B) {
	app := New()
	app.Use(func(c Ctx) error {
		return c.Next()
	})
	registerDummyRoutes(app)
	appHandler := app.Handler()
	c := &fasthttp.RequestCtx{}

	c.Request.Header.SetMethod("DELETE")
	c.URI().SetPath("/this/route/does/not/exist")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		appHandler(c)
	}
	require.Equal(b, 404, c.Response.StatusCode())
	require.Equal(b, "Cannot DELETE /this/route/does/not/exist", string(c.Response.Body()))
}

// go test -v ./... -run=^$ -bench=Benchmark_Router_Handler -benchmem -count=4
func BenchmarkRouterHandler(b *testing.B) {
	app := New()
	registerDummyRoutes(app)
	appHandler := app.Handler()

	c := &fasthttp.RequestCtx{}

	c.Request.Header.SetMethod("DELETE")
	c.URI().SetPath(literal_4567)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		appHandler(c)
	}
}

func BenchmarkRouterHandlerStrictCase(b *testing.B) {
	app := New(Config{
		StrictRouting: true,
		CaseSensitive: true,
	})
	registerDummyRoutes(app)
	appHandler := app.Handler()

	c := &fasthttp.RequestCtx{}

	c.Request.Header.SetMethod("DELETE")
	c.URI().SetPath(literal_4567)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		appHandler(c)
	}
}

// go test -v ./... -run=^$ -bench=Benchmark_Router_Chain -benchmem -count=4
func BenchmarkRouterChain(b *testing.B) {
	app := New()
	handler := func(c Ctx) error {
		return c.Next()
	}
	app.Get("/", handler, handler, handler, handler, handler, handler)

	appHandler := app.Handler()

	c := &fasthttp.RequestCtx{}

	c.Request.Header.SetMethod(MethodGet)
	c.URI().SetPath("/")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		appHandler(c)
	}
}

// go test -v ./... -run=^$ -bench=Benchmark_Router_WithCompression -benchmem -count=4
func BenchmarkRouterWithCompression(b *testing.B) {
	app := New()
	handler := func(c Ctx) error {
		return c.Next()
	}
	app.Get("/", handler)
	app.Get("/", handler)
	app.Get("/", handler)
	app.Get("/", handler)
	app.Get("/", handler)
	app.Get("/", handler)

	appHandler := app.Handler()
	c := &fasthttp.RequestCtx{}

	c.Request.Header.SetMethod(MethodGet)
	c.URI().SetPath("/")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		appHandler(c)
	}
}

// go test -run=^$ -bench=Benchmark_Startup_Process -benchmem -count=9
func BenchmarkStartupProcess(b *testing.B) {
	for n := 0; n < b.N; n++ {
		app := New()
		registerDummyRoutes(app)
		app.startupProcess()
	}
}

// go test -v ./... -run=^$ -bench=Benchmark_Router_Next -benchmem -count=4
func BenchmarkRouterNext(b *testing.B) {
	app := New()
	registerDummyRoutes(app)
	app.startupProcess()

	request := &fasthttp.RequestCtx{}

	request.Request.Header.SetMethod("DELETE")
	request.URI().SetPath(literal_4567)
	var res bool
	var err error

	c := app.AcquireCtx(request).(*DefaultCtx) //nolint:errcheck, forcetypeassert // not needed

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.indexRoute = -1
		res, err = app.next(c)
	}
	require.NoError(b, err)
	require.True(b, res)
	require.Equal(b, 4, c.indexRoute)
}

// go test -v ./... -run=^$ -bench=Benchmark_Route_Match -benchmem -count=4
func BenchmarkRouteMatch(b *testing.B) {
	var match bool
	var params [maxParams]string

	parsed := parseRoute(literal_1783)
	route := &Route{
		use:         false,
		root:        false,
		star:        false,
		routeParser: parsed,
		Params:      parsed.params,
		path:        literal_1783,

		Path:   literal_1783,
		Method: "DELETE",
	}
	route.Handlers = append(route.Handlers, func(_ Ctx) error {
		return nil
	})
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		match = route.match(literal_4567, literal_4567, &params)
	}

	require.True(b, match)
	require.Equal(b, []string{"1337"}, params[0:len(parsed.params)])
}

// go test -v ./... -run=^$ -bench=Benchmark_Route_Match_Star -benchmem -count=4
func BenchmarkRouteMatchStar(b *testing.B) {
	var match bool
	var params [maxParams]string

	parsed := parseRoute("/*")
	route := &Route{
		use:         false,
		root:        false,
		star:        true,
		routeParser: parsed,
		Params:      parsed.params,
		path:        literal_9127,

		Path:   literal_9127,
		Method: "DELETE",
	}
	route.Handlers = append(route.Handlers, func(_ Ctx) error {
		return nil
	})
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		match = route.match(literal_9127, literal_9127, &params)
	}

	require.True(b, match)
	require.Equal(b, []string{"user/keys/bla"}, params[0:len(parsed.params)])
}

// go test -v ./... -run=^$ -bench=Benchmark_Route_Match_Root -benchmem -count=4
func BenchmarkRouteMatchRoot(b *testing.B) {
	var match bool
	var params [maxParams]string

	parsed := parseRoute("/")
	route := &Route{
		use:         false,
		root:        true,
		star:        false,
		path:        "/",
		routeParser: parsed,
		Params:      parsed.params,

		Path:   "/",
		Method: "DELETE",
	}
	route.Handlers = append(route.Handlers, func(_ Ctx) error {
		return nil
	})

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		match = route.match("/", "/", &params)
	}

	require.True(b, match)
	require.Equal(b, []string{}, params[0:len(parsed.params)])
}

// go test -v ./... -run=^$ -bench=Benchmark_Router_Handler_CaseSensitive -benchmem -count=4
func BenchmarkRouterHandlerCaseSensitive(b *testing.B) {
	app := New()
	app.config.CaseSensitive = true
	registerDummyRoutes(app)
	appHandler := app.Handler()

	c := &fasthttp.RequestCtx{}

	c.Request.Header.SetMethod("DELETE")
	c.URI().SetPath(literal_4567)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		appHandler(c)
	}
}

// go test -v ./... -run=^$ -bench=Benchmark_Router_Handler_Unescape -benchmem -count=4
func BenchmarkRouterHandlerUnescape(b *testing.B) {
	app := New()
	app.config.UnescapePath = true
	registerDummyRoutes(app)
	app.Delete(literal_0493, func(_ Ctx) error {
		return nil
	})

	appHandler := app.Handler()

	c := &fasthttp.RequestCtx{}

	c.Request.Header.SetMethod(MethodDelete)
	c.URI().SetPath(literal_2379)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		c.URI().SetPath(literal_2379)
		appHandler(c)
	}
}

// go test -run=^$ -bench=Benchmark_Router_Github_API -benchmem -count=16
func BenchmarkRouterGithubAPI(b *testing.B) {
	app := New()
	registerDummyRoutes(app)
	app.startupProcess()

	c := &fasthttp.RequestCtx{}
	var match bool
	var err error

	b.ResetTimer()

	for i := range routesFixture.TestRoutes {
		c.Request.Header.SetMethod(routesFixture.TestRoutes[i].Method)
		for n := 0; n < b.N; n++ {
			c.URI().SetPath(routesFixture.TestRoutes[i].Path)

			ctx := app.AcquireCtx(c).(*DefaultCtx) //nolint:errcheck, forcetypeassert // not needed

			match, err = app.next(ctx)
			app.ReleaseCtx(ctx)
		}

		require.NoError(b, err)
		require.True(b, match)
	}
}

type testRoute struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type routeJSON struct {
	TestRoutes []testRoute `json:"test_routes"`
	GithubAPI  []testRoute `json:"github_api"`
}

const literal_3769 = "app.Test(req)"

const literal_6917 = "Status code"

const literal_0493 = "/créer"

const literal_2379 = "/cr%C3%A9er"

const literal_4567 = "/user/keys/1337"

const literal_1783 = "/user/keys/:id"

const literal_9127 = "/user/keys/bla"
