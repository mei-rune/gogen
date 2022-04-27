package gengen

import (
	"io"
	"strings"

	"github.com/swaggo/swag"
)

var _ Plugin = &ginPlugin{}

type ginPlugin struct{}

func (gin *ginPlugin) GetSpecificTypeArgument(typeStr string) (string, bool) {
	args := map[string]string{
		"url.Values":          "ctx.Request.URL.Query()",
		"*http.Request":       "ctx.Request",
		"io.Reader":           "ctx.Request.Body",
		"http.ResponseWriter": "ctx.Writer",
		"io.Writer":           "ctx.Writer",
		"context.Context":     "ctx.Request.Context()",
		"*gin.Context":        "ctx",
	}
	s, ok := args[typeStr]
	return s, ok
}

func (gin *ginPlugin) Functions() []Function {
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
			Format:      "ctx.Query(\"%s\")",
			IsArray:     false,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
		{
			Required:    false,
			Format:      "ctx.QueryArray(\"%s\")",
			IsArray:     true,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
	}
}

func (gin *ginPlugin) Imports() map[string]string {
	return map[string]string{
		"github.com/gin-gonic/gin": "",
	}
}

func (gin *ginPlugin) PartyTypeName() string {
	return "gin.IRouter"
}

func (gin *ginPlugin) Features() Config {
	return Config{
		BuildTag:        "gin",
		EnableHttpCode:  true,
		BoolConvert:     "toBool({{.name}})",
		DatetimeConvert: "toDatetime({{.name}})",
	}
}
func (gin *ginPlugin) ReadBodyFunc(argName string) string {
	return "ctx.Bind(" + argName + ")"
}

func (gin *ginPlugin) GetBodyErrorText(method *Method, bodyName, err string) string {
	return getBodyErrorText(method, bodyName, err)
}

func (gin *ginPlugin) GetCastErrorText(method *Method, accessFields, err, value string) string {
	return getCastErrorText(method, accessFields, err, value)
}

func (gin *ginPlugin) RenderFuncHeader(out io.Writer, method *Method, route swag.RouteProperties) error {
	urlstr, err := ConvertURL(route.Path, false, Colon)
	if err != nil {
		return err
	}
	if urlstr == "/" {
		urlstr = ""
	}
	_, err = io.WriteString(out, "\r\nmux."+strings.ToUpper(route.HTTPMethod)+"(\""+urlstr+"\", func(ctx *gin.Context) {")
	return err
}

func (gin *ginPlugin) RenderReturnError(out io.Writer, method *Method, errCode, err string) error {
	if errCode == "" && gin.Features().EnableHttpCode {
		errCode = "httpCodeWith(err)"
	}

	renderFunc := "JSON"
	errText := ""
	if len(method.Operation.Produces) == 1 &&
		method.Operation.Produces[0] == "text/plain" {
		renderFunc = "String"
		errText = ".Error()"
	}

	s := renderString(`{{- if .hasRealErrorCode -}}
    ctx.`+renderFunc+`({{.errCode}}, {{.err}}`+errText+`)
  {{- else -}}
    ctx.`+renderFunc+`(http.StatusInternalServerError, {{.err}}`+errText+`)
  {{- end}}
    return`, map[string]interface{}{
		"err":              err,
		"hasRealErrorCode": errCode != "",
		"errCode":          errCode,
	})
	_, e := io.WriteString(out, s)
	return e
}

func (gin *ginPlugin) RenderReturnOK(out io.Writer, method *Method, statusCode, data string) error {
	args := map[string]interface{}{
		"noreturn": method.NoReturn(),
		"data":     data,
	}
	if statusCode != "" {
		args["statusCode"] = statusCode
	} else {
		args["statusCode"] = statusCodeLiteralByMethod(method.Operation.RouterProperties[0].HTTPMethod)
	}

	renderFunc := "JSON"
	if len(method.Operation.Produces) == 1 &&
		method.Operation.Produces[0] == "text/plain" {
		renderFunc = "String"
	}
	s := renderString(`{{- if .noreturn -}}
  return
{{- else -}}
  ctx.`+renderFunc+`({{.statusCode}}, {{.data}})
  return
{{- end}}`, args)
	_, e := io.WriteString(out, s)
	return e
}

func (gin *ginPlugin) RenderReturnEmpty(out io.Writer, method *Method) error {
	_, e := io.WriteString(out, "return")
	return e
}
