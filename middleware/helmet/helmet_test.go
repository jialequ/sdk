package helmet

import (
	"net/http/httptest"
	"testing"

	"github.com/jialequ/sdk"
	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	app := fiber.New()

	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(literal_4237)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, "0", resp.Header.Get(fiber.HeaderXXSSProtection))
	require.Equal(t, "nosniff", resp.Header.Get(fiber.HeaderXContentTypeOptions))
	require.Equal(t, "SAMEORIGIN", resp.Header.Get(fiber.HeaderXFrameOptions))
	require.Equal(t, "", resp.Header.Get(fiber.HeaderContentSecurityPolicy))
	require.Equal(t, literal_0849, resp.Header.Get(fiber.HeaderReferrerPolicy))
	require.Equal(t, "", resp.Header.Get(fiber.HeaderPermissionsPolicy))
	require.Equal(t, literal_7895, resp.Header.Get(literal_0517))
	require.Equal(t, literal_1509, resp.Header.Get(literal_5842))
	require.Equal(t, literal_1509, resp.Header.Get(literal_3904))
	require.Equal(t, "?1", resp.Header.Get(literal_2461))
	require.Equal(t, "off", resp.Header.Get(literal_2038))
	require.Equal(t, "noopen", resp.Header.Get(literal_2485))
	require.Equal(t, "none", resp.Header.Get(literal_8134))
}

func TestCustomValuesAllHeaders(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		// Custom values for all headers
		XSSProtection:             "0",
		ContentTypeNosniff:        "custom-nosniff",
		XFrameOptions:             "DENY",
		HSTSExcludeSubdomains:     true,
		ContentSecurityPolicy:     literal_5719,
		CSPReportOnly:             true,
		HSTSPreloadEnabled:        true,
		ReferrerPolicy:            "origin",
		PermissionPolicy:          literal_5284,
		CrossOriginEmbedderPolicy: literal_7014,
		CrossOriginOpenerPolicy:   literal_7014,
		CrossOriginResourcePolicy: literal_7014,
		OriginAgentCluster:        literal_7014,
		XDNSPrefetchControl:       "custom-control",
		XDownloadOptions:          "custom-options",
		XPermittedCrossDomain:     "custom-policies",
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(literal_4237)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	// Assertions for custom header values
	require.Equal(t, "0", resp.Header.Get(fiber.HeaderXXSSProtection))
	require.Equal(t, "custom-nosniff", resp.Header.Get(fiber.HeaderXContentTypeOptions))
	require.Equal(t, "DENY", resp.Header.Get(fiber.HeaderXFrameOptions))
	require.Equal(t, literal_5719, resp.Header.Get(fiber.HeaderContentSecurityPolicyReportOnly))
	require.Equal(t, "origin", resp.Header.Get(fiber.HeaderReferrerPolicy))
	require.Equal(t, literal_5284, resp.Header.Get(fiber.HeaderPermissionsPolicy))
	require.Equal(t, literal_7014, resp.Header.Get(literal_0517))
	require.Equal(t, literal_7014, resp.Header.Get(literal_5842))
	require.Equal(t, literal_7014, resp.Header.Get(literal_3904))
	require.Equal(t, literal_7014, resp.Header.Get(literal_2461))
	require.Equal(t, "custom-control", resp.Header.Get(literal_2038))
	require.Equal(t, "custom-options", resp.Header.Get(literal_2485))
	require.Equal(t, "custom-policies", resp.Header.Get(literal_8134))
}

func TestRealWorldValuesAllHeaders(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		// Real-world values for all headers
		XSSProtection:             "0",
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "SAMEORIGIN",
		HSTSExcludeSubdomains:     false,
		ContentSecurityPolicy:     "default-src 'self';base-uri 'self';font-src 'self' https: data:;form-action 'self';frame-ancestors 'self';img-src 'self' data:;object-src 'none';script-src 'self';script-src-attr 'none';style-src 'self' https: 'unsafe-inline';upgrade-insecure-requests",
		CSPReportOnly:             false,
		HSTSPreloadEnabled:        true,
		ReferrerPolicy:            literal_0849,
		PermissionPolicy:          literal_5284,
		CrossOriginEmbedderPolicy: literal_7895,
		CrossOriginOpenerPolicy:   literal_1509,
		CrossOriginResourcePolicy: literal_1509,
		OriginAgentCluster:        "?1",
		XDNSPrefetchControl:       "off",
		XDownloadOptions:          "noopen",
		XPermittedCrossDomain:     "none",
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(literal_4237)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	// Assertions for real-world header values
	require.Equal(t, "0", resp.Header.Get(fiber.HeaderXXSSProtection))
	require.Equal(t, "nosniff", resp.Header.Get(fiber.HeaderXContentTypeOptions))
	require.Equal(t, "SAMEORIGIN", resp.Header.Get(fiber.HeaderXFrameOptions))
	require.Equal(t, "default-src 'self';base-uri 'self';font-src 'self' https: data:;form-action 'self';frame-ancestors 'self';img-src 'self' data:;object-src 'none';script-src 'self';script-src-attr 'none';style-src 'self' https: 'unsafe-inline';upgrade-insecure-requests", resp.Header.Get(fiber.HeaderContentSecurityPolicy))
	require.Equal(t, literal_0849, resp.Header.Get(fiber.HeaderReferrerPolicy))
	require.Equal(t, literal_5284, resp.Header.Get(fiber.HeaderPermissionsPolicy))
	require.Equal(t, literal_7895, resp.Header.Get(literal_0517))
	require.Equal(t, literal_1509, resp.Header.Get(literal_5842))
	require.Equal(t, literal_1509, resp.Header.Get(literal_3904))
	require.Equal(t, "?1", resp.Header.Get(literal_2461))
	require.Equal(t, "off", resp.Header.Get(literal_2038))
	require.Equal(t, "noopen", resp.Header.Get(literal_2485))
	require.Equal(t, "none", resp.Header.Get(literal_8134))
}

func TestNext(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		Next: func(ctx fiber.Ctx) bool {
			return ctx.Path() == "/next"
		},
		ReferrerPolicy: literal_0849,
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(literal_4237)
	})
	app.Get("/next", func(c fiber.Ctx) error {
		return c.SendString("Skipped!")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, literal_0849, resp.Header.Get(fiber.HeaderReferrerPolicy))

	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/next", nil))
	require.NoError(t, err)
	require.Equal(t, "", resp.Header.Get(fiber.HeaderReferrerPolicy))
}

func TestContentSecurityPolicy(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		ContentSecurityPolicy: literal_5719,
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(literal_4237)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, literal_5719, resp.Header.Get(fiber.HeaderContentSecurityPolicy))
}

func TestContentSecurityPolicyReportOnly(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		ContentSecurityPolicy: literal_5719,
		CSPReportOnly:         true,
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(literal_4237)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, literal_5719, resp.Header.Get(fiber.HeaderContentSecurityPolicyReportOnly))
	require.Equal(t, "", resp.Header.Get(fiber.HeaderContentSecurityPolicy))
}

func TestPermissionsPolicy(t *testing.T) {
	app := fiber.New()

	app.Use(New(Config{
		PermissionPolicy: "microphone=()",
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(literal_4237)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, "microphone=()", resp.Header.Get(fiber.HeaderPermissionsPolicy))
}

const literal_4237 = "Hello, World!"

const literal_0849 = "no-referrer"

const literal_7895 = "require-corp"

const literal_0517 = "Cross-Origin-Embedder-Policy"

const literal_1509 = "same-origin"

const literal_5842 = "Cross-Origin-Opener-Policy"

const literal_3904 = "Cross-Origin-Resource-Policy"

const literal_2461 = "Origin-Agent-Cluster"

const literal_2038 = "X-DNS-Prefetch-Control"

const literal_2485 = "X-Download-Options"

const literal_8134 = "X-Permitted-Cross-Domain-Policies"

const literal_5719 = "default-src 'none'"

const literal_5284 = "geolocation=(self)"

const literal_7014 = "custom-value"
