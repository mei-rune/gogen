package gengen

import (
	"io"
	"strings"

	"github.com/swaggo/swag"
)

var _ Plugin = &ginPlugin{}

type ginPlugin struct{}

func (gin *ginPlugin) Invocations() []Invocation {
	return []Invocation{
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

func (gin *ginPlugin) RenderFuncHeader(out io.Writer, method *Method, route swag.RouteProperties) error {
	urlstr, err := ConvertURL(route.Path, false, Colon)
	if err != nil {
		return err
	}
	_, err = io.WriteString(out, "\r\nmux."+strings.ToUpper(route.HTTPMethod)+"(\""+urlstr+"\", func(ctx *gin.Context) {")
	return err
}

func (gin *ginPlugin) RenderReturnError(out io.Writer, method *Method, errCode, err string) error {
	if errCode == "" && gin.Features().EnableHttpCode {
		errCode = "httpCodeWith(err)"
	}

	s := renderString(`{{- if .hasRealErrorCode -}}
    ctx.JSON({{.errCode}}, {{.err}})
  {{- else -}}
    ctx.JSON(http.StatusInternalServerError, {{.err}})
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
		args["statusCode"] = "http.StatusOK"
	}
	s := renderString(`{{- if .noreturn -}}
  return
{{- else -}}
  ctx.JSON({{.statusCode}}, {{.data}})
  return
{{- end}}`, args)
	_, e := io.WriteString(out, s)
	return e
}
