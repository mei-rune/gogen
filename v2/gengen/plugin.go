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

	// 用于返回一个获取参数的函数调用，如
	//         ctx.PathParam("a")
	//      或 ctx.QueryParam("a")
	//      或 ctx.GetInt64PathParam("id")
	//      或 ctx.GetInt64PathParamWithDefault("id", 0)
	//      或 ctx.GetInt64QueryParam("id")
	//      或 ctx.GetInt64QueryParamWithDefault("id", 0)
	//      或 ctx.GetQueryArray("id")
	//
	// 参数说明：
	//      param        参数名，上面例子中的 "a" 或 "id"
	//      typ          表示期望函数调用返回值的数型，这个只能是一个建议，
	//                   因为像 gin 等不支持 GetInt64PathParam() 之类的函数
	//      isArray      表示期望函数调用返回值是一个数组
	//      defaultValue 表示期望函数调用中的默认值，就是上面的ctx.GetInt64PathParamWithDefault("id", 0) 中最后一个参数的 "0"
	//
	// 返回:
	//     content 返回函数调用呈现的字符串
	//     isTransformed 返回函数调用返回值的类型是不是已经转换好了, 见 typ 参数
	//     hasError 表示函数调用时是否会返回参数
	//     err      发生错误
	// RenderInvocation(param, typ string, required bool, isArray bool, defaultValue string) (content string, isTransformed, hasError bool, err error)

	Invocations() []Invocation

	RenderFuncHeader(out io.Writer, method *Method, route swag.RouteProperties) error
	RenderReturnOK(out io.Writer, method *Method, statusCode, data string) error
	RenderReturnError(out io.Writer, method *Method, errCode, err string) error
}

func renderBadArgument(out io.Writer, plugin Plugin, method *Method, param *Param, err string) error {
	txt := "NewBadArgument(" + err + ", \"" + method.Method.Clazz.Name + "." + method.Method.Name + "\", \"" + param.Param.Name + "\")"
	return plugin.RenderReturnError(out, method, "http.StatusBadRequest", txt)
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
