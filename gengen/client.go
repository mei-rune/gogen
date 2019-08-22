package gengen

import (
	"errors"
	"flag"
	"go/ast"
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

	fs.StringVar(&cmd.config.TagName, "tag", "json", "")
	fs.StringVar(&cmd.config.RestyField, "field", "Proxy", "")
	fs.StringVar(&cmd.config.RestyName, "resty", "*resty.Proxy", "")
	fs.StringVar(&cmd.config.ContextClassName, "context", "context.Context", "")
	fs.StringVar(&cmd.config.newRequest, "new-request", "resty.NewRequest({{.proxy}},{{.url}})", "")
	fs.StringVar(&cmd.config.releaseRequest, "free-request", "resty.ReleaseRequest({{.proxy}},{{.request}})", "")

	fs.BoolVar(&cmd.config.HasWrapper, "has-wrapper", false, "")
	fs.StringVar(&cmd.config.WrapperType, "wrapper-type", "loong.Result", "")
	fs.StringVar(&cmd.config.WrapperData, "wrapper-data", "Data", "")
	fs.StringVar(&cmd.config.WrapperError, "wrapper-error", "Error", "")
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
			"github.com/runner-mei/resty": "",
			"github.com/runner-mei/loong": "",
		}
	}

	TimeFormat = "client." + cmd.config.RestyField + ".TimeFormat"

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

	config.Classes = file.Classes
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
	Classes          []Class
	TagName          string
	RestyName        string
	RestyField       string
	ContextClassName string
	ClassName        string
	RecvClassName    string
	HasWrapper       bool
	WrapperType      string
	WrapperData      string
	WrapperError     string

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

func (c *ClientConfig) ResultName(method Method) string {
	resultName := "result"
	isNameExist := func(name string) bool {
		if method.Params != nil {
			for idx := range method.Params.List {
				if method.Params.List[idx].Name.Name == name {
					return true
				}
			}
		}

		if method.Results != nil {
			for idx := range method.Results.List {
				if method.Results.List[idx].Name == nil {
					continue
				}
				if method.Results.List[idx].Name.Name == name {
					return true
				}
			}
		}
		return false
	}

	for i := 0; i < 100; i++ {
		if !isNameExist(resultName) {
			return resultName
		}
		resultName = resultName + "_"
	}
	panic("xxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
}

func (c *ClientConfig) RouteFunc(method Method) string {
	ann := getAnnotation(method, false)
	return strings.ToUpper(strings.TrimPrefix(ann.Name, "http."))
}

func (c *ClientConfig) GetPath(method Method, paramList []ParamConfig) string {
	anno := getAnnotation(method, false)

	rawurl := anno.Attributes["path"]
	//if rawurl == "" {
	// log.Fatalln(errors.New(method.Ctx.PostionFor(method.Node.Pos()).String() + ": path(in annotation) of method '" + method.Clazz.Name.Name + ":" + method.Name.Name + "' is missing"))
	//}
	var replace = ReplaceFunc(func(segement PathSegement) string {
		for idx := range paramList {
			if paramList[idx].Name.Name == segement.Value {
				return "\" + " + convertToStringLiteral(paramList[idx].Param) + " + \""
			}
		}
		err := errors.New(method.Ctx.PostionFor(method.Node.Pos()).String() + ": path param '" + segement.Value + "' isnot found")
		log.Fatalln(err)
		panic(err)
	})
	segements, _, _ := parseURL(rawurl)
	return "\"" + JoinPathSegments(segements, false, replace) + "\""
}

type ParamConfig struct {
	Param

	QueryParamName   string
	IsSkipDeclared   bool
	IsSkipUse        bool
	IsCodeSegement   bool
	IsQueryParam     bool
	IsMultQueryValue bool
	IsPathParam      bool
	IsBodyParam      bool
	Values           []Param
}

func (c *ClientConfig) ToParamList(method Method) []ParamConfig {
	anno := getAnnotation(method, false)
	rawurl := anno.Attributes["path"]
	// if rawurl == "" {
	//	log.Fatalln(errors.New(method.Ctx.PostionFor(method.Node.Pos()).String() + ": path(in annotation) of method '" + method.Clazz.Name.Name + ":" + method.Name.Name + "' is missing"))
	//}

	data := anno.Attributes["data"]
	_, pathNames, queryNames := parseURL(rawurl)

	methodStr := strings.ToUpper(strings.TrimPrefix(anno.Name, "http."))
	isEdit := methodStr == "PUT" || methodStr == "POST"

	var inBody []Param
	var bodyExists bool

	add := func(param Param, queryName string, isSkipDeclared, isSkipUse bool) ParamConfig {
		cp := ParamConfig{
			Param:          param,
			QueryParamName: queryName,
			IsSkipDeclared: isSkipDeclared,
			IsSkipUse:      isSkipUse,
		}

		typ := typePrint(param.Typ)
		if strings.HasSuffix(typ, ".Context") {
			cp.IsSkipDeclared = true
			return cp
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
				inBody = append(inBody, param)
				cp.IsSkipUse = true
				cp.IsBodyParam = true
			}
		}

		if cp.IsQueryParam || !cp.IsPathParam && !cp.IsBodyParam {
			cp.IsQueryParam = true

			if IsSliceType(param.Typ) || IsArrayType(param.Typ) {
				cp.IsMultQueryValue = true
			}
		}
		return cp
	}

	paramList := make([]ParamConfig, 0, len(method.Params.List))
	for _, param := range method.Params.List {

		if data != "" && data == param.Name.Name {
			paramList = append(paramList, add(param, param.Name.Name, false, false))
			continue
		}

		var prefix = param.Name.Name + "."
		if newName, ok := queryNames[param.Name.Name]; ok && newName != "" {
			if newName == "<none>" {
				prefix = ""
			} else {
				prefix = newName + "."
			}
		}

		var stType *Class
		if starType, ok := param.Typ.(*ast.StarExpr); ok {
			if identType, ok := starType.X.(*ast.Ident); ok {
				stType = method.Ctx.GetClass(identType.Name)
			}
		} else if identType, ok := param.Typ.(*ast.Ident); ok {
			stType = method.Ctx.GetClass(identType.Name)
		}
		if stType != nil {

			paramList = append(paramList, add(param, "", false, true))

			isPtr := IsPtrType(param.Typ)
			if isPtr {
				start := param
				start.Name = &ast.Ident{}
				start.Name.Name = "\r\nif " + param.Name.Name + " != nil {"
				paramList = append(paramList, ParamConfig{
					Param:          start,
					IsSkipDeclared: true,
					IsCodeSegement: true,
				})
			}

			for _, field := range stType.Fields {
				fieldParam := param
				fieldParam.Name = &ast.Ident{}
				*fieldParam.Name = *param.Name
				fieldParam.Name.Name = param.Name.Name + "." + field.Name.Name
				fieldParam.Typ = field.Typ

				queryName := prefix + field.Name.Name
				if field.Tag != nil {
					tagValue, _ := reflect.StructTag(field.Tag.Value).Lookup(c.TagName)
					if tagValue != "" {
						ss := strings.Split(tagValue, ",")
						if len(ss) > 0 && ss[0] != "" {
							queryName = prefix + "." + ss[0]
						}
					}
				}

				paramList = append(paramList, add(fieldParam, queryName, true, false))
			}

			if isPtr {
				end := param
				end.Name = &ast.Ident{}
				end.Name.Name = "\r\n}"
				paramList = append(paramList, ParamConfig{
					Param:          end,
					IsSkipDeclared: true,
					IsCodeSegement: true,
				})
			}
			continue
		}

		paramList = append(paramList, add(param, param.Name.Name, false, false))
	}

	if len(inBody) > 0 {
		if bodyExists {
			var names []string
			for _, a := range inBody {
				names = append(names, a.Name.Name)
			}
			err := errors.New(method.Ctx.PostionFor(method.Node.Pos()).String() + ": params '" + strings.Join(names, ",") + "' method '" + method.Clazz.Name.Name + ":" + method.Name.Name + "' is invalid - ")
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
			cp.JSONName = Underscore(result.Name.Name)
		} else {
			hasAnonymous = true
		}
		resultList = append(resultList, cp)
	}

	if hasAnonymous && len(resultList) > 1 {
		err := errors.New(method.Ctx.PostionFor(method.Node.Pos()).String() + ": method '" +
			method.Clazz.Name.Name + ":" + method.Name.Name + "' is anonymous")
		log.Fatalln(err)
	}
	return resultList
}

var clientTpl *template.Template

func init() {
	clientTpl = template.Must(template.New("clientTpl").Funcs(Funcs).Parse(`

type {{.config.ClassName}} struct {
  {{.config.RestyField}} {{.config.RestyName}}
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
{{$resultName := $.config.ResultName $method}}
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
      var {{$resultName}} {{trimPrefix (typePrint $result.Typ) "*"}}
    {{- end}}
  {{- else}}
      var {{$resultName}} struct{
    {{- range $result := $resultList}}
         {{$result.FieldName}} {{typePrint $result.Typ}} ` + "`json:\"{{$result.JSONName}}\"`" + `
    {{- end}}
      }
  {{- end}}
  {{- $hasWrapper := $.config.HasWrapper}}
  {{- if eq 0 (len $resultList) }}
  {{- $hasWrapper = false}}
  {{- end}}
  {{- if $hasWrapper }}
  var {{$resultName}}Wrap {{$.config.WrapperType}}
  {{- if ne (len $resultList) 0}}
	{{$resultName}}Wrap.Data = &{{$resultName}}
  {{- end}}
  {{- end}}
  {{- if gt (len $resultList) 0}}
  {{/* empty line */}}
  {{- end}}
  request := {{$.config.NewRequest (concat "client." $.config.RestyField) ($.config.GetPath $method $paramList) }}
  {{- $needAssignment := false -}}
  {{- range $param := $paramList -}}
    {{- if $param.IsSkipUse -}}
    {{- else if $param.IsCodeSegement -}}
    {{$param.Name}}
    {{- $needAssignment = true -}}    
    {{- else if $param.IsBodyParam -}}
    {{- if $needAssignment}}
    request = request.
    {{- else -}}
    .
    {{end -}}
    {{- if $param.Values -}}
    SetBody(map[string]interface{}{
      {{- range $a := $param.Values}}
      "{{underscore $a.Name.Name}}": {{$a.Name.Name}},
      {{- end}}
    })
    {{- else -}}
    SetBody({{$param.Name}})
    {{- end}}
    {{- $needAssignment = false -}}
    {{- else if $param.IsQueryParam -}}
      {{- $typeName := typePrint $param.Param.Typ -}}
      
      
      {{- if startWith $typeName "*http.Request"}}
      {{- else if startWith $typeName "url.Values"}}

          {{- if $needAssignment}}
          request = request.
          {{- else -}}
          .
          {{end -}}
          SetParams({{$param.Param.Name}})
          {{- $needAssignment = false -}}

      {{- else if startWith $typeName "*"}}
        if {{$param.Name.Name}} != nil {
        	request = request.SetParam("{{underscore $param.QueryParamName}}", {{convertToStringLiteral $param.Param}})
        }
        {{- $needAssignment = true -}}
      {{- else -}}
      
        {{- if $param.IsMultQueryValue }}
          for idx := range {{$param.Param.Name.Name}} {
            request = request.SetParam("{{underscore $param.QueryParamName}}", {{convertToStringLiteral2 "[idx]" $param.Param  true}})
          }
        {{- $needAssignment = true -}}
        {{- else}}    
          {{- if $needAssignment}}
          request = request.
          {{- else -}}
          .
          {{end -}}
          SetParam("{{underscore $param.QueryParamName}}", {{convertToStringLiteral $param.Param}})
          {{- $needAssignment = false -}}
        {{- end -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}

  {{- if gt (len $resultList) 0 -}}
  {{- if $needAssignment}}
  request = request.Result(&result)
  {{- else -}}
  	.
  	Result(&{{$resultName}}{{if $hasWrapper }}Wrap{{end}})
  {{- end}}
  {{- end}}

  {{if and (not $hasWrapper) (eq 0 (len $resultList)) }}
  defer {{$.config.ReleaseRequest (concat "client." $.config.RestyField) "request"}}
  return {{else}}err := {{end}} request.{{$.config.RouteFunc $method}}(ctx)
  {{- if eq 0 (len $resultList) }}
    
    {{- if $hasWrapper }}
      if err != nil {
        return err
      }
      if !{{$resultName}}Wrap.Success {
          if {{$resultName}}Wrap.Error == nil {
            return errors.New("error is nil")
          }
          return {{$resultName}}Wrap.Error
      }
      return nil
    {{- end}}
  
  {{- else if eq 1 (len $resultList) }}
    {{$.config.ReleaseRequest (concat "client." $.config.RestyField) "request"}}
  
    {{- $isPtr := false}}
    {{- $result := false}}
    {{- range $resulta := $resultList}}
      {{- $result = $resulta}}
      {{- if startWith (typePrint $resulta.Typ) "*"}}
        {{- $isPtr = true}}
      {{- end}}
    {{- end}}
  
    {{- if $hasWrapper }}
      if err != nil {
        return {{if $isPtr}}nil{{else}}{{zeroValue $result.Typ}}{{end}}, err
      }
      if !{{$resultName}}Wrap.Success {
        if {{$resultName}}Wrap.Error == nil {
            return {{if $isPtr}}nil{{else}}{{zeroValue $result.Typ}}{{end}}, errors.New("error is nil!!! wrapper is missing?")
        }
        return {{if $isPtr}}nil{{else}}{{zeroValue $result.Typ}}{{end}}, {{$resultName}}Wrap.Error
      }
      return {{if $isPtr}}&{{end}}{{$resultName}}, nil
    {{- else}}
      return {{if $isPtr}}&{{end}}{{$resultName}}, err
    {{- end}}
  {{- else}}
    {{$.config.ReleaseRequest (concat "client." $.config.RestyField) "request"}}
  
    if err != nil {
      return {{range $result := $resultList -}}
             {{zeroValue $result.Typ}},
             {{- end -}}err
    }

    {{- if $hasWrapper }} 
    if !{{$resultName}}Wrap.Success {
      if {{$resultName}}Wrap.Error == nil {
          return {{range $result := $resultList -}}
           {{zeroValue $result.Typ}},
           {{- end -}} errors.New("error is nil!!! wrapper is missing?")
      }
        
      return {{range $result := $resultList -}}
             {{zeroValue $result.Typ}},
             {{- end -}} {{$resultName}}Wrap.Error
    }
    {{- end}}
    return {{range $result := $resultList -}}
           {{$resultName}}.{{$result.FieldName}},
          {{- end -}} nil
  {{- end}}
}
	{{end}}{{/* if $skipResult.IsSkipped */}}
{{end}}
`))
}
