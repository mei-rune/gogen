package gengen

import (
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type WebClientGenerator struct {
	GeneratorBase
	config ClientConfig
}

func (cmd *WebClientGenerator) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs = cmd.GeneratorBase.Flags(fs)
	fs.StringVar(&cmd.config.RestyName, "resty", "Proxy", "")
	fs.StringVar(&cmd.config.ContextClassName, "context", "context.Context", "")
	fs.StringVar(&cmd.config.newRequest, "new-request", "NewRequest({{.proxy}},{{.url}})", "")
	fs.StringVar(&cmd.config.releaseRequest, "free-request", "ReleaseRequest({{.proxy}},{{.request}})", "")
	return fs
}

func (cmd *WebClientGenerator) Run(args []string) error {
	if cmd.ext == "" {
		cmd.ext = ".clientgen.go"
	}

	var e error
	for _, file := range args {
		if err := cmd.runFile(file); err != nil {
			log.Println(err)
			e = err
		}
	}

	return e
}

func (cmd *WebClientGenerator) runFile(filename string) error {
	pa, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

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

	if err = cmd.generateHeader(out, file, nil); err != nil {
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

func (cmd *WebClientGenerator) generateClass(out io.Writer, file *SourceContext, class *Class) error {
	args := map[string]interface{}{"config": cmd.config, "class": class}
	err := clientTpl.Execute(out, args)
	if err != nil {
		return errors.New("generate classTpl for '" + class.Name.Name + "' fail, " + err.Error())
	}
	return nil
}

type ClientConfig struct {
	RestyName        string
	ClassName        string
	RecvClassName    string
	ContextClassName string

	newRequest     string
	releaseRequest string
}

func (c *ClientConfig) NewRequest(proxy, url string) string {
	return renderString(c.newRequest, map[string]interface{}{
		"proxy": proxy,
		"url":   url,
	})
}
func (c *ClientConfig) ReleaseRequest(proxy, request string) string {
	return renderString(c.newRequest, map[string]interface{}{
		"proxy":   proxy,
		"request": request,
	})
}

func (c *ClientConfig) GetPath(method Method, paramList []ParamConfig) string {
	return "notimplemented"
}

type ParamConfig struct {
	Param

	QueryParamName string
	IsSkipped      bool
	IsQueryParam   bool
	IsPathParam    bool
	IsBodyParam    bool
}

func (c *ClientConfig) ToParamList(method Method) []ParamConfig {
	panic("notimplemented")
}

type ResultConfig struct {
	Result

	JSONName  string
	FieldName string
}

func (c *ClientConfig) ToResultList(method Method) []ResultConfig {
	panic("notimplemented")
}

var clientTpl *template.Template

func init() {
	clientTpl = template.Must(template.New("clientTpl").Funcs(Funcs).Parse(`

type {{.config.ClassName}} struct {
  proxy {{.config.RestyName}}
}

{{range $method := .class.Methods}}
{{$paramList := ($.config.ToParamList $method) }}
{{$resultList := ($.config.ToResultList $method) }}
func (client {{$.config.RecvClassName}}) {{$method.Name}}(ctx {{$.config.ContextClassName}}{{- range $param := $paramList -}}
    {{$typeName := typePrint $param.Typ}}
    {{- if $param.IsSkipped -}}
    {{- else -}}
    , {{$param.Name}} {{$typeName}}
    {{- end}}
  {{- end}}) ({{- range $result := $resultList -}}
      {{ typePrint $result.Typ}},
    {{- end -}}error) {

  {{- if eq 0 (len $resultList) }}
  {{- else if eq 1 (len $resultList) }}
    {{- range $result := $resultList}}
      var result {{trimPrefix (typePrint $result.Typ) "*"}}
    {{- end}}
  {{- else}}
      var result struct{
    {{- range $result := $resultList}}
         {{$result.FieldName}} {{typePrint $result.Typ}} ` + "`json:\"{{$result.JSONName}}\"`" + `
    {{- end}}
      }
  {{- end}}

  request := {{$.config.NewRequest "client.proxy" (concat "\"" ($.config.GetPath $method $paramList) "\"") }})
  {{- if eq 0 (len $resultList) }}
  defer {{$.config.ReleaseRequest "client.proxy" "request"}}
  return {{else}}err := {{end}} request.
  {{- range $param := $method.Params.List -}}
    {{- if $param.IsBodyParam -}}
    SetBody({{$param.Name}}).
    {{- else if $param.IsQueryParam -}}
    SetParam("{{$param.QueryParamName}}", {{$param.Name}}).
    {{- end}}
  {{- end}}
    Result(&result).
    GET(ctx)
  {{- if eq 0 (len $resultList) }}
  {{- else if eq 1 (len $resultList) }}
  {{$.config.ReleaseRequest "client.proxy" "request"}}
    {{- range $result := $resultList}}
  return {{if startWith (typePrint $result.Typ) "*"}}&{{end}}result, err
    {{- end}}
  {{- else}}
  {{$.config.ReleaseRequest "client.proxy" "request"}}
  return {{range $result := $resultList -}}
         result.{{$result.FieldName}},
        {{- end -}}, err
  {{- end}}
}
{{end}}

`))
}
