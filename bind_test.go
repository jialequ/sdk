//nolint:wrapcheck,tagliatelle // We must not wrap errors in tests
package fiber

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/jialequ/sdk/binder"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

const helloWorld = "hello world"

// go test -run Test_Bind_Query -v
func TestBindQuery(t *testing.T) {
	t.Parallel()
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Query struct {
		ID    int
		Name  string
		Hobby []string
	}
	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")
	c.Request().URI().SetQueryString(literal_7016)
	q := new(Query)
	require.NoError(t, c.Bind().Query(q))
	require.Len(t, q.Hobby, 2)

	c.Request().URI().SetQueryString(literal_3476)
	q = new(Query)
	require.NoError(t, c.Bind().Query(q))
	require.Len(t, q.Hobby, 2)

	c.Request().URI().SetQueryString("id=1&name=tom&hobby=scoccer&hobby=basketball,football")
	q = new(Query)
	require.NoError(t, c.Bind().Query(q))
	require.Len(t, q.Hobby, 3)

	empty := new(Query)
	c.Request().URI().SetQueryString("")
	require.NoError(t, c.Bind().Query(empty))
	require.Empty(t, empty.Hobby)

	type Query2 struct {
		Bool            bool
		ID              int
		Name            string
		Hobby           string
		FavouriteDrinks []string
		Empty           []string
		Alloc           []string
		No              []int64
	}

	c.Request().URI().SetQueryString("id=1&name=tom&hobby=basketball,football&favouriteDrinks=milo,coke,pepsi&alloc=&no=1")
	q2 := new(Query2)
	q2.Bool = true
	q2.Name = helloWorld
	require.NoError(t, c.Bind().Query(q2))
	require.Equal(t, "basketball,football", q2.Hobby)
	require.True(t, q2.Bool)
	require.Equal(t, "tom", q2.Name) // check value get overwritten
	require.Equal(t, []string{"milo", "coke", "pepsi"}, q2.FavouriteDrinks)
	var nilSlice []string
	require.Equal(t, nilSlice, q2.Empty)
	require.Equal(t, []string{""}, q2.Alloc)
	require.Equal(t, []int64{1}, q2.No)

	type RequiredQuery struct {
		Name string `query:"name,required"`
	}
	rq := new(RequiredQuery)
	c.Request().URI().SetQueryString("")
	require.Equal(t, literal_4602, c.Bind().Query(rq).Error())

	type ArrayQuery struct {
		Data []string
	}
	aq := new(ArrayQuery)
	c.Request().URI().SetQueryString("data[]=john&data[]=doe")
	require.NoError(t, c.Bind().Query(aq))
	require.Len(t, aq.Data, 2)
}

// go test -run Test_Bind_Query_Map -v
func TestBindQueryMap(t *testing.T) {
	t.Parallel()

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")
	c.Request().URI().SetQueryString(literal_7016)
	q := make(map[string][]string)
	require.NoError(t, c.Bind().Query(&q))
	require.Len(t, q["hobby"], 2)

	c.Request().URI().SetQueryString(literal_3476)
	q = make(map[string][]string)
	require.NoError(t, c.Bind().Query(&q))
	require.Len(t, q["hobby"], 2)

	c.Request().URI().SetQueryString("id=1&name=tom&hobby=scoccer&hobby=basketball,football")
	q = make(map[string][]string)
	require.NoError(t, c.Bind().Query(&q))
	require.Len(t, q["hobby"], 3)

	c.Request().URI().SetQueryString("id=1&name=tom&hobby=scoccer")
	qq := make(map[string]string)
	require.NoError(t, c.Bind().Query(&qq))
	require.Equal(t, "1", qq["id"])

	empty := make(map[string][]string)
	c.Request().URI().SetQueryString("")
	require.NoError(t, c.Bind().Query(&empty))
	require.Empty(t, empty["hobby"])

	em := make(map[string][]int)
	c.Request().URI().SetQueryString("")
	require.ErrorIs(t, c.Bind().Query(&em), binder.ErrMapNotConvertable)
}

// go test -run Test_Bind_Query_WithSetParserDecoder -v
func TestBindQueryWithSetParserDecoder(t *testing.T) {
	type NonRFCTime time.Time

	nonRFCConverter := func(value string) reflect.Value {
		if v, err := time.Parse(literal_1423, value); err == nil {
			return reflect.ValueOf(v)
		}
		return reflect.Value{}
	}

	nonRFCTime := binder.ParserType{
		Customtype: NonRFCTime{},
		Converter:  nonRFCConverter,
	}

	binder.SetParserDecoder(binder.ParserConfig{
		IgnoreUnknownKeys: true,
		ParserType:        []binder.ParserType{nonRFCTime},
		ZeroEmpty:         true,
		SetAliasTag:       "query",
	})

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type NonRFCTimeInput struct {
		Date  NonRFCTime `query:"date"`
		Title string     `query:"title"`
		Body  string     `query:"body"`
	}

	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")
	q := new(NonRFCTimeInput)

	c.Request().URI().SetQueryString("date=2021-04-10&title=CustomDateTest&Body=October")
	require.NoError(t, c.Bind().Query(q))
	require.Equal(t, "CustomDateTest", q.Title)
	date := fmt.Sprintf("%v", q.Date)
	require.Equal(t, literal_0285, date)
	require.Equal(t, "October", q.Body)

	c.Request().URI().SetQueryString("date=2021-04-10&title&Body=October")
	q = &NonRFCTimeInput{
		Title: literal_7460,
		Body:  literal_8365,
	}
	require.NoError(t, c.Bind().Query(q))
	require.Equal(t, "", q.Title)
}

// go test -run Test_Bind_Query_Schema -v
func TestBindQuerySchema(t *testing.T) {
	t.Parallel()
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Query1 struct {
		Name   string `query:"name,required"`
		Nested struct {
			Age int `query:"age"`
		} `query:"nested,required"`
	}
	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")
	c.Request().URI().SetQueryString("name=tom&nested.age=10")
	q := new(Query1)
	require.NoError(t, c.Bind().Query(q))

	c.Request().URI().SetQueryString("namex=tom&nested.age=10")
	q = new(Query1)
	require.Equal(t, literal_4602, c.Bind().Query(q).Error())

	c.Request().URI().SetQueryString("name=tom&nested.agex=10")
	q = new(Query1)
	require.NoError(t, c.Bind().Query(q))

	c.Request().URI().SetQueryString("name=tom&test.age=10")
	q = new(Query1)
	require.Equal(t, "nested is empty", c.Bind().Query(q).Error())

	type Query2 struct {
		Name   string `query:"name"`
		Nested struct {
			Age int `query:"age,required"`
		} `query:"nested"`
	}
	c.Request().URI().SetQueryString("name=tom&nested.age=10")
	q2 := new(Query2)
	require.NoError(t, c.Bind().Query(q2))

	c.Request().URI().SetQueryString("nested.age=10")
	q2 = new(Query2)
	require.NoError(t, c.Bind().Query(q2))

	c.Request().URI().SetQueryString("nested.agex=10")
	q2 = new(Query2)
	require.Equal(t, "nested.age is empty", c.Bind().Query(q2).Error())

	c.Request().URI().SetQueryString("nested.agex=10")
	q2 = new(Query2)
	require.Equal(t, "nested.age is empty", c.Bind().Query(q2).Error())

	type Node struct {
		Value int   `query:"val,required"`
		Next  *Node `query:"next,required"`
	}
	c.Request().URI().SetQueryString("val=1&next.val=3")
	n := new(Node)
	require.NoError(t, c.Bind().Query(n))
	require.Equal(t, 1, n.Value)
	require.Equal(t, 3, n.Next.Value)

	c.Request().URI().SetQueryString("next.val=2")
	n = new(Node)
	require.Equal(t, "val is empty", c.Bind().Query(n).Error())

	c.Request().URI().SetQueryString("val=3&next.value=2")
	n = new(Node)
	n.Next = new(Node)
	require.NoError(t, c.Bind().Query(n))
	require.Equal(t, 3, n.Value)
	require.Equal(t, 0, n.Next.Value)

	type Person struct {
		Name string `query:"name"`
		Age  int    `query:"age"`
	}

	type CollectionQuery struct {
		Data []Person `query:"data"`
	}

	c.Request().URI().SetQueryString("data[0][name]=john&data[0][age]=10&data[1][name]=doe&data[1][age]=12")
	cq := new(CollectionQuery)
	require.NoError(t, c.Bind().Query(cq))
	require.Len(t, cq.Data, 2)
	require.Equal(t, "john", cq.Data[0].Name)
	require.Equal(t, 10, cq.Data[0].Age)
	require.Equal(t, "doe", cq.Data[1].Name)
	require.Equal(t, 12, cq.Data[1].Age)

	c.Request().URI().SetQueryString("data.0.name=john&data.0.age=10&data.1.name=doe&data.1.age=12")
	cq = new(CollectionQuery)
	require.NoError(t, c.Bind().Query(cq))
	require.Len(t, cq.Data, 2)
	require.Equal(t, "john", cq.Data[0].Name)
	require.Equal(t, 10, cq.Data[0].Age)
	require.Equal(t, "doe", cq.Data[1].Name)
	require.Equal(t, 12, cq.Data[1].Age)
}

// go test -run Test_Bind_Header -v
func TestBindHeader(t *testing.T) {
	t.Parallel()
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Header struct {
		ID    int
		Name  string
		Hobby []string
	}
	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Request().Header.Add("id", "1")
	c.Request().Header.Add("Name", literal_5627)
	c.Request().Header.Add("Hobby", literal_5749)
	q := new(Header)
	require.NoError(t, c.Bind().Header(q))
	require.Len(t, q.Hobby, 2)

	c.Request().Header.Del("hobby")
	c.Request().Header.Add("Hobby", literal_8460)
	q = new(Header)
	require.NoError(t, c.Bind().Header(q))
	require.Len(t, q.Hobby, 3)

	empty := new(Header)
	c.Request().Header.Del("hobby")
	require.NoError(t, c.Bind().Query(empty))
	require.Empty(t, empty.Hobby)

	type Header2 struct {
		Bool            bool
		ID              int
		Name            string
		Hobby           string
		FavouriteDrinks []string
		Empty           []string
		Alloc           []string
		No              []int64
	}

	c.Request().Header.Add("id", "2")
	c.Request().Header.Add("Name", literal_8712)
	c.Request().Header.Del("hobby")
	c.Request().Header.Add("Hobby", literal_3506)
	c.Request().Header.Add("favouriteDrinks", literal_2149)
	c.Request().Header.Add("alloc", "")
	c.Request().Header.Add("no", "1")

	h2 := new(Header2)
	h2.Bool = true
	h2.Name = helloWorld
	require.NoError(t, c.Bind().Header(h2))
	require.Equal(t, literal_3506, h2.Hobby)
	require.True(t, h2.Bool)
	require.Equal(t, literal_8712, h2.Name) // check value get overwritten
	require.Equal(t, []string{"milo", "coke", "pepsi"}, h2.FavouriteDrinks)
	var nilSlice []string
	require.Equal(t, nilSlice, h2.Empty)
	require.Equal(t, []string{""}, h2.Alloc)
	require.Equal(t, []int64{1}, h2.No)

	type RequiredHeader struct {
		Name string `header:"name,required"`
	}
	rh := new(RequiredHeader)
	c.Request().Header.Del("name")
	require.Equal(t, literal_4602, c.Bind().Header(rh).Error())
}

// go test -run Test_Bind_Header_Map -v
func TestBindHeaderMap(t *testing.T) {
	t.Parallel()

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Request().Header.Add("id", "1")
	c.Request().Header.Add("Name", literal_5627)
	c.Request().Header.Add("Hobby", literal_5749)
	q := make(map[string][]string, 0)
	require.NoError(t, c.Bind().Header(&q))
	require.Len(t, q["Hobby"], 2)

	c.Request().Header.Del("hobby")
	c.Request().Header.Add("Hobby", literal_8460)
	q = make(map[string][]string, 0)
	require.NoError(t, c.Bind().Header(&q))
	require.Len(t, q["Hobby"], 3)

	empty := make(map[string][]string, 0)
	c.Request().Header.Del("hobby")
	require.NoError(t, c.Bind().Query(&empty))
	require.Empty(t, empty["Hobby"])
}

// go test -run Test_Bind_Header_WithSetParserDecoder -v
func TestBindHeaderWithSetParserDecoder(t *testing.T) {
	type NonRFCTime time.Time

	nonRFCConverter := func(value string) reflect.Value {
		if v, err := time.Parse(literal_1423, value); err == nil {
			return reflect.ValueOf(v)
		}
		return reflect.Value{}
	}

	nonRFCTime := binder.ParserType{
		Customtype: NonRFCTime{},
		Converter:  nonRFCConverter,
	}

	binder.SetParserDecoder(binder.ParserConfig{
		IgnoreUnknownKeys: true,
		ParserType:        []binder.ParserType{nonRFCTime},
		ZeroEmpty:         true,
		SetAliasTag:       "req",
	})

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type NonRFCTimeInput struct {
		Date  NonRFCTime `req:"date"`
		Title string     `req:"title"`
		Body  string     `req:"body"`
	}

	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")
	r := new(NonRFCTimeInput)

	c.Request().Header.Add("Date", "2021-04-10")
	c.Request().Header.Add("Title", "CustomDateTest")
	c.Request().Header.Add("Body", "October")

	require.NoError(t, c.Bind().Header(r))
	require.Equal(t, "CustomDateTest", r.Title)
	date := fmt.Sprintf("%v", r.Date)
	require.Equal(t, literal_0285, date)
	require.Equal(t, "October", r.Body)

	c.Request().Header.Add("Title", "")
	r = &NonRFCTimeInput{
		Title: literal_7460,
		Body:  literal_8365,
	}
	require.NoError(t, c.Bind().Header(r))
	require.Equal(t, "", r.Title)
}

// go test -run Test_Bind_Header_Schema -v
func TestBindHeaderSchema(t *testing.T) {
	t.Parallel()
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Header1 struct {
		Name   string `header:"Name,required"`
		Nested struct {
			Age int `header:"Age"`
		} `header:"Nested,required"`
	}
	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Request().Header.Add("Name", "tom")
	c.Request().Header.Add(literal_9374, "10")
	q := new(Header1)
	require.NoError(t, c.Bind().Header(q))

	c.Request().Header.Del("Name")
	q = new(Header1)
	require.Equal(t, "Name is empty", c.Bind().Header(q).Error())

	c.Request().Header.Add("Name", "tom")
	c.Request().Header.Del(literal_9374)
	c.Request().Header.Add(literal_7506, "10")
	q = new(Header1)
	require.NoError(t, c.Bind().Header(q))

	c.Request().Header.Del(literal_7506)
	q = new(Header1)
	require.Equal(t, "Nested is empty", c.Bind().Header(q).Error())

	c.Request().Header.Del(literal_7506)
	c.Request().Header.Del("Name")

	type Header2 struct {
		Name   string `header:"Name"`
		Nested struct {
			Age int `header:"age,required"`
		} `header:"Nested"`
	}

	c.Request().Header.Add("Name", "tom")
	c.Request().Header.Add(literal_9374, "10")

	h2 := new(Header2)
	require.NoError(t, c.Bind().Header(h2))

	c.Request().Header.Del("Name")
	h2 = new(Header2)
	require.NoError(t, c.Bind().Header(h2))

	c.Request().Header.Del("Name")
	c.Request().Header.Del(literal_9374)
	c.Request().Header.Add(literal_7506, "10")
	h2 = new(Header2)
	require.Equal(t, "Nested.age is empty", c.Bind().Header(h2).Error())

	type Node struct {
		Value int   `header:"Val,required"`
		Next  *Node `header:"Next,required"`
	}
	c.Request().Header.Add("Val", "1")
	c.Request().Header.Add(literal_4179, "3")
	n := new(Node)
	require.NoError(t, c.Bind().Header(n))
	require.Equal(t, 1, n.Value)
	require.Equal(t, 3, n.Next.Value)

	c.Request().Header.Del("Val")
	n = new(Node)
	require.Equal(t, "Val is empty", c.Bind().Header(n).Error())

	c.Request().Header.Add("Val", "3")
	c.Request().Header.Del(literal_4179)
	c.Request().Header.Add("Next.Value", "2")
	n = new(Node)
	n.Next = new(Node)
	require.NoError(t, c.Bind().Header(n))
	require.Equal(t, 3, n.Value)
	require.Equal(t, 0, n.Next.Value)
}

// go test -run Test_Bind_Resp_Header -v
func TestBindRespHeader(t *testing.T) {
	t.Parallel()
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Header struct {
		ID    int
		Name  string
		Hobby []string
	}
	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Response().Header.Add("id", "1")
	c.Response().Header.Add("Name", literal_5627)
	c.Response().Header.Add("Hobby", literal_5749)
	q := new(Header)
	require.NoError(t, c.Bind().RespHeader(q))
	require.Len(t, q.Hobby, 2)

	c.Response().Header.Del("hobby")
	c.Response().Header.Add("Hobby", literal_8460)
	q = new(Header)
	require.NoError(t, c.Bind().RespHeader(q))
	require.Len(t, q.Hobby, 3)

	empty := new(Header)
	c.Response().Header.Del("hobby")
	require.NoError(t, c.Bind().Query(empty))
	require.Empty(t, empty.Hobby)

	type Header2 struct {
		Bool            bool
		ID              int
		Name            string
		Hobby           string
		FavouriteDrinks []string
		Empty           []string
		Alloc           []string
		No              []int64
	}

	c.Response().Header.Add("id", "2")
	c.Response().Header.Add("Name", literal_8712)
	c.Response().Header.Del("hobby")
	c.Response().Header.Add("Hobby", literal_3506)
	c.Response().Header.Add("favouriteDrinks", literal_2149)
	c.Response().Header.Add("alloc", "")
	c.Response().Header.Add("no", "1")

	h2 := new(Header2)
	h2.Bool = true
	h2.Name = helloWorld
	require.NoError(t, c.Bind().RespHeader(h2))
	require.Equal(t, literal_3506, h2.Hobby)
	require.True(t, h2.Bool)
	require.Equal(t, literal_8712, h2.Name) // check value get overwritten
	require.Equal(t, []string{"milo", "coke", "pepsi"}, h2.FavouriteDrinks)
	var nilSlice []string
	require.Equal(t, nilSlice, h2.Empty)
	require.Equal(t, []string{""}, h2.Alloc)
	require.Equal(t, []int64{1}, h2.No)

	type RequiredHeader struct {
		Name string `respHeader:"name,required"`
	}
	rh := new(RequiredHeader)
	c.Response().Header.Del("name")
	require.Equal(t, literal_4602, c.Bind().RespHeader(rh).Error())
}

// go test -run Test_Bind_RespHeader_Map -v
func TestBindRespHeaderMap(t *testing.T) {
	t.Parallel()

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Response().Header.Add("id", "1")
	c.Response().Header.Add("Name", literal_5627)
	c.Response().Header.Add("Hobby", literal_5749)
	q := make(map[string][]string, 0)
	require.NoError(t, c.Bind().RespHeader(&q))
	require.Len(t, q["Hobby"], 2)

	c.Response().Header.Del("hobby")
	c.Response().Header.Add("Hobby", literal_8460)
	q = make(map[string][]string, 0)
	require.NoError(t, c.Bind().RespHeader(&q))
	require.Len(t, q["Hobby"], 3)

	empty := make(map[string][]string, 0)
	c.Response().Header.Del("hobby")
	require.NoError(t, c.Bind().Query(&empty))
	require.Empty(t, empty["Hobby"])
}

// go test -v  -run=^$ -bench=Benchmark_Bind_Query -benchmem -count=4
func BenchmarkBindQuery(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Query struct {
		ID    int
		Name  string
		Hobby []string
	}
	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")
	c.Request().URI().SetQueryString(literal_7016)
	q := new(Query)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = c.Bind().Query(q)
	}
	require.NoError(b, err)
}

// go test -v  -run=^$ -bench=Benchmark_Bind_Query_Map -benchmem -count=4
func BenchmarkBindQueryMap(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")
	c.Request().URI().SetQueryString(literal_7016)
	q := make(map[string][]string)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = c.Bind().Query(&q)
	}
	require.NoError(b, err)
}

// go test -v  -run=^$ -bench=Benchmark_Bind_Query_WithParseParam -benchmem -count=4
func BenchmarkBindQueryWithParseParam(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Person struct {
		Name string `query:"name"`
		Age  int    `query:"age"`
	}

	type CollectionQuery struct {
		Data []Person `query:"data"`
	}

	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")
	c.Request().URI().SetQueryString("data[0][name]=john&data[0][age]=10")
	cq := new(CollectionQuery)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = c.Bind().Query(cq)
	}

	require.NoError(b, err)
}

// go test -v  -run=^$ -bench=Benchmark_Bind_Query_Comma -benchmem -count=4
func BenchmarkBindQueryComma(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Query struct {
		ID    int
		Name  string
		Hobby []string
	}
	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")
	c.Request().URI().SetQueryString(literal_3476)
	q := new(Query)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = c.Bind().Query(q)
	}
	require.NoError(b, err)
}

// go test -v  -run=^$ -bench=Benchmark_Bind_Header -benchmem -count=4
func BenchmarkBindHeader(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type ReqHeader struct {
		ID    int
		Name  string
		Hobby []string
	}
	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Request().Header.Add("id", "1")
	c.Request().Header.Add("Name", literal_5627)
	c.Request().Header.Add("Hobby", literal_5749)

	q := new(ReqHeader)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = c.Bind().Header(q)
	}
	require.NoError(b, err)
}

// go test -v  -run=^$ -bench=Benchmark_Bind_Header_Map -benchmem -count=4
func BenchmarkBindHeaderMap(b *testing.B) {
	var err error
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Request().Header.Add("id", "1")
	c.Request().Header.Add("Name", literal_5627)
	c.Request().Header.Add("Hobby", literal_5749)

	q := make(map[string][]string)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = c.Bind().Header(&q)
	}
	require.NoError(b, err)
}

// go test -v  -run=^$ -bench=Benchmark_Bind_RespHeader -benchmem -count=4
func BenchmarkBindRespHeader(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type ReqHeader struct {
		ID    int
		Name  string
		Hobby []string
	}
	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Response().Header.Add("id", "1")
	c.Response().Header.Add("Name", literal_5627)
	c.Response().Header.Add("Hobby", literal_5749)

	q := new(ReqHeader)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = c.Bind().RespHeader(q)
	}
	require.NoError(b, err)
}

// go test -v  -run=^$ -bench=Benchmark_Bind_RespHeader_Map -benchmem -count=4
func BenchmarkBindRespHeaderMap(b *testing.B) {
	var err error
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Response().Header.Add("id", "1")
	c.Response().Header.Add("Name", literal_5627)
	c.Response().Header.Add("Hobby", literal_5749)

	q := make(map[string][]string)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = c.Bind().RespHeader(&q)
	}
	require.NoError(b, err)
}

// go test -run Test_Bind_Body
func TestBindBody(t *testing.T) {
	t.Parallel()
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Demo struct {
		Name string `json:"name" xml:"name" form:"name" query:"name"`
	}

	{
		var gzipJSON bytes.Buffer
		w := gzip.NewWriter(&gzipJSON)
		_, err := w.Write([]byte(`{"name":"john"}`))
		require.NoError(t, err)
		err = w.Close()
		require.NoError(t, err)

		c.Request().Header.SetContentType(MIMEApplicationJSON)
		c.Request().Header.Set(HeaderContentEncoding, "gzip")
		c.Request().SetBody(gzipJSON.Bytes())
		c.Request().Header.SetContentLength(len(gzipJSON.Bytes()))
		d := new(Demo)
		require.NoError(t, c.Bind().Body(d))
		require.Equal(t, "john", d.Name)
		c.Request().Header.Del(HeaderContentEncoding)
	}

	testDecodeParser := func(contentType, body string) {
		c.Request().Header.SetContentType(contentType)
		c.Request().SetBody([]byte(body))
		c.Request().Header.SetContentLength(len(body))
		d := new(Demo)
		require.NoError(t, c.Bind().Body(d))
		require.Equal(t, "john", d.Name)
	}

	testDecodeParser(MIMEApplicationJSON, `{"name":"john"}`)
	testDecodeParser(MIMEApplicationXML, `<Demo><name>john</name></Demo>`)
	testDecodeParser(MIMEApplicationForm, literal_3210)
	testDecodeParser(MIMEMultipartForm+`;boundary="b"`, "--b\r\nContent-Disposition: form-data; name=\"name\"\r\n\r\njohn\r\n--b--")

	testDecodeParserError := func(contentType, body string) {
		c.Request().Header.SetContentType(contentType)
		c.Request().SetBody([]byte(body))
		c.Request().Header.SetContentLength(len(body))
		require.Error(t, c.Bind().Body(nil))
	}

	testDecodeParserError("invalid-content-type", "")
	testDecodeParserError(MIMEMultipartForm+`;boundary="b"`, "--b")

	type CollectionQuery struct {
		Data []Demo `query:"data"`
	}

	c.Request().Reset()
	c.Request().Header.SetContentType(MIMEApplicationForm)
	c.Request().SetBody([]byte("data[0][name]=john&data[1][name]=doe"))
	c.Request().Header.SetContentLength(len(c.Body()))
	cq := new(CollectionQuery)
	require.NoError(t, c.Bind().Body(cq))
	require.Len(t, cq.Data, 2)
	require.Equal(t, "john", cq.Data[0].Name)
	require.Equal(t, "doe", cq.Data[1].Name)

	c.Request().Reset()
	c.Request().Header.SetContentType(MIMEApplicationForm)
	c.Request().SetBody([]byte("data.0.name=john&data.1.name=doe"))
	c.Request().Header.SetContentLength(len(c.Body()))
	cq = new(CollectionQuery)
	require.NoError(t, c.Bind().Body(cq))
	require.Len(t, cq.Data, 2)
	require.Equal(t, "john", cq.Data[0].Name)
	require.Equal(t, "doe", cq.Data[1].Name)
}

// go test -run Test_Bind_Body_WithSetParserDecoder
func TestBindBodyWithSetParserDecoder(t *testing.T) {
	type CustomTime time.Time

	timeConverter := func(value string) reflect.Value {
		if v, err := time.Parse(literal_1423, value); err == nil {
			return reflect.ValueOf(v)
		}
		return reflect.Value{}
	}

	customTime := binder.ParserType{
		Customtype: CustomTime{},
		Converter:  timeConverter,
	}

	binder.SetParserDecoder(binder.ParserConfig{
		IgnoreUnknownKeys: true,
		ParserType:        []binder.ParserType{customTime},
		ZeroEmpty:         true,
		SetAliasTag:       "form",
	})

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Demo struct {
		Date  CustomTime `form:"date"`
		Title string     `form:"title"`
		Body  string     `form:"body"`
	}

	testDecodeParser := func(contentType, body string) {
		c.Request().Header.SetContentType(contentType)
		c.Request().SetBody([]byte(body))
		c.Request().Header.SetContentLength(len(body))
		d := Demo{
			Title: literal_7460,
			Body:  literal_8365,
		}
		require.NoError(t, c.Bind().Body(&d))
		date := fmt.Sprintf("%v", d.Date)
		require.Equal(t, "{0 63743587200 <nil>}", date)
		require.Equal(t, "", d.Title)
		require.Equal(t, "New Body", d.Body)
	}

	testDecodeParser(MIMEApplicationForm, "date=2020-12-15&title=&body=New Body")
	testDecodeParser(MIMEMultipartForm+`; boundary="b"`, "--b\r\nContent-Disposition: form-data; name=\"date\"\r\n\r\n2020-12-15\r\n--b\r\nContent-Disposition: form-data; name=\"title\"\r\n\r\n\r\n--b\r\nContent-Disposition: form-data; name=\"body\"\r\n\r\nNew Body\r\n--b--")
}

// go test -v -run=^$ -bench=Benchmark_Bind_Body_JSON -benchmem -count=4
func BenchmarkBindBodyJSON(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Demo struct {
		Name string `json:"name"`
	}
	body := []byte(`{"name":"john"}`)
	c.Request().SetBody(body)
	c.Request().Header.SetContentType(MIMEApplicationJSON)
	c.Request().Header.SetContentLength(len(body))
	d := new(Demo)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = c.Bind().Body(d)
	}
	require.NoError(b, err)
	require.Equal(b, "john", d.Name)
}

// go test -v -run=^$ -bench=Benchmark_Bind_Body_XML -benchmem -count=4
func BenchmarkBindBodyXML(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Demo struct {
		Name string `xml:"name"`
	}
	body := []byte("<Demo><name>john</name></Demo>")
	c.Request().SetBody(body)
	c.Request().Header.SetContentType(MIMEApplicationXML)
	c.Request().Header.SetContentLength(len(body))
	d := new(Demo)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = c.Bind().Body(d)
	}
	require.NoError(b, err)
	require.Equal(b, "john", d.Name)
}

// go test -v -run=^$ -bench=Benchmark_Bind_Body_Form -benchmem -count=4
func BenchmarkBindBodyForm(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Demo struct {
		Name string `form:"name"`
	}
	body := []byte(literal_3210)
	c.Request().SetBody(body)
	c.Request().Header.SetContentType(MIMEApplicationForm)
	c.Request().Header.SetContentLength(len(body))
	d := new(Demo)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = c.Bind().Body(d)
	}
	require.NoError(b, err)
	require.Equal(b, "john", d.Name)
}

// go test -v -run=^$ -bench=Benchmark_Bind_Body_MultipartForm -benchmem -count=4
func BenchmarkBindBodyMultipartForm(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Demo struct {
		Name string `form:"name"`
	}

	body := []byte("--b\r\nContent-Disposition: form-data; name=\"name\"\r\n\r\njohn\r\n--b--")
	c.Request().SetBody(body)
	c.Request().Header.SetContentType(MIMEMultipartForm + `;boundary="b"`)
	c.Request().Header.SetContentLength(len(body))
	d := new(Demo)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = c.Bind().Body(d)
	}
	require.NoError(b, err)
	require.Equal(b, "john", d.Name)
}

// go test -v -run=^$ -bench=Benchmark_Bind_Body_Form_Map -benchmem -count=4
func BenchmarkBindBodyFormMap(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	body := []byte(literal_3210)
	c.Request().SetBody(body)
	c.Request().Header.SetContentType(MIMEApplicationForm)
	c.Request().Header.SetContentLength(len(body))
	d := make(map[string]string)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = c.Bind().Body(&d)
	}
	require.NoError(b, err)
	require.Equal(b, "john", d["name"])
}

// go test -run Test_Bind_URI
func TestBindURI(t *testing.T) {
	t.Parallel()

	app := New()
	app.Get("/test1/userId/role/:roleId", func(c Ctx) error {
		type Demo struct {
			UserID uint `uri:"userId"`
			RoleID uint `uri:"roleId"`
		}
		d := new(Demo)
		if err := c.Bind().URI(d); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, uint(111), d.UserID)
		require.Equal(t, uint(222), d.RoleID)
		return nil
	})
	_, err := app.Test(httptest.NewRequest(MethodGet, "/test1/111/role/222", nil))
	require.NoError(t, err)
	_, err = app.Test(httptest.NewRequest(MethodGet, "/test2/111/role/222", nil))
	require.NoError(t, err)
}

// go test -run Test_Bind_URI_Map
func TestBindURIMap(t *testing.T) {
	t.Parallel()

	app := New()
	app.Get("/test1/userId/role/:roleId", func(c Ctx) error {
		d := make(map[string]string)

		if err := c.Bind().URI(&d); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, uint(111), d["userId"])
		require.Equal(t, uint(222), d["roleId"])
		return nil
	})
	_, err := app.Test(httptest.NewRequest(MethodGet, "/test1/111/role/222", nil))
	require.NoError(t, err)
	_, err = app.Test(httptest.NewRequest(MethodGet, "/test2/111/role/222", nil))
	require.NoError(t, err)
}

// go test -v -run=^$ -bench=Benchmark_Bind_URI -benchmem -count=4
func BenchmarkBindURI(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{}).(*DefaultCtx) //nolint:errcheck, forcetypeassert // not needed

	c.route = &Route{
		Params: []string{
			"param1", "param2", "param3", "param4",
		},
	}
	c.values = [maxParams]string{
		"john", "doe", "is", "awesome",
	}

	var res struct {
		Param1 string `uri:"param1"`
		Param2 string `uri:"param2"`
		Param3 string `uri:"param3"`
		Param4 string `uri:"param4"`
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = c.Bind().URI(&res)
	}

	require.NoError(b, err)
	require.Equal(b, "john", res.Param1)
	require.Equal(b, "doe", res.Param2)
	require.Equal(b, "is", res.Param3)
	require.Equal(b, "awesome", res.Param4)
}

// go test -v -run=^$ -bench=Benchmark_Bind_URI_Map -benchmem -count=4
func BenchmarkBindURIMap(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{}).(*DefaultCtx) //nolint:errcheck, forcetypeassert // not needed

	c.route = &Route{
		Params: []string{
			"param1", "param2", "param3", "param4",
		},
	}
	c.values = [maxParams]string{
		"john", "doe", "is", "awesome",
	}

	res := make(map[string]string)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = c.Bind().URI(&res)
	}

	require.NoError(b, err)
	require.Equal(b, "john", res["param1"])
	require.Equal(b, "doe", res["param2"])
	require.Equal(b, "is", res["param3"])
	require.Equal(b, "awesome", res["param4"])
}

// go test -run Test_Bind_Cookie -v
func TestBindCookie(t *testing.T) {
	t.Parallel()

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Cookie struct {
		ID    int
		Name  string
		Hobby []string
	}
	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Request().Header.SetCookie("id", "1")
	c.Request().Header.SetCookie("Name", literal_5627)
	c.Request().Header.SetCookie("Hobby", literal_5749)
	q := new(Cookie)
	require.NoError(t, c.Bind().Cookie(q))
	require.Len(t, q.Hobby, 2)

	c.Request().Header.DelCookie("hobby")
	c.Request().Header.SetCookie("Hobby", literal_8460)
	q = new(Cookie)
	require.NoError(t, c.Bind().Cookie(q))
	require.Len(t, q.Hobby, 3)

	empty := new(Cookie)
	c.Request().Header.DelCookie("hobby")
	require.NoError(t, c.Bind().Query(empty))
	require.Empty(t, empty.Hobby)

	type Cookie2 struct {
		Bool            bool
		ID              int
		Name            string
		Hobby           string
		FavouriteDrinks []string
		Empty           []string
		Alloc           []string
		No              []int64
	}

	c.Request().Header.SetCookie("id", "2")
	c.Request().Header.SetCookie("Name", literal_8712)
	c.Request().Header.DelCookie("hobby")
	c.Request().Header.SetCookie("Hobby", literal_3506)
	c.Request().Header.SetCookie("favouriteDrinks", literal_2149)
	c.Request().Header.SetCookie("alloc", "")
	c.Request().Header.SetCookie("no", "1")

	h2 := new(Cookie2)
	h2.Bool = true
	h2.Name = helloWorld
	require.NoError(t, c.Bind().Cookie(h2))
	require.Equal(t, literal_3506, h2.Hobby)
	require.True(t, h2.Bool)
	require.Equal(t, literal_8712, h2.Name) // check value get overwritten
	require.Equal(t, []string{"milo", "coke", "pepsi"}, h2.FavouriteDrinks)
	var nilSlice []string
	require.Equal(t, nilSlice, h2.Empty)
	require.Equal(t, []string{""}, h2.Alloc)
	require.Equal(t, []int64{1}, h2.No)

	type RequiredCookie struct {
		Name string `cookie:"name,required"`
	}
	rh := new(RequiredCookie)
	c.Request().Header.DelCookie("name")
	require.Equal(t, literal_4602, c.Bind().Cookie(rh).Error())
}

// go test -run Test_Bind_Cookie_Map -v
func TestBindCookieMap(t *testing.T) {
	t.Parallel()

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Request().Header.SetCookie("id", "1")
	c.Request().Header.SetCookie("Name", literal_5627)
	c.Request().Header.SetCookie("Hobby", literal_5749)
	q := make(map[string][]string)
	require.NoError(t, c.Bind().Cookie(&q))
	require.Len(t, q["Hobby"], 2)

	c.Request().Header.DelCookie("hobby")
	c.Request().Header.SetCookie("Hobby", literal_8460)
	q = make(map[string][]string)
	require.NoError(t, c.Bind().Cookie(&q))
	require.Len(t, q["Hobby"], 3)

	empty := make(map[string][]string)
	c.Request().Header.DelCookie("hobby")
	require.NoError(t, c.Bind().Query(&empty))
	require.Empty(t, empty["Hobby"])
}

// go test -run Test_Bind_Cookie_WithSetParserDecoder -v
func TestBindCookieWithSetParserDecoder(t *testing.T) {
	type NonRFCTime time.Time

	nonRFCConverter := func(value string) reflect.Value {
		if v, err := time.Parse(literal_1423, value); err == nil {
			return reflect.ValueOf(v)
		}
		return reflect.Value{}
	}

	nonRFCTime := binder.ParserType{
		Customtype: NonRFCTime{},
		Converter:  nonRFCConverter,
	}

	binder.SetParserDecoder(binder.ParserConfig{
		IgnoreUnknownKeys: true,
		ParserType:        []binder.ParserType{nonRFCTime},
		ZeroEmpty:         true,
		SetAliasTag:       "cerez",
	})

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type NonRFCTimeInput struct {
		Date  NonRFCTime `cerez:"date"`
		Title string     `cerez:"title"`
		Body  string     `cerez:"body"`
	}

	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")
	r := new(NonRFCTimeInput)

	c.Request().Header.SetCookie("Date", "2021-04-10")
	c.Request().Header.SetCookie("Title", "CustomDateTest")
	c.Request().Header.SetCookie("Body", "October")

	require.NoError(t, c.Bind().Cookie(r))
	require.Equal(t, "CustomDateTest", r.Title)
	date := fmt.Sprintf("%v", r.Date)
	require.Equal(t, literal_0285, date)
	require.Equal(t, "October", r.Body)

	c.Request().Header.SetCookie("Title", "")
	r = &NonRFCTimeInput{
		Title: literal_7460,
		Body:  literal_8365,
	}
	require.NoError(t, c.Bind().Cookie(r))
	require.Equal(t, "", r.Title)
}

// go test -run Test_Bind_Cookie_Schema -v
func TestBindCookieSchema(t *testing.T) {
	t.Parallel()

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Cookie1 struct {
		Name   string `cookie:"Name,required"`
		Nested struct {
			Age int `cookie:"Age"`
		} `cookie:"Nested,required"`
	}
	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Request().Header.SetCookie("Name", "tom")
	c.Request().Header.SetCookie(literal_9374, "10")
	q := new(Cookie1)
	require.NoError(t, c.Bind().Cookie(q))

	c.Request().Header.DelCookie("Name")
	q = new(Cookie1)
	require.Equal(t, "Name is empty", c.Bind().Cookie(q).Error())

	c.Request().Header.SetCookie("Name", "tom")
	c.Request().Header.DelCookie(literal_9374)
	c.Request().Header.SetCookie(literal_7506, "10")
	q = new(Cookie1)
	require.NoError(t, c.Bind().Cookie(q))

	c.Request().Header.DelCookie(literal_7506)
	q = new(Cookie1)
	require.Equal(t, "Nested is empty", c.Bind().Cookie(q).Error())

	c.Request().Header.DelCookie(literal_7506)
	c.Request().Header.DelCookie("Name")

	type Cookie2 struct {
		Name   string `cookie:"Name"`
		Nested struct {
			Age int `cookie:"Age,required"`
		} `cookie:"Nested"`
	}

	c.Request().Header.SetCookie("Name", "tom")
	c.Request().Header.SetCookie(literal_9374, "10")

	h2 := new(Cookie2)
	require.NoError(t, c.Bind().Cookie(h2))

	c.Request().Header.DelCookie("Name")
	h2 = new(Cookie2)
	require.NoError(t, c.Bind().Cookie(h2))

	c.Request().Header.DelCookie("Name")
	c.Request().Header.DelCookie(literal_9374)
	c.Request().Header.SetCookie(literal_7506, "10")
	h2 = new(Cookie2)
	require.Equal(t, "Nested.Age is empty", c.Bind().Cookie(h2).Error())

	type Node struct {
		Value int   `cookie:"Val,required"`
		Next  *Node `cookie:"Next,required"`
	}
	c.Request().Header.SetCookie("Val", "1")
	c.Request().Header.SetCookie(literal_4179, "3")
	n := new(Node)
	require.NoError(t, c.Bind().Cookie(n))
	require.Equal(t, 1, n.Value)
	require.Equal(t, 3, n.Next.Value)

	c.Request().Header.DelCookie("Val")
	n = new(Node)
	require.Equal(t, "Val is empty", c.Bind().Cookie(n).Error())

	c.Request().Header.SetCookie("Val", "3")
	c.Request().Header.DelCookie(literal_4179)
	c.Request().Header.SetCookie("Next.Value", "2")
	n = new(Node)
	n.Next = new(Node)
	require.NoError(t, c.Bind().Cookie(n))
	require.Equal(t, 3, n.Value)
	require.Equal(t, 0, n.Next.Value)
}

// go test -v  -run=^$ -bench=Benchmark_Bind_Cookie -benchmem -count=4
func BenchmarkBindCookie(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type Cookie struct {
		ID    int
		Name  string
		Hobby []string
	}
	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Request().Header.SetCookie("id", "1")
	c.Request().Header.SetCookie("Name", literal_5627)
	c.Request().Header.SetCookie("Hobby", literal_5749)

	q := new(Cookie)
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = c.Bind().Cookie(q)
	}
	require.NoError(b, err)
}

// go test -v  -run=^$ -bench=Benchmark_Bind_Cookie_Map -benchmem -count=4
func BenchmarkBindCookieMap(b *testing.B) {
	var err error

	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	c.Request().SetBody([]byte(``))
	c.Request().Header.SetContentType("")

	c.Request().Header.SetCookie("id", "1")
	c.Request().Header.SetCookie("Name", literal_5627)
	c.Request().Header.SetCookie("Hobby", literal_5749)

	q := make(map[string][]string)
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = c.Bind().Cookie(&q)
	}
	require.NoError(b, err)
}

// custom binder for testing
type customBinder struct{}

func (*customBinder) Name() string {
	return "custom"
}

func (*customBinder) MIMETypes() []string {
	return []string{"test", "test2"}
}

func (*customBinder) Parse(c Ctx, out any) error {
	return json.Unmarshal(c.Body(), out)
}

// go test -run Test_Bind_CustomBinder
func TestBindCustomBinder(t *testing.T) {
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	// Register binder
	customBinder := &customBinder{}
	app.RegisterCustomBinder(customBinder)

	type Demo struct {
		Name string `json:"name"`
	}
	body := []byte(`{"name":"john"}`)
	c.Request().SetBody(body)
	c.Request().Header.SetContentType("test")
	c.Request().Header.SetContentLength(len(body))
	d := new(Demo)

	require.NoError(t, c.Bind().Body(d))
	require.NoError(t, c.Bind().Custom("custom", d))
	require.Equal(t, ErrCustomBinderNotFound, c.Bind().Custom("not_custom", d))
	require.Equal(t, "john", d.Name)
}

// go test -run Test_Bind_Must
func TestBindMust(t *testing.T) {
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	type RequiredQuery struct {
		Name string `query:"name,required"`
	}
	rq := new(RequiredQuery)
	c.Request().URI().SetQueryString("")
	err := c.Bind().Must().Query(rq)
	require.Equal(t, StatusBadRequest, c.Response().StatusCode())
	require.Equal(t, "Bad request: name is empty", err.Error())
}

// simple struct validator for testing
type structValidator struct{}

func (*structValidator) Validate(out any) error {
	out = reflect.ValueOf(out).Elem().Interface()
	sq, ok := out.(simpleQuery)
	if !ok {
		return errors.New("failed to type-assert to simpleQuery")
	}

	if sq.Name != "john" {
		return errors.New("you should have entered right name")
	}

	return nil
}

type simpleQuery struct {
	Name string `query:"name"`
}

// go test -run Test_Bind_StructValidator
func TestBindStructValidator(t *testing.T) {
	app := New(Config{StructValidator: &structValidator{}})
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	rq := new(simpleQuery)
	c.Request().URI().SetQueryString("name=efe")
	require.Equal(t, "you should have entered right name", c.Bind().Query(rq).Error())

	rq = new(simpleQuery)
	c.Request().URI().SetQueryString(literal_3210)
	require.NoError(t, c.Bind().Query(rq))
}

// go test -run Test_Bind_RepeatParserWithSameStruct -v
func TestBindRepeatParserWithSameStruct(t *testing.T) {
	t.Parallel()
	app := New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	type Request struct {
		QueryParam  string `query:"query_param"`
		HeaderParam string `header:"header_param"`
		BodyParam   string `json:"body_param" xml:"body_param" form:"body_param"`
	}

	r := new(Request)

	c.Request().URI().SetQueryString("query_param=query_param")
	require.NoError(t, c.Bind().Query(r))
	require.Equal(t, "query_param", r.QueryParam)

	c.Request().Header.Add("header_param", "header_param")
	require.NoError(t, c.Bind().Header(r))
	require.Equal(t, "header_param", r.HeaderParam)

	var gzipJSON bytes.Buffer
	w := gzip.NewWriter(&gzipJSON)
	_, err := w.Write([]byte(`{"body_param":"body_param"}`))
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)
	c.Request().Header.SetContentType(MIMEApplicationJSON)
	c.Request().Header.Set(HeaderContentEncoding, "gzip")
	c.Request().SetBody(gzipJSON.Bytes())
	c.Request().Header.SetContentLength(len(gzipJSON.Bytes()))
	require.NoError(t, c.Bind().Body(r))
	require.Equal(t, "body_param", r.BodyParam)
	c.Request().Header.Del(HeaderContentEncoding)

	testDecodeParser := func(contentType, body string) {
		c.Request().Header.SetContentType(contentType)
		c.Request().SetBody([]byte(body))
		c.Request().Header.SetContentLength(len(body))
		require.NoError(t, c.Bind().Body(r))
		require.Equal(t, "body_param", r.BodyParam)
	}

	testDecodeParser(MIMEApplicationJSON, `{"body_param":"body_param"}`)
	testDecodeParser(MIMEApplicationXML, `<Demo><body_param>body_param</body_param></Demo>`)
	testDecodeParser(MIMEApplicationForm, "body_param=body_param")
	testDecodeParser(MIMEMultipartForm+`;boundary="b"`, "--b\r\nContent-Disposition: form-data; name=\"body_param\"\r\n\r\nbody_param\r\n--b--")
}

const literal_7016 = "id=1&name=tom&hobby=basketball&hobby=football"

const literal_3476 = "id=1&name=tom&hobby=basketball,football"

const literal_4602 = "name is empty"

const literal_1423 = "2006-01-02"

const literal_0285 = "{0 63753609600 <nil>}"

const literal_7460 = "Existing title"

const literal_8365 = "Existing Body"

const literal_5627 = "John Doe"

const literal_5749 = "golang,fiber"

const literal_8460 = "golang,fiber,go"

const literal_8712 = "Jane Doe"

const literal_3506 = "go,fiber"

const literal_2149 = "milo,coke,pepsi"

const literal_9374 = "Nested.Age"

const literal_7506 = "Nested.Agex"

const literal_4179 = "Next.Val"

const literal_3210 = "name=john"
