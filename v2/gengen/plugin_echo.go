package gengen

import (
	"io"
	"strings"

	"github.com/swaggo/swag"
)

var _ Plugin = &echoPlugin{}

type echoPlugin struct {
	cfg Config
}

func (echo *echoPlugin) GetSpecificTypeArgument(typeStr string) (string, bool) {
	args := map[string]string{
		"url.Values":          "ctx.QueryParams()",
		"*http.Request":       "ctx.Request()",
		"io.Reader":           "ctx.Request().Body",
		"http.ResponseWriter": "ctx.Response().Writer",
		"io.Writer":           "ctx.Response().Writer",
		"context.Context":     "ctx.Request().Context()",
		"echo.Context":        "ctx",
	}
	s, ok := args[typeStr]
	return s, ok
}

func (echo *echoPlugin) Functions() []Function {
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
			Format:      "ctx.QueryParams()[\"%s\"]",
			IsArray:     true,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
	}
}

func (echo *echoPlugin) Imports() map[string]string {
	return map[string]string{
		"github.com/labstack/echo/v4": "echo",
	}
}

func (echo *echoPlugin) PartyTypeName() string {
	return "*echo.Group"
}

func (echo *echoPlugin) IsPartyFluentStyle() bool {
	return true
}

func (echo *echoPlugin) ReadBodyFunc(argName string) string {
	return "ctx.Bind(" + argName + ")"
}

func (echo *echoPlugin) GetBodyErrorText(method *Method, bodyName, err string) string {
	return getBodyErrorText(method, bodyName, err)
}

func (echo *echoPlugin) GetCastErrorText(method *Method, accessFields, err, value string) string {
	return getCastErrorText(method, accessFields, err, value)
}

func (echo *echoPlugin) RenderFuncHeader(out io.Writer, method *Method, route swag.RouteProperties) error {
	urlstr, err := ConvertURL(route.Path, false, Colon)
	if err != nil {
		return err
	}
	if urlstr == "/" {
		urlstr = ""
	}
	_, err = io.WriteString(out, "\r\nmux."+strings.ToUpper(route.HTTPMethod)+"(\""+urlstr+"\", func(ctx echo.Context) error {")
	return err
}

func (echo *echoPlugin) RenderReturnError(out io.Writer, method *Method, errCode, err string) error {
	if errCode == "" && echo.cfg.HttpCodeWith != "" {
		errCode = echo.cfg.HttpCodeWith + "(" + err + ")"
	}

	renderFunc := "JSON"
	errText := ""
	if len(method.Operation.Produces) == 1 &&
		method.Operation.Produces[0] == "text/plain" {
		renderFunc = "String"
		errText = ".Error()"
	}

	s := renderString(`{{- if .hasRealErrorCode -}}
    return ctx.`+renderFunc+`({{.errCode}}, {{.err}}`+errText+`)
  {{- else -}}
    return ctx.`+renderFunc+`(http.StatusInternalServerError, {{.err}}`+errText+`)
  {{- end}}`, map[string]interface{}{
		"err":              err,
		"hasRealErrorCode": errCode != "",
		"errCode":          errCode,
	})
	_, e := io.WriteString(out, s)
	return e
}

func (echo *echoPlugin) RenderReturnOK(out io.Writer, method *Method, statusCode, dataType, data string) error {
	args := map[string]interface{}{
		"noreturn": method.NoReturn(),
		"data":     data,
	}
	if statusCode != "" {
		args["statusCode"] = statusCode
	} else {
		args["statusCode"] = statusCodeLiteralByMethod(method.Operation.RouterProperties[0].HTTPMethod)
	}
	s := renderString(`{{- if .noreturn -}}
  return nil
{{- else -}}
  return ctx.JSON({{.statusCode}}, {{.data}})
{{- end}}`, args)
	_, e := io.WriteString(out, s)
	return e
}

func (echo *echoPlugin) RenderReturnEmpty(out io.Writer, method *Method) error {
	_, e := io.WriteString(out, "return nil")
	return e
}
