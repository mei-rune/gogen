package gengen

import (
	"io"
	"strings"

	"github.com/swaggo/swag"
)

var _ Plugin = &loongPlugin{}

type loongPlugin struct {
	cfg Config
}

func (lng *loongPlugin) GetSpecificTypeArgument(typeStr string) (string, bool) {
	args := map[string]string{
		"url.Values":          "ctx.QueryParams()",
		"*http.Request":       "ctx.Request()",
		"io.Reader":           "ctx.Request().Body",
		"http.ResponseWriter": "ctx.Response().Writer",
		"io.Writer":           "ctx.Response().Writer",
		"context.Context":     "ctx.StdContext",
		"*loong.Context":      "ctx",
	}
	s, ok := args[typeStr]
	return s, ok
}

func (lng *loongPlugin) Functions() []Function {
	return []Function{
		{
			Required:    true,
			Format:      "ctx.Param(\"%s\")",
			IsArray:     false,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
		{
			Required:    false,
			Format:      "ctx.QueryParam(\"%s\")",
			IsArray:     false,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
		{
			Required:    false,
			Format:      "ctx.QueryParamArray(\"%s\")",
			IsArray:     true,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
	}
}

func (lng *loongPlugin) Imports() map[string]string {
	return map[string]string{
		"github.com/runner-mei/loong": "",
	}
}

func (lng *loongPlugin) PartyTypeName() string {
	return "loong.Party"
}

func (lng *loongPlugin) IsPartyFluentStyle() bool {
	return true
}

func (lng *loongPlugin) ReadBodyFunc(argName string) string {
	return "ctx.Bind(" + argName + ")"
}

func (lng *loongPlugin) GetBodyErrorText(method *Method, bodyName, err string) string {
	return "loong.ErrBadArgument(\"" + bodyName + "\", \"body\", " + err + ")"
}

func (lng *loongPlugin) GetCastErrorText(method *Method, accessFields, err, value string) string {
	return "loong.ErrBadArgument(\"" + accessFields + "\", " + value + ", " + err + ")"
}

func (lng *loongPlugin) RenderFuncHeader(out io.Writer, method *Method, route swag.RouteProperties) error {
	urlstr, err := ConvertURL(route.Path, false, Colon)
	if err != nil {
		return err
	}
	if urlstr == "/" {
		urlstr = ""
	}
	_, err = io.WriteString(out, "\r\nmux."+strings.ToUpper(route.HTTPMethod)+"(\""+urlstr+"\", func(ctx *loong.Context) error {")
	return err
}

func (lng *loongPlugin) RenderReturnError(out io.Writer, method *Method, errCode, err string) error {
	s := renderString(`return ctx.ReturnError({{.err}}{{if and .errCode .hasRealErrorCode}},{{.errCode}}{{end}})`,
		map[string]interface{}{
			"err":              err,
			"hasRealErrorCode": errCode != "",
			"errCode":          errCode,
		})
	_, e := io.WriteString(out, s)
	return e
}

func (lng *loongPlugin) RenderReturnOK(out io.Writer, method *Method, statusCode, data string) error {
	args := map[string]interface{}{
		"noreturn": method.NoReturn(),
		"data":     data,
	}
	if statusCode != "" {
		args["statusCode"] = statusCode
	} else {
		args["statusCode"] = statusCodeLiteralByMethod(method.Operation.RouterProperties[0].HTTPMethod)

		if withCode := WithCode(method); withCode != "" {
			args["withCode"] = withCode
		}
	}

	args["method"] = strings.ToUpper(method.Operation.RouterProperties[0].HTTPMethod)

	s := renderString(`{{- if .noreturn -}}
	return nil
	{{- else if .withCode -}} 
	return ctx.ReturnResult({{.withCode}}, {{.data}})
	{{- else if eq .method "POST" -}} 
	return ctx.ReturnCreatedResult({{.data}})
	{{- else if eq .method "PUT" -}}
	return ctx.ReturnUpdatedResult({{.data}})
	{{- else if eq .method "DELETE" -}}
	return ctx.ReturnDeletedResult({{.data}})
	{{- else if eq .method "GET" -}}
	return ctx.ReturnQueryResult({{.data}})
	{{- else -}}
	return ctx.ReturnResult({{.statusCode}}, {{.data}})
	{{end}}`, args)
	_, e := io.WriteString(out, s)
	return e
}

func (lng *loongPlugin) RenderReturnEmpty(out io.Writer, method *Method) error {
	_, e := io.WriteString(out, "return nil")
	return e
}
