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

	ToBindString(method Method, results []ServerParam) string
	ToParamList(method Method) []ServerParam

	FuncSignature() string
	RouteFunc(method Method) string
	BadArgumentFunc(method Method, err string, args ...string) string
	ErrorFunc(method Method, hasRealErrorCode bool, errCode, err string, args ...string) string
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

	var cfg map[string]interface{}
	if cmd.config != "" {
		var err error
		cfg, err = readConfig(cmd.config)
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
		mux.reinit(cfg)
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

	if mux := cmd.Mux.(*DefaultStye); mux != nil {
		mux.classes = file.Classes
	}

	for _, class := range file.Classes {

		count := 0
		for _, method := range class.Methods {
			if ann := getAnnotation(method, true); ann != nil {
				count++
			}
		}
		if count == 0 {
			out.WriteString("// " + class.Name.Name + " is skipped\r\n")
			continue
		}

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
  {{- $paramList := $.mux.ToParamList $method }}
  mux.{{$.mux.RouteFunc $method}}("{{$.mux.GetPath $method}}", {{$.mux.FuncSignature}}{
    {{- $hasInitParam := false}}
    {{- range $param := $paramList}}
      {{- if $param.InitString}}
      {{$param.InitString}}
      {{- $hasInitParam = true}}
      {{- end}}
    {{- end}}
    {{- $bindParams := $.mux.ToBindString $method $paramList}}
    {{- if $bindParams}}
      {{$bindParams}}
      {{- $hasInitParam = true}}
    {{- end}}
    {{- if $hasInitParam }}
    {{/* generate a empty line*/}}
    {{end}}
    {{- if gt (len $method.Results.List) 2 }}
      {{range $idx, $result := $method.Results.List -}}

        {{- if gt $idx 0 -}},{{- end -}}
        
        {{- if eq "error" (typePrint $result.Typ) -}}
          err
        {{- else -}}
          {{ $result.Name }}
        {{- end -}}
      
      {{- end -}}
    {{- else if eq 1 (len $method.Results.List) }}
      {{- $arg := index $method.Results.List 0}}
      {{- if eq "error" (typePrint $arg.Typ)}}
        resulterr 
      {{- else -}}
        result
      {{- end -}}
    {{- else -}}
    result, err
    {{- end -}} := svc.{{$method.Name}}(
      {{- $isFirst := true}}
      {{- range $idx, $param := $paramList}}
        {{- if $param.IsSkipUse -}}
        {{- else -}}
        {{- if $isFirst -}}{{- $isFirst = false -}}{{- else -}},{{- end -}}
        {{- $param.ParamName }}
        {{- end -}}
      {{- end -}})

    {{- if gt (len $method.Results.List) 2 }}
      if err != nil {
        {{$.mux.ErrorFunc $method false "httpCodeWith(err)" "err"}}
      }
      
      result := map[string]interface{}{
      {{- range $idx, $result := $method.Results.List}}        
        {{- if eq "error" (typePrint $result.Typ) -}}
        {{- else}}
          "{{ $result.Name }}": {{ $result.Name }},
        {{- end -}}
      {{- end}}
      }
      
      {{$.mux.OkFunc $method "result"}}
    
    {{- else if eq 1 (len $method.Results.List) }}
      {{- $arg := index $method.Results.List 0}}
      {{- if eq "error" (typePrint $arg.Typ)}}
          if resulterr != nil {
            {{$.mux.ErrorFunc $method false "httpCodeWith(err)" "resulterr"}}
          }
          {{$.mux.OkFunc $method "\"OK\""}}        
      {{- else}}
          {{$.mux.OkFunc $method "result"}}
      {{- end}}
    {{- else}}
    if err != nil {
      {{$.mux.ErrorFunc $method false "httpCodeWith(err)" "err"}}
    }
    {{$.mux.OkFunc $method "result"}}
    {{- end}}
  })
  {{- end}} {{/* isSkipped */}}
  {{- end}}
}
`))
}
