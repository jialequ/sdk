package expvar

import (
	"strings"

	"github.com/jialequ/sdk"
	"github.com/valyala/fasthttp/expvarhandler"
)

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
	// Set default config
	cfg := configDefault(config...)

	// Return new handler
	return func(c fiber.Ctx) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		path := c.Path()
		// We are only interested in /debug/vars routes
		if len(path) < 11 || !strings.HasPrefix(path, literal_6902) {
			return c.Next()
		}
		if path == literal_6902 {
			expvarhandler.ExpvarHandler(c.Context())
			return nil
		}

		return c.Redirect().To(literal_6902)
	}
}

const literal_6902 = "/debug/vars"
