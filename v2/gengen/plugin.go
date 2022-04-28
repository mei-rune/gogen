package gengen

import (
	"errors"
	"io"
	"log"
	"strings"
	"text/template"

	"github.com/swaggo/swag"
)

type Config struct {
	HttpCodeWith string
}

type Function struct {
	Required    bool
	WithDefault bool
	Format      string
	IsArray     bool
	ResultType  string
	ResultError bool
	ResultBool  bool
}

func createPlugin(plugin string, cfg Config) (Plugin, error) {
	switch plugin {
	case "gin":
		return &ginPlugin{cfg: cfg}, nil
	case "chi":
		return &chiPlugin{cfg: cfg}, nil
	case "echo":
		return &echoPlugin{cfg: cfg}, nil
	case "iris":
		return &irisPlugin{cfg: cfg}, nil
	case "loong":
		return &loongPlugin{cfg: cfg}, nil
	default:
		return nil, errors.New("plugin '" + plugin + "' is unsupported")
	}
}

type Plugin interface {
	Imports() map[string]string
	PartyTypeName() string

	GetSpecificTypeArgument(typeStr string) (string, bool)

	Functions() []Function

	ReadBodyFunc(argName string) string
	RenderFuncHeader(out io.Writer, method *Method, route swag.RouteProperties) error
	RenderReturnOK(out io.Writer, method *Method, statusCode, data string) error
	RenderReturnError(out io.Writer, method *Method, errCode, err string) error
	RenderReturnEmpty(out io.Writer, method *Method) error

	GetBodyErrorText(method *Method, bodyName, err string) string
	GetCastErrorText(method *Method, accessFields string, err, value string) string
}

func getBodyErrorText(method *Method, bodyName, err string) string {
	txt := "NewBadArgument(" + err + ", \"" + method.FullName() + "\", \"" + bodyName + "\")"
	// return "fmt.Errorf(\"argument %q is invalid - %q\", \""+bodyName+"\", \"body\", "+ err + ")"
	return txt
}

func getCastErrorText(method *Method, accessFields string, err, value string) string {
	// txt := "fmt.Errorf(\"argument %q is invalid - %q\", \""+param.WebParamName()+"\", "+value+", "+err+")"
	return "NewBadArgument(" + err + ", \"" + method.FullName() + "\", \"" + accessFields + "\")"
}

func renderCastError(ctx *GenContext, method *Method, accessFields, err, value string) error {
	txt := ctx.plugin.GetCastErrorText(method, accessFields, err, value)
	return ctx.plugin.RenderReturnError(ctx.out, method, "http.StatusBadRequest", txt)
}

func renderText(txt *template.Template, out io.Writer, renderArgs interface{}) {
	err := txt.Execute(out, renderArgs)
	if err != nil {
		log.Fatalln(err)
	}
}

var Funcs template.FuncMap

func renderString(txt string, renderArgs interface{}) string {
	var out strings.Builder
	err := template.Must(template.New("a").Funcs(Funcs).Parse(txt)).Execute(&out, renderArgs)
	if err != nil {
		log.Fatalln(err)
	}
	return out.String()
}

func statusCodeLiteralByMethod(op string) string {
	if strings.ToLower(op) == "post" {
		return "http.StatusCreated"
	}
	return "http.StatusOK"
}
