package main

import (
	"context"
	"io"
	"io/ioutil"
)

type StringSvc interface {
	// @http.GET(path="/echo")
	Echo(a string) string

	// @http.GET(path="/echo", data="body")
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

	// @http.GET(path="/add/:a/:b")
	Add(a, b int) (int, error)

	// @http.GET(path="/add2/:a/:b")
	Add2(a, b *int) (int, error)

	// @http.GET(path="/add3")
	Add3(a, b *int) (int, error)

	Misc() string
}

var _ StringSvc = &StringSvcImpl{}

type StringSvcImpl struct {
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
