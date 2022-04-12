package main

import (
	"strings"
	"time"
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
	fint    int
	fstring string
	ftime   time.Time

	fintptr    *int
	fstringptr *string
	ftimeptr   *time.Time
}

// @http.Client(name="TestClient", ref="true")
type StringSvc interface {

	// @Summary test by name
	// @Description test by int64 ID
	// @ID TestCase1
	// @Accept  json
	// @Produce  json
	// @Param   name      path   string     true  "Some ID" Format(string)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_name/{name} [get]
	TestCase1(name string) error


	// @Summary test by name
	// @Description test by int64 ID
	// @ID TestCase2_1
	// @Accept  json
	// @Produce  json
	// @Param   name      query   string     true  "Some ID" Format(string)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_name [get]
	TestCase2_1(name string) error


	// @Summary test by names
	// @Description test by int64 ID
	// @ID TestCase2_2
	// @Accept  json
	// @Produce  json
	// @Param   name      query   []string     true  "Some ID" Format(string)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_names [get]
	TestCase2_2(name []string) error


	// @Summary test by int64 ID
	// @Description test by int64 ID
	// @ID TestCase3_1
	// @Accept  json
	// @Produce  json
	// @Param   id      path   int     true  "Some ID" Format(int64)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id/{id} [get]
	TestCase3_1(id int64) error


	// @Summary test by int32 ID
	// @Description test by int64 ID
	// @ID TestCase3_2
	// @Accept  json
	// @Produce  json
	// @Param   id      path   int     true  "Some ID" Format(int32)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id/{id} [get]
	TestCase3_2(id int32) error


	// @Summary test by int ID
	// @Description test by int64 ID
	// @ID TestCase3_3
	// @Accept  json
	// @Produce  json
	// @Param   id      path   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id/{id} [get]
	TestCase3_3(id int) error


	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以不好测试， iris 才支持)
	// @Description test by int64 ID
	// @ID TestCase4
	// @Accept  json
	// @Produce  json
	// @Param   id      path   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id/{id} [get]
	TestCase4(id int) error


	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以可以正常测试， iris 无法测试)
	// @Description test by int64 ID
	// @ID TestCase5
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase5_1(id int64) error


	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以可以正常测试， iris 无法测试)
	// @Description test by int64 ID
	// @ID TestCase5
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase5_2(id int32) error


	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以可以正常测试， iris 无法测试)
	// @Description test by int64 ID
	// @ID TestCase5
	// @Accept  json
	// @Produce  json
	// @Param   idlist      query   []int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase5_3(idlist []int64) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以无法测试， iris 可以测试)
	// @Description test by int64 ID
	// @ID TestCase6
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase6(id int64) error


	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以可以测试， iris 不能正确测试)
	// @Description test by int64 ID
	// @ID TestCase7_1
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase7_1(id sql.NullInt64) error


	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以可以测试， iris 不能正确测试)
	// @Description test by int32 ID
	// @ID TestCase7_2
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase7_2(id sql.NullInt32) error


	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以不能正确地测试， iris 能正确测试)
	// @Description test by int64 ID
	// @ID TestCase8
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase8(id sql.NullInt64) error


	// @Summary test by int ID
	// @Description test by int64 ID
	// @ID TestCase9
	// @Accept  json
	// @Produce  json
	// @Param   id      path   string     true  "Some ID" Format(string)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id/{id} [get]
	TestCase9(id *string) error

		// @Summary test by id
	// @Description test by string ID
	// @ID TestCase10
	// @Accept  json
	// @Produce  json
	// @Param   id      query   string     true  "Some ID" Format(string)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_name [get]
	TestCase10(id *string) error


	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以不能正确地测试， iris 能正确测试)
	// @Description test by int64 ID
	// @ID TestCase12
	// @Accept  json
	// @Produce  json
	// @Param   id      path   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/{id} [get]
	TestCase12(id *int) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Param，所以可以正确地测试， iris 不能正确测试)
	// @Description test by int64 ID
	// @ID TestCase13
	// @Accept  json
	// @Produce  json
	// @Param   id      path   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/{id} [get]
	TestCase13(id *int) error


	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Query，所以可以正确地测试， iris 不能正确测试)
	// @Description test by int64 ID
	// @ID TestCase14_1
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_id [get]
	TestCase14_1(id *int) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Query，所以可以正确地测试， iris 不能正确测试)
	// @Description test by int64 ID
	// @ID TestCase14_2
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_name [get]
	TestCase14_2(id *int32) error

	// @Summary test by int ID (注意 gin, chi 不支持 GetInt64Query，所以不能正确地测试， iris 可以正确测试)
	// @Description test by int64 ID
	// @ID TestCase14_3
	// @Accept  json
	// @Produce  json
	// @Param   id      query   int     true  "Some ID" Format(int)
	// @Success 200 {string} string	"ok"
	// @Failure 400 {object} string "We need ID!!"
	// @Failure 404 {object} string "Can not find ID"
	// @Router /test64/by_name [get]
	TestCase14_3(id *int) error

	// // @http.GET(path="/files")
	// GetFiles(filenames []string) (list []string, total int64, err error)

	// // @http.GET(path="/times")
	// GetTimes(times []time.Time) (list []string, total int64, err error)

	// // @http.GET(path="/allfiles")
	// GetAllFiles() (list []string, total int64, err error)

	// // @http.GET(path="/test_by_key")
	// TestByKey(key Key) error

	// // @http.GET(path="/test64")
	// TestInt64Query(id int64) error

	// // @http.GET(path="/test_query_args1/:id")
	// TestQueryArgs1(id int64, args QueryArgs) error

	// // @http.GET(path="/test_query_args2/:id")
	// TestQueryArgs2(id int64, args *QueryArgs) error

	// // @http.GET(path="/test_query_args3/:id?args=<none>")
	// TestQueryArgs3(id int64, args QueryArgs) error

	// // @http.GET(path="/test_query_args4/:id?<none>=args")
	// TestQueryArgs4(id int64, args *QueryArgs) error

	// // @http.GET(path="/ping")
	// Ping() error

	// // @http.GET(path="/echo")
	// Echo(a string) string

	// // @http.POST(path="/echo2", data="body")
	// EchoBody(body io.Reader) (string, error)

	// // @http.POST(path="/echo3")
	// Echo3(context context.Context, a string) (string, error)

	// // @http.GET(path="/concat")
	// Concat(a, b string) (string, error)

	// // @http.GET(path="/concat1")
	// Concat1(a, b *string) (string, error)

	// // @http.GET(path="/concat2/:a/:b")
	// Concat2(a, b string) (string, error)

	// // @http.GET(path="/concat3/:a/:b")
	// Concat3(a, b *string) (string, error)

	// // @http.GET(path="/sub")
	// Sub(a string, start int64) (string, error)

	// // @http.POST(path="/save/:a", data="b")
	// Save(a, b string) (string, error)

	// // @http.POST(path="/save2/:a", data="b")
	// Save2(a, b *string) (string, error)

	// // @http.POST(path="/save3")
	// Save3(a, b *string) (string, error)

	// // @http.POST(path="/save4")
	// Save4(a, b string) (string, error)

	// // @http.POST(path="/save5")
	// Save5(context context.Context, a, b string) (string, error)

	// // @http.GET(path="/add/:a/:b")
	// Add(a, b int) (int, error)

	// // @http.GET(path="/add2/:a/:b")
	// Add2(a, b *int) (int, error)

	// // @http.GET(path="/add3")
	// Add3(a, b *int) (int, error)

	// // @http.GET(path="/query1")
	// Query1(a string, beginAt, endAt time.Time, isRaw bool) string

	// // @http.GET(path="/query2/:isRaw")
	// Query2(a string, beginAt, endAt time.Time, isRaw bool) string

	// // @http.GET(path="/query3/:isRaw")
	// Query3(a string, beginAt, endAt time.Time, isRaw *bool) string

	// // @http.GET(path="/query4/:isRaw")
	// Query4(a string, createdAt TimeRange, isRaw *bool) string

	// // @http.GET(path="/query5/:isRaw")
	// Query5(a string, createdAt *TimeRange, isRaw *bool) string

	// // @http.GET(path="/query6/:isRaw")
	// Query6(a string, createdAt TimeRange2, isRaw *bool) string

	// // @http.GET(path="/query7/:isRaw")
	// Query7(a string, createdAt *TimeRange2, isRaw *bool) string

	// // @http.GET(path="/query8", content_type="text")
	// Query8(ctx context.Context, itemID int64) (string, error)

	// // @http.POST(path="", noreturn="true")
	// Create3(ctx context.Context, request *http.Request, response http.ResponseWriter) error

	// // @http.GET(path="/query9", content_type="text")
	// Query9(ctx context.Context, itemID sql.NullInt64) (string, error)

	// // @http.GET(path="/query10", content_type="text")
	// Query10(ctx context.Context, itemID sql.NullString) (string, error)

	// // @http.GET(path="/query11", content_type="text")
	// Query11(ctx context.Context, itemID sql.NullBool) (string, error)

	// // @http.GET(path="/query12?Name=Name", content_type="text")
	// Query1WithUpName(ctx context.Context, Name string) (string, error)

	// // @http.POST(path="/query12", auto_underscore="false")
	// Set1WithUpName(ctx context.Context, Name string) error

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

// // 用于测试 Parse() 不会 panic
// func notpanic() {}

// type AliasRequestQuery = models.RequestQuery

// type RequestQueryEx1 struct {
// 	models.RequestQuery

// 	ExArg string
// }

// type RequestQueryEx2 struct {
// 	*models.RequestQuery

// 	ExArg string
// }

// type RequestQueryEx3 struct {
// 	Request models.RequestQuery `json:"request"`

// 	ExArg string
// }

// type RequestQueryEx4 struct {
// 	Request *models.RequestQuery `json:"request"`

// 	ExArg string
// }

// type Sub1 struct {
// 	A1 string `json:"a1"`
// 	A2 string `json:"a2"`
// }

// type Sub2 struct {
// 	B1 string `json:"b1"`
// 	B2 string `json:"b2"`
// }

// type Sub3 struct {
// 	Sub2
// }

// type SubTest1 struct {
// 	Sub1 Sub1 `json:"sub1"`
// 	Sub2 Sub2 `json:"sub2"`

// 	ExArg string
// }

// type SubTest2 struct {
// 	Sub1 *Sub1 `json:"sub1"`
// 	Sub2 *Sub2 `json:"sub2"`

// 	ExArg string
// }

// type SubTest3 struct {
// 	SubTest2

// 	ExArg string
// }

// type SubTest4 struct {
// 	*SubTest2

// 	ExArg string
// }

// type Requests interface {
// 	// @http.GET(path="/query")
// 	Query1(ctx context.Context, query *models.RequestQuery, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

// 	// @http.GET(path="/query2?query=<none>")
// 	Query2(ctx context.Context, query *models.RequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)

// 	// @http.GET(path="/query3?query=<none>")
// 	Query3(ctx context.Context, query *AliasRequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)

// 	// @http.GET(path="/queryex1")
// 	QueryEx1(ctx context.Context, query *RequestQueryEx1, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

// 	// @http.GET(path="/queryex2")
// 	QueryEx2(ctx context.Context, query *RequestQueryEx2, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

// 	// @http.GET(path="/queryex3")
// 	QueryEx3(ctx context.Context, query *RequestQueryEx3, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

// 	// @http.GET(path="/queryex4")
// 	QueryEx4(ctx context.Context, query *RequestQueryEx4, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

// 	// @http.GET(path="/queryex3/NoPrefix?query=<none>")
// 	QueryEx3NoPrefix(ctx context.Context, query *RequestQueryEx3, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

// 	// @http.GET(path="/queryex4/NoPrefix?query=<none>")
// 	QueryEx4NoPrefix(ctx context.Context, query *RequestQueryEx4, offset, limit int64, params map[string]string) (requests []map[string]interface{}, err error)

// 	// @http.GET(path="")
// 	List(ctx context.Context, query *models.RequestQuery, offset, limit int64) (requests []map[string]interface{}, err error)
// 	// @http.POST(path="", data="data")
// 	Create(ctx context.Context, data *models.Request) (int64, error)

// 	// @http.POST(path="")
// 	Create2(ctx context.Context, request *models.Request, testarg int64) (int64, error)

// 	// @http.PUT(path="/:id", data="data")
// 	UpdateByID(ctx context.Context, id int64, data *models.Request) (int64, error)

// 	// @http.PATCH(path="/:id")
// 	Set1ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

// 	// @http.PATCH(path="/:id", data="params")
// 	Set2ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

// 	// @http.PUT(path="/:id")
// 	Set3ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

// 	// @http.PUT(path="/:id", data="params")
// 	Set4ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

// 	// @http.POST(path="/:id/5")
// 	Set5ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

// 	// @http.POST(path="/:id/6", data="params")
// 	Set6ByID(ctx context.Context, id int64, params map[string]string) (int64, err error)

// 	// @http.POST(path="/:id/7", data="params", dataType="map[string]string")
// 	Set7ByID(ctx context.Context, id int64, params interface{}) (int64, err error)

// 	// @http.GET(path="/querysub1")
// 	QuerySubTest1(ctx context.Context, query *SubTest1) (requests []map[string]interface{}, err error)

// 	// @http.GET(path="/querysub2")
// 	QuerySubTest2(ctx context.Context, query *SubTest2) (requests []map[string]interface{}, err error)

// 	// @http.GET(path="/querysub3")
// 	QuerySubTest3(ctx context.Context, query *SubTest3) (requests []map[string]interface{}, err error)

// 	// @http.GET(path="/querysub4")
// 	QuerySubTest4(ctx context.Context, query *SubTest4) (requests []map[string]interface{}, err error)
// }
