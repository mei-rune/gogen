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

	"github.com/runner-mei/gogen/gentest/models"
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
	Start, End time.Time
}

type TimeRange2 struct {
	Start, End *time.Time
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
	// @http.GET(path="/files")
	GetFiles(filenames []string) (list []string, total int64, err error)

	// @http.GET(path="/times")
	GetTimes(times []time.Time) (list []string, total int64, err error)

	// @http.GET(path="/allfiles")
	GetAllFiles() (list []string, total int64, err error)

	// @http.GET(path="/test_by_key/:key")
	TestByKey1(key Key) error

	// @http.GET(path="/test_by_key")
	TestByKey2(key Key) error

	// @http.GET(path="/test_by_strkey/:key")
	TestByStrKey1(key StrKey) error

	// @http.GET(path="/test_by_strkey")
	TestByStrKey2(key StrKey) error

	// @http.GET(path="/test64/:id")
	TestInt64Path(id int64) error

	// @http.GET(path="/test64")
	TestInt64Query(id int64) error

	// @http.GET(path="/test_query_args1/:id")
	TestQueryArgs1(id int64, args QueryArgs) error

	// @http.GET(path="/test_query_args2/:id")
	TestQueryArgs2(id int64, args *QueryArgs) error

	// @http.GET(path="/test_query_args3/:id?args=<none>")
	TestQueryArgs3(id int64, args QueryArgs) error

	// @http.GET(path="/test_query_args4/:id?<none>=args")
	TestQueryArgs4(id int64, args *QueryArgs) error

	// @http.GET(path="/ping")
	Ping() error

	// @http.GET(path="/echo")
	Echo(a string) string

	// @http.POST(path="/echo2", data="body")
	EchoBody(body io.Reader) (string, error)

	// @http.POST(path="/echo3")
	Echo3(context context.Context, a string) (string, error)

	// @http.GET(path="/concat")
	Concat0(a, b string) (string, error)

	// @http.GET(path="/concat1")
	Concat1(a, b *string) (string, error)

	// @http.GET(path="/concat2/:a/:b")
	Concat2(a, b string) (string, error)

	// @http.GET(path="/concat3/:a/:b")
	Concat3(a, b *string) (string, error)

	// @http.GET(path="/sub")
	Sub(a string, start int64) (string, error)

	// @http.POST(path="/save/:a", data="b")
	Save(a, b string) (string, error)

	// @http.POST(path="/save2/:a", data="b")
	Save2(a, b *string) (string, error)

	// @http.POST(path="/save3")
	Save3(a, b *string) (string, error)

	// @http.POST(path="/save4")
	Save4(a, b string) (string, error)

	// @http.POST(path="/save5")
	Save5(context context.Context, a, b string) (string, error)

	// @http.GET(path="/add/:a/:b")
	Add1(a, b int) (int, error)

	// @http.GET(path="/add2/:a/:b")
	Add2(a, b *int) (int, error)

	// @http.GET(path="/add3")
	Add3(a, b *int) (int, error)

	// @http.GET(path="/query1")
	Query1(a string, beginAt, endAt time.Time, isRaw bool) string

	// @http.GET(path="/query2/:isRaw")
	Query2(a string, beginAt, endAt time.Time, isRaw bool) string

	// @http.GET(path="/query3/:isRaw")
	Query3(a string, beginAt, endAt time.Time, isRaw *bool) string

	// @http.GET(path="/query4/:isRaw")
	Query4(a string, createdAt TimeRange, isRaw *bool) string

	// @http.GET(path="/query5/:isRaw")
	Query5(a string, createdAt *TimeRange, isRaw *bool) string

	// @http.GET(path="/query6/:isRaw")
	Query6(a string, createdAt TimeRange2, isRaw *bool) string

	// @http.GET(path="/query7/:isRaw")
	Query7(a string, createdAt *TimeRange2, isRaw *bool) string

	// @http.GET(path="/query8", content_type="text")
	Query8(ctx context.Context, itemID int64) (string, error)

	// @http.POST(path="", noreturn="true")
	CreateWithNoReturn(ctx context.Context, request *http.Request, response http.ResponseWriter) error

	// @http.GET(path="/query9", content_type="text")
	Query9(ctx context.Context, itemID sql.NullInt64) (string, error)

	// @http.GET(path="/query10", content_type="text")
	Query10(ctx context.Context, itemID sql.NullString) (string, error)

	// @http.GET(path="/query11", content_type="text")
	Query11(ctx context.Context, itemID sql.NullBool) (string, error)

	// @http.GET(path="/query12?Name=Name", content_type="text")
	Query1WithUpName(ctx context.Context, Name string) (string, error)

	// @http.POST(path="/query12", auto_underscore="false")
	Set1WithUpName(ctx context.Context, Name string) error

	Misc() string
}

var _ StringSvc = &StringSvcImpl{}

type StringSvcImpl struct{}

// @http.GET(path="/impl/files")
func (svc *StringSvcImpl) GetFiles(filenames []string) (list []string, total int64, err error) {
	return []string{"a.txt"}, 10, nil
}

// @http.GET(path="/impl/times")
func (svc *StringSvcImpl) GetTimes(times []time.Time) (list []string, total int64, err error) {
	return []string{"a.txt"}, 10, nil
}

// @http.GET(path="/impl/allfiles")
func (svc *StringSvcImpl) GetAllFiles() (list []string, total int64, err error) {
	return []string{"abc"}, 1, nil
}

// @http.GET(path="/impl/test_by_key/:key")
func (svc *StringSvcImpl) TestByKey1(key Key) error {
	return nil
}

// @http.GET(path="/impl/test_by_key")
func (svc *StringSvcImpl) TestByKey2(key Key) error {
	return nil
}

// @http.GET(path="/impl/test_by_strkey/:key")
func (svc *StringSvcImpl) TestByStrKey1(key StrKey) error {
	return nil
}

// @http.GET(path="/impl/test_by_strkey/")
func (svc *StringSvcImpl) TestByStrKey2(key StrKey) error {
	return nil
}

// @http.GET(path="/impl/test64/:id")
func (svc *StringSvcImpl) TestInt64Path(id int64) error {
	return nil
}

// @http.GET(path="/impl/test64")
func (svc *StringSvcImpl) TestInt64Query(id int64) error {
	return nil
}

// @http.GET(path="/impl/test_query_args1/:id")
func (svc *StringSvcImpl) TestQueryArgs1(id int64, args QueryArgs) error {
	return nil
}

// @http.GET(path="/impl/test_query_args2/:id")
func (svc *StringSvcImpl) TestQueryArgs2(id int64, args *QueryArgs) error {
	return nil
}

// @http.GET(path="/impl/test_query_args3/:id?args=<none>")
func (svc *StringSvcImpl) TestQueryArgs3(id int64, args QueryArgs) error {
	return nil
}

// @http.GET(path="/impl/test_query_args4/:id?<none>=args")
func (svc *StringSvcImpl) TestQueryArgs4(id int64, args *QueryArgs) error {
	return nil
}

// @http.GET(path="/impl/ping")
func (svc *StringSvcImpl) Ping() error {
	return nil
}

// @http.GET(path="/impl/echo")
func (svc *StringSvcImpl) Echo(a string) string {
	return a
}

// @http.POST(path="/impl/echo2", data="body")
func (svc *StringSvcImpl) EchoBody(body io.Reader) (string, error) {
	bs, err := ioutil.ReadAll(body)
	return string(bs), err
}

// @http.POST(path="/impl/echo3")
func (svc *StringSvcImpl) Echo3(context context.Context, a string) (string, error) {
	return a, nil
}

// @http.GET(path="/impl/concat")
func (svc *StringSvcImpl) Concat0(a, b string) (string, error) {
	return a + b, nil
}

// @http.GET(path="/impl/concat1")
func (svc *StringSvcImpl) Concat1(a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.GET(path="/impl/concat2/:a/:b")
func (svc *StringSvcImpl) Concat2(a, b string) (string, error) {
	return a + b, nil
}

// @http.GET(path="/impl/concat3/:a/:b")
func (svc *StringSvcImpl) Concat3(a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.GET(path="/impl/sub")
func (svc *StringSvcImpl) Sub(a string, start int64) (string, error) {
	return a[start:], nil
}

// @http.POST(path="/impl/save/:a", data="b")
func (svc *StringSvcImpl) Save(a, b string) (string, error) {
	return "", nil
}

// @http.POST(path="/impl/save2/:a", data="b")
func (svc *StringSvcImpl) Save2(a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.POST(path="/impl/save3")
func (svc *StringSvcImpl) Save3(a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.POST(path="/impl/save4")
func (svc *StringSvcImpl) Save4(a, b string) (string, error) {
	return a + b, nil
}

// @http.POST(path="/impl/save5")
func (svc *StringSvcImpl) Save5(context context.Context, a, b string) (string, error) {
	return a + b, nil
}

// @http.GET(path="/impl/add/:a/:b")
func (svc *StringSvcImpl) Add1(a, b int) (int, error) {
	return a + b, nil
}

// @http.GET(path="/impl/add2/:a/:b")
func (svc *StringSvcImpl) Add2(a, b *int) (int, error) {
	return *a + *b, nil
}

// @http.GET(path="/impl/add3")
func (svc *StringSvcImpl) Add3(a, b *int) (int, error) {
	return *a + *b, nil
}

// @http.GET(path="/impl/query1")
func (svc *StringSvcImpl) Query1(a string, beginAt, endAt time.Time, isRaw bool) string {
	return "queue"
}

// @http.GET(path="/impl/query2/:isRaw")
func (svc *StringSvcImpl) Query2(a string, beginAt, endAt time.Time, isRaw bool) string {
	return "queue"
}

// @http.GET(path="/impl/query3/:isRaw")
func (svc *StringSvcImpl) Query3(a string, beginAt, endAt time.Time, isRaw *bool) string {
	return "queue"
}

// @http.GET(path="/impl/query4/:isRaw")
func (svc *StringSvcImpl) Query4(a string, createdAt TimeRange, isRaw *bool) string {
	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
}

// @http.GET(path="/impl/query5/:isRaw")
func (svc *StringSvcImpl) Query5(a string, createdAt *TimeRange, isRaw *bool) string {
	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
}

// @http.GET(path="/impl/query6/:isRaw")
func (svc *StringSvcImpl) Query6(a string, createdAt TimeRange2, isRaw *bool) string {
	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
}

// @http.GET(path="/impl/query7/:isRaw")
func (svc *StringSvcImpl) Query7(a string, createdAt *TimeRange2, isRaw *bool) string {
	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
}

// @http.GET(path="/impl/query8", content_type="text")
func (svc *StringSvcImpl) Query8(ctx context.Context, itemID int64) (string, error) {
	return "queue:" + strconv.FormatInt(itemID, 10), nil
}

// @http.POST(path="/impl", noreturn="true")
func (svc *StringSvcImpl) CreateWithNoReturn(ctx context.Context, request *http.Request, response http.ResponseWriter) error {
	return nil
}

// @http.GET(path="/impl/query9", content_type="text")
func (svc *StringSvcImpl) Query9(ctx context.Context, itemID sql.NullInt64) (string, error) {
	return "query9", nil
}

// @http.GET(path="/impl/query10", content_type="text")
func (svc *StringSvcImpl) Query10(ctx context.Context, itemID sql.NullString) (string, error) {
	return "query10", nil
}

// @http.GET(path="/impl/query11", content_type="text")
func (svc *StringSvcImpl) Query11(ctx context.Context, itemID sql.NullBool) (string, error) {
	return "query11", nil
}

// @http.GET(path="/impl/query12?Name=Name", content_type="text")
func (svc *StringSvcImpl) Query1WithUpName(ctx context.Context, Name string) (string, error) {
	return "Query1WithUpName", nil
}

// @http.POST(path="/impl/query12", auto_underscore="false")
func (svc *StringSvcImpl) Set1WithUpName(ctx context.Context, Name string) error {
	return nil
}

func (svc *StringSvcImpl) Misc() string {
	return ""
}

type StringSvcWithContext struct {
}

// @http.GET(path="/ctx/echo")
func (svc *StringSvcWithContext) Echo(ctx context.Context, a string) string {
	return a
}

// @http.POST(path="/ctx/echo2", data="body")
func (svc *StringSvcWithContext) EchoBody(ctx context.Context, body io.Reader) (string, error) {
	bs, err := ioutil.ReadAll(body)
	return string(bs), err
}

// @http.GET(path="/ctx/concat")
func (svc *StringSvcWithContext) Concat0(ctx context.Context, a, b string) (string, error) {
	return a + b, nil
}

// @http.GET(path="/ctx/concat1")
func (svc *StringSvcWithContext) Concat1(ctx context.Context, a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.GET(path="/ctx/concat2/:a/:b")
func (svc *StringSvcWithContext) Concat2(ctx context.Context, a, b string) (string, error) {
	return a + b, nil
}

// @http.GET(path="/ctx/concat3/:a/:b")
func (svc *StringSvcWithContext) Concat3(ctx context.Context, a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.GET(path="/ctx/sub")
func (svc *StringSvcWithContext) Sub(ctx context.Context, a string, start int64) (string, error) {
	return a[start:], nil
}

// @http.POST(path="/ctx/save1/:a", data="b")
func (svc *StringSvcWithContext) Save1(ctx context.Context, a, b string) (string, error) {
	return "", nil
}

// @http.POST(path="/ctx/save2/:a", data="b")
func (svc *StringSvcWithContext) Save2(ctx context.Context, a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.POST(path="/ctx/save3")
func (svc *StringSvcWithContext) Save3(a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.GET(path="/ctx/add1/:a/:b")
func (svc *StringSvcWithContext) Add1(ctx context.Context, a, b int) (int, error) {
	return a + b, nil
}

// @http.GET(path="/ctx/add2/:a/:b")
func (svc *StringSvcWithContext) Add2(ctx context.Context, a, b *int) (int, error) {
	return *a + *b, nil
}

// @http.GET(path="/ctx/add3")
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
	// @http.GET(path="/requests/query1")
	Query1(ctx context.Context, query *models.RequestQuery, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/requests/query2?query=<none>")
	Query2(ctx context.Context, query *models.RequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)

	// @http.GET(path="/requests/query3?query=<none>")
	Query3(ctx context.Context, query *AliasRequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)

	// @http.GET(path="/requests/queryex1")
	QueryEx1(ctx context.Context, query *RequestQueryEx1, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/requests/queryex2")
	QueryEx2(ctx context.Context, query *RequestQueryEx2, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/requests/queryex3")
	QueryEx3(ctx context.Context, query *RequestQueryEx3, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/requests/queryex4")
	QueryEx4(ctx context.Context, query *RequestQueryEx4, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/requests/queryex3/NoPrefix?query=<none>")
	QueryEx3NoPrefix(ctx context.Context, query *RequestQueryEx3, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/requests/queryex4/NoPrefix?query=<none>")
	QueryEx4NoPrefix(ctx context.Context, query *RequestQueryEx4, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/requests")
	List(ctx context.Context, query *models.RequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)
	// @http.POST(path="/requests", data="data")
	Create(ctx context.Context, data *models.Request) (int64, error)

	// @http.POST(path="/requests")
	Create2(ctx context.Context, request *models.Request, testarg int64) (int64, error)

	// @http.PUT(path="/requests/:id", data="data")
	UpdateByID(ctx context.Context, id int64, data *models.Request) (int64, error)

	// @http.PATCH(path="/requests/:id")
	Set1ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @http.PATCH(path="/requests/:id", data="params")
	Set2ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @http.PUT(path="/requests/:id")
	Set3ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @http.PUT(path="/requests/:id", data="params")
	Set4ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @http.POST(path="/requests/:id/5")
	Set5ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @http.POST(path="/requests/:id/6", data="params")
	Set6ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @http.POST(path="/requests/:id/7", data="params", dataType="map[string]string")
	Set7ByID(ctx context.Context, id int64, params interface{}) (int64, err error)

	// @http.GET(path="/requests/querysub1")
	QuerySubTest1(ctx context.Context, query *SubTest1) (requests []map[string]interface{}, err error)

	// @http.GET(path="/requests/querysub2")
	QuerySubTest2(ctx context.Context, query *SubTest2) (requests []map[string]interface{}, err error)

	// @http.GET(path="/requests/querysub3")
	QuerySubTest3(ctx context.Context, query *SubTest3) (requests []map[string]interface{}, err error)

	// @http.GET(path="/requests/querysub4")
	QuerySubTest4(ctx context.Context, query *SubTest4) (requests []map[string]interface{}, err error)
}
