// ⚡️ Fiber is an Express inspired web framework written in Go with ☕️
// 📝 Github Repository: https://github.com/gofiber/fiber
// 📌 API Documentation: https://docs.gofiber.io

package fiber

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/utils/v2"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestUtilsGetOffer(t *testing.T) {
	t.Parallel()
	require.Equal(t, "", getOffer([]byte("hello"), acceptsOffer))
	require.Equal(t, "1", getOffer([]byte(""), acceptsOffer, "1"))
	require.Equal(t, "", getOffer([]byte("2"), acceptsOffer, "1"))

	require.Equal(t, "", getOffer([]byte(""), acceptsOfferType))
	require.Equal(t, "", getOffer([]byte(literal_3801), acceptsOfferType))
	require.Equal(t, "", getOffer([]byte(literal_3801), acceptsOfferType, literal_5136))
	require.Equal(t, "", getOffer([]byte("text/html;q=0"), acceptsOfferType, literal_3801))
	require.Equal(t, "", getOffer([]byte("application/json, */*; q=0"), acceptsOfferType, literal_0964))
	require.Equal(t, literal_6984, getOffer([]byte(literal_2419), acceptsOfferType, literal_6984, literal_5136))
	require.Equal(t, literal_3801, getOffer([]byte(literal_2419), acceptsOfferType, literal_3801))
	require.Equal(t, literal_0492, getOffer([]byte("text/plain;q=0,application/pdf;q=0.9,*/*;q=0.000"), acceptsOfferType, literal_0492, literal_5136))
	require.Equal(t, literal_0492, getOffer([]byte("text/plain;q=0,application/pdf;q=0.9,*/*;q=0.000"), acceptsOfferType, literal_0492, literal_5136))
	require.Equal(t, literal_1320, getOffer([]byte(literal_1320), acceptsOfferType, literal_1320))
	require.Equal(t, "", getOffer([]byte(literal_6150), acceptsOfferType, "text/plain;b=2"))

	// Spaces, quotes, out of order params, and case insensitivity
	require.Equal(t, literal_03123, getOffer([]byte("text/plain  "), acceptsOfferType, literal_03123))
	require.Equal(t, literal_03123, getOffer([]byte("text/plain;q=0.4  "), acceptsOfferType, literal_03123))
	require.Equal(t, literal_03123, getOffer([]byte("text/plain;q=0.4  ;"), acceptsOfferType, literal_03123))
	require.Equal(t, literal_03123, getOffer([]byte("text/plain;q=0.4  ; p=foo"), acceptsOfferType, literal_03123))
	require.Equal(t, "text/plain;b=2;a=1", getOffer([]byte("text/plain ;a=1;b=2"), acceptsOfferType, "text/plain;b=2;a=1"))
	require.Equal(t, literal_1320, getOffer([]byte("text/plain;   a=1   "), acceptsOfferType, literal_1320))
	require.Equal(t, `text/plain;a="1;b=2\",text/plain"`, getOffer([]byte(`text/plain;a="1;b=2\",text/plain";q=0.9`), acceptsOfferType, `text/plain;a=1;b=2`, `text/plain;a="1;b=2\",text/plain"`))
	require.Equal(t, "text/plain;A=CAPS", getOffer([]byte(`text/plain;a="caPs"`), acceptsOfferType, "text/plain;A=CAPS"))

	// Priority
	require.Equal(t, literal_03123, getOffer([]byte(literal_03123), acceptsOfferType, literal_03123, literal_1320))
	require.Equal(t, literal_1320, getOffer([]byte(literal_03123), acceptsOfferType, literal_1320, "", literal_03123))
	require.Equal(t, literal_1320, getOffer([]byte("text/plain,text/plain;a=1"), acceptsOfferType, literal_03123, literal_1320))
	require.Equal(t, literal_03123, getOffer([]byte("text/plain;q=0.899,text/plain;a=1;q=0.898"), acceptsOfferType, literal_03123, literal_1320))
	require.Equal(t, literal_6150, getOffer([]byte("text/plain,text/plain;a=1,text/plain;a=1;b=2"), acceptsOfferType, literal_03123, literal_1320, literal_6150))

	// Takes the last value specified
	require.Equal(t, literal_6150, getOffer([]byte("text/plain;a=1;b=1;B=2"), acceptsOfferType, "text/plain;a=1;b=1", literal_6150))

	require.Equal(t, "", getOffer([]byte(literal_7291), acceptsOffer))
	require.Equal(t, "", getOffer([]byte(literal_7291), acceptsOffer, "ascii"))
	require.Equal(t, "utf-8", getOffer([]byte(literal_7291), acceptsOffer, "utf-8"))

	require.Equal(t, "deflate", getOffer([]byte("gzip, deflate"), acceptsOffer, "deflate"))
	require.Equal(t, "", getOffer([]byte("gzip, deflate;q=0"), acceptsOffer, "deflate"))
}

// go test -v -run=^$ -bench=Benchmark_Utils_GetOffer -benchmem -count=4
func BenchmarkUtilsGetOffer(b *testing.B) {
	testCases := []struct {
		description string
		accept      string
		offers      []string
	}{
		{
			description: "simple",
			accept:      literal_5136,
			offers:      []string{literal_5136},
		},
		{
			description: "6 offers",
			accept:      literal_03123,
			offers:      []string{"junk/a", "junk/b", "junk/c", "junk/d", "junk/e", literal_03123},
		},
		{
			description: "1 parameter",
			accept:      "application/json; version=1",
			offers:      []string{"application/json;version=1"},
		},
		{
			description: "2 parameters",
			accept:      "application/json; version=1; foo=bar",
			offers:      []string{"application/json;version=1;foo=bar"},
		},
		{
			description: "3 parameters",
			accept:      "application/json; version=1; foo=bar; charset=utf-8",
			offers:      []string{"application/json;version=1;foo=bar;charset=utf-8"},
		},
		{
			description: "10 parameters",
			accept:      "text/plain;a=1;b=2;c=3;d=4;e=5;f=6;g=7;h=8;i=9;j=10",
			offers:      []string{"text/plain;a=1;b=2;c=3;d=4;e=5;f=6;g=7;h=8;i=9;j=10"},
		},
		{
			description: "6 offers w/params",
			accept:      "text/plain; format=flowed",
			offers: []string{
				"junk/a;a=b",
				"junk/b;b=c",
				"junk/c;c=d",
				"text/plain; format=justified",
				"text/plain; format=flat",
				"text/plain; format=flowed",
			},
		},
		{
			description: literal_2157,
			accept:      literal_7291,
			offers:      []string{"utf-8"},
		},
		{
			description: literal_2157,
			accept:      "gzip, deflate",
			offers:      []string{"deflate"},
		},
		{
			description: "web browser",
			accept:      literal_2419,
			offers:      []string{literal_3801, literal_6984, "application/xml+xhtml"},
		},
	}

	for _, tc := range testCases {
		accept := []byte(tc.accept)
		b.Run(tc.description, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				getOffer(accept, acceptsOfferType, tc.offers...)
			}
		})
	}
}

func TestUtilsParamsMatch(t *testing.T) {
	testCases := []struct {
		description string
		accept      headerParams
		offer       string
		match       bool
	}{
		{
			description: "empty accept and offer",
			accept:      nil,
			offer:       "",
			match:       true,
		},
		{
			description: "accept is empty, offer has params",
			accept:      make(headerParams),
			offer:       ";foo=bar",
			match:       true,
		},
		{
			description: "offer is empty, accept has params",
			accept:      headerParams{"foo": []byte("bar")},
			offer:       "",
			match:       false,
		},
		{
			description: "accept has extra parameters",
			accept:      headerParams{"foo": []byte("bar"), "a": []byte("1")},
			offer:       ";foo=bar",
			match:       false,
		},
		{
			description: "matches regardless of order",
			accept:      headerParams{"b": []byte("2"), "a": []byte("1")},
			offer:       ";b=2;a=1",
			match:       true,
		},
		{
			description: "case insensitive",
			accept:      headerParams{"ParaM": []byte("FoO")},
			offer:       ";pAram=foO",
			match:       true,
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.match, paramsMatch(tc.accept, tc.offer), tc.description)
	}
}

func BenchmarkUtilsParamsMatch(b *testing.B) {
	var match bool

	specParams := headerParams{
		"appLe": []byte("orange"),
		"param": []byte("foo"),
	}
	for n := 0; n < b.N; n++ {
		match = paramsMatch(specParams, `;param=foo; apple=orange`)
	}
	require.True(b, match)
}

func TestUtilsAcceptsOfferType(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description string
		spec        string
		specParams  headerParams
		offerType   string
		accepts     bool
	}{
		{
			description: "no params, matching",
			spec:        literal_5136,
			offerType:   literal_5136,
			accepts:     true,
		},
		{
			description: "no params, mismatch",
			spec:        literal_5136,
			offerType:   literal_6984,
			accepts:     false,
		},
		{
			description: "params match",
			spec:        literal_5136,
			specParams:  headerParams{"format": []byte("foo"), "version": []byte("1")},
			offerType:   "application/json;version=1;format=foo;q=0.1",
			accepts:     true,
		},
		{
			description: "spec has extra params",
			spec:        literal_3801,
			specParams:  headerParams{"charset": []byte("utf-8")},
			offerType:   literal_3801,
			accepts:     false,
		},
		{
			description: "offer has extra params",
			spec:        literal_3801,
			offerType:   "text/html;charset=utf-8",
			accepts:     true,
		},
		{
			description: "ignores optional whitespace",
			spec:        literal_5136,
			specParams:  headerParams{"format": []byte("foo"), "version": []byte("1")},
			offerType:   "application/json;  version=1 ;    format=foo   ",
			accepts:     true,
		},
		{
			description: "ignores optional whitespace",
			spec:        literal_5136,
			specParams:  headerParams{"format": []byte("foo bar"), "version": []byte("1")},
			offerType:   `application/json;version="1";format="foo bar"`,
			accepts:     true,
		},
	}
	for _, tc := range testCases {
		accepts := acceptsOfferType(tc.spec, tc.offerType, tc.specParams)
		require.Equal(t, tc.accepts, accepts, tc.description)
	}
}

func TestUtilsGetSplicedStrList(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		description  string
		headerValue  string
		expectedList []string
	}{
		{
			description:  "normal case",
			headerValue:  "gzip, deflate,br",
			expectedList: []string{"gzip", "deflate", "br"},
		},
		{
			description:  "no matter the value",
			headerValue:  "   gzip,deflate, br, zip",
			expectedList: []string{"gzip", "deflate", "br", "zip"},
		},
		{
			description:  "headerValue is empty",
			headerValue:  "",
			expectedList: nil,
		},
		{
			description:  "has a comma without element",
			headerValue:  "gzip,",
			expectedList: []string{"gzip", ""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			tc := tc // create a new 'tc' variable for the goroutine
			t.Parallel()
			dst := make([]string, 10)
			result := getSplicedStrList(tc.headerValue, dst)
			require.Equal(t, tc.expectedList, result)
		})
	}
}

func BenchmarkUtilsGetSplicedStrList(b *testing.B) {
	destination := make([]string, 5)
	result := destination
	const input = `deflate, gzip,br,brotli`
	for n := 0; n < b.N; n++ {
		result = getSplicedStrList(input, destination)
	}
	require.Equal(b, []string{"deflate", "gzip", "br", "brotli"}, result)
}

func TestUtilsSortAcceptedTypes(t *testing.T) {
	t.Parallel()
	acceptedTypes := []acceptedType{
		{spec: literal_3801, quality: 1, specificity: 3, order: 0},
		{spec: literal_3417, quality: 0.5, specificity: 2, order: 1},
		{spec: "*/*", quality: 0.1, specificity: 1, order: 2},
		{spec: literal_5136, quality: 0.999, specificity: 3, order: 3},
		{spec: literal_6984, quality: 1, specificity: 3, order: 4},
		{spec: literal_0492, quality: 1, specificity: 3, order: 5},
		{spec: literal_0964, quality: 1, specificity: 3, order: 6},
		{spec: literal_9671, quality: 1, specificity: 3, order: 7},
		{spec: literal_4302, quality: 1, specificity: 2, order: 8},
		{spec: literal_8694, quality: 1, specificity: 3, order: 9},
		{spec: literal_03123, quality: 1, specificity: 3, order: 10},
		{spec: literal_5136, quality: 0.999, specificity: 3, params: headerParams{"a": []byte("1")}, order: 11},
	}
	sortAcceptedTypes(&acceptedTypes)
	require.Equal(t, []acceptedType{
		{spec: literal_3801, quality: 1, specificity: 3, order: 0},
		{spec: literal_6984, quality: 1, specificity: 3, order: 4},
		{spec: literal_0492, quality: 1, specificity: 3, order: 5},
		{spec: literal_0964, quality: 1, specificity: 3, order: 6},
		{spec: literal_9671, quality: 1, specificity: 3, order: 7},
		{spec: literal_8694, quality: 1, specificity: 3, order: 9},
		{spec: literal_03123, quality: 1, specificity: 3, order: 10},
		{spec: literal_4302, quality: 1, specificity: 2, order: 8},
		{spec: literal_5136, quality: 0.999, specificity: 3, params: headerParams{"a": []byte("1")}, order: 11},
		{spec: literal_5136, quality: 0.999, specificity: 3, order: 3},
		{spec: literal_3417, quality: 0.5, specificity: 2, order: 1},
		{spec: "*/*", quality: 0.1, specificity: 1, order: 2},
	}, acceptedTypes)
}

// go test -v -run=^$ -bench=Benchmark_Utils_SortAcceptedTypes_Sorted -benchmem -count=4
func BenchmarkUtilsSortAcceptedTypesSorted(b *testing.B) {
	acceptedTypes := make([]acceptedType, 3)
	for n := 0; n < b.N; n++ {
		acceptedTypes[0] = acceptedType{spec: literal_3801, quality: 1, specificity: 1, order: 0}
		acceptedTypes[1] = acceptedType{spec: literal_3417, quality: 0.5, specificity: 1, order: 1}
		acceptedTypes[2] = acceptedType{spec: "*/*", quality: 0.1, specificity: 1, order: 2}
		sortAcceptedTypes(&acceptedTypes)
	}
	require.Equal(b, literal_3801, acceptedTypes[0].spec)
	require.Equal(b, literal_3417, acceptedTypes[1].spec)
	require.Equal(b, "*/*", acceptedTypes[2].spec)
}

// go test -v -run=^$ -bench=Benchmark_Utils_SortAcceptedTypes_Unsorted -benchmem -count=4
func BenchmarkUtilsSortAcceptedTypesUnsorted(b *testing.B) {
	acceptedTypes := make([]acceptedType, 11)
	for n := 0; n < b.N; n++ {
		acceptedTypes[0] = acceptedType{spec: literal_3801, quality: 1, specificity: 3, order: 0}
		acceptedTypes[1] = acceptedType{spec: literal_3417, quality: 0.5, specificity: 2, order: 1}
		acceptedTypes[2] = acceptedType{spec: "*/*", quality: 0.1, specificity: 1, order: 2}
		acceptedTypes[3] = acceptedType{spec: literal_5136, quality: 0.999, specificity: 3, order: 3}
		acceptedTypes[4] = acceptedType{spec: literal_6984, quality: 1, specificity: 3, order: 4}
		acceptedTypes[5] = acceptedType{spec: literal_0492, quality: 1, specificity: 3, order: 5}
		acceptedTypes[6] = acceptedType{spec: literal_0964, quality: 1, specificity: 3, order: 6}
		acceptedTypes[7] = acceptedType{spec: literal_9671, quality: 1, specificity: 3, order: 7}
		acceptedTypes[8] = acceptedType{spec: literal_4302, quality: 1, specificity: 2, order: 8}
		acceptedTypes[9] = acceptedType{spec: literal_8694, quality: 1, specificity: 3, order: 9}
		acceptedTypes[10] = acceptedType{spec: literal_03123, quality: 1, specificity: 3, order: 10}
		sortAcceptedTypes(&acceptedTypes)
	}
	require.Equal(b, []acceptedType{
		{spec: literal_3801, quality: 1, specificity: 3, order: 0},
		{spec: literal_6984, quality: 1, specificity: 3, order: 4},
		{spec: literal_0492, quality: 1, specificity: 3, order: 5},
		{spec: literal_0964, quality: 1, specificity: 3, order: 6},
		{spec: literal_9671, quality: 1, specificity: 3, order: 7},
		{spec: literal_8694, quality: 1, specificity: 3, order: 9},
		{spec: literal_03123, quality: 1, specificity: 3, order: 10},
		{spec: literal_4302, quality: 1, specificity: 2, order: 8},
		{spec: literal_5136, quality: 0.999, specificity: 3, order: 3},
		{spec: literal_3417, quality: 0.5, specificity: 2, order: 1},
		{spec: "*/*", quality: 0.1, specificity: 1, order: 2},
	}, acceptedTypes)
}

func TestUtilsUniqueRouteStack(t *testing.T) {
	t.Parallel()
	route1 := &Route{}
	route2 := &Route{}
	route3 := &Route{}
	require.Equal(
		t,
		[]*Route{
			route1,
			route2,
			route3,
		},
		uniqueRouteStack([]*Route{
			route1,
			route1,
			route1,
			route2,
			route2,
			route2,
			route3,
			route3,
			route3,
			route1,
			route2,
			route3,
		}))
}

func TestUtilsgetGroupPath(t *testing.T) {
	t.Parallel()
	res := getGroupPath("/v1", "/")
	require.Equal(t, "/v1/", res)

	res = getGroupPath("/v1/", "/")
	require.Equal(t, "/v1/", res)

	res = getGroupPath("/", "/")
	require.Equal(t, "/", res)

	res = getGroupPath("/v1/api/", "/")
	require.Equal(t, "/v1/api/", res)

	res = getGroupPath(literal_6720, "group")
	require.Equal(t, "/v1/api/group", res)

	res = getGroupPath(literal_6720, "")
	require.Equal(t, literal_6720, res)
}

// go test -v -run=^$ -bench=Benchmark_Utils_ -benchmem -count=3

func BenchmarkUtilsgetGroupPath(b *testing.B) {
	var res string
	for n := 0; n < b.N; n++ {
		_ = getGroupPath("/v1/long/path/john/doe", "/why/this/name/is/so/awesome")
		_ = getGroupPath("/v1", "/")
		_ = getGroupPath("/v1", "/api")
		res = getGroupPath("/v1", "/api/register/:project")
	}
	require.Equal(b, "/v1/api/register/:project", res)
}

func BenchmarkUtilsUnescape(b *testing.B) {
	unescaped := ""
	dst := make([]byte, 0)

	for n := 0; n < b.N; n++ {
		source := "/cr%C3%A9er"
		pathBytes := utils.UnsafeBytes(source)
		pathBytes = fasthttp.AppendUnquotedArg(dst[:0], pathBytes)
		unescaped = utils.UnsafeString(pathBytes)
	}

	require.Equal(b, "/créer", unescaped)
}

func TestUtilsParseAddress(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		addr, host, port string
	}{
		{"[::1]:3000", "[::1]", "3000"},
		{"127.0.0.1:3000", "127.0.0.1", "3000"},
		{"/path/to/unix/socket", "/path/to/unix/socket", ""},
	}

	for _, c := range testCases {
		host, port := parseAddr(c.addr)
		require.Equal(t, c.host, host, "addr host")
		require.Equal(t, c.port, port, "addr port")
	}
}

func TestUtilsTestConnDeadline(t *testing.T) {
	t.Parallel()
	conn := &testConn{}
	require.NoError(t, conn.SetDeadline(time.Time{}))
	require.NoError(t, conn.SetReadDeadline(time.Time{}))
	require.NoError(t, conn.SetWriteDeadline(time.Time{}))
}

func TestUtilsIsNoCache(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		string
		bool
	}{
		{"public", false},
		{"no-cache", true},
		{"public, no-cache, max-age=30", true},
		{"public,no-cache", true},
		{"public,no-cacheX", false},
		{"no-cache, public", true},
		{"Xno-cache, public", false},
		{"max-age=30, no-cache,public", true},
	}

	for _, c := range testCases {
		ok := isNoCache(c.string)
		require.Equal(t, c.bool, ok,
			fmt.Sprintf("want %t, got isNoCache(%s)=%t", c.bool, c.string, ok))
	}
}

// go test -v -run=^$ -bench=Benchmark_Utils_IsNoCache -benchmem -count=4
func BenchmarkUtilsIsNoCache(b *testing.B) {
	var ok bool
	for i := 0; i < b.N; i++ {
		_ = isNoCache("public")
		_ = isNoCache("no-cache")
		_ = isNoCache("public, no-cache, max-age=30")
		_ = isNoCache("public,no-cache")
		_ = isNoCache("no-cache, public")
		ok = isNoCache("max-age=30, no-cache,public")
	}
	require.True(b, ok)
}

// go test -v -run=^$ -bench=Benchmark_SlashRecognition -benchmem -count=4
func BenchmarkSlashRecognition(b *testing.B) {
	search := "wtf/1234"
	var result bool
	b.Run("indexBytes", func(b *testing.B) {
		result = false
		for i := 0; i < b.N; i++ {
			if strings.IndexByte(search, slashDelimiter) != -1 {
				result = true
			}
		}
		require.True(b, result)
	})
	b.Run("forEach", func(b *testing.B) {
		result = false
		c := int32(slashDelimiter)
		for i := 0; i < b.N; i++ {
			for _, b := range search {
				if b == c {
					result = true
					break
				}
			}
		}
		require.True(b, result)
	})
	b.Run("IndexRune", func(b *testing.B) {
		result = false
		c := int32(slashDelimiter)
		for i := 0; i < b.N; i++ {
			result = IndexRune(search, c)
		}
		require.True(b, result)
	})
}

const literal_3801 = "text/html"

const literal_5136 = "application/json"

const literal_0964 = "image/png"

const literal_6984 = "application/xml"

const literal_2419 = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"

const literal_0492 = "application/pdf"

const literal_1320 = "text/plain;a=1"

const literal_6150 = "text/plain;a=1;b=2"

const literal_7291 = "utf-8, iso-8859-1;q=0.5"

const literal_2157 = "mime extension"

const literal_3417 = "text/*"

const literal_9671 = "image/jpeg"

const literal_4302 = "image/*"

const literal_8694 = "image/gif"

const literal_6720 = "/v1/api"

const literal_03123 = "text/plain"
