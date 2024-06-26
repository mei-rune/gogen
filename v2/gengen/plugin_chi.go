package gengen

import (
	"io"
	"os"

	"github.com/swaggo/swag"
)

// "features.buildTag":     "gin",
// "features.httpCodeWith": true,
// "features.boolConvert":     "toBool({{.name}})",
// "features.datetimeConvert": "toDatetime({{.name}})",

var _ Plugin = &chiPlugin{}

type chiPlugin struct {
	cfg Config
}

func (chi *chiPlugin) Functions() []Function {
	return []Function{
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

func (chi *chiPlugin) Imports() map[string]string {
	return map[string]string{
		"github.com/go-chi/chi":    "",
		"github.com/go-chi/render": "",
	}
}


func (chi *chiPlugin) PartyTypeName() string {
	return "chi.Router"
}

func (chi *chiPlugin) IsPartyFluentStyle() bool {
	return false
}

func (chi *chiPlugin) GetSpecificTypeArgument(typeStr string) (string, bool) {
	if typeStr == "context.Context" {
		ctx := os.Getenv("GOGEN_CONTEXT_GETTER")
		if ctx != "" {
			return ctx, true
		}
	}

	args := map[string]string{
		"url.Values":          "r.URL.Query()",
		"*http.Request":       "r",
		"http.ResponseWriter": "w",
		"io.Writer":           "w",
		"io.Reader":           "r.Body",
		"context.Context":     "r.Context()",
	}
	s, ok := args[typeStr]
	return s, ok
}

// func (chi *chiPlugin) GetBodyErrorText(method *Method, bodyName, err string) string {
// 	return getBodyErrorText(chi.cfg.NewBadArgument, method, bodyName, err)
// }

// func (chi *chiPlugin) GetCastErrorText(method *Method, accessFields, err, value string) string {
// 	return getCastErrorText(chi.cfg.NewBadArgument, method, accessFields, err, value)
// }


func (chi *chiPlugin) MiddlewaresDeclaration() string {
	return "handlers ...func(http.Handler) http.Handler"
}

func (chi *chiPlugin) RenderWithMiddlewares(mux string) string {
	return mux + " = "+mux+".With(handlers...)"
}

func (chi *chiPlugin) RenderMiddlewares(out io.Writer, fn func(out io.Writer) error) error {
	return fn(out)
}

func (chi *chiPlugin) ReadBodyFunc(argName string) string {
	return "render.Decode(r, " + argName + ")"
}

func (chi *chiPlugin) RenderFunc(out io.Writer, method *Method, route swag.RouteProperties, fn func(out io.Writer) error) error {
	urlstr, err := ConvertURL(route.Path, false, Colon)
	if err != nil {
		return err
	}
	// if urlstr == "/" {
	// 	urlstr = ""
	// }

	_, err = io.WriteString(out, "\r\nmux."+ConvertMethodNameToCamelCase(route.HTTPMethod)+"(\""+urlstr+"\", func(w http.ResponseWriter, r *http.Request) {")
	if err != nil {
		return err
	}

	if method.HasQueryParam() {
		_, err = io.WriteString(out, "\r\n\tqueryParams := r.URL.Query()")
		if err != nil {
			return err
		}
	}

	if err := fn(out); err != nil {
		return err
	}

	_, err = io.WriteString(out, "\r\n})")
	return err
}

func (chi *chiPlugin) GetErrorResult(err string) string {
	if chi.cfg.ErrorResult != "" {
		return chi.cfg.ErrorResult + "(" + err + ")"
	}
	return "NewErrorResult(" + err + ")"
}

func (chi *chiPlugin) GetOkResult() string {
	if chi.cfg.OkResult != "" {
		return chi.cfg.OkResult + "()"
	}
	return "NewOkResult()"
}

func (chi *chiPlugin) RenderReturnOK(out io.Writer, method *Method, statusCode, dataType, data string) error {
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
		renderFunc = "PlainText"
	}

	s := renderString(`{{- if .noreturn -}}
  return
{{- else -}}
  {{- if .statusCode -}}
		render.Status(r, {{.statusCode}})
  {{- end -}}
	render.`+renderFunc+`(w, r, {{.data}})
  return
{{- end}}`, args)
	_, err := io.WriteString(out, s)
	return err
}

func (chi *chiPlugin) RenderBodyError(out io.Writer, method *Method, bodyName, err string) error {
	txt := getBodyErrorText(chi.cfg.NewBadArgument, method, bodyName, err)
	return chi.RenderReturnError(out, method, "http.StatusBadRequest", txt, true)
}

func (chi *chiPlugin) RenderCastError(out io.Writer, method *Method, accessFields, value, err string) error {
	txt := getCastErrorText(chi.cfg.NewBadArgument, method, accessFields, err, value)
	return chi.RenderReturnError(out, method, "http.StatusBadRequest", txt, true)
}

func (chi *chiPlugin) RenderReturnError(out io.Writer, method *Method, errCode, err string, errwrapped ...bool) error {
	if errCode == "" && chi.cfg.HttpCodeWith != "" {
		errCode = chi.cfg.HttpCodeWith + "(" + err + ")"
	}

	renderFunc := "JSON"
	if len(method.Operation.Produces) == 1 &&
		method.Operation.Produces[0] == "text/plain" {
		renderFunc = "PlainText"
		err = err + ".Error()"
	} else if (len(errwrapped) == 0 || !errwrapped[0]) && chi.cfg.ErrorToJSONError != "" {
		err = chi.cfg.ErrorToJSONError + "(" + err + ")"
	}

	s := renderString(`{{- if .hasRealErrorCode -}}
    render.Status(r, {{.errCode}})
  {{else -}}
    render.Status(r, http.StatusInternalServerError)
  {{end -}}
  render.`+renderFunc+`(w, r, {{.err}})
  return`, map[string]interface{}{
		"err":              err,
		"hasRealErrorCode": errCode != "",
		"errCode":          errCode,
	})
	_, e := io.WriteString(out, s)
	return e
}

func (chi *chiPlugin) RenderReturnEmpty(out io.Writer, method *Method) error {
	_, e := io.WriteString(out, "return")
	return e
}
