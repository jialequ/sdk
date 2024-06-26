package timeout

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jialequ/sdk"
	"github.com/stretchr/testify/require"
)

// go test -run Test_WithContextTimeout
func TestWithContextTimeout(t *testing.T) {
	t.Parallel()
	// fiber instance
	app := fiber.New()
	h := New(func(c fiber.Ctx) error {
		sleepTime, err := time.ParseDuration(c.Params("sleepTime") + "ms")
		require.NoError(t, err)
		if err := sleepWithContext(c.UserContext(), sleepTime, context.DeadlineExceeded); err != nil {
			return fmt.Errorf("%w: l2 wrap", fmt.Errorf("%w: l1 wrap ", err))
		}
		return nil
	}, 100*time.Millisecond)
	app.Get("/test/:sleepTime", h)
	testTimeout := func(timeoutStr string) {
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_2964+timeoutStr, nil))
		require.NoError(t, err, literal_6473)
		require.Equal(t, fiber.StatusRequestTimeout, resp.StatusCode, literal_2078)
	}
	testSucces := func(timeoutStr string) {
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_2964+timeoutStr, nil))
		require.NoError(t, err, literal_6473)
		require.Equal(t, fiber.StatusOK, resp.StatusCode, literal_2078)
	}
	testTimeout("300")
	testTimeout("500")
	testSucces("50")
	testSucces("30")
}

var ErrFooTimeOut = errors.New("foo context canceled")

// go test -run Test_WithContextTimeoutWithCustomError
func TestWithContextTimeoutWithCustomError(t *testing.T) {
	t.Parallel()
	// fiber instance
	app := fiber.New()
	h := New(func(c fiber.Ctx) error {
		sleepTime, err := time.ParseDuration(c.Params("sleepTime") + "ms")
		require.NoError(t, err)
		if err := sleepWithContext(c.UserContext(), sleepTime, ErrFooTimeOut); err != nil {
			return fmt.Errorf("%w: execution error", err)
		}
		return nil
	}, 100*time.Millisecond, ErrFooTimeOut)
	app.Get("/test/:sleepTime", h)
	testTimeout := func(timeoutStr string) {
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_2964+timeoutStr, nil))
		require.NoError(t, err, literal_6473)
		require.Equal(t, fiber.StatusRequestTimeout, resp.StatusCode, literal_2078)
	}
	testSucces := func(timeoutStr string) {
		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, literal_2964+timeoutStr, nil))
		require.NoError(t, err, literal_6473)
		require.Equal(t, fiber.StatusOK, resp.StatusCode, literal_2078)
	}
	testTimeout("300")
	testTimeout("500")
	testSucces("50")
	testSucces("30")
}

func sleepWithContext(ctx context.Context, d time.Duration, te error) error {
	timer := time.NewTimer(d)
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		return te
	case <-timer.C:
	}
	return nil
}

const literal_2964 = "/test/"

const literal_6473 = "app.Test(req)"

const literal_2078 = "Status code"
