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

type irisPlugin struct {
	cfg Config
}

func (iris *irisPlugin) GetSpecificTypeArgument(typeStr string) (string, bool) {
	args := map[string]string{
		"url.Values":          "ctx.Request().URL.Query()",
		"*http.Request":       "ctx.Request()",
		"io.Reader":           "ctx.Request().Body",
		"http.ResponseWriter": "ctx.ResponseWriter()",
		"io.Writer":           "ctx.ResponseWriter()",
		"context.Context":     "ctx.Request().Context()",
		"*iris.Context":       "ctx",
	}
	s, ok := args[typeStr]
	return s, ok
}
func (iris *irisPlugin) Functions() []Function {
	return []Function{
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
			Format:      "ctx.URLParamSlice(\"%s\")",
			IsArray:     true,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
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

func (iris *irisPlugin) IsPartyFluentStyle() bool {
	return true
}

func (iris *irisPlugin) ReadBodyFunc(argName string) string {
	return "ctx.UnmarshalBody(" + argName + ", nil)"
}

func (iris *irisPlugin) GetBodyErrorText(method *Method, bodyName, err string) string {
	return getBodyErrorText(iris.cfg.NewBadArgument, method, bodyName, err)
}

func (iris *irisPlugin) GetCastErrorText(method *Method, accessFields, err, value string) string {
	return getCastErrorText(iris.cfg.NewBadArgument, method, accessFields, err, value)
}

func (iris *irisPlugin) RenderFuncHeader(out io.Writer, method *Method, route swag.RouteProperties) error {
	urlstr, err := ConvertURL(route.Path, false, Colon)
	if err != nil {
		return err
	}
	// if urlstr == "/" {
	// 	urlstr = ""
	// }

	_, err = io.WriteString(out, "\r\nmux."+ConvertMethodNameToCamelCase(route.HTTPMethod)+"(\""+urlstr+"\", func(ctx iris.Context) {")
	return err
}
func (iris *irisPlugin) RenderReturnOK(out io.Writer, method *Method, statusCode, dataType, data string) error {
	args := map[string]interface{}{
		"noreturn": method.NoReturn(),
		"data":     data,
	}
	if statusCode != "" {
		args["statusCode"] = statusCode
		// } else {
		// args["statusCode"] = statusCodeLiteralByMethod(method.Operation.RouterProperties[0].HTTPMethod)
	}
	renderFunc := "JSON"
	if len(method.Operation.Produces) == 1 &&
		method.Operation.Produces[0] == "text/plain" {
		renderFunc = "Text"
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
	if errCode == "" && iris.cfg.HttpCodeWith != "" {
		errCode = iris.cfg.HttpCodeWith + "(" + err + ")"
	}

	renderFunc := "JSON"
	errText := ""
	if len(method.Operation.Produces) == 1 &&
		method.Operation.Produces[0] == "text/plain" {
		renderFunc = "Text"
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

func (iris *irisPlugin) RenderReturnEmpty(out io.Writer, method *Method) error {
	_, e := io.WriteString(out, "return")
	return e
}
