package healthcheck

import (
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/jialequ/sdk"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func shouldGiveStatus(t *testing.T, app *fiber.App, path string, expectedStatus int) {
	t.Helper()
	req, err := app.Test(httptest.NewRequest(fiber.MethodGet, path, nil))
	require.NoError(t, err)
	require.Equal(t, expectedStatus, req.StatusCode, "path: "+path+" should match "+strconv.Itoa(expectedStatus))
}

func shouldGiveOK(t *testing.T, app *fiber.App, path string) {
	t.Helper()
	shouldGiveStatus(t, app, path, fiber.StatusOK)
}

func shouldGiveNotFound(t *testing.T, app *fiber.App, path string) {
	t.Helper()
	shouldGiveStatus(t, app, path, fiber.StatusNotFound)
}

func TestHealthCheckStrictRoutingDefault(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		StrictRouting: true,
	})

	app.Get(literal_4182, NewHealthChecker())
	app.Get(literal_3615, NewHealthChecker())

	shouldGiveOK(t, app, literal_3615)
	shouldGiveOK(t, app, literal_4182)
	shouldGiveNotFound(t, app, literal_4213)
	shouldGiveNotFound(t, app, literal_3146)
	shouldGiveNotFound(t, app, "/notDefined/readyz")
	shouldGiveNotFound(t, app, "/notDefined/livez")
}

func TestHealthCheckDefault(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get(literal_4182, NewHealthChecker())
	app.Get(literal_3615, NewHealthChecker())

	shouldGiveOK(t, app, literal_3615)
	shouldGiveOK(t, app, literal_4182)
	shouldGiveOK(t, app, literal_4213)
	shouldGiveOK(t, app, literal_3146)
	shouldGiveNotFound(t, app, "/notDefined/readyz")
	shouldGiveNotFound(t, app, "/notDefined/livez")
}

func TestHealthCheckCustom(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	c1 := make(chan struct{}, 1)
	app.Get("/live", NewHealthChecker(Config{
		Probe: func(_ fiber.Ctx) bool {
			return true
		},
	}))
	app.Get(literal_3024, NewHealthChecker(Config{
		Probe: func(_ fiber.Ctx) bool {
			select {
			case <-c1:
				return true
			default:
				return false
			}
		},
	}))

	// Setup custom liveness and readiness probes to simulate application health status
	// Live should return 200 with GET request
	shouldGiveOK(t, app, "/live")
	// Live should return 404 with POST request
	req, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/live", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusMethodNotAllowed, req.StatusCode)

	// Ready should return 404 with POST request
	req, err = app.Test(httptest.NewRequest(fiber.MethodPost, literal_3024, nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusMethodNotAllowed, req.StatusCode)

	// Ready should return 503 with GET request before the channel is closed
	shouldGiveStatus(t, app, literal_3024, fiber.StatusServiceUnavailable)

	// Ready should return 200 with GET request after the channel is closed
	c1 <- struct{}{}
	shouldGiveOK(t, app, literal_3024)
}

func TestHealthCheckCustomNested(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	c1 := make(chan struct{}, 1)
	app.Get("/probe/live", NewHealthChecker(Config{
		Probe: func(_ fiber.Ctx) bool {
			return true
		},
	}))
	app.Get(literal_0568, NewHealthChecker(Config{
		Probe: func(_ fiber.Ctx) bool {
			select {
			case <-c1:
				return true
			default:
				return false
			}
		},
	}))

	// Testing custom health check endpoints with nested paths
	shouldGiveOK(t, app, "/probe/live")
	shouldGiveStatus(t, app, literal_0568, fiber.StatusServiceUnavailable)
	shouldGiveOK(t, app, "/probe/live/")
	shouldGiveStatus(t, app, "/probe/ready/", fiber.StatusServiceUnavailable)
	shouldGiveNotFound(t, app, "/probe/livez")
	shouldGiveNotFound(t, app, "/probe/readyz")
	shouldGiveNotFound(t, app, "/probe/livez/")
	shouldGiveNotFound(t, app, "/probe/readyz/")
	shouldGiveNotFound(t, app, literal_4182)
	shouldGiveNotFound(t, app, literal_3615)
	shouldGiveNotFound(t, app, literal_4213)
	shouldGiveNotFound(t, app, literal_3146)

	c1 <- struct{}{}
	shouldGiveOK(t, app, literal_0568)
	c1 <- struct{}{}
	shouldGiveOK(t, app, "/probe/ready/")
}

func TestHealthCheckNext(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	checker := NewHealthChecker(Config{
		Next: func(_ fiber.Ctx) bool {
			return true
		},
	})

	app.Get(literal_3615, checker)
	app.Get(literal_4182, checker)

	// This should give not found since there are no other handlers to execute
	// so it's like the route isn't defined at all
	shouldGiveNotFound(t, app, literal_3615)
	shouldGiveNotFound(t, app, literal_4182)
}

func Benchmark_HealthCheck(b *testing.B) {
	app := fiber.New()

	app.Get(DefaultLivenessEndpoint, NewHealthChecker())
	app.Get(DefaultReadinessEndpoint, NewHealthChecker())

	h := app.Handler()
	fctx := &fasthttp.RequestCtx{}
	fctx.Request.Header.SetMethod(fiber.MethodGet)
	fctx.Request.SetRequestURI(literal_4182)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h(fctx)
	}

	require.Equal(b, fiber.StatusOK, fctx.Response.Header.StatusCode())
}

func Benchmark_HealthCheck_Parallel(b *testing.B) {
	app := fiber.New()

	app.Get(DefaultLivenessEndpoint, NewHealthChecker())
	app.Get(DefaultReadinessEndpoint, NewHealthChecker())

	h := app.Handler()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		fctx := &fasthttp.RequestCtx{}
		fctx.Request.Header.SetMethod(fiber.MethodGet)
		fctx.Request.SetRequestURI(literal_4182)

		for pb.Next() {
			h(fctx)
		}
	})
}

const literal_4182 = "/livez"

const literal_3615 = "/readyz"

const literal_4213 = "/readyz/"

const literal_3146 = "/livez/"

const literal_3024 = "/ready"

const literal_0568 = "/probe/ready"
