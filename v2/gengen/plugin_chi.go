package gengen

import (
	"io"

	"github.com/swaggo/swag"
)

// "features.buildTag":     "gin",
// "features.httpCodeWith": true,
// "features.boolConvert":     "toBool({{.name}})",
// "features.datetimeConvert": "toDatetime({{.name}})",

var _ Plugin = &chiPlugin{}

type chiPlugin struct {
}

func (chi *chiPlugin) Invocations() []Invocation {
	return []Invocation{
		{
			Required:    true,
			Format:      "chi.URLParam(r, \"%s\")",
			IsArray:     false,
			ResultType:  "string",
			ResultError: false,
			ResultBool:  false,
		},
		{
			Required:    false,
			Format:      "queryParams.Get(\"%s\")",
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

func (chi *chiPlugin) Features() Config {
	return Config{
		BuildTag:        "chi",
		EnableHttpCode:  true,
		BoolConvert:     "toBool({{.name}})",
		DatetimeConvert: "toDatetime({{.name}})",
	}
}

func (chi *chiPlugin) Imports() map[string]string {
	return map[string]string{
		"github.com/go-chi/chi":    "",
		"github.com/go-chi/render": "",
	}
}

func (chi *chiPlugin) PartyTypeName() string {
	return "chi.Router"
}

func (chi *chiPlugin) TypeInContext(name string) (string, bool) {
	args := map[string]string{
		"url.Values":          "r.URL.Query()",
		"*http.Request":       "r",
		"http.ResponseWriter": "w",
		"io.Writer":           "w",
		"io.Reader":           "r.Body",
		"context.Context":     "r.Context()",
	}
	s, ok := args[name]
	return s, ok
}

func (chi *chiPlugin)	ReadBodyFunc(argName string) string {
	return "render.Decode(r, "+argName+")"
}

func (chi *chiPlugin) RenderFuncHeader(out io.Writer, method *Method, route swag.RouteProperties) error {
	urlstr, err := ConvertURL(route.Path, false, Colon)
	if err != nil {
		return err
	}

	io.WriteString(out, "\r\nmux."+ConvertMethodNameToCamelCase(route.HTTPMethod)+"(\""+urlstr+"\", func(w http.ResponseWriter, r *http.Request) {")
	params, err := method.GetParams(chi)
	if err != nil {
		return err
	}
	for _, param := range params {
		if param.Option.In == "query" {
			_, err = io.WriteString(out, "\r\n\tqueryParams := r.URL.Query()")
			break
		}
	}
	return err
}

func (chi *chiPlugin) RenderReturnOK(out io.Writer, method *Method, statusCode, data string) error {
	args := map[string]interface{}{
		"noreturn": method.NoReturn(),
		"data":     data,
	}
	if statusCode != "" {
		args["statusCode"] = statusCode
	}
	s := renderString(`{{- if .noreturn -}}
  return
{{- else -}}
  {{- if .statusCode -}}
		render.Status(r, {{.statusCode}})
  {{- end -}}
	render.JSON(w, r, {{.data}})
  return
{{- end}}`, args)
	_, err := io.WriteString(out, s)
	return err
}

func (chi *chiPlugin) RenderReturnError(out io.Writer, method *Method, errCode, err string) error {
	if errCode == "" && chi.Features().EnableHttpCode {
		errCode = "httpCodeWith(err)"
	}

	s := renderString(`{{- if .hasRealErrorCode -}}
    render.Status(r, {{.errCode}})
  {{else -}}
    render.Status(r, http.StatusInternalServerError)
  {{end -}}
  render.JSON(w, r, {{.err}})
  return`, map[string]interface{}{
		"err":              err,
		"hasRealErrorCode": errCode != "",
		"errCode":          errCode,
	})
	_, e := io.WriteString(out, s)
	return e
}
