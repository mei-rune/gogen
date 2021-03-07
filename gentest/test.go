package main

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/runner-mei/gogen/gentest/models"
)

// 用于测试 parse() 方法
type Key int64

func (key Key) String() string {
	return ""
}

const TimeFormat = time.RFC3339

func BoolToString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func toBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true"
}

func toDatetime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}

type TimeRange struct {
	Start, End time.Time
}

type TimeRange2 struct {
	Start, End *time.Time
}

type QueryArgs struct {
	fint    int
	fstring string
	ftime   time.Time

	fintptr    *int
	fstringptr *string
	ftimeptr   *time.Time
}

// @http.Client(name="TestClient", ref="true")
type StringSvc interface {
	// @http.GET(path="/allfiles")
	GetAllFiles() (list []string, total int64, err error)

	// @http.GET(path="/test_by_key")
	TestByKey(key Key) error

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
	Concat(a, b string) (string, error)

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
	Add(a, b int) (int, error)

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
	Create3(ctx context.Context, request *http.Request, response http.ResponseWriter) error

	Misc() string
}

var _ StringSvc = &StringSvcImpl{}

type StringSvcImpl struct {
}

// @http.GET(path="/test_by_key")
func (svc *StringSvcImpl) TestByKey(key Key) error {
	return nil
}

// @http.GET(path="/allfiles")
func (svc *StringSvcImpl) GetAllFiles() (list []string, total int64, err error) {
	return []string{"abc"}, 1, nil
}

// @http.GET(path="/test64/:id")
func (svc *StringSvcImpl) TestInt64Path(id int64) error {
	return nil
}

// @http.GET(path="/test64")
func (svc *StringSvcImpl) TestInt64Query(id int64) error {
	return nil
}

// @http.GET(path="/test_query_args1/:id")
func (svc *StringSvcImpl) TestQueryArgs1(id int64, args QueryArgs) error {
	return nil
}

// @http.GET(path="/test_query_args2/:id")
func (svc *StringSvcImpl) TestQueryArgs2(id int64, args *QueryArgs) error {
	return nil
}

// @http.GET(path="/test_query_args3/:id?args=<none>")
func (svc *StringSvcImpl) TestQueryArgs3(id int64, args QueryArgs) error {
	return nil
}

// @http.GET(path="/test_query_args4/:id?<none>=args")
func (svc *StringSvcImpl) TestQueryArgs4(id int64, args *QueryArgs) error {
	return nil
}

// @http.GET(path="/ping")
func (svc *StringSvcImpl) Ping() error {
	return nil
}

// @http.GET(path="/echo")
func (svc *StringSvcImpl) Echo(a string) string {
	return a
}

// @http.GET(path="/echo_body1", data="body")
func (svc *StringSvcImpl) EchoBody(body io.Reader) (string, error) {
	bs, err := ioutil.ReadAll(body)
	return string(bs), err
}

// @http.POST(path="/echo3")
func (svc *StringSvcImpl) Echo3(context context.Context, a string) (string, error) {
	return a, nil
}

// @http.GET(path="/concat")
func (svc *StringSvcImpl) Concat(a, b string) (string, error) {
	return a + b, nil
}

// @http.GET(path="/concat1")
func (svc *StringSvcImpl) Concat1(a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.GET(path="/concat2/:a/:b")
func (svc *StringSvcImpl) Concat2(a, b string) (string, error) {
	return a + b, nil
}

// @http.GET(path="/concat3/:a/:b")
func (svc *StringSvcImpl) Concat3(a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.GET(path="/sub")
func (svc *StringSvcImpl) Sub(a string, start int64) (string, error) {
	return a[start:], nil
}

// @http.POST(path="/save/:a", data="b")
func (svc *StringSvcImpl) Save(a, b string) (string, error) {
	return "", nil
}

// @http.POST(path="/save2/:a", data="b")
func (svc *StringSvcImpl) Save2(a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.POST(path="/save3")
func (svc *StringSvcImpl) Save3(a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.POST(path="/save4")
func (svc *StringSvcImpl) Save4(a, b string) (string, error) {
	return a + b, nil
}

// @http.POST(path="/save5")
func (svc *StringSvcImpl) Save5(context context.Context, a, b string) (string, error) {
	return a + b, nil
}

// @http.GET(path="/add/:a/:b")
func (svc *StringSvcImpl) Add(a, b int) (int, error) {
	return a + b, nil
}

// @http.GET(path="/add2/:a/:b")
func (svc *StringSvcImpl) Add2(a, b *int) (int, error) {
	return *a + *b, nil
}

// @http.GET(path="/add3")
func (svc *StringSvcImpl) Add3(a, b *int) (int, error) {
	return *a + *b, nil
}

// @http.GET(path="/query1")
func (svc *StringSvcImpl) Query1(a string, beginAt, endAt time.Time, isRaw bool) string {
	return "queue"
}

// @http.GET(path="/query2/:isRaw")
func (svc *StringSvcImpl) Query2(a string, beginAt, endAt time.Time, isRaw bool) string {
	return "queue"
}

// @http.GET(path="/query3/:isRaw")
func (svc *StringSvcImpl) Query3(a string, beginAt, endAt time.Time, isRaw *bool) string {
	return "queue"
}

// @http.GET(path="/query4/:isRaw")
func (svc *StringSvcImpl) Query4(a string, createdAt TimeRange, isRaw *bool) string {
	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
}

// @http.GET(path="/query5/:isRaw")
func (svc *StringSvcImpl) Query5(a string, createdAt *TimeRange, isRaw *bool) string {
	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
}

// @http.GET(path="/query6/:isRaw")
func (svc *StringSvcImpl) Query6(a string, createdAt TimeRange2, isRaw *bool) string {
	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
}

// @http.GET(path="/query7/:isRaw")
func (svc *StringSvcImpl) Query7(a string, createdAt *TimeRange2, isRaw *bool) string {
	return "queue:" + a + ":" + createdAt.Start.Format(time.RFC3339) + "-" + createdAt.End.Format(time.RFC3339)
}

func (svc *StringSvcImpl) Misc() string {
	return ""
}

type StringSvcWithContext struct {
}

// @http.GET(path="/echo")
func (svc *StringSvcWithContext) Echo(ctx context.Context, a string) string {
	return a
}

// @http.GET(path="/echo", data="body")
func (svc *StringSvcWithContext) EchoBody(ctx context.Context, body io.Reader) (string, error) {
	bs, err := ioutil.ReadAll(body)
	return string(bs), err
}

// @http.GET(path="/concat")
func (svc *StringSvcWithContext) Concat(ctx context.Context, a, b string) (string, error) {
	return a + b, nil
}

// @http.GET(path="/concat1")
func (svc *StringSvcWithContext) Concat1(ctx context.Context, a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.GET(path="/concat2/:a/:b")
func (svc *StringSvcWithContext) Concat2(ctx context.Context, a, b string) (string, error) {
	return a + b, nil
}

// @http.GET(path="/concat3/:a/:b")
func (svc *StringSvcWithContext) Concat3(ctx context.Context, a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.GET(path="/sub")
func (svc *StringSvcWithContext) Sub(ctx context.Context, a string, start int64) (string, error) {
	return a[start:], nil
}

// @http.POST(path="/save/:a", data="b")
func (svc *StringSvcWithContext) Save(ctx context.Context, a, b string) (string, error) {
	return "", nil
}

// @http.POST(path="/save2/:a", data="b")
func (svc *StringSvcWithContext) Save2(ctx context.Context, a, b *string) (string, error) {
	return *a + *b, nil
}

// @http.GET(path="/add/:a/:b")
func (svc *StringSvcWithContext) Add(ctx context.Context, a, b int) (int, error) {
	return a + b, nil
}

// @http.GET(path="/add2/:a/:b")
func (svc *StringSvcWithContext) Add2(ctx context.Context, a, b *int) (int, error) {
	return *a + *b, nil
}

// @http.GET(path="/add3")
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

type Requests interface {
	// @http.GET(path="/query")
	Query1(ctx context.Context, query *models.RequestQuery, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/query2?query=<none>")
	Query2(ctx context.Context, query *models.RequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)

	// @http.GET(path="/query3?query=<none>")
	Query3(ctx context.Context, query *AliasRequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)

	// @http.GET(path="/queryex1")
	QueryEx1(ctx context.Context, query *RequestQueryEx1, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/queryex2")
	QueryEx2(ctx context.Context, query *RequestQueryEx2, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/queryex3")
	QueryEx3(ctx context.Context, query *RequestQueryEx3, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/queryex4")
	QueryEx4(ctx context.Context, query *RequestQueryEx4, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/queryex3/NoPrefix?query=<none>")
	QueryEx3NoPrefix(ctx context.Context, query *RequestQueryEx3, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="/queryex4/NoPrefix?query=<none>")
	QueryEx4NoPrefix(ctx context.Context, query *RequestQueryEx4, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

	// @http.GET(path="")
	List(ctx context.Context, query *models.RequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)
	// @http.POST(path="", data="data")
	Create(ctx context.Context, data *models.Request) (int64, error)

	// @http.POST(path="")
	Create2(ctx context.Context, request *models.Request, testarg int64) (int64, error)

	// @http.PUT(path="/:id", data="data")
	UpdateByID(ctx context.Context, id int64, data *models.Request) (int64, error)

	// @http.PATCH(path="/:id")
	Set1ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @http.PATCH(path="/:id", data="params")
	Set2ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @http.PUT(path="/:id")
	Set3ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @http.PUT(path="/:id", data="params")
	Set4ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @http.POST(path="/:id/5")
	Set5ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @http.POST(path="/:id/6", data="params")
	Set6ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

	// @http.POST(path="/:id/7", data="params", dataType="map[string]string")
	Set7ByID(ctx context.Context, id int64, params interface{}) (int64, err error)
}
