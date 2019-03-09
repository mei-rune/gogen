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
	"strings"
	"text/template"
)

type SkippedResult struct {
	IsSkipped bool
	Message   string
}

type MuxStye interface {
	// CtxName() string
	// CtxType() string
	// IsReserved(param Param) bool
	// ToReserved(param Param) string
	// ReadParam(param Param, name string) string

	RouteParty() string
	InitParam(param Param) string
	UseParam(param Param) string

	FuncSignature() string
	RouteFunc(method Method) string
	BadArgumentFunc(method Method, err string, args ...string) string
	ErrorFunc(method Method, err string, args ...string) string
	// OkCode(method Method) int
	OkFunc(method Method, args ...string) string
	GetPath(method Method) string
	IsSkipped(method Method) SkippedResult
}

type WebServerGenerator struct {
	GeneratorBase

	config             string
	enableHttpCodeWith bool
	Mux                MuxStye
}

func (cmd *WebServerGenerator) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs = cmd.GeneratorBase.Flags(fs)

	fs.StringVar(&cmd.config, "config", "", "配置文件名")
	fs.BoolVar(&cmd.enableHttpCodeWith, "httpCodeWith", false, "生成 enableHttpCodeWith 函数")
	return fs
}

func (cmd *WebServerGenerator) Run(args []string) error {
	if cmd.Mux == nil {
		cmd.Mux = NewEchoStye()
	}

	if cmd.config != "" {
		cfg, err := readConfig(cmd.config)
		if err != nil {
			log.Fatalln(err)
			return err
		}

		if err := toStruct(cmd.Mux, cfg); err != nil {
			log.Fatalln(err)
			return err
		}

		cmd.enableHttpCodeWith = boolWith(cfg, "features.httpCodeWith", cmd.enableHttpCodeWith)
		if cmd.buildTag == "" {
			cmd.buildTag = stringWith(cfg, "features.buildTag", cmd.buildTag)
		}
		cmd.imports = readImports(cfg)
	}

	if mux := cmd.Mux.(*DefaultStye); mux != nil {
		mux.reinit(nil)
	}

	if cmd.ext == "" {
		cmd.ext = ".gogen.go"
	}

	for _, file := range args {
		if err := cmd.runFile(file); err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (cmd *WebServerGenerator) runFile(filename string) error {
	pa, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	//dir := filepath.Dir(pa)

	file, err := ParseFile(pa)
	if err != nil {
		return err
	}

	targetFile := strings.TrimSuffix(pa, ".go") + cmd.ext

	if len(file.Classes) == 0 {
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

	if err = cmd.generateHeader(out, file, func(w io.Writer) error {
		if cmd.enableHttpCodeWith {
			io.WriteString(out, `func httpCodeWith(err error) int {
  if herr, ok := err.(interface{
    HTTPCode() int
    }); ok {
      return herr.HTTPCode()
    }
  return http.StatusInternalServerError
}
`)
		}
		return nil
	}); err != nil {
		return err
	}

	for _, class := range file.Classes {
		if err := cmd.generateClass(out, file, &class); err != nil {
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

func (cmd *WebServerGenerator) generateClass(out io.Writer, file *SourceContext, class *Class) error {
	args := map[string]interface{}{"mux": cmd.Mux, "class": class}
	err := initFunc.Execute(out, args)
	if err != nil {
		return errors.New("generate initFunc for '" + class.Name.Name + "' fail, " + err.Error())
	}
	return nil
}

var initFunc *template.Template

func init() {
	initFunc = template.Must(template.New("InitFunc").Funcs(Funcs).Parse(`
func Init{{.class.Name}}(mux {{.mux.RouteParty}}, svc {{if not .class.IsInterface}}*{{end}}{{.class.Name}}) {
  {{- range $method := .class.Methods}}
  {{- $skipResult := $.mux.IsSkipped $method }}
  {{- if $skipResult.IsSkipped }}
  {{- if $skipResult.Message}} 
  // {{$method.Name.Name}}: {{$skipResult.Message}} 
  {{- end}}
  {{- else}}
  mux.{{$.mux.RouteFunc $method}}("{{$.mux.GetPath $method}}", {{$.mux.FuncSignature}}{
    {{- $hasInitParam := false}}
    {{- range $param := $method.Params.List}}
      {{- $initStatment := $.mux.InitParam $param }}
      {{- if $initStatment}}
      {{$initStatment}}
      {{- $hasInitParam = true}}
      {{- end}}
    {{- end}}
    {{- if $hasInitParam }}
    {{/* generate a empty line*/}}
    {{end}}
    {{- if eq 1 (len $method.Results.List) }}
      {{- $arg := index $method.Results.List 0}}
      {{- if eq "error" (typePrint $arg.Typ)}}
        resulterr 
      {{- else -}}
        result
      {{- end -}}
    {{- else -}}
    result, err
    {{- end -}} := svc.{{$method.Name}}(
      {{- range $idx, $param := $method.Params.List -}}
        {{- $.mux.UseParam $param }}
        {{- if isLast $method.Params.List $idx | not -}},{{- end -}}
      {{- end -}})

    {{- if eq 1 (len $method.Results.List) }}
      {{- $arg := index $method.Results.List 0}}
      {{- if eq "error" (typePrint $arg.Typ)}}
          if err != nil {
            {{$.mux.ErrorFunc $method "resulterr"}}
          }
          {{$.mux.OkFunc $method "\"OK\""}}        
      {{- else}}
          {{$.mux.OkFunc $method "result"}}
      {{- end}}
    {{- else}}
    if err != nil {
      {{$.mux.ErrorFunc $method "err"}}
    }
    {{$.mux.OkFunc $method "result"}}
    {{- end}}
  })
  {{- end}} {{/* isSkipped */}}
  {{- end}}
}
`))
}
