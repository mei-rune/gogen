package gengen

import (
	"io"
	"os"
	"strings"

	"github.com/swaggo/swag"
)

var _ Plugin = &echoPlugin{}

type echoPlugin struct {
	cfg Config
	isV5 bool

	customReturnFunc bool
	returnResultFuncName string
	returnCreatedResultFuncName string
	returnUpdatedResultFuncName  string
	returnDeletedResultFuncName string
	returnQueryResultFuncName string
	returnErrorResultFuncName string
}

func (echo *echoPlugin) initCustomReturnFunc(ns string) {
	echo.customReturnFunc = true
	echo.returnResultFuncName = ns + "ReturnResult"
	echo.returnCreatedResultFuncName = ns + "ReturnCreatedResult"
	echo.returnUpdatedResultFuncName = ns + "ReturnUpdatedResult"
	echo.returnDeletedResultFuncName = ns + "ReturnDeletedResult"
	echo.returnQueryResultFuncName = ns + "ReturnQueryResult"
	echo.returnErrorResultFuncName = ns + "ReturnError"
}

func (echo *echoPlugin) GetSpecificTypeArgument(typeStr string) (string, bool) {
	if typeStr == "context.Context" {
		ctx := os.Getenv("GOGEN_CONTEXT_GETTER")
		if ctx != "" {
			return ctx, true
		}
	}

	args := map[string]string{
		"url.Values":          "ctx.QueryParams()",
		"*http.Request":       "ctx.Request()",
		"io.Reader":           "ctx.Request().Body",

		"http.ResponseWriter": "ctx.Response().Writer",
		"io.Writer":           "ctx.Response().Writer",
		"context.Context":     "ctx.Request().Context()",
		"echo.Context":        "ctx",
		"*echo.Context":        "ctx",
	}
	if echo.isV5 {
		args["http.ResponseWriter"] = "ctx.Response()"
		args["io.Writer"] =           "ctx.Response()"
	}
	s, ok := args[typeStr]
	return s, ok
}

func (chi *echoPlugin) HeaderFunctions() []Function {
	return []Function{
		{
			Required:    true,
			Format:      "ctx.Request().Header.Get(r, \"%s\")",
			IsArray:     false,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
		{
			Required:    false,
			Format:      "ctx.Request().Header.Get(\"%s\")",
			IsArray:     false,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
		{
			Required:    false,
			Format:      "ctx.Request().Header[\"%s\"]",
			IsArray:     true,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
	}
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
	if echo.isV5 {
		return map[string]string{
			"github.com/labstack/echo/v5": "echo",
		}
	}

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

func (echo *echoPlugin) MiddlewaresDeclaration() string {
	return "handlers ...echo.MiddlewareFunc"
}

func (echo *echoPlugin) RenderWithMiddlewares(mux string) string {
	return ""
}

func (echo *echoPlugin) ReadBodyFunc(argName string) string {
	return "ctx.Bind(" + argName + ")"
}

// func (echo *echoPlugin) GetBodyErrorText(method *Method, bodyName, err string) string {
// 	return getBodyErrorText(echo.cfg.NewBadArgument, method, bodyName, err)
// }

// func (echo *echoPlugin) GetCastErrorText(method *Method, accessFields, err, value string) string {
// 	return getCastErrorText(echo.cfg.NewBadArgument, method, accessFields, err, value)
// }

func (echo *echoPlugin) RenderFunc(out io.Writer, method *Method, route swag.RouteProperties, fn func(out io.Writer) error) error {
	urlstr, err := ConvertURL(route.Path, false, Colon)
	if err != nil {
		return err
	}
	if urlstr == "/" {
		urlstr = ""
	}
	_, err = io.WriteString(out, "\r\nmux."+strings.ToUpper(route.HTTPMethod)+"(\""+urlstr+"\", ")
	if err != nil {
		return err
	}
	if echo.isV5 {
	_, err = io.WriteString(out, "func(ctx *echo.Context) error {")
	} else {
	_, err = io.WriteString(out, "func(ctx echo.Context) error {")
	}
	if err != nil {
		return err
	}
	if err := fn(out); err != nil {
		return err
	}
	_, err = io.WriteString(out, "\r\n}, handlers...)")
	return err
}

func (echo *echoPlugin) RenderBodyError(out io.Writer, method *Method, bodyName, err string) error {
	txt := getBodyErrorText(echo.cfg.NewBadArgument, method, bodyName, err)

	return echo.RenderReturnError(out, method, "http.StatusBadRequest", txt, true)
}

func (echo *echoPlugin) RenderCastError(out io.Writer, method *Method, accessFields, value, err string) error {
	txt := getCastErrorText(echo.cfg.NewBadArgument, method, accessFields, err, value)
	return echo.RenderReturnError(out, method, "http.StatusBadRequest", txt, true)
}

func (echo *echoPlugin) RenderReturnError(out io.Writer, method *Method, errCode, err string, errwrapped ...bool) error {
	if errCode == "" && echo.cfg.HttpCodeWith != "" {
		errCode = echo.cfg.HttpCodeWith + "(" + err + ")"
	}

	renderFunc := "JSON"
	if len(method.Operation.Produces) == 1 &&
		method.Operation.Produces[0] == "text/plain" {
		renderFunc = "String"
		err = err + ".Error()"
	} else if (len(errwrapped) == 0 || !errwrapped[0]) && echo.cfg.ErrorToJSONError != "" {
		err = echo.cfg.ErrorToJSONError + "(" + err + ")"
	}

	var text string
	if echo.customReturnFunc {
		text = `return `+echo.returnErrorResultFuncName+`(ctx, {{.err}}{{if and .errCode .hasRealErrorCode}},{{.errCode}}{{end}})`
	} else {
		text = `return ctx.`+renderFunc+`(`+
			`{{if .hasRealErrorCode -}}{{.errCode}}{{else}}http.StatusInternalServerError{{end}},`+
			` {{.err}})`
	}

	s := renderString(text,
			map[string]interface{}{
			"err":              err,
			"hasRealErrorCode": errCode != "",
			"errCode":          errCode,
		})
	_, e := io.WriteString(out, s)
	return e
}

func (echo *echoPlugin) GetErrorResult(err string) string {
	if echo.cfg.ErrorResult != "" {
		return echo.cfg.ErrorResult + "(" + err + ")"
	}
	return "NewErrorResult(" + err + ")"
}

func (echo *echoPlugin) GetOkResult() string {
	if echo.cfg.OkResult != "" {
		return echo.cfg.OkResult + "()"
	}
	return "NewOkResult()"
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

	if withCode := WithCode(method); withCode != "" {
			args["withCode"] = withCode
	}

	args["method"] = strings.ToUpper(method.Operation.RouterProperties[0].HTTPMethod)

	if len(method.Operation.Produces) == 1 &&
		method.Operation.Produces[0] == "text/plain" {

		if dataType == "[]byte" {
			s := renderString(`{{- if .noreturn -}}
			return nil
			{{- else if .withCode -}} 
			return ctx.Blob({{.withCode}}, "text/plain", {{.data}})
			{{- else -}}
			return ctx.Blob({{.statusCode}}, "text/plain", {{.data}})
			{{- end}}`, args)
			_, e := io.WriteString(out, s)
			return e
		}

		s := renderString(`{{- if .noreturn -}}
		return nil
		{{- else if .withCode -}} 
		return ctx.String({{.withCode}}, {{.data}})
		{{- else -}}
		return ctx.String({{.statusCode}}, {{.data}})
		{{- end}}`, args)
		_, e := io.WriteString(out, s)
		return e
	}




	text := `{{- if .noreturn -}}
  return nil
{{- else -}}
  return ctx.JSON({{.statusCode}}, {{.data}})
{{- end}}`


	if echo.customReturnFunc {
		text = `{{- if .noreturn -}}
	return nil
	{{- else if .withCode -}} 
	return `+echo.returnResultFuncName + `(ctx, {{.withCode}}, {{.data}})
	{{- else if eq .method "POST" -}} 
	return `+echo.returnCreatedResultFuncName + `(ctx, {{.data}})
	{{- else if eq .method "PUT" -}}
	return `+echo.returnUpdatedResultFuncName + `(ctx, {{.data}})
	{{- else if eq .method "DELETE" -}}
	return `+echo.returnDeletedResultFuncName + `(ctx, {{.data}})
	{{- else if eq .method "GET" -}}
	return `+echo.returnQueryResultFuncName + `(ctx, {{.data}})
	{{- else -}}
	return `+echo.returnResultFuncName + `(ctx, {{.statusCode}}, {{.data}})
	{{- end}}`
	}

	s := renderString(text, args)
	_, e := io.WriteString(out, s)
	return e
}

func (echo *echoPlugin) RenderReturnEmpty(out io.Writer, method *Method) error {
	_, e := io.WriteString(out, "return nil")
	return e
}
