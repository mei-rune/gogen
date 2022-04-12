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
	ToParamList(method Method) ServerMethod

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
	preInitObject      bool
  ConvertNamespace string
}

func (cmd *WebServerGenerator) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs = cmd.GeneratorBase.Flags(fs)

	fs.StringVar(&cmd.config, "config", "", "配置文件名")
	fs.BoolVar(&cmd.preInitObject, "pre_init_object", false, "生成 enableHttpCodeWith 函数")
	fs.BoolVar(&cmd.enableHttpCodeWith, "httpCodeWith", false, "生成 enableHttpCodeWith 函数")
  fs.StringVar(&cmd.ConvertNamespace, "convert_ns", "", "转换函数的前缀")
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
		if cmd.preInitObject {
			cfg["pre_init_object"] = true
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

    mux.ConvertNamespace = cmd.ConvertNamespace


		mux.reinit(cfg)
	}



	if cmd.ext == "" {
		cmd.ext = ".gogen.go"
	}

	var includeFiles []*SourceContext
	for _, filename := range strings.Split(cmd.includes, ",") {
		if filename == "" {
			continue
		}
		pa, err := filepath.Abs(filename)
		if err != nil {
			return err
		}

		file, err := ParseFile(pa)
		if err != nil {
			return err
		}
		includeFiles = append(includeFiles, file)
	}

	if mux := cmd.Mux.(*DefaultStye); mux != nil {
		mux.includeFiles = includeFiles
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

  {{- $methodParams := $.mux.ToParamList $method }}
  {{- $paramList := $methodParams.ParamList }}
  {{- $IsErrorDefined := $methodParams.IsErrorDefined }}
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





		{{- /* invoke begin  */ -}}

    {{- if gt (len $method.Results.List) 2 }}
      {{range $idx, $result := $method.Results.List -}}

        {{- if gt $idx 0 -}},{{- end -}}
        
        {{- if eq "error" (typePrint $result.Typ) -}}
          err 
        {{- else -}}
          {{ $result.Name }} 
        {{- end -}}
      
      {{- end -}} :=
    {{- else if eq 1 (len $method.Results.List) }}
      {{- $arg := index $method.Results.List 0}}
      {{- if eq "error" (typePrint $arg.Typ)}}
        err {{if $IsErrorDefined }}={{else}}:={{end}}
      {{- else -}}
        result :=
      {{- end -}}
    {{- else -}}
    result, err :=
    {{- end -}} svc.{{$method.Name}}(
      {{- $isFirst := true}}
      {{- range $idx, $param := $paramList}}
        {{- if $param.IsSkipUse -}}
        {{- else -}}
        {{- if $isFirst -}}{{- $isFirst = false -}}{{- else -}},{{- end -}}
        {{- $param.ParamName }}{{if isEllipsisType $param.Param.Typ}}...{{- end -}}
        {{- end -}}
      {{- end -}})


		{{- /* invoke end  */ -}}




    {{- if gt (len $method.Results.List) 2 }}
      if err != nil {
        {{$.mux.ErrorFunc $method false "httpCodeWith(err)" "err"}}
      }
      
      result := map[string]interface{}{
      {{- range $idx, $result := $method.Results.List}}        
        {{- if eq "error" (typePrint $result.Typ) -}}
        {{- else}}
          "{{underscore $result.Name.Name }}": {{ $result.Name }},
        {{- end -}}
      {{- end}}
      }
      
      {{$.mux.OkFunc $method "result"}}
    
    {{- else if eq 1 (len $method.Results.List) }}
      {{- $arg := index $method.Results.List 0}}
      {{- if eq "error" (typePrint $arg.Typ)}}
          if err != nil {
            {{$.mux.ErrorFunc $method false "httpCodeWith(err)" "err"}}
          }
          {{$.mux.OkFunc $method "\"OK\""}}        
      {{- else}}
		    {{- if $methodParams.IsPlainText }}
		     	{{$.mux.PlainTextFunc $method "result"}}
		   	{{- else}}
          		{{$.mux.OkFunc $method "result"}}
      		{{- end}}
      {{- end}}
    {{- else}}
    if err != nil {
      {{$.mux.ErrorFunc $method false "httpCodeWith(err)" "err"}}
    }
    {{- if $methodParams.IsPlainText }}
      {{$.mux.PlainTextFunc $method "result"}}
    {{- else}}
      {{$.mux.OkFunc $method "result"}}
    {{- end}}
    {{- end}}
  })
  {{- end}} {{/* isSkipped */}}
  {{- end}}
}
`))
}
