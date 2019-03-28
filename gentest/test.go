package main

import (
	"context"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

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

// @http.Client(name="TestClient", ref="true")
type StringSvc interface {
	// @http.GET(path="/ping")
	Ping() error

	// @http.GET(path="/echo")
	Echo(a string) string

	// @http.POST(path="/echo", data="body")
	EchoBody(body io.Reader) (string, error)

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

	// @http.POST(path="/echo5")
	Echo5(context context.Context, a string) (string, error)

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

	Misc() string
}

var _ StringSvc = &StringSvcImpl{}

type StringSvcImpl struct {
}

// @http.GET(path="/ping")
func (svc *StringSvcImpl) Ping() error {
	return nil
}

// @http.GET(path="/echo")
func (svc *StringSvcImpl) Echo(a string) string {
	return a
}

// @http.GET(path="/echo_body", data="body")
func (svc *StringSvcImpl) EchoBody(body io.Reader) (string, error) {
	bs, err := ioutil.ReadAll(body)
	return string(bs), err
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
