package cors

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jialequ/sdk"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestCORSDefaults(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New())

	testDefaultOrEmptyConfig(t, app)
}

func TestCORSEmptyConfig(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{}))

	testDefaultOrEmptyConfig(t, app)
}

func TestCORSWildcardHeaders(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		AllowMethods:  []string{"*"},
		AllowHeaders:  []string{"*"},
		ExposeHeaders: []string{"*"},
	}))

	h := app.Handler()

	// Test preflight request
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_3627)
	h(ctx)

	require.Equal(t, "*", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))
	require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowCredentials)))
	require.Equal(t, "*", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowMethods)))
	require.Equal(t, "*", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowHeaders)))
	require.Equal(t, "*", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlExposeHeaders)))
}

func TestCORSNegativeMaxAge(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(New(Config{MaxAge: -1}))

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_3627)
	app.Handler()(ctx)

	require.Equal(t, "0", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlMaxAge)))
}

func testDefaultOrEmptyConfig(t *testing.T, app *fiber.App) {
	t.Helper()

	h := app.Handler()

	// Test default GET response headers
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_3627)
	h(ctx)

	require.Equal(t, "*", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))
	require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowCredentials)))
	require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlExposeHeaders)))

	// Test default OPTIONS (preflight) response headers
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_3627)
	h(ctx)

	require.Equal(t, literal_3958, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowMethods)))
	require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowHeaders)))
	require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlMaxAge)))
}

func TestCORSAllowOriginsVary(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(
		Config{
			AllowOrigins: []string{literal_3627},
		},
	))

	h := app.Handler()

	// Test Vary header non-Cors request
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fiber.MethodGet)
	h(ctx)
	require.Contains(t, string(ctx.Response.Header.Peek(fiber.HeaderVary)), fiber.HeaderOrigin, literal_6982)

	// Test Vary header Cors request
	ctx.Request.Reset()
	ctx.Response.Reset()
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_3627)
	h(ctx)
	require.Contains(t, string(ctx.Response.Header.Peek(fiber.HeaderVary)), fiber.HeaderOrigin, literal_6982)
}

// go test -run -v Test_CORS_Wildcard
func TestCORSWildcard(t *testing.T) {
	t.Parallel()
	// New fiber instance
	app := fiber.New()
	// OPTIONS (preflight) response headers when AllowOrigins is *
	app.Use(New(Config{
		MaxAge:        3600,
		ExposeHeaders: []string{literal_3472},
		AllowHeaders:  []string{"Authentication"},
	}))
	// Get handler pointer
	handler := app.Handler()

	// Make request
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_3627)
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)

	// Perform request
	handler(ctx)

	// Check result
	require.Equal(t, "*", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin))) // Validates request is not reflecting origin in the response
	require.Contains(t, string(ctx.Response.Header.Peek(fiber.HeaderVary)), fiber.HeaderOrigin, literal_6982)
	require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowCredentials)))
	require.Equal(t, "3600", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlMaxAge)))
	require.Equal(t, "Authentication", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowHeaders)))

	// Test non OPTIONS (preflight) response headers
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_3627)
	handler(ctx)

	require.NotContains(t, string(ctx.Response.Header.Peek(fiber.HeaderVary)), fiber.HeaderOrigin, "Vary header should not be set")
	require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowCredentials)))
	require.Equal(t, literal_3472, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlExposeHeaders)))
}

// go test -run -v Test_CORS_Origin_AllowCredentials
func TestCORSOriginAllowCredentials(t *testing.T) {
	t.Parallel()
	// New fiber instance
	app := fiber.New()
	// OPTIONS (preflight) response headers when AllowOrigins is *
	app.Use(New(Config{
		AllowOrigins:     []string{literal_3627},
		AllowCredentials: true,
		MaxAge:           3600,
		ExposeHeaders:    []string{literal_3472},
		AllowHeaders:     []string{"Authentication"},
	}))
	// Get handler pointer
	handler := app.Handler()

	// Make request
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_3627)
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)

	// Perform request
	handler(ctx)

	// Check result
	require.Equal(t, literal_3627, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))
	require.Equal(t, "true", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowCredentials)))
	require.Equal(t, "3600", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlMaxAge)))
	require.Equal(t, "Authentication", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowHeaders)))

	// Test non OPTIONS (preflight) response headers
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_3627)
	ctx.Request.Header.SetMethod(fiber.MethodGet)
	handler(ctx)

	require.Equal(t, "true", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowCredentials)))
	require.Equal(t, literal_3472, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlExposeHeaders)))
}

// go test -run -v Test_CORS_Wildcard_AllowCredentials_Panic
// Test for fiber-ghsa-fmg4-x8pw-hjhg
func TestCORSWildcardAllowCredentialsPanic(t *testing.T) {
	t.Parallel()
	// New fiber instance
	app := fiber.New()

	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()

		app.Use(New(Config{
			AllowOrigins:     []string{"*"},
			AllowCredentials: true,
		}))
	}()

	if !didPanic {
		t.Errorf("Expected a panic when AllowOrigins is '*' and AllowCredentials is true")
	}
}

// go test -run -v Test_CORS_Invalid_Origin_Panic
func TestCORSInvalidOriginsPanic(t *testing.T) {
	t.Parallel()

	invalidOrigins := []string{
		"localhost",
		"http://foo.[a-z]*.example.com",
		"http://*",
		"https://*",
		"http://*.com*",
		"invalid url",
		"*",
		"http://origin.com,invalid url",
		// add more invalid origins as needed
	}

	for _, origin := range invalidOrigins {
		// New fiber instance
		app := fiber.New()

		didPanic := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					didPanic = true
				}
			}()

			app.Use(New(Config{
				AllowOrigins:     []string{origin},
				AllowCredentials: true,
			}))
		}()

		if !didPanic {
			t.Errorf("Expected a panic for invalid origin: %s", origin)
		}
	}
}

// go test -run -v Test_CORS_Subdomain
func TestCORSSubdomain(t *testing.T) {
	t.Parallel()
	// New fiber instance
	app := fiber.New()
	// OPTIONS (preflight) response headers when AllowOrigins is set to a subdomain
	app.Use("/", New(Config{
		AllowOrigins: []string{literal_8206},
	}))

	// Get handler pointer
	handler := app.Handler()

	// Make request with disallowed origin
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_9653)

	// Perform request
	handler(ctx)

	// Allow-Origin header should be "" because http://google.com does not satisfy http://*.example.com
	require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))

	ctx.Request.Reset()
	ctx.Response.Reset()

	// Make request with domain only (disallowed)
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_0543)

	handler(ctx)

	require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))

	ctx.Request.Reset()
	ctx.Response.Reset()

	// Make request with allowed origin
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, "http://test.example.com")

	handler(ctx)

	require.Equal(t, "http://test.example.com", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))
}

func TestCORSAllowOriginScheme(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pattern           []string
		reqOrigin         string
		shouldAllowOrigin bool
	}{
		{
			pattern:           []string{literal_0543},
			reqOrigin:         literal_0543,
			shouldAllowOrigin: true,
		},
		{
			pattern:           []string{"HTTP://EXAMPLE.COM"},
			reqOrigin:         literal_0543,
			shouldAllowOrigin: true,
		},
		{
			pattern:           []string{literal_7286},
			reqOrigin:         literal_7286,
			shouldAllowOrigin: true,
		},
		{
			pattern:           []string{literal_0543},
			reqOrigin:         literal_7286,
			shouldAllowOrigin: false,
		},
		{
			pattern:           []string{literal_8206},
			reqOrigin:         "http://aaa.example.com",
			shouldAllowOrigin: true,
		},
		{
			pattern:           []string{literal_8206},
			reqOrigin:         "http://bbb.aaa.example.com",
			shouldAllowOrigin: true,
		},
		{
			pattern:           []string{"http://*.aaa.example.com"},
			reqOrigin:         "http://bbb.aaa.example.com",
			shouldAllowOrigin: true,
		},
		{
			pattern:           []string{"http://*.example.com:8080"},
			reqOrigin:         "http://aaa.example.com:8080",
			shouldAllowOrigin: true,
		},
		{
			pattern:           []string{literal_8206},
			reqOrigin:         "http://1.2.aaa.example.com",
			shouldAllowOrigin: true,
		},
		{
			pattern:           []string{literal_0543},
			reqOrigin:         "http://gofiber.com",
			shouldAllowOrigin: false,
		},
		{
			pattern:           []string{"http://*.aaa.example.com"},
			reqOrigin:         literal_5340,
			shouldAllowOrigin: false,
		},
		{
			pattern:           []string{literal_8206},
			reqOrigin:         "http://1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.1234567890.example.com",
			shouldAllowOrigin: true,
		},
		{
			pattern:           []string{literal_0543},
			reqOrigin:         literal_5340,
			shouldAllowOrigin: false,
		},
		{
			pattern:           []string{"https://--aaa.bbb.com"},
			reqOrigin:         "https://prod-preview--aaa.bbb.com",
			shouldAllowOrigin: false,
		},
		{
			pattern:           []string{literal_8206},
			reqOrigin:         literal_5340,
			shouldAllowOrigin: true,
		},
		{
			pattern:           []string{literal_4608, literal_0543},
			reqOrigin:         literal_0543,
			shouldAllowOrigin: true,
		},
		{
			pattern:           []string{literal_4608, literal_0543},
			reqOrigin:         "http://domain-2.com",
			shouldAllowOrigin: false,
		},
		{
			pattern:           []string{literal_4608, literal_0543},
			reqOrigin:         literal_0543,
			shouldAllowOrigin: true,
		},
		{
			pattern:           []string{literal_4608, literal_0543},
			reqOrigin:         "http://domain-2.com",
			shouldAllowOrigin: false,
		},
		{
			pattern:           []string{literal_4608, literal_0543},
			reqOrigin:         literal_4608,
			shouldAllowOrigin: true,
		},
	}

	for _, tt := range tests {
		app := fiber.New()
		app.Use("/", New(Config{AllowOrigins: tt.pattern}))

		handler := app.Handler()

		ctx := &fasthttp.RequestCtx{}
		ctx.Request.SetRequestURI("/")
		ctx.Request.Header.SetMethod(fiber.MethodOptions)
		ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
		ctx.Request.Header.Set(fiber.HeaderOrigin, tt.reqOrigin)

		handler(ctx)

		if tt.shouldAllowOrigin {
			require.Equal(t, tt.reqOrigin, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))
		} else {
			require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))
		}
	}
}

func TestCORSAllowOriginHeaderNoMatch(t *testing.T) {
	t.Parallel()
	// New fiber instance
	app := fiber.New()
	app.Use("/", New(Config{
		AllowOrigins: []string{literal_2783, "https://example-1.com"},
	}))

	// Get handler pointer
	handler := app.Handler()

	// Make request with disallowed origin
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_9653)

	// Perform request
	handler(ctx)

	var headerExists bool
	ctx.Response.Header.VisitAll(func(key, _ []byte) {
		if string(key) == fiber.HeaderAccessControlAllowOrigin {
			headerExists = true
		}
	})
	require.False(t, headerExists, "Access-Control-Allow-Origin header should not be set")
}

// go test -run Test_CORS_Next
func TestCORSNext(t *testing.T) {
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

// go test -run Test_CORS_Headers_BasedOnRequestType
func TestCORSHeadersBasedOnRequestType(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New())
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	methods := []string{
		fiber.MethodGet,
		fiber.MethodPost,
		fiber.MethodPut,
		fiber.MethodDelete,
		fiber.MethodPatch,
		fiber.MethodHead,
	}

	// Get handler pointer
	handler := app.Handler()

	t.Run("Without origin", func(t *testing.T) {
		t.Parallel()
		// Make request without origin header, and without Access-Control-Request-Method
		for _, method := range methods {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.SetMethod(method)
			ctx.Request.SetRequestURI(literal_7395)
			handler(ctx)
			require.Equal(t, 200, ctx.Response.StatusCode(), "Status code should be 200")
			require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)), "Access-Control-Allow-Origin header should not be set")
		}
	})

	t.Run("Preflight request with origin and Access-Control-Request-Method headers", func(t *testing.T) {
		t.Parallel()
		// Make preflight request with origin header and with Access-Control-Request-Method
		for _, method := range methods {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.SetMethod(fiber.MethodOptions)
			ctx.Request.SetRequestURI(literal_7395)
			ctx.Request.Header.Set(fiber.HeaderOrigin, literal_0543)
			ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, method)
			handler(ctx)
			require.Equal(t, 204, ctx.Response.StatusCode(), "Status code should be 204")
			require.Equal(t, "*", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)), literal_3547)
			require.Equal(t, literal_3958, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowMethods)), "Access-Control-Allow-Methods header should be set (preflight request)")
			require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowHeaders)), "Access-Control-Allow-Headers header should be set (preflight request)")
		}
	})

	t.Run("Non-preflight request with origin", func(t *testing.T) {
		t.Parallel()
		// Make non-preflight request with origin header and with Access-Control-Request-Method
		for _, method := range methods {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.SetMethod(method)
			ctx.Request.SetRequestURI("https://example.com/api/action")
			ctx.Request.Header.Set(fiber.HeaderOrigin, literal_0543)
			handler(ctx)
			require.Equal(t, 200, ctx.Response.StatusCode(), "Status code should be 200")
			require.Equal(t, "*", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)), literal_3547)
			require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowMethods)), "Access-Control-Allow-Methods header should not be set (non-preflight request)")
			require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowHeaders)), "Access-Control-Allow-Headers header should not be set (non-preflight request)")
		}
	})

	t.Run("Preflight with Access-Control-Request-Headers", func(t *testing.T) {
		t.Parallel()
		// Make preflight request with origin header and with Access-Control-Request-Method
		for _, method := range methods {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.SetMethod(fiber.MethodOptions)
			ctx.Request.SetRequestURI(literal_7395)
			ctx.Request.Header.Set(fiber.HeaderOrigin, literal_0543)
			ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, method)
			ctx.Request.Header.Set(fiber.HeaderAccessControlRequestHeaders, "X-Custom-Header")
			handler(ctx)
			require.Equal(t, 204, ctx.Response.StatusCode(), "Status code should be 204")
			require.Equal(t, "*", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)), literal_3547)
			require.Equal(t, literal_3958, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowMethods)), "Access-Control-Allow-Methods header should be set (preflight request)")
			require.Equal(t, "X-Custom-Header", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowHeaders)), "Access-Control-Allow-Headers header should be set (preflight request)")
		}
	})
}

func TestCORSAllowOriginsAndAllowOriginsFunc(t *testing.T) {
	t.Parallel()
	// New fiber instance
	app := fiber.New()
	app.Use("/", New(Config{
		AllowOrigins: []string{literal_2783},
		AllowOriginsFunc: func(origin string) bool {
			return strings.Contains(origin, "example-2")
		},
	}))

	// Get handler pointer
	handler := app.Handler()

	// Make request with disallowed origin
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_9653)

	// Perform request
	handler(ctx)

	// Allow-Origin header should be "" because http://google.com does not satisfy http://example-1.com or 'strings.Contains(origin, "example-2")'
	require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))

	ctx.Request.Reset()
	ctx.Response.Reset()

	// Make request with allowed origin
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_2783)

	handler(ctx)

	require.Equal(t, literal_2783, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))

	ctx.Request.Reset()
	ctx.Response.Reset()

	// Make request with allowed origin
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_5063)

	handler(ctx)

	require.Equal(t, literal_5063, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))
}

func TestCORSAllowOriginsFunc(t *testing.T) {
	t.Parallel()
	// New fiber instance
	app := fiber.New()
	app.Use("/", New(Config{
		AllowOriginsFunc: func(origin string) bool {
			return strings.Contains(origin, "example-2")
		},
	}))

	// Get handler pointer
	handler := app.Handler()

	// Make request with disallowed origin
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_9653)

	// Perform request
	handler(ctx)

	// Allow-Origin header should be empty because http://google.com does not satisfy 'strings.Contains(origin, "example-2")'
	// and AllowOrigins has not been set
	require.Equal(t, "", string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))

	ctx.Request.Reset()
	ctx.Response.Reset()

	// Make request with allowed origin
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_5063)

	handler(ctx)

	// Allow-Origin header should be literal_5063
	require.Equal(t, literal_5063, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))
}

func TestCORSAllowOriginsAndAllowOriginsFuncAllUseCases(t *testing.T) {
	testCases := []struct {
		Name           string
		Config         Config
		RequestOrigin  string
		ResponseOrigin string
	}{
		{
			Name: "AllowOriginsDefined/AllowOriginsFuncUndefined/OriginAllowed",
			Config: Config{
				AllowOrigins:     []string{literal_9126},
				AllowOriginsFunc: nil,
			},
			RequestOrigin:  literal_9126,
			ResponseOrigin: literal_9126,
		},
		{
			Name: "AllowOriginsDefined/AllowOriginsFuncUndefined/MultipleOrigins/OriginAllowed",
			Config: Config{
				AllowOrigins:     []string{literal_9126, literal_8931},
				AllowOriginsFunc: nil,
			},
			RequestOrigin:  literal_8931,
			ResponseOrigin: literal_8931,
		},
		{
			Name: "AllowOriginsDefined/AllowOriginsFuncUndefined/MultipleOrigins/OriginNotAllowed",
			Config: Config{
				AllowOrigins:     []string{literal_9126, literal_8931},
				AllowOriginsFunc: nil,
			},
			RequestOrigin:  "http://ccc.com",
			ResponseOrigin: "",
		},
		{
			Name: "AllowOriginsDefined/AllowOriginsFuncUndefined/MultipleOrigins/Whitespace/OriginAllowed",
			Config: Config{
				AllowOrigins:     []string{" http://aaa.com ", " http://bbb.com "},
				AllowOriginsFunc: nil,
			},
			RequestOrigin:  literal_9126,
			ResponseOrigin: literal_9126,
		},
		{
			Name: "AllowOriginsDefined/AllowOriginsFuncUndefined/OriginNotAllowed",
			Config: Config{
				AllowOrigins:     []string{literal_9126},
				AllowOriginsFunc: nil,
			},
			RequestOrigin:  literal_8931,
			ResponseOrigin: "",
		},
		{
			Name: "AllowOriginsDefined/AllowOriginsFuncReturnsTrue/OriginAllowed",
			Config: Config{
				AllowOrigins: []string{literal_9126},
				AllowOriginsFunc: func(_ string) bool {
					return true
				},
			},
			RequestOrigin:  literal_9126,
			ResponseOrigin: literal_9126,
		},
		{
			Name: "AllowOriginsDefined/AllowOriginsFuncReturnsTrue/OriginNotAllowed",
			Config: Config{
				AllowOrigins: []string{literal_9126},
				AllowOriginsFunc: func(_ string) bool {
					return true
				},
			},
			RequestOrigin:  literal_8931,
			ResponseOrigin: literal_8931,
		},
		{
			Name: "AllowOriginsDefined/AllowOriginsFuncReturnsFalse/OriginAllowed",
			Config: Config{
				AllowOrigins: []string{literal_9126},
				AllowOriginsFunc: func(_ string) bool {
					return false
				},
			},
			RequestOrigin:  literal_9126,
			ResponseOrigin: literal_9126,
		},
		{
			Name: "AllowOriginsDefined/AllowOriginsFuncReturnsFalse/OriginNotAllowed",
			Config: Config{
				AllowOrigins: []string{literal_9126},
				AllowOriginsFunc: func(_ string) bool {
					return false
				},
			},
			RequestOrigin:  literal_8931,
			ResponseOrigin: "",
		},
		{
			Name: "AllowOriginsEmpty/AllowOriginsFuncUndefined/OriginAllowed",
			Config: Config{
				AllowOriginsFunc: nil,
			},
			RequestOrigin:  literal_9126,
			ResponseOrigin: "*",
		},
		{
			Name: "AllowOriginsEmpty/AllowOriginsFuncReturnsTrue/OriginAllowed",
			Config: Config{
				AllowOriginsFunc: func(_ string) bool {
					return true
				},
			},
			RequestOrigin:  literal_9126,
			ResponseOrigin: literal_9126,
		},
		{
			Name: "AllowOriginsEmpty/AllowOriginsFuncReturnsFalse/OriginNotAllowed",
			Config: Config{
				AllowOriginsFunc: func(_ string) bool {
					return false
				},
			},
			RequestOrigin:  literal_9126,
			ResponseOrigin: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := fiber.New()
			app.Use("/", New(tc.Config))

			handler := app.Handler()

			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetRequestURI("/")
			ctx.Request.Header.SetMethod(fiber.MethodOptions)
			ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
			ctx.Request.Header.Set(fiber.HeaderOrigin, tc.RequestOrigin)

			handler(ctx)

			require.Equal(t, tc.ResponseOrigin, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))
		})
	}
}

// The fix for issue #2422
func TestCORSAllowCredentials(t *testing.T) {
	testCases := []struct {
		Name                string
		Config              Config
		RequestOrigin       string
		ResponseOrigin      string
		ResponseCredentials string
	}{
		{
			Name: "AllowOriginsFuncDefined",
			Config: Config{
				AllowCredentials: true,
				AllowOriginsFunc: func(_ string) bool {
					return true
				},
			},
			RequestOrigin: literal_9126,
			// The AllowOriginsFunc config was defined, should use the real origin of the function
			ResponseOrigin:      literal_9126,
			ResponseCredentials: "true",
		},
		{
			Name: "fiber-ghsa-fmg4-x8pw-hjhg-wildcard-credentials",
			Config: Config{
				AllowCredentials: true,
				AllowOriginsFunc: func(_ string) bool {
					return true
				},
			},
			RequestOrigin:  "*",
			ResponseOrigin: "*",
			// Middleware will validate that wildcard wont set credentials to true
			ResponseCredentials: "",
		},
		{
			Name: "AllowOriginsFuncNotDefined",
			Config: Config{
				// Setting this to true will cause the middleware to panic since default AllowOrigins is "*"
				AllowCredentials: false,
			},
			RequestOrigin: literal_9126,
			// None of the AllowOrigins or AllowOriginsFunc config was defined, should use the default origin of "*"
			// which will cause the CORS error in the client:
			// The value of the 'Access-Control-Allow-Origin' header in the response must not be the wildcard '*'
			// when the request's credentials mode is 'include'.
			ResponseOrigin:      "*",
			ResponseCredentials: "",
		},
		{
			Name: "AllowOriginsDefined",
			Config: Config{
				AllowCredentials: true,
				AllowOrigins:     []string{literal_9126},
			},
			RequestOrigin:       literal_9126,
			ResponseOrigin:      literal_9126,
			ResponseCredentials: "true",
		},
		{
			Name: "AllowOriginsDefined/UnallowedOrigin",
			Config: Config{
				AllowCredentials: true,
				AllowOrigins:     []string{literal_9126},
			},
			RequestOrigin:       literal_8931,
			ResponseOrigin:      "",
			ResponseCredentials: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			app := fiber.New()
			app.Use("/", New(tc.Config))

			handler := app.Handler()

			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetRequestURI("/")
			ctx.Request.Header.SetMethod(fiber.MethodOptions)
			ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
			ctx.Request.Header.Set(fiber.HeaderOrigin, tc.RequestOrigin)

			handler(ctx)

			require.Equal(t, tc.ResponseCredentials, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowCredentials)))
			require.Equal(t, tc.ResponseOrigin, string(ctx.Response.Header.Peek(fiber.HeaderAccessControlAllowOrigin)))
		})
	}
}

// The Enhancement for issue #2804
func TestCORSAllowPrivateNetwork(t *testing.T) {
	t.Parallel()

	// Test scenario where AllowPrivateNetwork is enabled
	app := fiber.New()
	app.Use(New(Config{
		AllowPrivateNetwork: true,
	}))
	handler := app.Handler()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_7286)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(literal_2091, "true")
	handler(ctx)

	// Verify the Access-Control-Allow-Private-Network header is set to "true"
	require.Equal(t, "true", string(ctx.Response.Header.Peek(literal_2895)), literal_0684)

	// Non-preflight request should not have Access-Control-Allow-Private-Network header
	ctx.Request.Reset()
	ctx.Response.Reset()
	ctx.Request.Header.SetMethod(fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_7286)
	ctx.Request.Header.Set(literal_2091, "true")
	handler(ctx)

	require.Equal(t, "", string(ctx.Response.Header.Peek(literal_2895)), literal_0684)

	// Non-preflight GET request should not have Access-Control-Allow-Private-Network header
	require.Equal(t, "", string(ctx.Response.Header.Peek(literal_2895)), literal_0684)

	// Non-preflight OPTIONS request should not have Access-Control-Allow-Private-Network header
	ctx.Request.Reset()
	ctx.Response.Reset()
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_7286)
	ctx.Request.Header.Set(literal_2091, "true")
	handler(ctx)

	require.Equal(t, "", string(ctx.Response.Header.Peek(literal_2895)), literal_0684)

	// Reset ctx for next test
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_7286)

	// Test scenario where AllowPrivateNetwork is disabled (default)
	app = fiber.New()
	app.Use(New())
	handler = app.Handler()
	handler(ctx)

	// Verify the Access-Control-Allow-Private-Network header is not present
	require.Equal(t, "", string(ctx.Response.Header.Peek(literal_2895)), literal_0786)

	// Test scenario where AllowPrivateNetwork is disabled but client sends header
	app = fiber.New()
	app.Use(New())
	handler = app.Handler()

	ctx = &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_7286)
	ctx.Request.Header.Set(literal_2091, "true")
	handler(ctx)

	// Verify the Access-Control-Allow-Private-Network header is not present
	require.Equal(t, "", string(ctx.Response.Header.Peek(literal_2895)), literal_0786)

	// Test scenario where AllowPrivateNetwork is enabled and client does NOT send header
	app = fiber.New()
	app.Use(New(Config{
		AllowPrivateNetwork: true,
	}))
	handler = app.Handler()

	ctx = &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_7286)
	handler(ctx)

	// Verify the Access-Control-Allow-Private-Network header is not present
	require.Equal(t, "", string(ctx.Response.Header.Peek(literal_2895)), literal_0786)

	// Test scenario where AllowPrivateNetwork is enabled and client sends header with false value
	app = fiber.New()
	app.Use(New(Config{
		AllowPrivateNetwork: true,
	}))
	handler = app.Handler()

	ctx = &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(fiber.MethodOptions)
	ctx.Request.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
	ctx.Request.Header.Set(fiber.HeaderOrigin, literal_7286)
	ctx.Request.Header.Set(literal_2091, "false")
	handler(ctx)

	// Verify the Access-Control-Allow-Private-Network header is not present
	require.Equal(t, "", string(ctx.Response.Header.Peek(literal_2895)), literal_0786)
}

// go test -v -run=^$ -bench=Benchmark_CORS_NewHandler -benchmem -count=4
func Benchmark_CORS_NewHandler(b *testing.B) {
	app := fiber.New()
	c := New(Config{
		AllowOrigins:     []string{literal_3627, literal_0543},
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept},
		AllowCredentials: true,
		MaxAge:           600,
	})

	app.Use(c)
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := app.Handler()
	ctx := &fasthttp.RequestCtx{}

	req := &fasthttp.Request{}
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI("/")
	req.Header.Set(fiber.HeaderOrigin, literal_3627)
	req.Header.Set(fiber.HeaderAccessControlRequestHeaders, literal_2407)

	ctx.Init(req, nil, nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h(ctx)
	}
}

// go test -v -run=^$ -bench=Benchmark_CORS_NewHandlerParallel -benchmem -count=4
func Benchmark_CORS_NewHandlerParallel(b *testing.B) {
	app := fiber.New()
	c := New(Config{
		AllowOrigins:     []string{literal_3627, literal_0543},
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept},
		AllowCredentials: true,
		MaxAge:           600,
	})

	app.Use(c)
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := app.Handler()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ctx := &fasthttp.RequestCtx{}

		req := &fasthttp.Request{}
		req.Header.SetMethod(fiber.MethodGet)
		req.SetRequestURI("/")
		req.Header.Set(fiber.HeaderOrigin, literal_3627)
		req.Header.Set(fiber.HeaderAccessControlRequestHeaders, literal_2407)

		ctx.Init(req, nil, nil)

		for pb.Next() {
			h(ctx)
		}
	})
}

// go test -v -run=^$ -bench=Benchmark_CORS_NewHandlerSingleOrigin -benchmem -count=4
func Benchmark_CORS_NewHandlerSingleOrigin(b *testing.B) {
	app := fiber.New()
	c := New(Config{
		AllowOrigins:     []string{literal_0543},
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept},
		AllowCredentials: true,
		MaxAge:           600,
	})

	app.Use(c)
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := app.Handler()
	ctx := &fasthttp.RequestCtx{}

	req := &fasthttp.Request{}
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI("/")
	req.Header.Set(fiber.HeaderOrigin, literal_0543)
	req.Header.Set(fiber.HeaderAccessControlRequestHeaders, literal_2407)

	ctx.Init(req, nil, nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h(ctx)
	}
}

// go test -v -run=^$ -bench=Benchmark_CORS_NewHandlerSingleOriginParallel -benchmem -count=4
func Benchmark_CORS_NewHandlerSingleOriginParallel(b *testing.B) {
	app := fiber.New()
	c := New(Config{
		AllowOrigins:     []string{literal_0543},
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept},
		AllowCredentials: true,
		MaxAge:           600,
	})

	app.Use(c)
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := app.Handler()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ctx := &fasthttp.RequestCtx{}

		req := &fasthttp.Request{}
		req.Header.SetMethod(fiber.MethodGet)
		req.SetRequestURI("/")
		req.Header.Set(fiber.HeaderOrigin, literal_0543)
		req.Header.Set(fiber.HeaderAccessControlRequestHeaders, literal_2407)

		ctx.Init(req, nil, nil)

		for pb.Next() {
			h(ctx)
		}
	})
}

// go test -v -run=^$ -bench=Benchmark_CORS_NewHandlerWildcard -benchmem -count=4
func Benchmark_CORS_NewHandlerWildcard(b *testing.B) {
	app := fiber.New()
	c := New(Config{
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept},
		AllowCredentials: false,
		MaxAge:           600,
	})

	app.Use(c)
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := app.Handler()
	ctx := &fasthttp.RequestCtx{}

	req := &fasthttp.Request{}
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI("/")
	req.Header.Set(fiber.HeaderOrigin, literal_0543)
	req.Header.Set(fiber.HeaderAccessControlRequestHeaders, literal_2407)

	ctx.Init(req, nil, nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h(ctx)
	}
}

// go test -v -run=^$ -bench=Benchmark_CORS_NewHandlerWildcardParallel -benchmem -count=4
func Benchmark_CORS_NewHandlerWildcardParallel(b *testing.B) {
	app := fiber.New()
	c := New(Config{
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept},
		AllowCredentials: false,
		MaxAge:           600,
	})

	app.Use(c)
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := app.Handler()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ctx := &fasthttp.RequestCtx{}

		req := &fasthttp.Request{}
		req.Header.SetMethod(fiber.MethodGet)
		req.SetRequestURI("/")
		req.Header.Set(fiber.HeaderOrigin, literal_0543)
		req.Header.Set(fiber.HeaderAccessControlRequestHeaders, literal_2407)

		ctx.Init(req, nil, nil)

		for pb.Next() {
			h(ctx)
		}
	})
}

// go test -v -run=^$ -bench=Benchmark_CORS_NewHandlerPreflight -benchmem -count=4
func Benchmark_CORS_NewHandlerPreflight(b *testing.B) {
	app := fiber.New()
	c := New(Config{
		AllowOrigins:     []string{literal_3627, literal_0543},
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept},
		AllowCredentials: true,
		MaxAge:           600,
	})

	app.Use(c)
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := app.Handler()
	ctx := &fasthttp.RequestCtx{}

	// Preflight request
	req := &fasthttp.Request{}
	req.Header.SetMethod(fiber.MethodOptions)
	req.SetRequestURI("/")
	req.Header.Set(fiber.HeaderOrigin, literal_0543)
	req.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodPost)
	req.Header.Set(fiber.HeaderAccessControlRequestHeaders, literal_2407)

	ctx.Init(req, nil, nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h(ctx)
	}
}

// go test -v -run=^$ -bench=Benchmark_CORS_NewHandlerPreflightParallel -benchmem -count=4
func Benchmark_CORS_NewHandlerPreflightParallel(b *testing.B) {
	app := fiber.New()
	c := New(Config{
		AllowOrigins:     []string{literal_3627, literal_0543},
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept},
		AllowCredentials: true,
		MaxAge:           600,
	})

	app.Use(c)
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := app.Handler()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ctx := &fasthttp.RequestCtx{}

		req := &fasthttp.Request{}
		req.Header.SetMethod(fiber.MethodOptions)
		req.SetRequestURI("/")
		req.Header.Set(fiber.HeaderOrigin, literal_0543)
		req.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodPost)
		req.Header.Set(fiber.HeaderAccessControlRequestHeaders, literal_2407)

		ctx.Init(req, nil, nil)

		for pb.Next() {
			h(ctx)
		}
	})
}

// go test -v -run=^$ -bench=Benchmark_CORS_NewHandlerPreflightSingleOrigin -benchmem -count=4
func Benchmark_CORS_NewHandlerPreflightSingleOrigin(b *testing.B) {
	app := fiber.New()
	c := New(Config{
		AllowOrigins:     []string{literal_0543},
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept},
		AllowCredentials: true,
		MaxAge:           600,
	})

	app.Use(c)
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := app.Handler()
	ctx := &fasthttp.RequestCtx{}

	req := &fasthttp.Request{}
	req.Header.SetMethod(fiber.MethodOptions)
	req.SetRequestURI("/")
	req.Header.Set(fiber.HeaderOrigin, literal_0543)
	req.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodPost)
	req.Header.Set(fiber.HeaderAccessControlRequestHeaders, literal_2407)

	ctx.Init(req, nil, nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h(ctx)
	}
}

// go test -v -run=^$ -bench=Benchmark_CORS_NewHandlerPreflightSingleOriginParallel -benchmem -count=4
func Benchmark_CORS_NewHandlerPreflightSingleOriginParallel(b *testing.B) {
	app := fiber.New()
	c := New(Config{
		AllowOrigins:     []string{literal_0543},
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept},
		AllowCredentials: true,
		MaxAge:           600,
	})

	app.Use(c)
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := app.Handler()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ctx := &fasthttp.RequestCtx{}

		req := &fasthttp.Request{}
		req.Header.SetMethod(fiber.MethodOptions)
		req.SetRequestURI("/")
		req.Header.Set(fiber.HeaderOrigin, literal_0543)
		req.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodPost)
		req.Header.Set(fiber.HeaderAccessControlRequestHeaders, literal_2407)

		ctx.Init(req, nil, nil)

		for pb.Next() {
			h(ctx)
		}
	})
}

// go test -v -run=^$ -bench=Benchmark_CORS_NewHandlerPreflightWildcard -benchmem -count=4
func Benchmark_CORS_NewHandlerPreflightWildcard(b *testing.B) {
	app := fiber.New()
	c := New(Config{
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept},
		AllowCredentials: false,
		MaxAge:           600,
	})

	app.Use(c)
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := app.Handler()
	ctx := &fasthttp.RequestCtx{}

	req := &fasthttp.Request{}
	req.Header.SetMethod(fiber.MethodOptions)
	req.SetRequestURI("/")
	req.Header.Set(fiber.HeaderOrigin, literal_0543)
	req.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodPost)
	req.Header.Set(fiber.HeaderAccessControlRequestHeaders, literal_2407)

	ctx.Init(req, nil, nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h(ctx)
	}
}

// go test -v -run=^$ -bench=Benchmark_CORS_NewHandlerPreflightWildcardParallel -benchmem -count=4
func Benchmark_CORS_NewHandlerPreflightWildcardParallel(b *testing.B) {
	app := fiber.New()
	c := New(Config{
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodPut, fiber.MethodDelete},
		AllowHeaders:     []string{fiber.HeaderOrigin, fiber.HeaderContentType, fiber.HeaderAccept},
		AllowCredentials: false,
		MaxAge:           600,
	})

	app.Use(c)
	app.Use(func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := app.Handler()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ctx := &fasthttp.RequestCtx{}

		req := &fasthttp.Request{}
		req.Header.SetMethod(fiber.MethodOptions)
		req.SetRequestURI("/")
		req.Header.Set(fiber.HeaderOrigin, literal_0543)
		req.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodPost)
		req.Header.Set(fiber.HeaderAccessControlRequestHeaders, literal_2407)

		ctx.Init(req, nil, nil)

		for pb.Next() {
			h(ctx)
		}
	})
}

const literal_3627 = "http://localhost"

const literal_3958 = "GET, POST, HEAD, PUT, DELETE, PATCH"

const literal_6982 = "Vary header should be set"

const literal_3472 = "X-Request-ID"

const literal_8206 = "http://*.example.com"

const literal_9653 = "http://google.com"

const literal_0543 = "http://example.com"

const literal_7286 = "https://example.com"

const literal_5340 = "http://ccc.bbb.example.com"

const literal_4608 = "http://domain-1.com"

const literal_2783 = "http://example-1.com"

const literal_7395 = "https://example.com/"

const literal_3547 = "Access-Control-Allow-Origin header should be set"

const literal_5063 = "http://example-2.com"

const literal_9126 = "http://aaa.com"

const literal_8931 = "http://bbb.com"

const literal_2091 = "Access-Control-Request-Private-Network"

const literal_2895 = "Access-Control-Allow-Private-Network"

const literal_0684 = "The Access-Control-Allow-Private-Network header should be set to 'true' when AllowPrivateNetwork is enabled"

const literal_0786 = "The Access-Control-Allow-Private-Network header should not be present by default"

const literal_2407 = "Origin,Content-Type,Accept"
