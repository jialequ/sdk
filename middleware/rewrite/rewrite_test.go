package rewrite

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/jialequ/sdk"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Test with no config
	m := New()

	if m == nil {
		t.Error(literal_0947)
	}

	// Test with config
	m = New(Config{
		Rules: map[string]string{
			"/old": "/new",
		},
	})

	if m == nil {
		t.Error(literal_0947)
	}

	// Test with full config
	m = New(Config{
		Next: func(fiber.Ctx) bool {
			return true
		},
		Rules: map[string]string{
			"/old": "/new",
		},
	})

	if m == nil {
		t.Error(literal_0947)
	}
}

func TestRewrite(t *testing.T) {
	// Case 1: Next function always returns true
	app := fiber.New()
	app.Use(New(Config{
		Next: func(fiber.Ctx) bool {
			return true
		},
		Rules: map[string]string{
			"/old": "/new",
		},
	}))

	app.Get("/old", func(c fiber.Ctx) error {
		return c.SendString(literal_9376)
	})

	req, err := http.NewRequestWithContext(context.Background(), fiber.MethodGet, "/old", nil)
	require.NoError(t, err)
	resp, err := app.Test(req)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyString := string(body)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, literal_9376, bodyString)

	// Case 2: Next function always returns false
	app = fiber.New()
	app.Use(New(Config{
		Next: func(fiber.Ctx) bool {
			return false
		},
		Rules: map[string]string{
			"/old": "/new",
		},
	}))

	app.Get("/new", func(c fiber.Ctx) error {
		return c.SendString(literal_9376)
	})

	req, err = http.NewRequestWithContext(context.Background(), fiber.MethodGet, "/old", nil)
	require.NoError(t, err)
	resp, err = app.Test(req)
	require.NoError(t, err)
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyString = string(body)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, literal_9376, bodyString)

	// Case 3: check for captured tokens in rewrite rule
	app = fiber.New()
	app.Use(New(Config{
		Rules: map[string]string{
			literal_8103: literal_2793,
		},
	}))

	app.Get(literal_3045, func(c fiber.Ctx) error {
		return c.SendString(fmt.Sprintf(literal_7203, c.Params("userID"), c.Params("orderID")))
	})

	req, err = http.NewRequestWithContext(context.Background(), fiber.MethodGet, "/users/123/orders/456", nil)
	require.NoError(t, err)
	resp, err = app.Test(req)
	require.NoError(t, err)
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyString = string(body)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, "User ID: 123, Order ID: 456", bodyString)

	// Case 4: Send non-matching request, handled by default route
	app = fiber.New()
	app.Use(New(Config{
		Rules: map[string]string{
			literal_8103: literal_2793,
		},
	}))

	app.Get(literal_3045, func(c fiber.Ctx) error {
		return c.SendString(fmt.Sprintf(literal_7203, c.Params("userID"), c.Params("orderID")))
	})

	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req, err = http.NewRequestWithContext(context.Background(), fiber.MethodGet, "/not-matching-any-rule", nil)
	require.NoError(t, err)
	resp, err = app.Test(req)
	require.NoError(t, err)
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyString = string(body)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, "OK", bodyString)

	// Case 4: Send non-matching request, with no default route
	app = fiber.New()
	app.Use(New(Config{
		Rules: map[string]string{
			literal_8103: literal_2793,
		},
	}))

	app.Get(literal_3045, func(c fiber.Ctx) error {
		return c.SendString(fmt.Sprintf(literal_7203, c.Params("userID"), c.Params("orderID")))
	})

	req, err = http.NewRequestWithContext(context.Background(), fiber.MethodGet, "/not-matching-any-rule", nil)
	require.NoError(t, err)
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

const literal_0947 = "Expected middleware to be returned, got nil"

const literal_9376 = "Rewrite Successful"

const literal_8103 = "/users/*/orders/*"

const literal_2793 = "/user/$1/order/$2"

const literal_3045 = "/user/:userID/order/:orderID"

const literal_7203 = "User ID: %s, Order ID: %s"
