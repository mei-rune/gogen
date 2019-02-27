package gengen

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

type SkippedResult struct {
	IsSkipped bool
	Message   string
}

type MuxStye interface {
	CtxName() string
	CtxType() string
	IsReserved(param Param) bool
	ToReserved(param Param) string
	ReadParam(param Param, name string) string
	FuncSignature() string
	RouteFunc(method Method) string
	BadArgumentFunc(method Method, args ...string) string
	ErrorFunc(method Method, args ...string) string
	OkFunc(method Method, args ...string) string
	GetPath(method Method) string
	IsSkipped(method Method) SkippedResult
}

type Generator struct {
	Ext string
	Mux MuxStye
}

func (cmd *Generator) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.Ext, "ext", ".gogen.go", "文件后缀名")
	return fs
}

func (cmd *Generator) Run(args []string) error {
	if cmd.Mux == nil {
		cmd.Mux = NewEchoStye()
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

	for _, class := range file.Interfaces {
		if err := cmd.generateInterface(out, file, &class); err != nil {
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

func (cmd *Generator) generateInterface(out io.Writer, file *SourceContext, class *Class) error {
	args := map[string]interface{}{"mux": cmd.Mux, "class": class}
	err := initFunc.Execute(out, args)
	if err != nil {
		return errors.New("generate initFunc for '" + class.Name.Name + "' fail, " + err.Error())
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
func Init{{.class.Name}}(mux {{.mux.RoutePartyName}}, svc {{if not .class.IsInterface}}*{{end}}{{.class.Name}}) {
	{{- range $method := .class.Methods}}
	{{- $skipResult := $.mux.IsSkipped $method }}
	{{- if $skipResult.IsSkipped }}
	{{if $skipResult.Message}} 
	// {{$method.Name.Name}}: {{$skipResult.Message}} 
	{{end}}

	{{- else}}
	mux.{{$.mux.RouteFunc $method}}("{{$.mux.GetPath $method}}", {{$.mux.FuncSignature}}{
		{{- range $param := $method.Params.List}}
				{{- if $.mux.IsReserved $param -}}
				{{- else if eq $param.Name.Name $.mux.CtxName}}
					{{- if ne (typePrint $param.Typ) $.mux.CtxType}}
					var _{{$param.Name.Name}} {{$.mux.ReadParam $param (concat "_" $param.Name.Name)}}
					{{- end}}
				{{- else}}
				var {{$param.Name.Name}} {{$.mux.ReadParam $param  $param.Name.Name}}
				{{- end}}
		{{- end}}

		result, err := svc.{{$method.Name}}(
			{{- range $idx, $param := $method.Params.List -}}
				{{- if eq $param.Name.Name $.mux.CtxName -}}
					{{- if $.mux.IsReserved $param -}}
					{{$.mux.ToReserved $param}}
					{{- else if eq (typePrint $param.Typ) $.mux.CtxType -}}
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
	{{- end}} {{/* isSkipped */}}
	{{- end}}
}
`))
}
