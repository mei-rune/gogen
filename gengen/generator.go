package gengen

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type Generator struct {
	Ext string
	Mux interface{}
}

func (cmd *Generator) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.Ext, "ext", ".gogen.go", "文件后缀名")
	return fs
}

func (cmd *Generator) Run(args []string) error {
	if cmd.Mux == nil {
		cmd.Mux = NewMux()
	}

	if cmd.Ext == "" {
		cmd.Ext = ".gogen.go"
	}

	for _, file := range args {
		if err := cmd.runFile(file); err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (cmd *Generator) runFile(filename string) error {
	pa, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	//dir := filepath.Dir(pa)

	file, err := ParseFile(pa)
	if err != nil {
		return err
	}

	targetFile := strings.TrimSuffix(pa, ".go") + cmd.Ext

	if len(file.Interfaces) == 0 {
		err = os.Remove(targetFile)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	out, err := os.Create(targetFile + ".tmp")
	if err != nil {
		return err
	}
	defer func() {
		out.Close()
		os.Remove(targetFile + ".tmp")
	}()

	if err = cmd.generateHeader(out, file); err != nil {
		return err
	}

	for _, itf := range file.Interfaces {
		if err := cmd.generateInterface(out, file, &itf); err != nil {
			return err
		}
	}

	if err = out.Close(); err != nil {
		os.Remove(targetFile + ".tmp")
		return err
	}
	err = os.Rename(targetFile+".tmp", targetFile)
	if err != nil {
		return err
	}

	// 不知为什么，有时运行两次 goimports 才起效
	exec.Command("goimports", "-w", targetFile).Run()
	return goImports(targetFile)
}

func goImports(src string) error {
	cmd := exec.Command("goimports", "-w", src)
	cmd.Dir = filepath.Dir(src)
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		fmt.Println("goimports -w", src)
		fmt.Println(string(out))
	}
	if err != nil {
		fmt.Println(err)
	} else if len(out) == 0 {
		fmt.Println("run `" + cmd.Path + " -w " + src + "` ok")
	}
	return err
}

func (cmd *Generator) generateHeader(out io.Writer, file *SourceContext) error {
	io.WriteString(out, "// Please don't edit this file!\r\npackage ")
	io.WriteString(out, file.Pkg.Name)
	io.WriteString(out, "\r\n\r\nimport (")
	io.WriteString(out, "\r\n\t\"errors\"")
	for _, pa := range file.Imports {
		io.WriteString(out, "\r\n\t\"")
		io.WriteString(out, pa.Path.Value)
		io.WriteString(out, "\"")
	}
	io.WriteString(out, "\r\n)\r\n")
	return nil
}

func (cmd *Generator) generateInterface(out io.Writer, file *SourceContext, itf *Interface) error {
	args := map[string]interface{}{"mux": cmd.Mux, "itf": itf}
	err := initFunc.Execute(out, args)
	if err != nil {
		return errors.New("generate initFunc for '" + itf.Name.Name + "' fail, " + err.Error())
	}
	return nil
}

var Funcs = template.FuncMap{
	"concat":        strings.Join,
	"containSubstr": strings.Contains,
	"startWith":     strings.HasPrefix,
	"endWith":       strings.HasSuffix,
	"trimPrefix":    strings.TrimPrefix,
	"trimSuffix":    strings.TrimSuffix,

	"sub": func(a, b int) int {
		return a - b
	},
	"sum": func(a, b int) int {
		return a + b
	},
	"default": func(value, defvalue interface{}) interface{} {
		if nil == value {
			return defvalue
		}
		if s, ok := value.(string); ok && "" == s {
			return defvalue
		}
		return value
	},
	"set": func(args map[string]interface{}, name string, value interface{}) string {
		args[name] = value
		return ""
	},
	"arg": func(name string, value interface{}, args map[string]interface{}) map[string]interface{} {
		args[name] = value
		return args
	},
	"last": func(objects interface{}) interface{} {
		if objects == nil {
			return nil
		}

		rv := reflect.ValueOf(objects)
		if rv.Kind() == reflect.Array {
			return rv.Index(rv.Len() - 1).Interface()
		}
		if rv.Kind() == reflect.Slice {
			return rv.Index(rv.Len() - 1).Interface()
		}
		return nil
	},

	"isLast": func(objects interface{}, idx int) bool {
		if objects == nil {
			return true
		}

		rv := reflect.ValueOf(objects)
		if rv.Kind() == reflect.Array {
			return (rv.Len() - 1) == idx
		}
		if rv.Kind() == reflect.Slice {
			return (rv.Len() - 1) == idx
		}
		return false
	},
	"typePrint": typePrint,
}

var initFunc *template.Template

func init() {
	initFunc = template.Must(template.New("InitFunc").Funcs(Funcs).Parse(`
func Init{{.itf.Name}}(mux {{.mux.RoutePartyName}}, svc {{.itf.Name}}) {
	{{- range $method := .itf.Methods}}
	mux.{{$.mux.RouteFunc $method}}({{$.mux.FuncSignature}}{
		{{- range $param := $method.Params.List}}
				{{- if eq $param.Name.Name $.mux.CtxName}}
					{{- if ne (typePrint $param.Typ) $.mux.CtxType}}
					var _{{$param.Name.Name}} {{$.mux.ReadParam $param (concat "_" $param.Name.Name)}}
					{{- end}}
				{{- else}}
				var {{$param.Name.Name}} {{$.mux.ReadParam $param  $param.Name.Name}}
				{{- end}}
		{{- end}}

		result, err := {{$method.Name}}(
			{{- range $idx, $param := $method.Params.List -}}
				{{- if eq $param.Name.Name $.mux.CtxName -}}
					{{- if eq (typePrint $param.Typ) $.mux.CtxType -}}
					{{$param.Name.Name}}
					{{- else -}}
					_{{$param.Name.Name}}
					{{- end -}}
				{{- else -}}
				{{$param.Name.Name}}
				{{- end -}}
				{{- if isLast $method.Params.List $idx | not -}},{{- end -}}
			{{- end -}})
		if err != nil {
			{{$.mux.ErrorFunc $method "err"}}
		}
		{{$.mux.OkFunc $method "result"}}
	})
	{{- end}}
}
`))

}

func NewMux() *Mux {
	mux := &Mux{}
	mux.Init()
	return mux
}

type Mux struct {
	RoutePartyName    string
	PathParam         string
	QueryParam        string
	ReadBody          string
	BadArgumentFormat string
}

func (mux *Mux) Init() {
	mux.RoutePartyName = "*echo.Group"
	mux.PathParam = "Param"
	mux.QueryParam = "QueryParam"
	mux.ReadBody = "Bind"
	mux.BadArgumentFormat = "errors.New(\"argument %%q is invalid - %%q\", %s, %s, %s)"
}

func (mux *Mux) CtxName() string {
	return `ctx`
}

func (mux *Mux) CtxType() string {
	return `ctx`
}

func typePrint(typ ast.Node) string {
	fset := token.NewFileSet()
	var buf strings.Builder
	if err := format.Node(&buf, fset, typ); err != nil {
		panic(err)
	}
	return buf.String()
}

func (mux *Mux) ReadParam(param Param, name string) string {
	typeStr := typePrint(param.Typ)

	var optional bool = true
	var readParam = mux.QueryParam

	var sb strings.Builder
	sb.WriteString(" ")
	// sb.WriteString(typeStr)

	switch typeStr {
	case "string":
		sb.WriteString(" = ")
		sb.WriteString(mux.CtxName())
		sb.WriteString(".")
		sb.WriteString(readParam)
		sb.WriteString("(\"")
		sb.WriteString(param.Name.Name)
		sb.WriteString("\")")
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		sb.WriteString(typeStr)

		if optional {
			sb.WriteString("\r\n    if s := ")
			sb.WriteString(mux.CtxName())
			sb.WriteString(".")
			sb.WriteString(readParam)
			sb.WriteString("(\"")
			sb.WriteString(param.Name.Name)
			sb.WriteString("\"); s != \"\" {\r\n      ")
		} else {
			sb.WriteString("\r\n    ")
		}

		conv := "strconv.ParseInt"
		if strings.HasPrefix(typeStr, "u") {
			conv = "strconv.ParseUint"
		}

		sb.WriteString("if v64, err := ")
		sb.WriteString(conv)
		sb.WriteString("(s, 10, 64); err != nil {")
		sb.WriteString("\r\n          ")
		sb.WriteString(mux.BadArgumentFunc(*param.Method, fmt.Sprintf(mux.BadArgumentFormat, param.Name.Name, "s", "err")))
		sb.WriteString("\r\n        }")
		sb.WriteString("\r\n        ")
		sb.WriteString(name)
		sb.WriteString(" = ")
		if strings.HasSuffix(typeStr, "64") {
			sb.WriteString("v64")
		} else {
			sb.WriteString(typeStr)
			sb.WriteString("(v64)")
		}

		if optional {
			sb.WriteString("\r\n    }")
		}

	case "*int", "*int8", "*int16", "*int32", "*int64", "*uint", "*uint8", "*uint16", "*uint32", "*uint64":
		sb.WriteString(typeStr)

		if optional {
			sb.WriteString("\r\n    if s := ")
			sb.WriteString(mux.CtxName())
			sb.WriteString(".")
			sb.WriteString(readParam)
			sb.WriteString("(\"")
			sb.WriteString(param.Name.Name)
			sb.WriteString("\"); s != \"\" {\r\n      ")
		} else {
			sb.WriteString("\r\n    ")
		}

		conv := "strconv.ParseInt"
		if strings.HasPrefix(typeStr, "u") {
			conv = "strconv.ParseUint"
		}

		sb.WriteString("if v64, err := ")
		sb.WriteString(conv)
		sb.WriteString("(")

		if optional {
			sb.WriteString("s")
		} else {
			sb.WriteString(mux.CtxName())
			sb.WriteString(".")
			sb.WriteString(readParam)
			sb.WriteString("(\"")
			sb.WriteString(param.Name.Name)
			sb.WriteString("\")")
		}
		sb.WriteString(", 10, 64); err != nil {")
		sb.WriteString("\r\n          ")
		sb.WriteString(mux.BadArgumentFunc(*param.Method, fmt.Sprintf(mux.BadArgumentFormat, param.Name.Name, "s", "err")))
		sb.WriteString("\r\n        }")
		sb.WriteString("\r\n        ")
		sb.WriteString(name)
		sb.WriteString(" = ")
		if strings.HasSuffix(typeStr, "64") {
			sb.WriteString("&v64")
		} else {
			sb.WriteString("new(")
			sb.WriteString(typeStr)
			sb.WriteString(")\r\n        *")
			sb.WriteString(name)
			sb.WriteString(" = ")
			sb.WriteString(typeStr)
			sb.WriteString("(v64)")
		}

		if optional {
			sb.WriteString("\r\n    }")
		}
	}

	// ann := mux.GetAnnotation(method)
	// if ann == nil {
	// 	panic(errors.New(strconv.FormatInt(method.Node.Pos(), 10) + ": Annotation of method '" + method.Itf.Name + ":" + method.Name + "' is missing"))
	// }
	// return strings.TrimPrefix(ann.Name, "http.")
	return sb.String()
}

func (mux *Mux) FuncSignature() string {
	return `func(` + mux.CtxName() + ` echo.Context) error `
}

func (mux *Mux) RouteFunc(method Method) string {
	ann := mux.GetAnnotation(method)
	if ann == nil {
		panic(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": Annotation of method '" + method.Itf.Name.Name + ":" + method.Name.Name + "' is missing"))
	}
	return strings.TrimPrefix(ann.Name, "http.")
}

func (mux *Mux) GetAnnotation(method Method) *Annotation {
	var annotation *Annotation
	for idx := range method.Annotations {
		if !strings.HasPrefix(method.Annotations[idx].Name, "http.") {
			continue
		}

		if annotation != nil {
			panic(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": Annotation of method '" + method.Itf.Name.Name + ":" + method.Name.Name + "' is duplicated"))
		}
		annotation = &method.Annotations[idx]
	}
	return annotation
}

func (mux *Mux) BadArgumentFunc(method Method, args ...string) string {
	return mux.ErrorFunc(method, args...)
}

func (mux *Mux) ErrorFunc(method Method, args ...string) string {
	return "return " + mux.CtxName() + ".Error(" + strings.Join(args, ",") + ")"
}

func (mux *Mux) OkFunc(method Method, args ...string) string {
	return "return " + mux.CtxName() + ".JSON(" + strings.Join(args, ",") + ")"
}
