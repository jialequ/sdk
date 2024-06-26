package csrf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -run -v Test_normalizeOrigin
func TestNormalizeOrigin(t *testing.T) {
	testCases := []struct {
		origin         string
		expectedValid  bool
		expectedOrigin string
	}{
		{literal_4307, true, literal_4307},                                       // Simple case should work.
		{"HTTP://EXAMPLE.COM", true, literal_4307},                               // Case should be normalized.
		{"http://example.com/", true, literal_4307},                              // Trailing slash should be removed.
		{literal_4719, true, literal_4719},                                       // Port should be preserved.
		{"http://example.com:3000/", true, literal_4719},                         // Trailing slash should be removed.
		{"http://", false, ""},                                                   // Invalid origin should not be accepted.
		{"file:///etc/passwd", false, ""},                                        // File scheme should not be accepted.
		{"https://*example.com", false, ""},                                      // Wildcard domain should not be accepted.
		{"http://*.example.com", false, ""},                                      // Wildcard subdomain should not be accepted.
		{"http://example.com/path", false, ""},                                   // Path should not be accepted.
		{"http://example.com?query=123", false, ""},                              // Query should not be accepted.
		{"http://example.com#fragment", false, ""},                               // Fragment should not be accepted.
		{"http://localhost", true, "http://localhost"},                           // Localhost should be accepted.
		{"http://127.0.0.1", true, "http://127.0.0.1"},                           // IPv4 address should be accepted.
		{"http://[::1]", true, "http://[::1]"},                                   // IPv6 address should be accepted.
		{literal_6531, true, literal_6531},                                       // IPv6 address with port should be accepted.
		{"http://[::1]:8080/", true, literal_6531},                               // IPv6 address with port and trailing slash should be accepted.
		{"http://[::1]:8080/path", false, ""},                                    // IPv6 address with port and path should not be accepted.
		{"http://[::1]:8080?query=123", false, ""},                               // IPv6 address with port and query should not be accepted.
		{"http://[::1]:8080#fragment", false, ""},                                // IPv6 address with port and fragment should not be accepted.
		{"http://[::1]:8080/path?query=123#fragment", false, ""},                 // IPv6 address with port, path, query, and fragment should not be accepted.
		{"http://[::1]:8080/path?query=123#fragment/", false, ""},                // IPv6 address with port, path, query, fragment, and trailing slash should not be accepted.
		{"http://[::1]:8080/path?query=123#fragment/invalid", false, ""},         // IPv6 address with port, path, query, fragment, trailing slash, and invalid segment should not be accepted.
		{"http://[::1]:8080/path?query=123#fragment/invalid/", false, ""},        // IPv6 address with port, path, query, fragment, trailing slash, and invalid segment with trailing slash should not be accepted.
		{"http://[::1]:8080/path?query=123#fragment/invalid/segment", false, ""}, // IPv6 address with port, path, query, fragment, trailing slash, and invalid segment with additional segment should not be accepted.
	}

	for _, tc := range testCases {
		valid, normalizedOrigin := normalizeOrigin(tc.origin)

		if valid != tc.expectedValid {
			t.Errorf("Expected origin '%s' to be valid: %v, but got: %v", tc.origin, tc.expectedValid, valid)
		}

		if normalizedOrigin != tc.expectedOrigin {
			t.Errorf("Expected normalized origin '%s' for origin '%s', but got: '%s'", tc.expectedOrigin, tc.origin, normalizedOrigin)
		}
	}
}

// go test -run -v TestSubdomainMatch
func TestSubdomainMatch(t *testing.T) {
	tests := []struct {
		name     string
		sub      subdomain
		origin   string
		expected bool
	}{
		{
			name:     "match with different scheme",
			sub:      subdomain{prefix: "http://api.", suffix: literal_2651},
			origin:   "https://api.service.example.com",
			expected: false,
		},
		{
			name:     "match with different scheme",
			sub:      subdomain{prefix: literal_7495, suffix: literal_2651},
			origin:   "http://api.service.example.com",
			expected: false,
		},
		{
			name:     "match with valid subdomain",
			sub:      subdomain{prefix: literal_7495, suffix: literal_2651},
			origin:   "https://api.service.example.com",
			expected: true,
		},
		{
			name:     "match with valid nested subdomain",
			sub:      subdomain{prefix: literal_7495, suffix: literal_2651},
			origin:   "https://1.2.api.service.example.com",
			expected: true,
		},

		{
			name:     "no match with invalid prefix",
			sub:      subdomain{prefix: "https://abc.", suffix: literal_2651},
			origin:   "https://service.example.com",
			expected: false,
		},
		{
			name:     "no match with invalid suffix",
			sub:      subdomain{prefix: literal_7495, suffix: literal_2651},
			origin:   "https://api.example.org",
			expected: false,
		},
		{
			name:     "no match with empty origin",
			sub:      subdomain{prefix: literal_7495, suffix: literal_2651},
			origin:   "",
			expected: false,
		},
		{
			name:     "partial match not considered a match",
			sub:      subdomain{prefix: "https://service.", suffix: literal_2651},
			origin:   "https://api.example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sub.match(tt.origin)
			assert.Equal(t, tt.expected, got, "subdomain.match()")
		})
	}
}

// go test -v -run=^$ -bench=Benchmark_CSRF_SubdomainMatch -benchmem -count=4
func BenchmarkCSRFSubdomainMatch(b *testing.B) {
	s := subdomain{
		prefix: "www",
		suffix: literal_2651,
	}

	o := "www.example.com"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		s.match(o)
	}
}

const literal_4307 = "http://example.com"

const literal_4719 = "http://example.com:3000"

const literal_6531 = "http://[::1]:8080"

const literal_2651 = ".example.com"

const literal_7495 = "https://"
