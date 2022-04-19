package gengen

import (
	"io"

	"github.com/swaggo/swag"
)

// "features.buildTag":     "gin",
// "features.httpCodeWith": true,
// "features.boolConvert":     "toBool({{.name}})",
// "features.datetimeConvert": "toDatetime({{.name}})",

var _ Plugin = &irisPlugin{}

type irisPlugin struct{}

func (iris *irisPlugin) TypeInContext(name string) (string, bool) {
	args := map[string]string{
		"url.Values":          "ctx.Request().URL.Query()",
		"*http.Request":       "ctx.Request()",
		"io.Reader":           "ctx.Request().Body",
		"http.ResponseWriter": "ctx.ResponseWriter()",
		"io.Writer":           "ctx.ResponseWriter()",
		"context.Context":     "ctx.Request.Context()",
		"*iris.Context":       "ctx",
	}
	s, ok := args[name]
	return s, ok
}
func (iris *irisPlugin) Invocations() []Invocation {
	return []Invocation{
		{
			Required:    true,
			Format:      "ctx.Params().GetBool(\"%s\")",
			IsArray:     false,
			ResultType:  "bool",
			ResultError: true,
			ResultBool:  false,
		},
		{
			Required:    true,
			Format:      "ctx.Params().GetInt64(\"%s\")",
			IsArray:     false,
			ResultType:  "int64",
			ResultError: true,
			ResultBool:  false,
		},
		{
			Required:    true,
			Format:      "ctx.Params().GetInt(\"%s\")",
			IsArray:     false,
			ResultType:  "int",
			ResultError: true,
			ResultBool:  false,
		},
		{
			Required:    true,
			Format:      "ctx.Params().GetString(\"%s\")",
			IsArray:     false,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
		{
			Required:    false,
			Format:      "ctx.URLParam(\"%s\")",
			IsArray:     false,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
		{
			Required:    false,
			Format:      "queryParams[\"%s\"]",
			IsArray:     true,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
	}
}
func (iris *irisPlugin) Features() Config {
	return Config{
		BuildTag:        "iris",
		EnableHttpCode:  true,
		BoolConvert:     "toBool({{.name}})",
		DatetimeConvert: "toDatetime({{.name}})",
	}
}
func (iris *irisPlugin) Imports() map[string]string {
	return map[string]string{
		"github.com/kataras/iris/v12": "iris",
	}
}
func (iris *irisPlugin) PartyTypeName() string {
	return "iris.Party"
}
func (iris *irisPlugin) ReadBodyFunc(argName string) string {
	return "ctx.ReadBody(" + argName + ")"
}
func (iris *irisPlugin) RenderFuncHeader(out io.Writer, method *Method, route swag.RouteProperties) error {
	urlstr, err := ConvertURL(route.Path, false, Colon)
	if err != nil {
		return err
	}
	if urlstr == "/" {
		urlstr = ""
	}

	io.WriteString(out, "\r\nmux."+ConvertMethodNameToCamelCase(route.HTTPMethod)+"(\""+urlstr+"\", func(ctx iris.Context) {")
	params, err := method.GetParams(iris)
	if err != nil {
		return err
	}
	for _, param := range params {
		if param.Option.In == "query" && param.Option.SimpleSchema.Type == swag.ARRAY {
			_, err = io.WriteString(out, "\r\n\tqueryParams := ctx.Request().URL.Query()")
			break
		}
	}
	return err
}
func (iris *irisPlugin) RenderReturnOK(out io.Writer, method *Method, statusCode, data string) error {
	args := map[string]interface{}{
		"noreturn": method.NoReturn(),
		"data":     data,
	}
	if statusCode != "" {
		args["statusCode"] = statusCode
	}
	renderFunc := "JSON"
	if len(method.Operation.Produces) == 1 &&
		method.Operation.Produces[0] == "text/plain" {
		renderFunc = "TEXT"
	}

	s := renderString(`{{- if .noreturn -}}
  return
{{- else -}}
  {{- if .statusCode -}}
	ctx.StatusCode(r, {{.statusCode}})
  {{- end -}}
	ctx.`+renderFunc+`({{.data}})
  return
{{- end}}`, args)
	_, err := io.WriteString(out, s)
	return err
}
func (iris *irisPlugin) RenderReturnError(out io.Writer, method *Method, errCode, err string) error {
	if errCode == "" && iris.Features().EnableHttpCode {
		errCode = "httpCodeWith(err)"
	}

	renderFunc := "JSON"
	errText := ""
	if len(method.Operation.Produces) == 1 &&
		method.Operation.Produces[0] == "text/plain" {
		renderFunc = "TEXT"
		errText = ".Error()"
	}

	s := renderString(`{{- if .hasRealErrorCode -}}
    ctx.StatusCode({{.errCode}})
  {{else -}}
    ctx.StatusCode(http.StatusInternalServerError)
  {{end -}}
  ctx.`+renderFunc+`({{.err}}`+errText+`)
  return`, map[string]interface{}{
		"err":              err,
		"hasRealErrorCode": errCode != "",
		"errCode":          errCode,
	})
	_, e := io.WriteString(out, s)
	return e
}
