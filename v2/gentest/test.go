package main

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/runner-mei/gogen/v2/gentest/models"
)

// 用于测试 parse() 方法
type Key int64

func (key Key) String() string {
	return ""
}

type StrKey string

func (key StrKey) String() string {
	return string(key)
}

const TimeFormat = time.RFC3339

func BoolToString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func ToBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true"
}

func ToDatetime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type TimeRange2 struct {
	Start *time.Time `json:"start"`
	End   *time.Time `json:"end"`
}

type QueryArgs struct {
	Fint    int
	Fstring string
	Ftime   time.Time

	Fintptr    *int
	Fstringptr *string
	Ftimeptr   *time.Time
}

// @http.Client(name="TestClient", ref="true")
type StringSvc interface {

	// @Summary get files
	// @Description get files
	// @ID GetFiles
	// @Accept  json
	// @Produce  json
	// @Param   filenames      query   []string     false  "Some ID" Format(string)
	// @Router /files [get]
	GetFiles(filenames []string) (list []string, total int64, err error)

	// @Summary get files
	// @Description get files
	// @ID GetTimes
	// @Accept  json
	// @Produce  json
	// @Param   times      query   []string     false  "Some ID" Format(datetime)
	// @Router /times [get]
	GetTimes(times []time.Time) (list []string, total int64, err error)

	// @Summary get files
	// @Description get files
	// @ID GetAllFiles
	// @Accept  json
	// @Produce  json
	// @Router /allfiles [get]
	GetAllFiles() (list []string, total int64, err error)

	// @Summary test by int key
	// @Description test by key
	// @ID TestByKey1
	// @Param   key      path   int     true  "Some ID" Format(int)
	// @Accept  json
	// @Produce  json
	// @Router /test_by_key/:key [get]
	TestByKey1(key Key) error

	// @Summary test by int key
	// @Description test by key
	// @ID TestByKey2
	// @Param   key      query   int     false  "Some ID" Format(int)
	// @Accept  json
	// @Produce  json
	// @Router /test_by_key [get]
	TestByKey2(key Key) error

	// @Summary test by str key
	// @Description test by key
	// @ID TestByStrKey1
	// @Param   key      path   string     true  "Some ID" Format(string)
	// @Accept  json
	// @Produce  json
	// @Router /test_by_strkey/:key [get]
	TestByStrKey1(key StrKey) error

	// @Summary test by str key
	// @Description test by key
	// @ID TestByStrKey2
	// @Param   key      query   string     false  "Some ID" Format(string)
	// @Accept  json
	// @Produce  json
	// @Router /test_by_strkey [get]
	TestByStrKey2(key StrKey) error

	// @Summary test by query
	// @Description test by query
	// @ID TestInt64Query
	// @Param   id      query   int     true  "Some ID" Format(int64)
	// @Accept  json
	// @Produce  json
	// @Router /test64 [get]
	TestInt64Query(id int64) error

	// // @http.GET(path="/test_query_args1/:id")

	// @Summary test by query
	// @Description test by query
	// @ID TestQueryArgs1
	// @Param   id      query   int     true  "Some ID" Format(int64)
	// @Param   args    query   QueryArgs     false  "Some ID" Format(int64)
	// @Accept  json
	// @Produce  json
	// @Router /test_query_args1/{id} [get]
	TestQueryArgs1(id int64, args QueryArgs) error

	// @Summary test by query
	// @Description test by query
	// @ID TestQueryArgs2
	// @Param   id      query   int     true  "Some ID" Format(int64)
	// @Param   args    query   QueryArgs     false  "Some ID" Format(int64)
	// @Accept  json
	// @Produce  json
	// @Router /test_query_args2/{id} [get]
	TestQueryArgs2(id int64, args *QueryArgs) error

	// @Summary test by query
	// @Description test by query
	// @ID TestQueryArgs3
	// @Param   id      query   int     true  "Some ID" Format(int64)
	// @Param   args    query   QueryArgs     false  "Some ID" extensions(x-gogen-extend=inline)
	// @Accept  json
	// @Produce  json
	// @Router /test_query_args3/{id} [get]
	TestQueryArgs3(id int64, args QueryArgs) error

	// @Summary test by query
	// @Description test by query
	// @ID TestQueryArgs4
	// @Param   id      query   int     true  "Some ID" Format(int64)
	// @Param   args    query   QueryArgs     false  "Some ID" extensions(x-gogen-extend=inline)
	// @Accept  json
	// @Produce  json
	// @Router /test_query_args4/{id} [get]
	TestQueryArgs4(id int64, args *QueryArgs) error

	// @Summary test by query
	// @Description test by query
	// @ID Ping
	// @Accept  json
	// @Produce  json
	// @Router /ping [get]
	Ping() error

	// @Summary test by query
	// @Description test by query
	// @ID Echo
	// @Param   a      query   string     false  "Some ID" Format(int64)
	// @Accept  json
	// @Produce  json
	// @Router /echo [get]
	Echo(a string) string

	// @Summary test by query
	// @Description test by query
	// @ID EchoBody
	// @Param   body      body   string     false  "Some ID"
	// @Accept  json
	// @Produce  json
	// @Router /echo2 [post]
	EchoBody(body io.Reader) (string, error)

	// @Summary test by body
	// @Description test by body
	// @ID Echo3
	// @Param   a      body   string     false  "Some ID"
	// @Accept  json
	// @Produce  json
	// @Router /echo3 [post]
	Echo3(context context.Context, a string) (string, error)

	// @Summary test by query
	// @Description test by query
	// @ID Concat0
	// @Param   a      query   string     false  "Some ID"
	// @Param   b      query   string     false  "Some ID"
	// @Accept  json
	// @Produce  json
	// @Router /concat [get]
	Concat0(a, b string) (string, error)

	// @Summary test by query
	// @Description test by query
	// @ID Concat1
	// @Param   a      query   string     false  "arg a"
	// @Param   b      query   string     false  "arg b"
	// @Accept  json
	// @Produce  json
	// @Router /concat1 [get]
	Concat1(a, b *string) (string, error)

	// @Summary test by query
	// @Description test by query
	// @ID Concat2
	// @Param   a      path   string     true  "arg a"
	// @Param   b      path   string    true  "arg b"
	// @Accept  json
	// @Produce  json
	// @Router /concat2/{a}/{b} [get]
	Concat2(a, b string) (string, error)

	// @Summary test by query
	// @Description test by query
	// @ID Concat3
	// @Param   a      path   string     true  "arg a"
	// @Param   b      path   string    true  "arg b"
	// @Accept  json
	// @Produce  json
	// @Router /concat3/{a}/{b} [get]
	Concat3(a, b *string) (string, error)

	// @Summary test by query
	// @Description test by query
	// @ID Sub
	// @Param   a      query   string     true  "arg a"
	// @Param   start      query   int64    true  "arg start"
	// @Accept  json
	// @Produce  json
	// @Router /sub [get]
	Sub(a string, start int64) (string, error)

	// @Summary test save
	// @Description test save
	// @ID Save1
	// @Param   a      path   string     true  "arg a"
	// @Param   b      body   string    true  "arg b" extensions(x-gogen-entire-body=true)
	// @Accept  json
	// @Produce  json
	// @Router /save/{a} [get]
	Save(a, b string) (string, error)

	// @Summary test by query
	// @Description test by query
	// @ID Save2
	// @Param   a      path   string     true  "arg a"
	// @Param   b      body   string    true  "arg b" extensions(x-gogen-entire-body=true)
	// @Accept  json
	// @Produce  json
	// @Router /save2/{a} [post]
	Save2(a, b *string) (string, error)

	// @Summary test by query
	// @Description test by query
	// @ID Save3
	// @Param   a      body   string     true  "arg a"
	// @Param   b      body   string    true  "arg b"
	// @Accept  json
	// @Produce  json
	// @Router /save3 [post]
	Save3(a, b *string) (string, error)

	// @Summary test by query
	// @Description test by query
	// @ID Save4
	// @Param   a      body   string     true  "arg a"
	// @Param   b      body   string    true  "arg b"
	// @Accept  json
	// @Produce  json
	// @Router /save4 [post]
	Save4(a, b string) (string, error)

	// @Summary test by query
	// @Description test by query
	// @ID Save5
	// @Param   a      body   string     true  "arg a"
	// @Param   b      body   string    true  "arg b"
	// @Accept  json
	// @Produce  json
	// @Router /save5 [post]
	Save5(context context.Context, a, b string) (string, error)

	// @Summary add by path
	// @Description add by path
	// @ID Add1
	// @Param   a      path   int     true  "arg a"
	// @Param   b      path   int    true  "arg b"
	// @Accept  json
	// @Produce  json
	// @Router /add/{a}/{b} [get]
	Add1(a, b int) (int, error)

	// @Summary add by path
	// @Description add by path
	// @ID Add2
	// @Param   a      path   int     true  "arg a"
	// @Param   b      path   int    true  "arg b"
	// @Accept  json
	// @Produce  json
	// @Router /add2/{a}/{b} [get]
	Add2(a, b *int) (int, error)

	// // @http.GET(path="/add3")

	// @Summary add by path
	// @Description add by path
	// @ID Add3
	// @Param   a      query   int     true  "arg a"
	// @Param   b      query   int    true  "arg b"
	// @Accept  json
	// @Produce  json
	// @Router /add3 [get]
	Add3(a, b *int) (int, error)

	// @Summary add by path
	// @Description add by path
	// @ID Query1
	// @Param   a      query   int     true  "arg a"
	// @Param   begin_at      query   string    true  "arg beginAt" Format(time)
	// @Param   end_at      query   string    true  "arg endAt" Format(time)
	// @Param   is_raw      query   boolean    true  "arg isRaw"
	// @Accept  json
	// @Produce  json
	// @Router /query1 [get]
	Query1(a string, beginAt, endAt time.Time, isRaw bool) string

	// @Summary add by path
	// @Description add by path
	// @ID Query2
	// @Param   a      query   int     true  "arg a"
	// @Param   begin_at      query   string    true  "arg beginAt" Format(time)
	// @Param   end_at      query   string    true  "arg endAt" Format(time)
	// @Param   is_raw      path   boolean    true  "arg isRaw"
	// @Accept  json
	// @Produce  json
	// @Router /query2/{isRaw} [get]
	Query2(a string, beginAt, endAt time.Time, isRaw bool) string

	// @Summary add by path
	// @Description add by path
	// @ID Query3
	// @Param   a      query   int     true  "arg a"
	// @Param   begin_at      query   string    true  "arg beginAt" Format(time)
	// @Param   end_at      query   string    true  "arg endAt" Format(time)
	// @Param   is_raw      path   boolean    true  "arg isRaw"
	// @Accept  json
	// @Produce  json
	// @Router /query3/{isRaw} [get]
	Query3(a string, beginAt, endAt time.Time, isRaw *bool) string

	// @Summary add by path
	// @Description add by path
	// @ID Query4
	// @Param   a      query   int     true  "arg a"
	// @Param   created_at      query   TimeRange    true  "arg beginAt" Format(time)
	// @Param   is_raw      path   boolean    true  "arg isRaw"
	// @Accept  json
	// @Produce  json
	// @Router /query4/{isRaw} [get]
	Query4(a string, createdAt TimeRange, isRaw *bool) string

	// @Summary add by path
	// @Description add by path
	// @ID Query5
	// @Param   a      query   int     true  "arg a"
	// @Param   created_at      query   TimeRange    true  "arg beginAt" Format(time)
	// @Param   is_raw      path   boolean    true  "arg isRaw"
	// @Accept  json
	// @Produce  json
	// @Router /query5/{isRaw} [get]
	Query5(a string, createdAt *TimeRange, isRaw *bool) string

	// @Summary add by path
	// @Description add by path
	// @ID Query6
	// @Param   a      query   int     true  "arg a"
	// @Param   created_at      query   TimeRange2    true  "arg beginAt" Format(time)
	// @Param   is_raw      path   boolean    true  "arg isRaw"
	// @Accept  json
	// @Produce  json
	// @Router /query6/{isRaw} [get]
	Query6(a string, createdAt TimeRange2, isRaw *bool) string

	// @Summary add by path
	// @Description add by path
	// @ID Query7
	// @Param   a      query   int     true  "arg a"
	// @Param   created_at      query   TimeRange2    true  "arg beginAt" Format(time)
	// @Param   is_raw      path   boolean    true  "arg isRaw"
	// @Accept  json
	// @Produce  json
	// @Router /query7/{isRaw} [get]
	Query7(a string, createdAt *TimeRange2, isRaw *bool) string

	// @Summary content_type="text"
	// @Description content_type="text"
	// @ID Query8
	// @Param   item_id      query   int     true  "arg a"
	// @Accept  json
	// @Produce  plain
	// @Router /query8 [get]
	Query8(ctx context.Context, itemID int64) (string, error)

	// @Summary noreturn="true"
	// @Description noreturn="true"
	// @ID CreateWithNoReturn
	// @Accept  json
	// @Produce  json
	// @Router / [post]
	// @x-gogen-noreturn true
	CreateWithNoReturn(ctx context.Context, request *http.Request, response http.ResponseWriter) error

	// @Summary query9 content_type="text"
	// @Description query9 content_type="text"
	// @ID Query9
	// @Param   item_id      query   int     true  "arg a"
	// @Accept  json
	// @Produce  plain
	// @Router /query9 [get]
	Query9(ctx context.Context, itemID sql.NullInt64) (string, error)

	// @Summary query10 content_type="text"
	// @Description query10 content_type="text"
	// @ID Query10
	// @Param   item_id      query   int     true  "arg a"
	// @Accept  json
	// @Produce  plain
	// @Router /query10 [get]
	Query10(ctx context.Context, itemID sql.NullString) (string, error)

	// @Summary query11 content_type="text"
	// @Description query11 content_type="text"
	// @ID Query11
	// @Param   item_id      query   bool     true  "arg a"
	// @Accept  json
	// @Produce  plain
	// @Router /query11 [get]
	Query11(ctx context.Context, itemID sql.NullBool) (string, error)

	// @Summary query12 Name is Upper
	// @Description query12 Name is Upper
	// @ID Query1WithUpName
	// @Param   Name      query   string     true  "arg a"
	// @Accept  json
	// @Produce  plain
	// @Router /query12 [get]
	Query1WithUpName(ctx context.Context, Name string) (string, error)

	// @Summary query12 Name is Upper
	// @Description query12 Name is Upper
	// @ID Query1WithUpName
	// @Param   Name      body   string     true  "arg a"
	// @Accept  json
	// @Produce  json
	// @Router /query12 [post]
	Set1WithUpName(ctx context.Context, Name string) error

	// Misc() string
}

// var _ StringSvc = &StringSvcImpl{}

// type StringSvcImpl struct {
// }

// // @http.GET(path="/test_by_key")
// func (svc *StringSvcImpl) TestByKey(key Key) error {
// 	return nil
// }

// // @http.GET(path="/allfiles")
// func (svc *StringSvcImpl) GetAllFiles() (list []string, total int64, err error) {
// 	return []string{"abc"}, 1, nil
// }

// // @http.GET(path="/test64/:id")
// func (svc *StringSvcImpl) TestInt64Path(id int64) error {
// 	return nil
// }

// // @http.GET(path="/test64")
// func (svc *StringSvcImpl) TestInt64Query(id int64) error {
// 	return nil
// }

// // @http.GET(path="/test_query_args1/:id")
// func (svc *StringSvcImpl) TestQueryArgs1(id int64, args QueryArgs) error {
// 	return nil
// }

// // @http.GET(path="/test_query_args2/:id")
// func (svc *StringSvcImpl) TestQueryArgs2(id int64, args *QueryArgs) error {
// 	return nil
// }

// // @http.GET(path="/test_query_args3/:id?args=<none>")
// func (svc *StringSvcImpl) TestQueryArgs3(id int64, args QueryArgs) error {
// 	return nil
// }

// // @http.GET(path="/test_query_args4/:id?<none>=args")
// func (svc *StringSvcImpl) TestQueryArgs4(id int64, args *QueryArgs) error {
// 	return nil
// }

// // @http.GET(path="/ping")
// func (svc *StringSvcImpl) Ping() error {
// 	return nil
// }

// // @http.GET(path="/echo")
// func (svc *StringSvcImpl) Echo(a string) string {
// 	return a
// }

// // @http.GET(path="/echo_body1", data="body")
// func (svc *StringSvcImpl) EchoBody(body io.Reader) (string, error) {
// 	bs, err := ioutil.ReadAll(body)
// 	return string(bs), err
// }

// // @http.POST(path="/echo3")
// func (svc *StringSvcImpl) Echo3(context context.Context, a string) (string, error) {
// 	return a, nil
// }

// // @http.GET(path="/concat")
// func (svc *StringSvcImpl) Concat(a, b string) (string, error) {
// 	return a + b, nil
// }

// // @http.GET(path="/concat1")
// func (svc *StringSvcImpl) Concat1(a, b *string) (string, error) {
// 	return *a + *b, nil
// }

// // @http.GET(path="/concat2/:a/:b")
// func (svc *StringSvcImpl) Concat2(a, b string) (string, error) {
// 	return a + b, nil
// }

// // @http.GET(path="/concat3/:a/:b")
// func (svc *StringSvcImpl) Concat3(a, b *string) (string, error) {
// 	return *a + *b, nil
// }

// // @http.GET(path="/sub")
// func (svc *StringSvcImpl) Sub(a string, start int64) (string, error) {
// 	return a[start:], nil
// }

// // @http.POST(path="/save/:a", data="b")
// func (svc *StringSvcImpl) Save(a, b string) (string, error) {
// 	return "", nil
// }

// // @http.POST(path="/save2/:a", data="b")
// func (svc *StringSvcImpl) Save2(a, b *string) (string, error) {
// 	return *a + *b, nil
// }

// // @http.POST(path="/save3")
// func (svc *StringSvcImpl) Save3(a, b *string) (string, error) {
// 	return *a + *b, nil
// }

// // @http.POST(path="/save4")
// func (svc *StringSvcImpl) Save4(a, b string) (string, error) {
// 	return a + b, nil
// }

// // @http.POST(path="/save5")
// func (svc *StringSvcImpl) Save5(context context.Context, a, b string) (string, error) {
// 	return a + b, nil
// }

// // @http.GET(path="/add/:a/:b")
// func (svc *StringSvcImpl) Add(a, b int) (int, error) {
// 	return a + b, nil
// }

// // @http.GET(path="/add2/:a/:b")
// func (svc *StringSvcImpl) Add2(a, b *int) (int, error) {
// 	return *a + *b, nil
// }

// // @http.GET(path="/add3")
// func (svc *StringSvcImpl) Add3(a, b *int) (int, error) {
// 	return *a + *b, nil
// }

// // @http.GET(path="/query1")
// func (svc *StringSvcImpl) Query1(a string, beginAt, endAt time.Time, isRaw bool) string {
// 	return "queue"
// }

// // @http.GET(path="/query2/:isRaw")
// func (svc *StringSvcImpl) Query2(a string, beginAt, endAt time.Time, isRaw bool) string {
// 	return "queue"
// }

// // @http.GET(path="/query3/:isRaw")
// func (svc *StringSvcImpl) Query3(a string, beginAt, endAt time.Time, isRaw *bool) string {
// 	return "queue"
// }

// // @http.GET(path="/query4/:isRaw")
// func (svc *StringSvcImpl) Query4(a string, createdAt TimeRange, isRaw *bool) string {
// 	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
// }

// // @http.GET(path="/query5/:isRaw")
// func (svc *StringSvcImpl) Query5(a string, createdAt *TimeRange, isRaw *bool) string {
// 	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
// }

// // @http.GET(path="/query6/:isRaw")
// func (svc *StringSvcImpl) Query6(a string, createdAt TimeRange2, isRaw *bool) string {
// 	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
// }

// // @http.GET(path="/query7/:isRaw")
// func (svc *StringSvcImpl) Query7(a string, createdAt *TimeRange2, isRaw *bool) string {
// 	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
// }

// func (svc *StringSvcImpl) Misc() string {
// 	return ""
// }

// type StringSvcWithContext struct {
// }

// // @http.GET(path="/echo")
// func (svc *StringSvcWithContext) Echo(ctx context.Context, a string) string {
// 	return a
// }

// // @http.GET(path="/echo", data="body")
// func (svc *StringSvcWithContext) EchoBody(ctx context.Context, body io.Reader) (string, error) {
// 	bs, err := ioutil.ReadAll(body)
// 	return string(bs), err
// }

// // @http.GET(path="/concat")
// func (svc *StringSvcWithContext) Concat(ctx context.Context, a, b string) (string, error) {
// 	return a + b, nil
// }

// // @http.GET(path="/concat1")
// func (svc *StringSvcWithContext) Concat1(ctx context.Context, a, b *string) (string, error) {
// 	return *a + *b, nil
// }

// // @http.GET(path="/concat2/:a/:b")
// func (svc *StringSvcWithContext) Concat2(ctx context.Context, a, b string) (string, error) {
// 	return a + b, nil
// }

// // @http.GET(path="/concat3/:a/:b")
// func (svc *StringSvcWithContext) Concat3(ctx context.Context, a, b *string) (string, error) {
// 	return *a + *b, nil
// }

// // @http.GET(path="/sub")
// func (svc *StringSvcWithContext) Sub(ctx context.Context, a string, start int64) (string, error) {
// 	return a[start:], nil
// }

// // @http.POST(path="/save/:a", data="b")
// func (svc *StringSvcWithContext) Save(ctx context.Context, a, b string) (string, error) {
// 	return "", nil
// }

// // @http.POST(path="/save2/:a", data="b")
// func (svc *StringSvcWithContext) Save2(ctx context.Context, a, b *string) (string, error) {
// 	return *a + *b, nil
// }

// // @http.GET(path="/add/:a/:b")
// func (svc *StringSvcWithContext) Add(ctx context.Context, a, b int) (int, error) {
// 	return a + b, nil
// }

// // @http.GET(path="/add2/:a/:b")
// func (svc *StringSvcWithContext) Add2(ctx context.Context, a, b *int) (int, error) {
// 	return *a + *b, nil
// }

// // @http.GET(path="/add3")
// func (svc *StringSvcWithContext) Add3(ctx context.Context, a, b *int) (int, error) {
// 	return *a + *b, nil
// }

// func (svc *StringSvcWithContext) Misc() string {
// 	return ""
// }

// 用于测试 Parse() 不会 panic
func notpanic() {}

type AliasRequestQuery = models.RequestQuery

type RequestQueryEx1 struct {
	models.RequestQuery

	ExArg string
}

type RequestQueryEx2 struct {
	*models.RequestQuery

	ExArg string
}

type RequestQueryEx3 struct {
	Request models.RequestQuery `json:"request"`

	ExArg string
}

type RequestQueryEx4 struct {
	Request *models.RequestQuery `json:"request"`

	ExArg string
}

type Sub1 struct {
	A1 string `json:"a1"`
	A2 string `json:"a2"`
}

type Sub2 struct {
	B1 string `json:"b1"`
	B2 string `json:"b2"`
}

type Sub3 struct {
	Sub2
}

type SubTest1 struct {
	Sub1 Sub1 `json:"sub1"`
	Sub2 Sub2 `json:"sub2"`

	ExArg string
}

type SubTest2 struct {
	Sub1 *Sub1 `json:"sub1"`
	Sub2 *Sub2 `json:"sub2"`

	ExArg string
}

type SubTest3 struct {
	SubTest2

	ExArg string
}

type SubTest4 struct {
	*SubTest2

	ExArg string
}

type Requests interface {

	// @Summary Query1
	// @Description Query1
	// @ID Requests.Query1
	// @Param   query    query   models.RequestQuery     false  "request query param"
	// @Param   offset   query   int     false  "offset"
	// @Param   limit    query   int     false  "limit"
	// @Accept  json
	// @Produce  json
	// @Router /requests/query1 [get]
	Query1(ctx context.Context, query *models.RequestQuery, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// // @http.GET(path="/query2?query=<none>")
	// Query2(ctx context.Context, query *models.RequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)

	// // @http.GET(path="/query3?query=<none>")
	// Query3(ctx context.Context, query *AliasRequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)

	// // @http.GET(path="/queryex1")
	// QueryEx1(ctx context.Context, query *RequestQueryEx1, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// // @http.GET(path="/queryex2")
	// QueryEx2(ctx context.Context, query *RequestQueryEx2, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// // @http.GET(path="/queryex3")
	// QueryEx3(ctx context.Context, query *RequestQueryEx3, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// // @http.GET(path="/queryex4")
	// QueryEx4(ctx context.Context, query *RequestQueryEx4, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// // @http.GET(path="/queryex3/NoPrefix?query=<none>")
	// QueryEx3NoPrefix(ctx context.Context, query *RequestQueryEx3, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// // @http.GET(path="/queryex4/NoPrefix?query=<none>")
	// QueryEx4NoPrefix(ctx context.Context, query *RequestQueryEx4, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// // @http.GET(path="")
	// List(ctx context.Context, query *models.RequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)
	// // @http.POST(path="", data="data")
	// Create(ctx context.Context, data *models.Request) (int64, error)

	// // @http.POST(path="")
	// Create2(ctx context.Context, request *models.Request, testarg int64) (int64, error)

	// // @http.PUT(path="/:id", data="data")
	// UpdateByID(ctx context.Context, id int64, data *models.Request) (int64, error)

	// // @http.PATCH(path="/:id")
	// Set1ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// // @http.PATCH(path="/:id", data="params")
	// Set2ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// // @http.PUT(path="/:id")
	// Set3ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// // @http.PUT(path="/:id", data="params")
	// Set4ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// // @http.POST(path="/:id/5")
	// Set5ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// // @http.POST(path="/:id/6", data="params")
	// Set6ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// // @http.POST(path="/:id/7", data="params", dataType="map[string]string")
	// Set7ByID(ctx context.Context, id int64, params interface{}) (int64, err error)

	// // @http.GET(path="/querysub1")
	// QuerySubTest1(ctx context.Context, query *SubTest1) (requests []map[string]interface{}, err error)

	// // @http.GET(path="/querysub2")
	// QuerySubTest2(ctx context.Context, query *SubTest2) (requests []map[string]interface{}, err error)

	// // @http.GET(path="/querysub3")
	// QuerySubTest3(ctx context.Context, query *SubTest3) (requests []map[string]interface{}, err error)

	// // @http.GET(path="/querysub4")
	// QuerySubTest4(ctx context.Context, query *SubTest4) (requests []map[string]interface{}, err error)
}
