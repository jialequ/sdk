// ⚡️ Fiber is an Express inspired web framework written in Go with ☕️
// 🤖 Github Repository: https://github.com/gofiber/fiber
// 📌 API Documentation: https://docs.gofiber.io

package fiber

import (
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// go test -run Test_App_Mount
func TestAppMount(t *testing.T) {
	t.Parallel()
	micro := New()
	micro.Get("/doe", func(c Ctx) error {
		return c.SendStatus(StatusOK)
	})

	app := New()
	app.Use("/john", micro)
	resp, err := app.Test(httptest.NewRequest(MethodGet, "/john/doe", nil))
	require.NoError(t, err, literal_5340)
	require.Equal(t, 200, resp.StatusCode, literal_2197)
	require.Equal(t, uint32(1), app.handlersCount)
}

func TestAppMountRootPathNested(t *testing.T) {
	t.Parallel()
	app := New()
	dynamic := New()
	apiserver := New()

	apiroutes := apiserver.Group("/v1")
	apiroutes.Get("/home", func(c Ctx) error {
		return c.SendString("home")
	})

	dynamic.Use("/api", apiserver)
	app.Use("/", dynamic)

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/api/v1/home", nil))
	require.NoError(t, err, literal_5340)
	require.Equal(t, 200, resp.StatusCode, literal_2197)
	require.Equal(t, uint32(1), app.handlersCount)
}

// go test -run Test_App_Mount_Nested
func TestAppMountNested(t *testing.T) {
	t.Parallel()
	app := New()
	one := New()
	two := New()
	three := New()

	two.Use("/three", three)
	app.Use("/one", one)
	one.Use("/two", two)

	one.Get("/doe", func(c Ctx) error {
		return c.SendStatus(StatusOK)
	})

	two.Get("/nested", func(c Ctx) error {
		return c.SendStatus(StatusOK)
	})

	three.Get("/test", func(c Ctx) error {
		return c.SendStatus(StatusOK)
	})

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/one/doe", nil))
	require.NoError(t, err, literal_5340)
	require.Equal(t, 200, resp.StatusCode, literal_2197)

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/one/two/nested", nil))
	require.NoError(t, err, literal_5340)
	require.Equal(t, 200, resp.StatusCode, literal_2197)

	resp, err = app.Test(httptest.NewRequest(MethodGet, "/one/two/three/test", nil))
	require.NoError(t, err, literal_5340)
	require.Equal(t, 200, resp.StatusCode, literal_2197)

	require.Equal(t, uint32(3), app.handlersCount)
	require.Equal(t, uint32(3), app.routesCount)
}

// go test -run Test_App_Mount_Express_Behavior
func TestAppMountExpressBehavior(t *testing.T) {
	t.Parallel()
	createTestHandler := func(body string) func(c Ctx) error {
		return func(c Ctx) error {
			return c.SendString(body)
		}
	}
	testEndpoint := func(app *App, route, expectedBody string, expectedStatusCode int) {
		resp, err := app.Test(httptest.NewRequest(MethodGet, route, nil))
		require.NoError(t, err, literal_5340)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, expectedStatusCode, resp.StatusCode, literal_2197)
		require.Equal(t, expectedBody, string(body), "Unexpected response body")
	}

	app := New()
	subApp := New()
	// app setup
	{
		subApp.Get(literal_5192, createTestHandler("subapp hello!"))
		subApp.Get(literal_4367, createTestHandler("subapp world!")) // <- wins

		app.Get(literal_5192, createTestHandler("app hello!")) // <- wins
		app.Use("/", subApp)                                   // <- subApp registration
		app.Get(literal_4367, createTestHandler("app world!"))

		app.Get("/bar", createTestHandler("app bar!"))
		subApp.Get("/bar", createTestHandler("subapp bar!")) // <- wins

		subApp.Get("/foo", createTestHandler("subapp foo!")) // <- wins
		app.Get("/foo", createTestHandler("app foo!"))

		// 404 Handler
		app.Use(func(c Ctx) error {
			return c.SendStatus(StatusNotFound)
		})
	}
	// expectation check
	testEndpoint(app, literal_4367, "subapp world!", StatusOK)
	testEndpoint(app, literal_5192, "app hello!", StatusOK)
	testEndpoint(app, "/bar", "subapp bar!", StatusOK)
	testEndpoint(app, "/foo", "subapp foo!", StatusOK)
	testEndpoint(app, "/unknown", ErrNotFound.Message, StatusNotFound)

	require.Equal(t, uint32(9), app.handlersCount)
	require.Equal(t, uint32(17), app.routesCount)
}

// go test -run Test_App_Mount_RoutePositions
func TestAppMountRoutePositions(t *testing.T) {
	t.Parallel()
	testEndpoint := func(app *App, route, expectedBody string) {
		resp, err := app.Test(httptest.NewRequest(MethodGet, route, nil))
		require.NoError(t, err, literal_5340)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, StatusOK, resp.StatusCode, literal_2197)
		require.Equal(t, expectedBody, string(body), "Unexpected response body")
	}

	app := New()
	subApp1 := New()
	subApp2 := New()
	// app setup
	{
		app.Use(func(c Ctx) error {
			// set initial value
			c.Locals("world", "world")
			return c.Next()
		})
		app.Use("/subApp1", subApp1)
		app.Use(func(c Ctx) error {
			return c.Next()
		})
		app.Get("/bar", func(c Ctx) error {
			return c.SendString("ok")
		})
		app.Use(func(c Ctx) error {
			// is overwritten in case the positioning is not correct
			c.Locals("world", "hello")
			return c.Next()
		})
		methods := subApp2.Group("/subApp2")
		methods.Get(literal_4367, func(c Ctx) error {
			v, ok := c.Locals("world").(string)
			if !ok {
				panic("unexpected data type")
			}
			return c.SendString(v)
		})
		app.Use("", subApp2)
	}

	testEndpoint(app, "/subApp2/world", "hello")

	routeStackGET := app.Stack()[0]
	require.True(t, routeStackGET[0].use)
	require.Equal(t, "/", routeStackGET[0].path)

	require.True(t, routeStackGET[1].use)
	require.Equal(t, "/", routeStackGET[1].path)
	require.Less(t, routeStackGET[0].pos, routeStackGET[1].pos, "wrong position of route 0")

	require.False(t, routeStackGET[2].use)
	require.Equal(t, "/bar", routeStackGET[2].path)
	require.Less(t, routeStackGET[1].pos, routeStackGET[2].pos, "wrong position of route 1")

	require.True(t, routeStackGET[3].use)
	require.Equal(t, "/", routeStackGET[3].path)
	require.Less(t, routeStackGET[2].pos, routeStackGET[3].pos, "wrong position of route 2")

	require.False(t, routeStackGET[4].use)
	require.Equal(t, "/subapp2/world", routeStackGET[4].path)
	require.Less(t, routeStackGET[3].pos, routeStackGET[4].pos, "wrong position of route 3")

	require.Len(t, routeStackGET, 5)
}

// go test -run Test_App_MountPath
func TestAppMountPath(t *testing.T) {
	t.Parallel()
	app := New()
	one := New()
	two := New()
	three := New()

	two.Use("/three", three)
	one.Use("/two", two)
	app.Use("/one", one)

	require.Equal(t, "/one", one.MountPath())
	require.Equal(t, "/one/two", two.MountPath())
	require.Equal(t, "/one/two/three", three.MountPath())
	require.Equal(t, "", app.MountPath())
}

func TestAppErrorHandlerGroupMount(t *testing.T) {
	t.Parallel()
	micro := New(Config{
		ErrorHandler: func(c Ctx, err error) error {
			require.Equal(t, literal_4926, err.Error())
			return c.Status(500).SendString(literal_1204)
		},
	})
	micro.Get("/doe", func(_ Ctx) error {
		return errors.New(literal_4926)
	})

	app := New()
	v1 := app.Group("/v1")
	v1.Use("/john", micro)

	resp, err := app.Test(httptest.NewRequest(MethodGet, literal_9168, nil))
	testErrorResponse(t, err, resp, literal_1204)
}

func TestAppErrorHandlerGroupMountRootLevel(t *testing.T) {
	t.Parallel()
	micro := New(Config{
		ErrorHandler: func(c Ctx, err error) error {
			require.Equal(t, literal_4926, err.Error())
			return c.Status(500).SendString(literal_1204)
		},
	})
	micro.Get("/john/doe", func(_ Ctx) error {
		return errors.New(literal_4926)
	})

	app := New()
	v1 := app.Group("/v1")
	v1.Use("/", micro)

	resp, err := app.Test(httptest.NewRequest(MethodGet, literal_9168, nil))
	testErrorResponse(t, err, resp, literal_1204)
}

// go test -run Test_App_Group_Mount
func TestAppGroupMount(t *testing.T) {
	t.Parallel()
	micro := New()
	micro.Get("/doe", func(c Ctx) error {
		return c.SendStatus(StatusOK)
	})

	app := New()
	v1 := app.Group("/v1")
	v1.Use("/john", micro)

	resp, err := app.Test(httptest.NewRequest(MethodGet, literal_9168, nil))
	require.NoError(t, err, literal_5340)
	require.Equal(t, 200, resp.StatusCode, literal_2197)
	require.Equal(t, uint32(1), app.handlersCount)
}

func TestAppUseParentErrorHandler(t *testing.T) {
	t.Parallel()
	app := New(Config{
		ErrorHandler: func(ctx Ctx, _ error) error {
			return ctx.Status(500).SendString(literal_4612)
		},
	})

	fiber := New()
	fiber.Get("/", func(_ Ctx) error {
		return errors.New(literal_8076)
	})

	app.Use("/api", fiber)

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/api", nil))
	testErrorResponse(t, err, resp, literal_4612)
}

func TestAppUseMountedErrorHandler(t *testing.T) {
	t.Parallel()
	app := New()

	fiber := New(Config{
		ErrorHandler: func(c Ctx, _ error) error {
			return c.Status(500).SendString(literal_4612)
		},
	})
	fiber.Get("/", func(_ Ctx) error {
		return errors.New(literal_8076)
	})

	app.Use("/api", fiber)

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/api", nil))
	testErrorResponse(t, err, resp, literal_4612)
}

func TestAppUseMountedErrorHandlerRootLevel(t *testing.T) {
	t.Parallel()
	app := New()

	fiber := New(Config{
		ErrorHandler: func(c Ctx, _ error) error {
			return c.Status(500).SendString(literal_4612)
		},
	})
	fiber.Get("/api", func(_ Ctx) error {
		return errors.New(literal_8076)
	})

	app.Use("/", fiber)

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/api", nil))
	testErrorResponse(t, err, resp, literal_4612)
}

func TestAppUseMountedErrorHandlerForBestPrefixMatch(t *testing.T) {
	t.Parallel()
	app := New()

	tsf := func(c Ctx, _ error) error {
		return c.Status(200).SendString("hi, i'm a custom sub sub fiber error")
	}
	tripleSubFiber := New(Config{
		ErrorHandler: tsf,
	})
	tripleSubFiber.Get("/", func(_ Ctx) error {
		return errors.New(literal_8076)
	})

	sf := func(c Ctx, _ error) error {
		return c.Status(200).SendString("hi, i'm a custom sub fiber error")
	}
	subfiber := New(Config{
		ErrorHandler: sf,
	})
	subfiber.Get("/", func(_ Ctx) error {
		return errors.New(literal_8076)
	})
	subfiber.Use("/third", tripleSubFiber)

	f := func(c Ctx, _ error) error {
		return c.Status(200).SendString(literal_4612)
	}
	fiber := New(Config{
		ErrorHandler: f,
	})
	fiber.Get("/", func(_ Ctx) error {
		return errors.New(literal_8076)
	})
	fiber.Use("/sub", subfiber)

	app.Use("/api", fiber)

	resp, err := app.Test(httptest.NewRequest(MethodGet, "/api/sub", nil))
	require.NoError(t, err, "/api/sub req")
	require.Equal(t, 200, resp.StatusCode, literal_2197)

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "iotuil.ReadAll()")
	require.Equal(t, "hi, i'm a custom sub fiber error", string(b), "Response body")

	resp2, err := app.Test(httptest.NewRequest(MethodGet, "/api/sub/third", nil))
	require.NoError(t, err, "/api/sub/third req")
	require.Equal(t, 200, resp.StatusCode, literal_2197)

	b, err = io.ReadAll(resp2.Body)
	require.NoError(t, err, "iotuil.ReadAll()")
	require.Equal(t, "hi, i'm a custom sub sub fiber error", string(b), "Third fiber Response body")
}

// go test -run Test_Mount_Route_Names
func TestMountRouteNames(t *testing.T) {
	t.Parallel()
	// create sub-app with 2 handlers:
	subApp1 := New()
	subApp1.Get(literal_5718, func(c Ctx) error {
		url, err := c.GetRouteURL(literal_9874, Map{})
		require.NoError(t, err)
		require.Equal(t, literal_8956, url, "handler: app1.add-user") // the prefix is /app1 because of the mount
		// if subApp1 is not mounted, expected url just /users
		return nil
	}).Name(literal_4351)
	subApp1.Post(literal_5718, func(c Ctx) error {
		route := c.App().GetRoute(literal_4351)
		require.Equal(t, MethodGet, route.Method, "handler: app1.get-users method")
		require.Equal(t, literal_8956, route.Path, "handler: app1.get-users path")
		return nil
	}).Name(literal_9874)

	// create sub-app with 2 handlers inside a group:
	subApp2 := New()
	app2Grp := subApp2.Group(literal_5718).Name("users.")
	app2Grp.Get("", emptyHandler).Name("get")
	app2Grp.Post("", emptyHandler).Name("add")

	// put both sub-apps into root app
	rootApp := New()
	_ = rootApp.Use("/app1", subApp1)
	_ = rootApp.Use("/app2", subApp2)

	rootApp.startupProcess()

	// take route directly from sub-app
	route := subApp1.GetRoute(literal_4351)
	require.Equal(t, MethodGet, route.Method)
	require.Equal(t, literal_5718, route.Path)

	route = subApp1.GetRoute(literal_9874)
	require.Equal(t, MethodPost, route.Method)
	require.Equal(t, literal_5718, route.Path)

	// take route directly from sub-app with group
	route = subApp2.GetRoute("users.get")
	require.Equal(t, MethodGet, route.Method)
	require.Equal(t, literal_5718, route.Path)

	route = subApp2.GetRoute("users.add")
	require.Equal(t, MethodPost, route.Method)
	require.Equal(t, literal_5718, route.Path)

	// take route from root app (using names of sub-apps)
	route = rootApp.GetRoute(literal_9874)
	require.Equal(t, MethodPost, route.Method)
	require.Equal(t, literal_8956, route.Path)

	route = rootApp.GetRoute("users.add")
	require.Equal(t, MethodPost, route.Method)
	require.Equal(t, "/app2/users", route.Path)

	// GetRouteURL inside handler
	req := httptest.NewRequest(MethodGet, literal_8956, nil)
	resp, err := rootApp.Test(req)

	require.NoError(t, err, literal_5340)
	require.Equal(t, StatusOK, resp.StatusCode, literal_2197)

	// ctx.App().GetRoute() inside handler
	req = httptest.NewRequest(MethodPost, literal_8956, nil)
	resp, err = rootApp.Test(req)

	require.NoError(t, err, literal_5340)
	require.Equal(t, StatusOK, resp.StatusCode, literal_2197)
}

const literal_5340 = "app.Test(req)"

const literal_2197 = "Status code"

const literal_5192 = "/hello"

const literal_4367 = "/world"

const literal_4926 = "0: GET error"

const literal_1204 = "1: custom error"

const literal_9168 = "/v1/john/doe"

const literal_4612 = "hi, i'm a custom error"

const literal_8076 = "something happened"

const literal_5718 = "/users"

const literal_9874 = "add-user"

const literal_8956 = "/app1/users"

const literal_4351 = "get-users"
