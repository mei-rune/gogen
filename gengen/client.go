package gengen

import (
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

type WebClientGenerator struct {
	GeneratorBase
	config ClientConfig
	file   string
}

func (cmd *WebClientGenerator) Flags(fs *flag.FlagSet) *flag.FlagSet {
	if cmd.ext == "" {
		cmd.ext = ".client-gen.go"
	}

	fs = cmd.GeneratorBase.Flags(fs)
	fs.StringVar(&cmd.file, "config", "", "配置文件名")
	fs.StringVar(&cmd.config.RestyName, "resty", "*resty.Proxy", "")
	fs.StringVar(&cmd.config.ContextClassName, "context", "context.Context", "")
	fs.StringVar(&cmd.config.newRequest, "new-request", "resty.NewRequest({{.proxy}},{{.url}})", "")
	fs.StringVar(&cmd.config.releaseRequest, "free-request", "resty.ReleaseRequest({{.proxy}},{{.request}})", "")
	return fs
}

func (cmd *WebClientGenerator) Run(args []string) error {
	if cmd.ext == "" {
		cmd.ext = ".client-gen.go"
	}

	if cmd.file != "" {
		cfg, err := readConfig(cmd.file)
		if err != nil {
			log.Fatalln(err)
			return err
		}

		if err := toStruct(cmd.config, cfg); err != nil {
			log.Fatalln(err)
			return err
		}

		if cmd.buildTag == "" {
			cmd.buildTag = stringWith(cfg, "features.buildTag", cmd.buildTag)
		}
		cmd.imports = readImports(cfg)
	} else {
		cmd.imports = map[string]string{
			"github.com/runner-mei/loong/resty": "",
		}
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

func (cmd *WebClientGenerator) generateClass(out io.Writer, file *SourceContext, class *Class) error {

	config := cmd.config

	config.ClassName = class.Name.Name + "Client"
	config.RecvClassName = config.ClassName
	ann, err := findAnnotation(class.Annotations, "http.Client")
	if err != nil {
		if ErrDuplicated == err {
			log.Fatalln(errors.New(strconv.Itoa(int(class.Node.Pos())) + ": annotations of class '" + class.Name.Name + "' is duplicated"))
		} else {
			log.Fatalln(err)
		}
		return err
	}
	if ann != nil {
		if name := ann.Attributes["name"]; name != "" {
			config.ClassName = name
			config.RecvClassName = config.ClassName
		}

		if reference := ann.Attributes["reference"]; reference == "true" {
			config.RecvClassName = "*" + config.ClassName
		}
	}

	args := map[string]interface{}{"config": &config, "class": class}
	err = clientTpl.Execute(out, args)
	if err != nil {
		return errors.New("generate classTpl for '" + class.Name.Name + "' fail, " + err.Error())
	}
	return nil
}

type ClientConfig struct {
	RestyName        string
	ContextClassName string
	ClassName        string
	RecvClassName    string

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
	return renderString(c.releaseRequest, map[string]interface{}{
		"proxy":   proxy,
		"request": request,
	})
}

func (c *ClientConfig) IsSkipped(method Method) SkippedResult {
	anno := getAnnotation(method, true)
	res := SkippedResult{
		IsSkipped: anno == nil,
	}
	if res.IsSkipped {
		res.Message = "annotation is missing"
	}
	return res
}

func (c *ClientConfig) RouteFunc(method Method) string {
	ann := getAnnotation(method, false)
	return strings.ToUpper(strings.TrimPrefix(ann.Name, "http."))
}

func (c *ClientConfig) GetPath(method Method, paramList []ParamConfig) string {
	anno := getAnnotation(method, false)

	rawurl := anno.Attributes["path"]
	if rawurl == "" {
		log.Fatalln(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": path(in annotation) of method '" + method.Itf.Name.Name + ":" + method.Name.Name + "' is missing"))
	}
	var replace = ReplaceFunc(func(segement PathSegement) string {
		for idx := range paramList {
			if paramList[idx].Name.Name == segement.Value {
				return "\" + " + convertToStringLiteral(paramList[idx].Param) + " + \""
			}
		}
		err := errors.New("path param '" + segement.Value + "' isnot found")
		log.Fatalln(err)
		panic(err)
	})
	segements, _, _ := parseURL(rawurl)
	return "\"" + JoinPathSegments(segements, replace) + "\""
}

type ParamConfig struct {
	Param

	QueryParamName string
	IsSkipDeclared bool
	IsSkipUse      bool
	IsQueryParam   bool
	IsPathParam    bool
	IsBodyParam    bool
	Values         []Param
}

func (c *ClientConfig) ToParamList(method Method) []ParamConfig {
	anno := getAnnotation(method, false)
	rawurl := anno.Attributes["path"]
	if rawurl == "" {
		log.Fatalln(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": path(in annotation) of method '" + method.Itf.Name.Name + ":" + method.Name.Name + "' is missing"))
	}

	data := anno.Attributes["data"]
	_, pathNames, queryNames := parseURL(rawurl)

	methodStr := strings.ToUpper(strings.TrimPrefix(anno.Name, "http."))
	isEdit := methodStr == "PUT" || methodStr == "POST"

	var inBody []Param
	var bodyExists bool

	paramList := make([]ParamConfig, 0, len(method.Params.List))
	for idx, param := range method.Params.List {
		cp := ParamConfig{
			Param:          param,
			QueryParamName: param.Name.Name,
		}

		typ := typePrint(param.Typ)
		if strings.HasSuffix(typ, ".Context") {
			cp.IsSkipDeclared = true
			paramList = append(paramList, cp)
			continue
		}

		for _, pname := range pathNames {
			if param.Name.Name == pname {
				cp.IsPathParam = true
				break
			}
		}
		if !cp.IsPathParam {
			if data != "" && data == param.Name.Name {
				cp.IsBodyParam = true
				bodyExists = true
			} else if newName, ok := queryNames[param.Name.Name]; ok && newName != "" {
				cp.QueryParamName = newName
			} else if isEdit {
				inBody = append(inBody, method.Params.List[idx])
				cp.IsSkipUse = true
				cp.IsBodyParam = true
			}
		}

		if !cp.IsPathParam && !cp.IsBodyParam {
			cp.IsQueryParam = true
		}

		paramList = append(paramList, cp)
	}

	if len(inBody) > 0 {
		if bodyExists {
			var names []string
			for _, a := range inBody {
				names = append(names, a.Name.Name)
			}
			err := errors.New(strconv.Itoa(int(method.Node.Pos())) + ": params '" + strings.Join(names, ",") + "' method '" + method.Itf.Name.Name + ":" + method.Name.Name + "' is invalid - ")
			log.Fatalln(err)
		}

		paramList = append(paramList, ParamConfig{
			IsSkipDeclared: true,
			IsBodyParam:    true,
			Values:         inBody,
		})
	}
	return paramList
}

type ResultConfig struct {
	Result

	JSONName  string
	FieldName string
}

func (c *ClientConfig) ToResultList(method Method) []ResultConfig {
	resultList := make([]ResultConfig, 0, len(method.Results.List))
	hasAnonymous := false
	for _, result := range method.Results.List {
		typ := typePrint(result.Typ)
		if typ == "error" {
			continue
		}

		cp := ResultConfig{
			Result: result,
		}
		if result.Name != nil {
			cp.FieldName = "E" + result.Name.Name
			cp.JSONName = result.Name.Name
			hasAnonymous = true
		}
		resultList = append(resultList, cp)
	}

	if hasAnonymous && len(resultList) > 1 {
		err := errors.New(strconv.Itoa(int(method.Node.Pos())) + ": method '" +
			method.Itf.Name.Name + ":" + method.Name.Name + "' is anonymous")
		log.Fatalln(err)
	}
	return resultList
}

var clientTpl *template.Template

func init() {
	clientTpl = template.Must(template.New("clientTpl").Funcs(Funcs).Parse(`

type {{.config.ClassName}} struct {
  proxy {{.config.RestyName}}
}

{{range $method := .class.Methods}}
  {{- $skipResult := $.config.IsSkipped $method }}
  {{- if $skipResult.IsSkipped }}
  {{- if $skipResult.Message}} 
  // {{$method.Name.Name}}: {{$skipResult.Message}} 
  {{- end}}
  {{- else}}

{{$paramList := ($.config.ToParamList $method) }}
{{$resultList := ($.config.ToResultList $method) }}
func (client {{$.config.RecvClassName}}) {{$method.Name}}(ctx {{$.config.ContextClassName}}{{- range $param := $paramList -}}
    {{- if $param.IsSkipDeclared -}}
    {{- else -}}
    {{- $typeName := typePrint $param.Typ -}}
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

  request := {{$.config.NewRequest "client.proxy" ($.config.GetPath $method $paramList) }}
  {{- $needAssignment := false -}}
  {{- range $param := $paramList -}}
    {{- if $param.IsSkipUse -}}
    {{- else if $param.IsBodyParam -}}
    {{- if $needAssignment}}
    request = request.
    {{- else -}}
    .
    {{end -}}
    {{- if $param.Values -}}
    SetBody(map[string]interface{}{
      {{- range $a := $param.Values}}
      "{{$a.Name.Name}}": {{$a.Name.Name}},
      {{- end}}
    })
    {{- else -}}
    SetBody({{$param.Name}})
    {{- end}}
    {{- $needAssignment = false -}}
    {{- else if $param.IsQueryParam -}}
    {{- $typeName := typePrint $param.Param.Typ -}}
    {{- if startWith $typeName "*"}}
    if {{$param.Name.Name}} != nil {
    	request = request.SetParam("{{$param.QueryParamName}}", {{convertToStringLiteral $param.Param}})
    }
    {{- $needAssignment = true -}}
    {{- else -}}
    {{- if $needAssignment}}
    request = request.
    {{- else -}}
    .
    {{end -}}
    SetParam("{{$param.QueryParamName}}", {{convertToStringLiteral $param.Param}})
    {{- $needAssignment = false -}}
    {{- end -}}
    {{- end -}}
  {{- end -}}

  {{- if $needAssignment}}
  request = request.Result(&result)
  {{- else -}}
  	.
  	Result(&result)
  {{- end}}

  {{if eq 0 (len $resultList) }}
  defer {{$.config.ReleaseRequest "client.proxy" "request"}}
  return {{else}}err := {{end}} request.{{$.config.RouteFunc $method}}(ctx)
  {{if eq 0 (len $resultList) }}
  {{else if eq 1 (len $resultList) }}
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
	{{end}}{{/* if $skipResult.IsSkipped */}}
{{end}}
`))
}
