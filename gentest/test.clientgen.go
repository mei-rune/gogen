// Please don't edit this file!
package main

import (
	"context"
	"io"
	"strconv"
	"time"
)

type StringSvcClient struct {
	proxy Proxy
}

func (client StringSvcClient) Echo(ctx context.Context, a string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/echo").
		SetParam("a", a).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) EchoBody(ctx context.Context, body io.Reader) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/echo").
		SetBody(body).
		Result(&result)

	err := request.POST(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Concat(ctx context.Context, a string, b string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/concat").
		SetParam("a", a).
		SetParam("b", b).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Concat1(ctx context.Context, a *string, b *string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/concat1")
	if a != nil {
		request = request.SetParam("a", *a)
	}
	if b != nil {
		request = request.SetParam("b", *b)
	}
	request = request.Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Concat2(ctx context.Context, a string, b string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/concat2/"+a+"/"+b+"").
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Concat3(ctx context.Context, a *string, b *string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/concat3/"+*a+"/"+*b+"").
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Sub(ctx context.Context, a string, start int64) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/sub").
		SetParam("a", a).
		SetParam("start", strconv.FormatInt(start, 10)).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Save(ctx context.Context, a string, b string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/save/"+a+"").
		SetBody(b).
		Result(&result)

	err := request.POST(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Save2(ctx context.Context, a *string, b *string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/save2/"+*a+"").
		SetBody(b).
		Result(&result)

	err := request.POST(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Add(ctx context.Context, a int, b int) (int, error) {
	var result int

	request := NewRequest(client.proxy, "/add/"+strconv.FormatInt(int64(a), 10)+"/"+strconv.FormatInt(int64(b), 10)+"").
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Add2(ctx context.Context, a *int, b *int) (int, error) {
	var result int

	request := NewRequest(client.proxy, "/add2/"+strconv.FormatInt(int64(*a), 10)+"/"+strconv.FormatInt(int64(*b), 10)+"").
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Add3(ctx context.Context, a *int, b *int) (int, error) {
	var result int

	request := NewRequest(client.proxy, "/add3")
	if a != nil {
		request = request.SetParam("a", strconv.FormatInt(int64(*a), 10))
	}
	if b != nil {
		request = request.SetParam("b", strconv.FormatInt(int64(*b), 10))
	}
	request = request.Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Query1(ctx context.Context, a string, beginAt time.Time, endAt time.Time, isRaw bool) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/query1").
		SetParam("a", a).
		SetParam("beginAt", beginAt.Format(TimeFormat)).
		SetParam("endAt", endAt.Format(TimeFormat)).
		SetParam("isRaw", BoolToString(isRaw, 10)).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Query2(ctx context.Context, a string, beginAt time.Time, endAt time.Time, isRaw bool) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/query2/"+BoolToString(isRaw, 10)+"").
		SetParam("a", a).
		SetParam("beginAt", beginAt.Format(TimeFormat)).
		SetParam("endAt", endAt.Format(TimeFormat)).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcClient) Query3(ctx context.Context, a string, beginAt time.Time, endAt time.Time, isRaw *bool) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/query3/"+BoolToString(*isRaw, 10)+"").
		SetParam("a", a).
		SetParam("beginAt", beginAt.Format(TimeFormat)).
		SetParam("endAt", endAt.Format(TimeFormat)).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

// Misc: annotation is missing

type StringSvcImplClient struct {
	proxy Proxy
}

func (client StringSvcImplClient) Echo(ctx context.Context, a string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/echo").
		SetParam("a", a).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) EchoBody(ctx context.Context, body io.Reader) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/echo_body").
		SetBody(body).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Concat(ctx context.Context, a string, b string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/concat").
		SetParam("a", a).
		SetParam("b", b).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Concat1(ctx context.Context, a *string, b *string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/concat1")
	if a != nil {
		request = request.SetParam("a", *a)
	}
	if b != nil {
		request = request.SetParam("b", *b)
	}
	request = request.Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Concat2(ctx context.Context, a string, b string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/concat2/"+a+"/"+b+"").
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Concat3(ctx context.Context, a *string, b *string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/concat3/"+*a+"/"+*b+"").
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Sub(ctx context.Context, a string, start int64) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/sub").
		SetParam("a", a).
		SetParam("start", strconv.FormatInt(start, 10)).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Save(ctx context.Context, a string, b string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/save/"+a+"").
		SetBody(b).
		Result(&result)

	err := request.POST(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Save2(ctx context.Context, a *string, b *string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/save2/"+*a+"").
		SetBody(b).
		Result(&result)

	err := request.POST(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Add(ctx context.Context, a int, b int) (int, error) {
	var result int

	request := NewRequest(client.proxy, "/add/"+strconv.FormatInt(int64(a), 10)+"/"+strconv.FormatInt(int64(b), 10)+"").
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Add2(ctx context.Context, a *int, b *int) (int, error) {
	var result int

	request := NewRequest(client.proxy, "/add2/"+strconv.FormatInt(int64(*a), 10)+"/"+strconv.FormatInt(int64(*b), 10)+"").
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Add3(ctx context.Context, a *int, b *int) (int, error) {
	var result int

	request := NewRequest(client.proxy, "/add3")
	if a != nil {
		request = request.SetParam("a", strconv.FormatInt(int64(*a), 10))
	}
	if b != nil {
		request = request.SetParam("b", strconv.FormatInt(int64(*b), 10))
	}
	request = request.Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Query1(ctx context.Context, a string, beginAt time.Time, endAt time.Time, isRaw bool) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/query1").
		SetParam("a", a).
		SetParam("beginAt", beginAt.Format(TimeFormat)).
		SetParam("endAt", endAt.Format(TimeFormat)).
		SetParam("isRaw", BoolToString(isRaw, 10)).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Query2(ctx context.Context, a string, beginAt time.Time, endAt time.Time, isRaw bool) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/query2/"+BoolToString(isRaw, 10)+"").
		SetParam("a", a).
		SetParam("beginAt", beginAt.Format(TimeFormat)).
		SetParam("endAt", endAt.Format(TimeFormat)).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcImplClient) Query3(ctx context.Context, a string, beginAt time.Time, endAt time.Time, isRaw *bool) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/query3/"+BoolToString(*isRaw, 10)+"").
		SetParam("a", a).
		SetParam("beginAt", beginAt.Format(TimeFormat)).
		SetParam("endAt", endAt.Format(TimeFormat)).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

// Misc: annotation is missing

type StringSvcWithContextClient struct {
	proxy Proxy
}

func (client StringSvcWithContextClient) Echo(ctx context.Context, a string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/echo").
		SetParam("a", a).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcWithContextClient) EchoBody(ctx context.Context, body io.Reader) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/echo").
		SetBody(body).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcWithContextClient) Concat(ctx context.Context, a string, b string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/concat").
		SetParam("a", a).
		SetParam("b", b).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcWithContextClient) Concat1(ctx context.Context, a *string, b *string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/concat1")
	if a != nil {
		request = request.SetParam("a", *a)
	}
	if b != nil {
		request = request.SetParam("b", *b)
	}
	request = request.Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcWithContextClient) Concat2(ctx context.Context, a string, b string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/concat2/"+a+"/"+b+"").
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcWithContextClient) Concat3(ctx context.Context, a *string, b *string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/concat3/"+*a+"/"+*b+"").
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcWithContextClient) Sub(ctx context.Context, a string, start int64) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/sub").
		SetParam("a", a).
		SetParam("start", strconv.FormatInt(start, 10)).
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcWithContextClient) Save(ctx context.Context, a string, b string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/save/"+a+"").
		SetBody(b).
		Result(&result)

	err := request.POST(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcWithContextClient) Save2(ctx context.Context, a *string, b *string) (string, error) {
	var result string

	request := NewRequest(client.proxy, "/save2/"+*a+"").
		SetBody(b).
		Result(&result)

	err := request.POST(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcWithContextClient) Add(ctx context.Context, a int, b int) (int, error) {
	var result int

	request := NewRequest(client.proxy, "/add/"+strconv.FormatInt(int64(a), 10)+"/"+strconv.FormatInt(int64(b), 10)+"").
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcWithContextClient) Add2(ctx context.Context, a *int, b *int) (int, error) {
	var result int

	request := NewRequest(client.proxy, "/add2/"+strconv.FormatInt(int64(*a), 10)+"/"+strconv.FormatInt(int64(*b), 10)+"").
		Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client StringSvcWithContextClient) Add3(ctx context.Context, a *int, b *int) (int, error) {
	var result int

	request := NewRequest(client.proxy, "/add3")
	if a != nil {
		request = request.SetParam("a", strconv.FormatInt(int64(*a), 10))
	}
	if b != nil {
		request = request.SetParam("b", strconv.FormatInt(int64(*b), 10))
	}
	request = request.Result(&result)

	err := request.GET(ctx)

	ReleaseRequest(client.proxy, request)
	return result, err
}

// Misc: annotation is missing
