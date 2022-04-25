package main

import (
	"context"
	"database/sql"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
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

// @http.Client name="TestClient" reference="false"
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
	// @Router /test_by_key/{key} [get]
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
	// @Router /test_by_strkey/{key} [get]
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
	// @ID TestInt64Path
	// @Param   id      path   int     true  "Some ID" Format(int64)
	// @Accept  json
	// @Produce  json
	// @Router /test64/{id} [get]
	TestInt64Path(id int64) error

	// @Summary test by query
	// @Description test by query
	// @ID TestInt64Query
	// @Param   id      query   int     true  "Some ID" Format(int64)
	// @Accept  json
	// @Produce  json
	// @Router /test64 [get]
	TestInt64Query(id int64) error

	// @Summary test by query
	// @Description test by query
	// @ID TestQueryArgs1
	// @Param   id      path   int     true  "Some ID" Format(int64)
	// @Param   args    query   QueryArgs     false  "Some ID" Format(int64)
	// @Accept  json
	// @Produce  json
	// @Router /test_query_args1/{id} [get]
	TestQueryArgs1(id int64, args QueryArgs) error

	// @Summary test by query
	// @Description test by query
	// @ID TestQueryArgs2
	// @Param   id      path   int     true  "Some ID" Format(int64)
	// @Param   args    query   QueryArgs     false  "Some ID" Format(int64)
	// @Accept  json
	// @Produce  json
	// @Router /test_query_args2/{id} [get]
	TestQueryArgs2(id int64, args *QueryArgs) error

	// @Summary test by query
	// @Description test by query
	// @ID TestQueryArgs3
	// @Param   id      path   int     true  "Some ID" Format(int64)
	// @Param   args    query   QueryArgs     false  "Some ID" extensions(x-gogen-extend=inline)
	// @Accept  json
	// @Produce  json
	// @Router /test_query_args3/{id} [get]
	TestQueryArgs3(id int64, args QueryArgs) error

	// @Summary test by query
	// @Description test by query
	// @ID TestQueryArgs4
	// @Param   id      path   int     true  "Some ID" Format(int64)
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
	// @Router /save/{a} [post]
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
	// @Router /query2/{is_raw} [get]
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
	// @Router /query3/{is_raw} [get]
	Query3(a string, beginAt, endAt time.Time, isRaw *bool) string

	// @Summary add by path
	// @Description add by path
	// @ID Query4
	// @Param   a      query   int     true  "arg a"
	// @Param   created_at      query   TimeRange    true  "arg beginAt" Format(time)
	// @Param   is_raw      path   boolean    true  "arg isRaw"
	// @Accept  json
	// @Produce  json
	// @Router /query4/{is_raw} [get]
	Query4(a string, createdAt TimeRange, isRaw *bool) string

	// @Summary add by path
	// @Description add by path
	// @ID Query5
	// @Param   a      query   int     true  "arg a"
	// @Param   created_at      query   TimeRange    true  "arg beginAt" Format(time)
	// @Param   is_raw      path   boolean    true  "arg isRaw"
	// @Accept  json
	// @Produce  json
	// @Router /query5/{is_raw} [get]
	Query5(a string, createdAt *TimeRange, isRaw *bool) string

	// @Summary add by path
	// @Description add by path
	// @ID Query6
	// @Param   a      query   int     true  "arg a"
	// @Param   created_at      query   TimeRange2    true  "arg beginAt" Format(time)
	// @Param   is_raw      path   boolean    true  "arg isRaw"
	// @Accept  json
	// @Produce  json
	// @Router /query6/{is_raw} [get]
	Query6(a string, createdAt TimeRange2, isRaw *bool) string

	// @Summary add by path
	// @Description add by path
	// @ID Query7
	// @Param   a      query   int     true  "arg a"
	// @Param   created_at      query   TimeRange2    true  "arg beginAt" Format(time)
	// @Param   is_raw      path   boolean    true  "arg isRaw"
	// @Accept  json
	// @Produce  json
	// @Router /query7/{is_raw} [get]
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

	Misc() string
}

var _ StringSvc = &StringSvcImpl{}

type StringSvcImpl struct{}

// @Summary get files
// @Description get files
// @ID StringSvcImpl.GetFiles
// @Accept  json
// @Produce  json
// @Param   filenames      query   []string     false  "Some ID" Format(string)
// @Router /impl/files [get]
func (svc *StringSvcImpl) GetFiles(filenames []string) (list []string, total int64, err error) {
	return []string{"a.txt"}, 10, nil
}

// @Summary get files
// @Description get files
// @ID StringSvcImpl.GetTimes
// @Accept  json
// @Produce  json
// @Param   times      query   []string     false  "Some ID" Format(datetime)
// @Router /impl/times [get]
func (svc *StringSvcImpl) GetTimes(times []time.Time) (list []string, total int64, err error) {
	return []string{"a.txt"}, 10, nil
}

// @Summary get files
// @Description get files
// @ID StringSvcImpl.GetAllFiles
// @Accept  json
// @Produce  json
// @Router /impl/allfiles [get]
func (svc *StringSvcImpl) GetAllFiles() (list []string, total int64, err error) {
	return []string{"abc"}, 1, nil
}

// @Summary test by int key
// @Description test by key
// @ID StringSvcImpl.TestByKey1
// @Param   key      path   int     true  "Some ID" Format(int)
// @Accept  json
// @Produce  json
// @Router /impl/test_by_key/{key} [get]
func (svc *StringSvcImpl) TestByKey1(key Key) error {
	return nil
}

// @Summary test by int key
// @Description test by key
// @ID StringSvcImpl.TestByKey2
// @Param   key      query   int     false  "Some ID" Format(int)
// @Accept  json
// @Produce  json
// @Router /impl/test_by_key [get]
func (svc *StringSvcImpl) TestByKey2(key Key) error {
	return nil
}

// @Summary test by str key
// @Description test by key
// @ID StringSvcImpl.TestByStrKey1
// @Param   key      path   string     true  "Some ID" Format(string)
// @Accept  json
// @Produce  json
// @Router /impl/test_by_strkey/{key} [get]
func (svc *StringSvcImpl) TestByStrKey1(key StrKey) error {
	return nil
}

// @Summary test by str key
// @Description test by key
// @ID StringSvcImpl.TestByStrKey2
// @Param   key      query   string     false  "Some ID" Format(string)
// @Accept  json
// @Produce  json
// @Router /impl/test_by_strkey [get]
func (svc *StringSvcImpl) TestByStrKey2(key StrKey) error {
	return nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.TestInt64Path
// @Param   id      path   int     true  "Some ID" Format(int64)
// @Accept  json
// @Produce  json
// @Router /impl/test64/{id} [get]
func (svc *StringSvcImpl) TestInt64Path(id int64) error {
	return nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.TestInt64Query
// @Param   id      query   int     true  "Some ID" Format(int64)
// @Accept  json
// @Produce  json
// @Router /impl/test64 [get]
func (svc *StringSvcImpl) TestInt64Query(id int64) error {
	return nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.TestQueryArgs1
// @Param   id      path   int     true  "Some ID" Format(int64)
// @Param   args    query   QueryArgs     false  "Some ID" Format(int64)
// @Accept  json
// @Produce  json
// @Router /impl/test_query_args1/{id} [get]
func (svc *StringSvcImpl) TestQueryArgs1(id int64, args QueryArgs) error {
	return nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.TestQueryArgs2
// @Param   id      path   int     true  "Some ID" Format(int64)
// @Param   args    query   QueryArgs     false  "Some ID" Format(int64)
// @Accept  json
// @Produce  json
// @Router /impl/test_query_args2/{id} [get]
func (svc *StringSvcImpl) TestQueryArgs2(id int64, args *QueryArgs) error {
	return nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.TestQueryArgs3
// @Param   id      path   int     true  "Some ID" Format(int64)
// @Param   args    query   QueryArgs     false  "Some ID" extensions(x-gogen-extend=inline)
// @Accept  json
// @Produce  json
// @Router /impl/test_query_args3/{id} [get]
func (svc *StringSvcImpl) TestQueryArgs3(id int64, args QueryArgs) error {
	return nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.TestQueryArgs4
// @Param   id      path   int     true  "Some ID" Format(int64)
// @Param   args    query   QueryArgs     false  "Some ID" extensions(x-gogen-extend=inline)
// @Accept  json
// @Produce  json
// @Router /impl/test_query_args4/{id} [get]
func (svc *StringSvcImpl) TestQueryArgs4(id int64, args *QueryArgs) error {
	return nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.Ping
// @Accept  json
// @Produce  json
// @Router /impl/ping [get]
func (svc *StringSvcImpl) Ping() error {
	return nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.Echo
// @Param   a      query   string     false  "Some ID" Format(int64)
// @Accept  json
// @Produce  json
// @Router /impl/echo [get]
func (svc *StringSvcImpl) Echo(a string) string {
	return a
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.EchoBody
// @Param   body      body   string     false  "Some ID"
// @Accept  json
// @Produce  json
// @Router /impl/echo2 [post]
func (svc *StringSvcImpl) EchoBody(body io.Reader) (string, error) {
	bs, err := ioutil.ReadAll(body)
	return string(bs), err
}

// @Summary test by body
// @Description test by body
// @ID StringSvcImpl.Echo3
// @Param   a      body   string     false  "Some ID"
// @Accept  json
// @Produce  json
// @Router /impl/echo3 [post]
func (svc *StringSvcImpl) Echo3(context context.Context, a string) (string, error) {
	return a, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.Concat0
// @Param   a      query   string     false  "Some ID"
// @Param   b      query   string     false  "Some ID"
// @Accept  json
// @Produce  json
// @Router /impl/concat [get]
func (svc *StringSvcImpl) Concat0(a, b string) (string, error) {
	return a + b, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.Concat1
// @Param   a      query   string     false  "arg a"
// @Param   b      query   string     false  "arg b"
// @Accept  json
// @Produce  json
// @Router /impl/concat1 [get]
func (svc *StringSvcImpl) Concat1(a, b *string) (string, error) {
	return *a + *b, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.Concat2
// @Param   a      path   string     true  "arg a"
// @Param   b      path   string    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /impl/concat2/{a}/{b} [get]
func (svc *StringSvcImpl) Concat2(a, b string) (string, error) {
	return a + b, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.Concat3
// @Param   a      path   string     true  "arg a"
// @Param   b      path   string    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /impl/concat3/{a}/{b} [get]
func (svc *StringSvcImpl) Concat3(a, b *string) (string, error) {
	return *a + *b, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.Sub
// @Param   a      query   string     true  "arg a"
// @Param   start      query   int64    true  "arg start"
// @Accept  json
// @Produce  json
// @Router /impl/sub [get]
func (svc *StringSvcImpl) Sub(a string, start int64) (string, error) {
	return a[start:], nil
}

// @Summary test save
// @Description test save
// @ID StringSvcImpl.Save1
// @Param   a      path   string     true  "arg a"
// @Param   b      body   string    true  "arg b" extensions(x-gogen-entire-body=true)
// @Accept  json
// @Produce  json
// @Router /impl/save/{a} [post]
func (svc *StringSvcImpl) Save(a, b string) (string, error) {
	return "", nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.Save2
// @Param   a      path   string     true  "arg a"
// @Param   b      body   string    true  "arg b" extensions(x-gogen-entire-body=true)
// @Accept  json
// @Produce  json
// @Router /impl/save2/{a} [post]
func (svc *StringSvcImpl) Save2(a, b *string) (string, error) {
	return *a + *b, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.Save3
// @Param   a      body   string     true  "arg a"
// @Param   b      body   string    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /impl/save3 [post]
func (svc *StringSvcImpl) Save3(a, b *string) (string, error) {
	return *a + *b, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.Save4
// @Param   a      body   string     true  "arg a"
// @Param   b      body   string    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /impl/save4 [post]
func (svc *StringSvcImpl) Save4(a, b string) (string, error) {
	return a + b, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcImpl.Save5
// @Param   a      body   string     true  "arg a"
// @Param   b      body   string    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /impl/save5 [post]
func (svc *StringSvcImpl) Save5(context context.Context, a, b string) (string, error) {
	return a + b, nil
}

// @Summary add by path
// @Description add by path
// @ID StringSvcImpl.Add1
// @Param   a      path   int     true  "arg a"
// @Param   b      path   int    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /impl/add/{a}/{b} [get]
func (svc *StringSvcImpl) Add1(a, b int) (int, error) {
	return a + b, nil
}

// @Summary add by path
// @Description add by path
// @ID StringSvcImpl.Add2
// @Param   a      path   int     true  "arg a"
// @Param   b      path   int    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /impl/add2/{a}/{b} [get]
func (svc *StringSvcImpl) Add2(a, b *int) (int, error) {
	return *a + *b, nil
}

// @Summary add by path
// @Description add by path
// @ID StringSvcImpl.Add3
// @Param   a      query   int     true  "arg a"
// @Param   b      query   int    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /impl/add3 [get]
func (svc *StringSvcImpl) Add3(a, b *int) (int, error) {
	return *a + *b, nil
}

// @Summary add by path
// @Description add by path
// @ID StringSvcImpl.Query1
// @Param   a      query   int     true  "arg a"
// @Param   begin_at      query   string    true  "arg beginAt" Format(time)
// @Param   end_at      query   string    true  "arg endAt" Format(time)
// @Param   is_raw      query   boolean    true  "arg isRaw"
// @Accept  json
// @Produce  json
// @Router /impl/query1 [get]
func (svc *StringSvcImpl) Query1(a string, beginAt, endAt time.Time, isRaw bool) string {
	return "queue"
}

// @Summary add by path
// @Description add by path
// @ID StringSvcImpl.Query2
// @Param   a      query   int     true  "arg a"
// @Param   begin_at      query   string    true  "arg beginAt" Format(time)
// @Param   end_at      query   string    true  "arg endAt" Format(time)
// @Param   is_raw      path   boolean    true  "arg isRaw"
// @Accept  json
// @Produce  json
// @Router /impl/query2/{is_raw} [get]
func (svc *StringSvcImpl) Query2(a string, beginAt, endAt time.Time, isRaw bool) string {
	return "queue"
}

// @Summary add by path
// @Description add by path
// @ID StringSvcImpl.Query3
// @Param   a      query   int     true  "arg a"
// @Param   begin_at      query   string    true  "arg beginAt" Format(time)
// @Param   end_at      query   string    true  "arg endAt" Format(time)
// @Param   is_raw      path   boolean    true  "arg isRaw"
// @Accept  json
// @Produce  json
// @Router /impl/query3/{is_raw} [get]
func (svc *StringSvcImpl) Query3(a string, beginAt, endAt time.Time, isRaw *bool) string {
	return "queue"
}

// @Summary add by path
// @Description add by path
// @ID StringSvcImpl.Query4
// @Param   a      query   int     true  "arg a"
// @Param   created_at      query   TimeRange    true  "arg beginAt" Format(time)
// @Param   is_raw      path   boolean    true  "arg isRaw"
// @Accept  json
// @Produce  json
// @Router /impl/query4/{is_raw} [get]
func (svc *StringSvcImpl) Query4(a string, createdAt TimeRange, isRaw *bool) string {
	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
}

// @Summary add by path
// @Description add by path
// @ID StringSvcImpl.Query5
// @Param   a      query   int     true  "arg a"
// @Param   created_at      query   TimeRange    true  "arg beginAt" Format(time)
// @Param   is_raw      path   boolean    true  "arg isRaw"
// @Accept  json
// @Produce  json
// @Router /impl/query5/{is_raw} [get]
func (svc *StringSvcImpl) Query5(a string, createdAt *TimeRange, isRaw *bool) string {
	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
}

// @Summary add by path
// @Description add by path
// @ID StringSvcImpl.Query6
// @Param   a      query   int     true  "arg a"
// @Param   created_at      query   TimeRange2    true  "arg beginAt" Format(time)
// @Param   is_raw      path   boolean    true  "arg isRaw"
// @Accept  json
// @Produce  json
// @Router /impl/query6/{is_raw} [get]
func (svc *StringSvcImpl) Query6(a string, createdAt TimeRange2, isRaw *bool) string {
	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
}

// @Summary add by path
// @Description add by path
// @ID StringSvcImpl.Query7
// @Param   a      query   int     true  "arg a"
// @Param   created_at      query   TimeRange2    true  "arg beginAt" Format(time)
// @Param   is_raw      path   boolean    true  "arg isRaw"
// @Accept  json
// @Produce  json
// @Router /impl/query7/{is_raw} [get]
func (svc *StringSvcImpl) Query7(a string, createdAt *TimeRange2, isRaw *bool) string {
	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
}

// @Summary content_type="text"
// @Description content_type="text"
// @ID StringSvcImpl.Query8
// @Param   item_id      query   int     true  "arg a"
// @Accept  json
// @Produce  plain
// @Router /impl/query8 [get]
func (svc *StringSvcImpl) Query8(ctx context.Context, itemID int64) (string, error) {
	return "queue:" + strconv.FormatInt(itemID, 10), nil
}

// @Summary noreturn="true"
// @Description noreturn="true"
// @ID StringSvcImpl.CreateWithNoReturn
// @Accept  json
// @Produce  json
// @Router /impl/ [post]
// @x-gogen-noreturn true
func (svc *StringSvcImpl) CreateWithNoReturn(ctx context.Context, request *http.Request, response http.ResponseWriter) error {
	return nil
}

// @Summary query9 content_type="text"
// @Description query9 content_type="text"
// @ID StringSvcImpl.Query9
// @Param   item_id      query   int     true  "arg a"
// @Accept  json
// @Produce  plain
// @Router /impl/query9 [get]
func (svc *StringSvcImpl) Query9(ctx context.Context, itemID sql.NullInt64) (string, error) {
	return "query9", nil
}

// @Summary query10 content_type="text"
// @Description query10 content_type="text"
// @ID StringSvcImpl.Query10
// @Param   item_id      query   int     true  "arg a"
// @Accept  json
// @Produce  plain
// @Router /impl/query10 [get]
func (svc *StringSvcImpl) Query10(ctx context.Context, itemID sql.NullString) (string, error) {
	return "query10", nil
}

// @Summary query11 content_type="text"
// @Description query11 content_type="text"
// @ID StringSvcImpl.Query11
// @Param   item_id      query   bool     true  "arg a"
// @Accept  json
// @Produce  plain
// @Router /impl/query11 [get]
func (svc *StringSvcImpl) Query11(ctx context.Context, itemID sql.NullBool) (string, error) {
	return "query11", nil
}

// @Summary query12 Name is Upper
// @Description query12 Name is Upper
// @ID StringSvcImpl.Query1WithUpName
// @Param   Name      query   string     true  "arg a"
// @Accept  json
// @Produce  plain
// @Router /impl/query12 [get]
func (svc *StringSvcImpl) Query1WithUpName(ctx context.Context, Name string) (string, error) {
	return "Query1WithUpName", nil
}

// @Summary query12 Name is Upper
// @Description query12 Name is Upper
// @ID StringSvcImpl.Query1WithUpName
// @Param   Name      body   string     true  "arg a"
// @Accept  json
// @Produce  json
// @Router /impl/query12 [post]
func (svc *StringSvcImpl) Set1WithUpName(ctx context.Context, Name string) error {
	return nil
}

func (svc *StringSvcImpl) Misc() string {
	return ""
}

type StringSvcWithContext struct {
}

// @Summary test by query
// @Description test by query
// @ID StringSvcWithContext.Echo
// @Param   a      query   string     false  "Some ID" Format(int64)
// @Accept  json
// @Produce  json
// @Router /ctx/echo [get]
func (svc *StringSvcWithContext) Echo(ctx context.Context, a string) string {
	return a
}

// @Summary test by query
// @Description test by query
// @ID StringSvcWithContext.EchoBody
// @Param   body      body   string     false  "Some ID"
// @Accept  json
// @Produce  json
// @Router /ctx/echo2 [post]
func (svc *StringSvcWithContext) EchoBody(ctx context.Context, body io.Reader) (string, error) {
	bs, err := ioutil.ReadAll(body)
	return string(bs), err
}

// @Summary test by query
// @Description test by query
// @ID StringSvcWithContext.Concat0
// @Param   a      query   string     false  "Some ID"
// @Param   b      query   string     false  "Some ID"
// @Accept  json
// @Produce  json
// @Router /ctx/concat [get]
func (svc *StringSvcWithContext) Concat0(ctx context.Context, a, b string) (string, error) {
	return a + b, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcWithContext.Concat1
// @Param   a      query   string     false  "arg a"
// @Param   b      query   string     false  "arg b"
// @Accept  json
// @Produce  json
// @Router /ctx/concat1 [get]
func (svc *StringSvcWithContext) Concat1(ctx context.Context, a, b *string) (string, error) {
	return *a + *b, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcWithContext.Concat2
// @Param   a      path   string     true  "arg a"
// @Param   b      path   string    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /ctx/concat2/{a}/{b} [get]
func (svc *StringSvcWithContext) Concat2(ctx context.Context, a, b string) (string, error) {
	return a + b, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcWithContext.Concat3
// @Param   a      path   string     true  "arg a"
// @Param   b      path   string    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /ctx/concat3/{a}/{b} [get]
func (svc *StringSvcWithContext) Concat3(ctx context.Context, a, b *string) (string, error) {
	return *a + *b, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcWithContext.Sub
// @Param   a      query   string     true  "arg a"
// @Param   start      query   int64    true  "arg start"
// @Accept  json
// @Produce  json
// @Router /ctx/sub [get]
func (svc *StringSvcWithContext) Sub(ctx context.Context, a string, start int64) (string, error) {
	return a[start:], nil
}

// @Summary test save
// @Description test save
// @ID StringSvcWithContext.Save1
// @Param   a      path   string     true  "arg a"
// @Param   b      body   string    true  "arg b" extensions(x-gogen-entire-body=true)
// @Accept  json
// @Produce  json
// @Router /ctx/save1/{a} [post]
func (svc *StringSvcWithContext) Save1(ctx context.Context, a, b string) (string, error) {
	return "", nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcWithContext.Save2
// @Param   a      path   string     true  "arg a"
// @Param   b      body   string    true  "arg b" extensions(x-gogen-entire-body=true)
// @Accept  json
// @Produce  json
// @Router /ctx/save2/{a} [post]
func (svc *StringSvcWithContext) Save2(ctx context.Context, a, b *string) (string, error) {
	return *a + *b, nil
}

// @Summary test by query
// @Description test by query
// @ID StringSvcWithContext.Save3
// @Param   a      body   string     true  "arg a"
// @Param   b      body   string    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /ctx/save3 [post]
func (svc *StringSvcWithContext) Save3(a, b *string) (string, error) {
	return *a + *b, nil
}

// @Summary add by path
// @Description add by path
// @ID StringSvcWithContext.Add1
// @Param   a      path   int     true  "arg a"
// @Param   b      path   int    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /ctx/add1/{a}/{b} [get]
func (svc *StringSvcWithContext) Add1(ctx context.Context, a, b int) (int, error) {
	return a + b, nil
}

// @Summary add by path
// @Description add by path
// @ID StringSvcWithContext.Add2
// @Param   a      path   int     true  "arg a"
// @Param   b      path   int    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /ctx/add2/{a}/{b} [get]
func (svc *StringSvcWithContext) Add2(ctx context.Context, a, b *int) (int, error) {
	return *a + *b, nil
}

// @Summary add by path
// @Description add by path
// @ID StringSvcWithContext.Add3
// @Param   a      query   int     true  "arg a"
// @Param   b      query   int    true  "arg b"
// @Accept  json
// @Produce  json
// @Router /ctx/add3 [get]
func (svc *StringSvcWithContext) Add3(ctx context.Context, a, b *int) (int, error) {
	return *a + *b, nil
}

func (svc *StringSvcWithContext) Misc() string {
	return ""
}

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

	// @Summary Query2
	// @Description Query2
	// @ID Requests.Query2
	// @Param   query    query   models.RequestQuery     false  "request query param" extensions(x-gogen-extend=inline)
	// @Param   offset   query   int     false  "offset"
	// @Param   limit    query   int     false  "limit"
	// @Accept  json
	// @Produce  json
	// @Router /requests/query2 [get]
	Query2(ctx context.Context, query *models.RequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)

	// @Summary Query3
	// @Description Query3
	// @ID Requests.Query3
	// @Param   query    query   models.RequestQuery     false  "request query param" extensions(x-gogen-extend=inline)
	// @Param   offset   query   int     false  "offset"
	// @Param   limit    query   int     false  "limit"
	// @Accept  json
	// @Produce  json
	// @Router /requests/query3 [get]
	Query3(ctx context.Context, query *AliasRequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)

	// @Summary QueryEx1
	// @Description QueryEx1
	// @ID Requests.QueryEx1
	// @Param   query    query   RequestQueryEx1     false  "request query param"
	// @Param   offset   query   int     false  "offset"
	// @Param   limit    query   int     false  "limit"
	// @Accept  json
	// @Produce  json
	// @Router /requests/queryex1 [get]
	QueryEx1(ctx context.Context, query *RequestQueryEx1, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @Summary QueryEx2
	// @Description QueryEx2
	// @ID Requests.QueryEx2
	// @Param   query    query   RequestQueryEx2     false  "request query param"
	// @Param   offset   query   int     false  "offset"
	// @Param   limit    query   int     false  "limit"
	// @Accept  json
	// @Produce  json
	// @Router /requests/queryex2 [get]
	QueryEx2(ctx context.Context, query *RequestQueryEx2, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @Summary QueryEx3
	// @Description QueryEx3
	// @ID Requests.QueryEx3
	// @Param   query    query   RequestQueryEx3     false  "request query param"
	// @Param   offset   query   int     false  "offset"
	// @Param   limit    query   int     false  "limit"
	// @Accept  json
	// @Produce  json
	// @Router /requests/queryex3 [get]
	QueryEx3(ctx context.Context, query *RequestQueryEx3, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @Summary QueryEx4
	// @Description QueryEx4
	// @ID Requests.QueryEx4
	// @Param   query    query   RequestQueryEx4     false  "request query param"
	// @Param   offset   query   int     false  "offset"
	// @Param   limit    query   int     false  "limit"
	// @Accept  json
	// @Produce  json
	// @Router /requests/queryex4 [get]
	QueryEx4(ctx context.Context, query *RequestQueryEx4, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @Summary QueryEx3NoPrefix
	// @Description QueryEx3NoPrefix
	// @ID Requests.QueryEx3NoPrefix
	// @Param   query    query   RequestQueryEx3     false  "request query param" extensions(x-gogen-extend=inline)
	// @Param   offset   query   int     false  "offset"
	// @Param   limit    query   int     false  "limit"
	// @Accept  json
	// @Produce  json
	// @Router /requests/queryex3/NoPrefix [get]
	QueryEx3NoPrefix(ctx context.Context, query *RequestQueryEx3, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @Summary QueryEx4NoPrefix
	// @Description QueryEx4NoPrefix
	// @ID Requests.QueryEx4NoPrefix
	// @Param   query    query   RequestQueryEx3     false  "request query param" extensions(x-gogen-extend=inline)
	// @Param   offset   query   int     false  "offset"
	// @Param   limit    query   int     false  "limit"
	// @Accept  json
	// @Produce  json
	// @Router /requests/queryex4/NoPrefix [get]
	QueryEx4NoPrefix(ctx context.Context, query *RequestQueryEx4, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @Summary List
	// @Description List
	// @ID Requests.List
	// @Param   query    query   models.RequestQuery     false  "request query param"
	// @Param   offset   query   int     false  "offset"
	// @Param   limit    query   int     false  "limit"
	// @Accept  json
	// @Produce  json
	// @Router /requests [get]
	List(ctx context.Context, query *models.RequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)

	// @Summary Create1
	// @Description Create1
	// @ID Requests.Create1
	// @Param   data    body   models.Request     false  "request query param" extensions(x-gogen-entire-body=true)
	// @Accept  json
	// @Produce  json
	// @Router /requests/create1 [post]
	Create1(ctx context.Context, data *models.Request) (int64, error)

	// @Summary Create2
	// @Description Create2
	// @ID Requests.Create2
	// @Param   request    body   models.Request     false  "request query param"
	// @Param   testarg  body   int     false  "testarg"
	// @Accept  json
	// @Produce  json
	// @Router /requests/create2 [post]
	Create2(ctx context.Context, request *models.Request, testarg int64) (int64, error)

	// @Summary UpdateByID
	// @Description UpdateByID
	// @ID Requests.UpdateByID
	// @Param   data    body   models.Request     false  "request query param" extensions(x-gogen-entire-body=true)
	// @Param   id  path   int     false  "id"
	// @Accept  json
	// @Produce  json
	// @Router /requests/{id} [put]
	UpdateByID(ctx context.Context, id int64, data *models.Request) (int64, error)

	// @Summary Set1ByID
	// @Description Set1ByID
	// @ID Requests.Set1ByID
	// @Param   params     body   map[string]string     false  "params"
	// @Param   id  path   int     false  "id"
	// @Accept  json
	// @Produce  json
	// @Router /requests/set1/{id} [patch]
	Set1ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @Summary Set2ByID
	// @Description Set2ByID
	// @ID Requests.Set2ByID
	// @Param   id         path   int     false  "id"
	// @Param   params     body   map[string]string     false  "params" extensions(x-gogen-entire-body=true)
	// @Accept  json
	// @Produce  json
	// @Router /requests/set2/{id} [patch]
	Set2ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @Summary Set3ByID
	// @Description Set3ByID
	// @ID Requests.Set3ByID
	// @Param   id         path   int     false  "id"
	// @Param   params     body   map[string]string     false  "params"
	// @Accept  json
	// @Produce  json
	// @Router /requests/set3/{id} [put]
	Set3ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @Summary Set4ByID
	// @Description Set4ByID
	// @ID Requests.Set4ByID
	// @Param   id         path   int     false  "id"
	// @Param   params     body   map[string]string     false  "params" extensions(x-gogen-entire-body=true)
	// @Accept  json
	// @Produce  json
	// @Router /requests/set4/{id} [put]
	Set4ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @Summary Set5ByID
	// @Description Set5ByID
	// @ID Requests.Set5ByID
	// @Param   id         path   int     false  "id"
	// @Param   params     body   map[string]string     false  "params"
	// @Accept  json
	// @Produce  json
	// @Router /requests/set5/{id} [post]
	Set5ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @Summary Set6ByID
	// @Description Set6ByID
	// @ID Requests.Set6ByID
	// @Param   id         path   int     false  "id"
	// @Param   params     body   map[string]string    false  "params" extensions(x-gogen-entire-body=true)
	// @Accept  json
	// @Produce  json
	// @Router /requests/set6/{id} [post]
	Set6ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @Summary Set7ByID, dataType="map[string]string"
	// @Description Set7ByID
	// @ID Requests.Set7ByID
	// @Param   id         path   int     false  "id"
	// @Param   params     body   map[string]interface{}     false  "params" extensions(x-gogen-entire-body=true)
	// @Accept  json
	// @Produce  json
	// @Router /requests/set7/{id} [post]
	Set7ByID(ctx context.Context, id int64, params interface{}) (int64, err error)

	// @Summary QuerySubTest1
	// @Description QuerySubTest1
	// @ID Requests.QuerySubTest1
	// @Param   id         path   int     false  "id"
	// @Param   query     query   SubTest1     false  "params"
	// @Accept  json
	// @Produce  json
	// @Router /requests/querysub1 [get]
	QuerySubTest1(ctx context.Context, query *SubTest1) (requests []map[string]interface{}, err error)

	// @Summary QuerySubTest2
	// @Description QuerySubTest2
	// @ID Requests.QuerySubTest2
	// @Param   id         path   int     false  "id"
	// @Param   query     query   SubTest2     false  "params"
	// @Accept  json
	// @Produce  json
	// @Router /requests/querysub2 [get]
	QuerySubTest2(ctx context.Context, query *SubTest2) (requests []map[string]interface{}, err error)

	// @Summary QuerySubTest3
	// @Description QuerySubTest3
	// @ID Requests.QuerySubTest3
	// @Param   id         path   int     false  "id"
	// @Param   query     query   SubTest3     false  "params"
	// @Accept  json
	// @Produce  json
	// @Router /requests/querysub3 [get]
	QuerySubTest3(ctx context.Context, query *SubTest3) (requests []map[string]interface{}, err error)

	// @Summary QuerySubTest4
	// @Description QuerySubTest4
	// @ID Requests.QuerySubTest4
	// @Param   id         path   int     false  "id"
	// @Param   query     query   SubTest4     false  "params"
	// @Accept  json
	// @Produce  json
	// @Router /requests/querysub4 [get]
	QuerySubTest4(ctx context.Context, query *SubTest4) (requests []map[string]interface{}, err error)
}
