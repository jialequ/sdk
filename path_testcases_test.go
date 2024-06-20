// ⚡️ Fiber is an Express inspired web framework written in Go with ☕️
// 📝 Github Repository: https://github.com/gofiber/fiber
// 📌 API Documentation: https://docs.gofiber.io

package fiber

import (
	"strings"
)

type routeTestCase struct {
	url          string
	match        bool
	params       []string
	partialCheck bool
}

type routeCaseCollection struct {
	pattern   string
	testCases []routeTestCase
}

var (
	benchmarkCases []routeCaseCollection
	routeTestCases []routeCaseCollection
)

func init() {
	// smaller list for benchmark cases
	benchmarkCases = []routeCaseCollection{
		{
			pattern: "/api/v1/const",
			testCases: []routeTestCase{
				{url: "/api/v1/const", params: []string{}, match: true},
				{url: literal_8745, params: nil, match: false},
				{url: literal_3501, params: nil, match: false},
				{url: "/api/v1/something", params: nil, match: false},
			},
		},
		{
			pattern: "/api/:param/fixedEnd",
			testCases: []routeTestCase{
				{url: "/api/abc/fixedEnd", params: []string{"abc"}, match: true},
				{url: "/api/abc/def/fixedEnd", params: nil, match: false},
			},
		},
		{
			pattern: "/api/v1/:param/*",
			testCases: []routeTestCase{
				{url: literal_2675, params: []string{"entity", ""}, match: true},
				{url: "/api/v1/entity/", params: []string{"entity", ""}, match: true},
				{url: "/api/v1/entity/1", params: []string{"entity", "1"}, match: true},
				{url: literal_03126, params: nil, match: false},
				{url: literal_8207, params: nil, match: false},
				{url: literal_3501, params: nil, match: false},
			},
		},
	}

	// combine benchmark cases and other cases
	routeTestCases = benchmarkCases
	routeTestCases = append(
		routeTestCases,
		[]routeCaseCollection{
			{
				pattern: "/api/v1/:param/+",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: "/api/v1/entity/", params: nil, match: false},
					{url: "/api/v1/entity/1", params: []string{"entity", "1"}, match: true},
					{url: literal_03126, params: nil, match: false},
					{url: literal_8207, params: nil, match: false},
					{url: literal_3501, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param?",
				testCases: []routeTestCase{
					{url: literal_8745, params: []string{""}, match: true},
					{url: literal_3501, params: []string{""}, match: true},
					{url: "/api/v1/optional", params: []string{"optional"}, match: true},
					{url: literal_03126, params: nil, match: false},
					{url: literal_8207, params: nil, match: false},
					{url: "/api/xyz", params: nil, match: false},
				},
			},
			{
				pattern: `/v1/some/resource/name\:customVerb`,
				testCases: []routeTestCase{
					{url: "/v1/some/resource/name:customVerb", params: nil, match: true},
					{url: "/v1/some/resource/name:test", params: nil, match: false},
				},
			},
			{
				pattern: `/v1/some/resource/:name\:customVerb`,
				testCases: []routeTestCase{
					{url: "/v1/some/resource/test:customVerb", params: []string{"test"}, match: true},
					{url: "/v1/some/resource/test:test", params: nil, match: false},
				},
			},
			{
				pattern: `/v1/some/resource/name\\:customVerb?\?/:param/*`,
				testCases: []routeTestCase{
					{url: "/v1/some/resource/name:customVerb??/test/optionalWildCard/character", params: []string{"test", "optionalWildCard/character"}, match: true},
					{url: "/v1/some/resource/name:customVerb??/test", params: []string{"test", ""}, match: true},
				},
			},
			{
				pattern: "/api/v1/*",
				testCases: []routeTestCase{
					{url: literal_8745, params: []string{""}, match: true},
					{url: literal_3501, params: []string{""}, match: true},
					{url: literal_2675, params: []string{"entity"}, match: true},
					{url: "/api/v1/entity/1/2", params: []string{"entity/1/2"}, match: true},
					{url: "/api/v1/Entity/1/2", params: []string{"Entity/1/2"}, match: true},
					{url: literal_03126, params: nil, match: false},
					{url: literal_8207, params: nil, match: false},
					{url: "/api/abc", params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param",
				testCases: []routeTestCase{
					{url: literal_2675, params: []string{"entity"}, match: true},
					{url: "/api/v1/entity/8728382", params: nil, match: false},
					{url: literal_8745, params: nil, match: false},
					{url: literal_3501, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param-:param2",
				testCases: []routeTestCase{
					{url: "/api/v1/entity-entity2", params: []string{"entity", "entity2"}, match: true},
					{url: "/api/v1/entity/8728382", params: nil, match: false},
					{url: "/api/v1/entity-8728382", params: []string{"entity", "8728382"}, match: true},
					{url: literal_8745, params: nil, match: false},
					{url: literal_3501, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:filename.:extension",
				testCases: []routeTestCase{
					{url: "/api/v1/test.pdf", params: []string{"test", "pdf"}, match: true},
					{url: "/api/v1/test/pdf", params: nil, match: false},
					{url: "/api/v1/test-pdf", params: nil, match: false},
					{url: "/api/v1/test_pdf", params: nil, match: false},
					{url: literal_8745, params: nil, match: false},
					{url: literal_3501, params: nil, match: false},
				},
			},
			{
				pattern: "/shop/product/::filter/color::color/size::size",
				testCases: []routeTestCase{
					{url: "/shop/product/:test/color:blue/size:xs", params: []string{"test", "blue", "xs"}, match: true},
					{url: "/shop/product/test/color:blue/size:xs", params: nil, match: false},
				},
			},
			{
				pattern: "/::param?",
				testCases: []routeTestCase{
					{url: "/:hello", params: []string{"hello"}, match: true},
					{url: "/:", params: []string{""}, match: true},
					{url: "/", params: nil, match: false},
				},
			},
			// successive parameters, each take one character and the last parameter gets everything
			{
				pattern: "/test:sign:param",
				testCases: []routeTestCase{
					{url: "/test-abc", params: []string{"-", "abc"}, match: true},
					{url: "/test", params: nil, match: false},
				},
			},
			// optional parameters are not greedy
			{
				pattern: "/:param1:param2?:param3",
				testCases: []routeTestCase{
					{url: "/abbbc", params: []string{"a", "b", "bbc"}, match: true},
					// {url: "/ac", testCases: []string{"a", "", "c"}, match: true}, // : fix it
					{url: "/test", params: []string{"t", "e", "st"}, match: true},
				},
			},
			{
				pattern: "/test:optional?:mandatory",
				testCases: []routeTestCase{
					// {url: "/testo", testCases: []string{"", "o"}, match: true}, // : fix it
					{url: "/testoaaa", params: []string{"o", "aaa"}, match: true},
					{url: "/test", params: nil, match: false},
				},
			},
			{
				pattern: "/test:optional?:optional2?",
				testCases: []routeTestCase{
					{url: "/testo", params: []string{"o", ""}, match: true},
					{url: "/testoaaa", params: []string{"o", "aaa"}, match: true},
					{url: "/test", params: []string{"", ""}, match: true},
					{url: "/tes", params: nil, match: false},
				},
			},
			{
				pattern: "/foo:param?bar",
				testCases: []routeTestCase{
					{url: literal_4796, params: []string{"fasel"}, match: true},
					{url: literal_4132, params: []string{""}, match: true},
					{url: "/fooba", params: nil, match: false},
					{url: "/fobar", params: nil, match: false},
				},
			},
			{
				pattern: "/foo*bar",
				testCases: []routeTestCase{
					{url: literal_4796, params: []string{"fasel"}, match: true},
					{url: literal_4132, params: []string{""}, match: true},
					{url: "/", params: nil, match: false},
				},
			},
			{
				pattern: "/foo+bar",
				testCases: []routeTestCase{
					{url: literal_4796, params: []string{"fasel"}, match: true},
					{url: literal_4132, params: nil, match: false},
					{url: "/", params: nil, match: false},
				},
			},
			{
				pattern: "/a*cde*g/",
				testCases: []routeTestCase{
					{url: "/abbbcdefffg", params: []string{"bbb", "fff"}, match: true},
					{url: "/acdeg", params: []string{"", ""}, match: true},
					{url: "/", params: nil, match: false},
				},
			},
			{
				pattern: "/*v1*/proxy",
				testCases: []routeTestCase{
					{url: "/customer/v1/cart/proxy", params: []string{"customer/", "/cart"}, match: true},
					{url: "/v1/proxy", params: []string{"", ""}, match: true},
					{url: "/v1/", params: nil, match: false},
				},
			},
			// successive wildcard -> first wildcard is greedy
			{
				pattern: "/foo***bar",
				testCases: []routeTestCase{
					{url: "/foo*abar", params: []string{"*a", "", ""}, match: true},
					{url: "/foo*bar", params: []string{"*", "", ""}, match: true},
					{url: literal_4132, params: []string{"", "", ""}, match: true},
					{url: "/fooba", params: nil, match: false},
				},
			},
			// chars in front of an parameter
			{
				pattern: "/name::name",
				testCases: []routeTestCase{
					{url: "/name:john", params: []string{"john"}, match: true},
				},
			},
			{
				pattern: "/@:name",
				testCases: []routeTestCase{
					{url: "/@john", params: []string{"john"}, match: true},
				},
			},
			{
				pattern: "/-:name",
				testCases: []routeTestCase{
					{url: "/-john", params: []string{"john"}, match: true},
				},
			},
			{
				pattern: "/.:name",
				testCases: []routeTestCase{
					{url: "/.john", params: []string{"john"}, match: true},
				},
			},
			{
				pattern: "/api/v1/:param/abc/*",
				testCases: []routeTestCase{
					{url: "/api/v1/well/abc/wildcard", params: []string{"well", "wildcard"}, match: true},
					{url: "/api/v1/well/abc/", params: []string{"well", ""}, match: true},
					{url: "/api/v1/well/abc", params: []string{"well", ""}, match: true},
					{url: "/api/v1/well/ttt", params: nil, match: false},
				},
			},
			{
				pattern: "/api/:day/:month?/:year?",
				testCases: []routeTestCase{
					{url: literal_7204, params: []string{"1", "", ""}, match: true},
					{url: literal_8071, params: []string{"1", "", ""}, match: true},
					{url: "/api/1//", params: []string{"1", "", ""}, match: true},
					{url: "/api/1/-/", params: []string{"1", "-", ""}, match: true},
					{url: "/api/1-", params: []string{"1-", "", ""}, match: true},
					{url: "/api/1.", params: []string{"1.", "", ""}, match: true},
					{url: "/api/1/2", params: []string{"1", "2", ""}, match: true},
					{url: "/api/1/2/3", params: []string{"1", "2", "3"}, match: true},
					{url: "/api/", params: nil, match: false},
				},
			},
			{
				pattern: "/api/:day.:month?.:year?",
				testCases: []routeTestCase{
					{url: literal_7204, params: nil, match: false},
					{url: literal_8071, params: nil, match: false},
					{url: "/api/1.", params: nil, match: false},
					{url: "/api/1..", params: []string{"1", "", ""}, match: true},
					{url: "/api/1.2", params: nil, match: false},
					{url: "/api/1.2.", params: []string{"1", "2", ""}, match: true},
					{url: "/api/1.2.3", params: []string{"1", "2", "3"}, match: true},
					{url: "/api/", params: nil, match: false},
				},
			},
			{
				pattern: "/api/:day-:month?-:year?",
				testCases: []routeTestCase{
					{url: literal_7204, params: nil, match: false},
					{url: literal_8071, params: nil, match: false},
					{url: "/api/1-", params: nil, match: false},
					{url: "/api/1--", params: []string{"1", "", ""}, match: true},
					{url: "/api/1-/", params: nil, match: false},
					// {url: "/api/1-/-", testCases: nil, match: false}, // : fix this part
					{url: "/api/1-2", params: nil, match: false},
					{url: "/api/1-2-", params: []string{"1", "2", ""}, match: true},
					{url: "/api/1-2-3", params: []string{"1", "2", "3"}, match: true},
					{url: "/api/", params: nil, match: false},
				},
			},
			{
				pattern: "/api/*",
				testCases: []routeTestCase{
					{url: "/api/", params: []string{""}, match: true},
					{url: literal_4310, params: []string{"joker"}, match: true},
					{url: "/api", params: []string{""}, match: true},
					{url: literal_2675, params: []string{"v1/entity"}, match: true},
					{url: "/api2/v1/entity", params: nil, match: false},
					{url: "/api_ignore/v1/entity", params: nil, match: false},
				},
			},
			{
				pattern: "/partialCheck/foo/bar/:param",
				testCases: []routeTestCase{
					{url: "/partialCheck/foo/bar/test", params: []string{"test"}, match: true, partialCheck: true},
					{url: "/partialCheck/foo/bar/test/test2", params: []string{"test"}, match: true, partialCheck: true},
					{url: "/partialCheck/foo/bar", params: nil, match: false, partialCheck: true},
					{url: "/partiaFoo", params: nil, match: false, partialCheck: true},
				},
			},
			{
				pattern: "/",
				testCases: []routeTestCase{
					{url: "/api", params: nil, match: false},
					{url: "", params: []string{}, match: true},
					{url: "/", params: []string{}, match: true},
				},
			},
			{
				pattern: literal_2059,
				testCases: []routeTestCase{
					{url: literal_2059, params: []string{}, match: true},
					{url: literal_03127, params: nil, match: false},
					{url: literal_7460, params: nil, match: false},
					{url: literal_03128, params: nil, match: false},
				},
			},
			{
				pattern: "/config/*.json",
				testCases: []routeTestCase{
					{url: literal_2059, params: []string{"abc"}, match: true},
					{url: literal_7460, params: []string{"efg"}, match: true},
					{url: "/config/.json", params: []string{""}, match: true},
					{url: "/config/efg.csv", params: nil, match: false},
					{url: literal_03127, params: nil, match: false},
					{url: literal_03128, params: nil, match: false},
				},
			},
			{
				pattern: "/config/+.json",
				testCases: []routeTestCase{
					{url: literal_2059, params: []string{"abc"}, match: true},
					{url: "/config/.json", params: nil, match: false},
					{url: literal_7460, params: []string{"efg"}, match: true},
					{url: "/config/efg.csv", params: nil, match: false},
					{url: literal_03127, params: nil, match: false},
					{url: literal_03128, params: nil, match: false},
				},
			},
			{
				pattern: "/xyz",
				testCases: []routeTestCase{
					{url: "xyz", params: nil, match: false},
					{url: "xyz/", params: nil, match: false},
				},
			},
			{
				pattern: "/api/*/:param?",
				testCases: []routeTestCase{
					{url: "/api/", params: []string{"", ""}, match: true},
					{url: literal_4310, params: []string{"joker", ""}, match: true},
					{url: literal_0843, params: []string{"joker", "batman"}, match: true},
					{url: "/api/joker//batman", params: []string{"joker/", "batman"}, match: true},
					{url: literal_2901, params: []string{literal_03129, "robin"}, match: true},
					{url: literal_6713, params: []string{literal_03130, "1"}, match: true},
					{url: "/api/joker/batman/robin/1/", params: []string{"joker/batman/robin/1", ""}, match: true},
					{url: "/api/joker-batman/robin/1", params: []string{"joker-batman/robin", "1"}, match: true},
					{url: "/api/joker-batman-robin/1", params: []string{"joker-batman-robin", "1"}, match: true},
					{url: literal_2365, params: []string{"joker-batman-robin-1", ""}, match: true},
					{url: "/api", params: []string{"", ""}, match: true},
				},
			},
			{
				pattern: "/api/*/:param",
				testCases: []routeTestCase{
					{url: literal_3024, params: []string{"test", "abc"}, match: true},
					{url: literal_0843, params: []string{"joker", "batman"}, match: true},
					{url: literal_2901, params: []string{literal_03129, "robin"}, match: true},
					{url: literal_6713, params: []string{literal_03130, "1"}, match: true},
					{url: "/api/joker/batman-robin/1", params: []string{"joker/batman-robin", "1"}, match: true},
					{url: literal_2365, params: nil, match: false},
					{url: "/api", params: nil, match: false},
				},
			},
			{
				pattern: "/api/+/:param",
				testCases: []routeTestCase{
					{url: literal_3024, params: []string{"test", "abc"}, match: true},
					{url: literal_6713, params: []string{literal_03130, "1"}, match: true},
					{url: literal_4310, params: nil, match: false},
					{url: "/api", params: nil, match: false},
				},
			},
			{
				pattern: "/api/*/:param/:param2",
				testCases: []routeTestCase{
					{url: "/api/test/abc/1", params: []string{"test", "abc", "1"}, match: true},
					{url: literal_0843, params: nil, match: false},
					{url: "/api/joker/batman-robin/1", params: []string{"joker", "batman-robin", "1"}, match: true},
					{url: literal_2365, params: nil, match: false},
					{url: literal_3024, params: nil, match: false},
					{url: literal_2901, params: []string{"joker", "batman", "robin"}, match: true},
					{url: literal_6713, params: []string{literal_03129, "robin", "1"}, match: true},
					{url: "/api/joker/batman/robin/1/2", params: []string{literal_03130, "1", "2"}, match: true},
					{url: "/api", params: nil, match: false},
					{url: "/api/:test", params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param<int>",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_7854, params: []string{"8728382"}, match: true},
					{url: literal_5940, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param<bool>",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_7854, params: nil, match: false},
					{url: literal_5940, params: []string{"true"}, match: true},
				},
			},
			{
				pattern: "/api/v1/:param<float>",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_7854, params: []string{"8728382"}, match: true},
					{url: "/api/v1/8728382.5", params: []string{"8728382.5"}, match: true},
					{url: literal_5940, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param<alpha>",
				testCases: []routeTestCase{
					{url: literal_2675, params: []string{"entity"}, match: true},
					{url: "/api/v1/#!?", params: nil, match: false},
					{url: literal_7854, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param<guid>",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_7854, params: nil, match: false},
					{url: "/api/v1/f0fa66cc-d22e-445b-866d-1d76e776371d", params: []string{"f0fa66cc-d22e-445b-866d-1d76e776371d"}, match: true},
				},
			},
			{
				pattern: "/api/v1/:param<minLen>",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_7854, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param<minLen(5)>",
				testCases: []routeTestCase{
					{url: literal_2675, params: []string{"entity"}, match: true},
					{url: literal_3571, params: nil, match: false},
					{url: literal_7854, params: []string{"8728382"}, match: true},
					{url: literal_8974, params: nil, match: false},
					{url: literal_5176, params: []string{"12345"}, match: true},
				},
			},
			{
				pattern: "/api/v1/:param<maxLen(5)>",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_3571, params: []string{"ent"}, match: true},
					{url: literal_7854, params: nil, match: false},
					{url: literal_8974, params: []string{"123"}, match: true},
					{url: literal_5176, params: []string{"12345"}, match: true},
				},
			},
			{
				pattern: "/api/v1/:param<len(5)>",
				testCases: []routeTestCase{
					{url: literal_3571, params: nil, match: false},
					{url: literal_8974, params: nil, match: false},
					{url: literal_5176, params: []string{"12345"}, match: true},
				},
			},
			{
				pattern: "/api/v1/:param<betweenLen(1)>",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_3571, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param<betweenLen(2,5)>",
				testCases: []routeTestCase{
					{url: "/api/v1/e", params: nil, match: false},
					{url: "/api/v1/en", params: []string{"en"}, match: true},
					{url: literal_7854, params: nil, match: false},
					{url: literal_8974, params: []string{"123"}, match: true},
					{url: literal_5176, params: []string{"12345"}, match: true},
				},
			},
			{
				pattern: "/api/v1/:param<betweenLen(2,5)>",
				testCases: []routeTestCase{
					{url: "/api/v1/e", params: nil, match: false},
					{url: "/api/v1/en", params: []string{"en"}, match: true},
					{url: literal_7854, params: nil, match: false},
					{url: literal_8974, params: []string{"123"}, match: true},
					{url: literal_5176, params: []string{"12345"}, match: true},
				},
			},
			{
				pattern: "/api/v1/:param<min(5)>",
				testCases: []routeTestCase{
					{url: literal_3571, params: nil, match: false},
					{url: "/api/v1/1", params: nil, match: false},
					{url: literal_8219, params: []string{"5"}, match: true},
				},
			},
			{
				pattern: "/api/v1/:param<max(5)>",
				testCases: []routeTestCase{
					{url: literal_3571, params: nil, match: false},
					{url: "/api/v1/1", params: []string{"1"}, match: true},
					{url: literal_8219, params: []string{"5"}, match: true},
					{url: literal_8671, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param<range(5,10)>",
				testCases: []routeTestCase{
					{url: literal_3571, params: nil, match: false},
					{url: "/api/v1/9", params: []string{"9"}, match: true},
					{url: literal_8219, params: []string{"5"}, match: true},
					{url: literal_8671, params: nil, match: false},
				},
			},
			{
				pattern: `/api/v1/:param<datetime(2006\-01\-02)>`,
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_7854, params: nil, match: false},
					{url: "/api/v1/2005-11-01", params: []string{"2005-11-01"}, match: true},
				},
			},
			{
				pattern: "/api/v1/:param<regex(p([a-z]+)ch)>",
				testCases: []routeTestCase{
					{url: literal_3571, params: nil, match: false},
					{url: literal_8671, params: nil, match: false},
					{url: "/api/v1/peach", params: []string{"peach"}, match: true},
					{url: "/api/v1/p34ch", params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param<regex(^[a-z0-9]([a-z0-9-]{1,61}[a-z0-9])?$)>",
				testCases: []routeTestCase{
					{url: "/api/v1/12", params: nil, match: false},
					{url: "/api/v1/xy", params: nil, match: false},
					{url: "/api/v1/test", params: []string{"test"}, match: true},
					{url: literal_3501 + strings.Repeat("a", 64), params: nil, match: false},
				},
			},
			{
				pattern: `/api/v1/:param<regex(\d{4}-\d{2}-\d{2})}>`,
				testCases: []routeTestCase{
					{url: literal_3571, params: nil, match: false},
					{url: literal_8671, params: nil, match: false},
					{url: "/api/v1/2022-08-27", params: []string{"2022-08-27"}, match: true},
					{url: "/api/v1/2022/08-27", params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param<int;bool((>",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_7854, params: []string{"8728382"}, match: true},
					{url: literal_5940, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param<int;max(3000)>",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_7854, params: nil, match: false},
					{url: literal_8974, params: []string{"123"}, match: true},
					{url: literal_5940, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param<int;maxLen(10)>",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_2081, params: nil, match: false},
					{url: literal_8974, params: []string{"123"}, match: true},
					{url: literal_5940, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:param<int;range(10,30)>",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_2081, params: nil, match: false},
					{url: literal_4650, params: []string{"25"}, match: true},
					{url: literal_5940, params: nil, match: false},
				},
			},
			{
				pattern: `/api/v1/:param<int\;range(10,30)>`,
				testCases: []routeTestCase{
					{url: literal_2675, params: []string{"entity"}, match: true},
					{url: literal_2081, params: []string{"87283827683"}, match: true},
					{url: literal_4650, params: []string{"25"}, match: true},
					{url: literal_5940, params: []string{"true"}, match: true},
				},
			},
			{
				pattern: `/api/v1/:param<range(10\,30,1500)>`,
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_2081, params: nil, match: false},
					{url: literal_4650, params: []string{"25"}, match: true},
					{url: "/api/v1/1200", params: []string{"1200"}, match: true},
					{url: literal_5940, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:lang<len(2)>/videos/:page<range(100,1500)>",
				testCases: []routeTestCase{
					{url: "/api/v1/try/videos/200", params: nil, match: false},
					{url: "/api/v1/tr/videos/1800", params: nil, match: false},
					{url: "/api/v1/tr/videos/100", params: []string{"tr", "100"}, match: true},
					{url: "/api/v1/e/videos/10", params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:lang<len(2)>/:page<range(100,1500)>",
				testCases: []routeTestCase{
					{url: literal_7286, params: nil, match: false},
					{url: literal_0618, params: nil, match: false},
					{url: literal_5026, params: []string{"tr", "100"}, match: true},
					{url: literal_8510, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:lang/:page<range(100,1500)>",
				testCases: []routeTestCase{
					{url: literal_7286, params: []string{"try", "200"}, match: true},
					{url: literal_0618, params: nil, match: false},
					{url: literal_5026, params: []string{"tr", "100"}, match: true},
					{url: literal_8510, params: nil, match: false},
				},
			},
			{
				pattern: "/api/v1/:lang<len(2)>/:page",
				testCases: []routeTestCase{
					{url: literal_7286, params: nil, match: false},
					{url: literal_0618, params: []string{"tr", "1800"}, match: true},
					{url: literal_5026, params: []string{"tr", "100"}, match: true},
					{url: literal_8510, params: nil, match: false},
				},
			},
			{
				pattern: `/api/v1/:date<datetime(2006\-01\-02)>/:regex<regex(p([a-z]+)ch)>`,
				testCases: []routeTestCase{
					{url: "/api/v1/2005-11-01/a", params: nil, match: false},
					{url: "/api/v1/2005-1101/paach", params: nil, match: false},
					{url: "/api/v1/2005-11-01/peach", params: []string{"2005-11-01", "peach"}, match: true},
				},
			},
			{
				pattern: "/api/v1/:param<int>?",
				testCases: []routeTestCase{
					{url: literal_2675, params: nil, match: false},
					{url: literal_7854, params: []string{"8728382"}, match: true},
					{url: literal_5940, params: nil, match: false},
					{url: literal_3501, params: []string{""}, match: true},
				},
			},
		}...,
	)
}

const literal_8745 = "/api/v1"

const literal_3501 = "/api/v1/"

const literal_2675 = "/api/v1/entity"

const literal_8207 = "/api/v2"

const literal_4796 = "/foofaselbar"

const literal_4132 = "/foobar"

const literal_7204 = "/api/1"

const literal_8071 = "/api/1/"

const literal_4310 = "/api/joker"

const literal_2059 = "/config/abc.json"

const literal_0843 = "/api/joker/batman"

const literal_2901 = "/api/joker/batman/robin"

const literal_6713 = "/api/joker/batman/robin/1"

const literal_2365 = "/api/joker-batman-robin-1"

const literal_3024 = "/api/test/abc"

const literal_7854 = "/api/v1/8728382"

const literal_5940 = "/api/v1/true"

const literal_3571 = "/api/v1/ent"

const literal_8974 = "/api/v1/123"

const literal_5176 = "/api/v1/12345"

const literal_8219 = "/api/v1/5"

const literal_8671 = "/api/v1/15"

const literal_2081 = "/api/v1/87283827683"

const literal_4650 = "/api/v1/25"

const literal_7286 = "/api/v1/try/200"

const literal_0618 = "/api/v1/tr/1800"

const literal_5026 = "/api/v1/tr/100"

const literal_8510 = "/api/v1/e/10"

const literal_03126 = "/api/v"

const literal_03127 = "config/abc.json"

const literal_03128 = "/config"

const literal_03129 = "joker/batman"

const literal_03130 = "joker/batman/robin"
