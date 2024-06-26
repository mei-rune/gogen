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
	HttpCodeWith     string
	NewBadArgument   string
	ErrorToJSONError string
	ErrorResult      string
	OkResult         string
	EnableResultWrap bool
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
	IsPartyFluentStyle() bool

	GetSpecificTypeArgument(typeStr string) (string, bool)

	Functions() []Function

	ReadBodyFunc(argName string) string
	RenderFunc(out io.Writer, method *Method, route swag.RouteProperties, fn func(io.Writer) error) error
	RenderReturnOK(out io.Writer, method *Method, statusCode, dataType, data string) error
	RenderReturnEmpty(out io.Writer, method *Method) error
	RenderReturnError(out io.Writer, method *Method, errCode, err string, errwrapped ...bool) error

	RenderBodyError(out io.Writer, method *Method, accessFields, err string) error
	RenderCastError(out io.Writer, method *Method, accessFields, err, value string) error

	GetErrorResult(err string) string
	GetOkResult() string


	MiddlewaresDeclaration() string
	RenderWithMiddlewares(mux string) string
	// RenderMiddlewares(out io.Writer, fn func(out io.Writer) error) error
}

func getBodyErrorText(badArg string, method *Method, bodyName, err string) string {
	txt := badArg + "(" + err + ", \"" + method.FullName() + "\", \"" + bodyName + "\")"
	// return "fmt.Errorf(\"argument %q is invalid - %q\", \""+bodyName+"\", \"body\", "+ err + ")"
	return txt
}

func getCastErrorText(badArg string, method *Method, accessFields string, err, value string) string {
	// txt := "fmt.Errorf(\"argument %q is invalid - %q\", \""+param.WebParamName()+"\", "+value+", "+err+")"
	return badArg + "(" + err + ", \"" + method.FullName() + "\", \"" + accessFields + "\")"
}

// func renderCastError(ctx *GenContext, method *Method, accessFields, err, value string) error {
// 	txt := ctx.plugin.GetCastErrorText(method, accessFields, err, value)
// 	return ctx.plugin.RenderCastError(ctx.out, method, "http.StatusBadRequest", txt)
// }

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
