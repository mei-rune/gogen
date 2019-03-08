package network

import (
	"context"
	"io"
)

type Proxy interface {
	New() Request
}

type Request interface {
	SetParam(key, value string) Request
	AddParam(key, value string) Request
	SetBody(body interface{}) Request
	Body(body interface{}) Request
	ExceptedCode(code int) Request
	GET(context.Context) error
	POST(context.Context) error
}

func NewRequest(proxy Proxy, url string) Request {
	return nil
}

func ReleaseRequest(proxy Proxy, request Request) {
}

type TestClient struct {
	proxy Proxy
}

func (client TestClient) Echo(ctx context.Context, a string) (string, error) {
	var result string
	request := NewRequest(client.proxy, "")
	err := request.
		Result(&result).
		GET(ctx)
	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client TestClient) EchoBody(ctx context.Context, body io.Reader) (string, error) {
	var result string

	request := NewRequest(client.proxy, "")
	err := request.
		SetBody(body).
		POST(ctx)
	ReleaseRequest(client.proxy, request)
	return result, err
}

func (client TestClient) Concat(ctx context.Context, a, b string) (string, error) {
	var result string
	request := NewRequest(client.proxy, "")
	err := request.
		SetParam("a", "b").
		SetParam("b", "b").
		GET(ctx)
	ReleaseRequest(client.proxy, request)
	return result, err
}
