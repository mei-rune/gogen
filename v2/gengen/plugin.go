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
	BuildTag        string
	EnableHttpCode  bool
	BoolConvert     string
	DatetimeConvert string
}

type Invocation struct {
	Required    bool
	WithDefault bool
	Format      string
	IsArray     bool
	ResultType  string
	ResultError bool
	ResultBool  bool
}

func createPlugin(plugin string) (Plugin, error) {
	switch plugin {
	case "gin":
		return &ginPlugin{}, nil
	case "chi":
		return &chiPlugin{}, nil
	case "echo":
		return &echoPlugin{}, nil
	case "iris":
		return &irisPlugin{}, nil
	default:
		return nil, errors.New("plugin '" + plugin + "' is unsupported")
	}
}

type Plugin interface {
	Features() Config
	Imports() map[string]string
	PartyTypeName() string

	TypeInContext(name string) (string, bool)

	Invocations() []Invocation

	ReadBodyFunc(argName string) string
	RenderFuncHeader(out io.Writer, method *Method, route swag.RouteProperties) error
	RenderReturnOK(out io.Writer, method *Method, statusCode, data string) error
	RenderReturnError(out io.Writer, method *Method, errCode, err string) error
}

func renderBadArgument(out io.Writer, plugin Plugin, param *Param, err string) error {
	txt := "NewBadArgument(" + err + ", \"" + param.Method.Method.Clazz.Name + "." + param.Method.Method.Name + "\", \"" + param.WebParamName() + "\")"
	return plugin.RenderReturnError(out, param.Method, "http.StatusBadRequest", txt)
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
