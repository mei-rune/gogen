package gengen

import (
	"errors"
	"fmt"
	"go/ast"
	"log"
	"runtime/debug"
	"strconv"
	"strings"
	"text/template"
)

type ReadArgs struct {
	Format string   `json:"format"`
	Name   string   `json:"name"`
	Args   []string `json:"args"`
}

type ConvertArgs struct {
	Format        string `json:"format"`
	NeedTransform bool   `json:"needTransform"`
	HasError      bool   `json:"hasError"`
}

type DefaultStye struct {
	includeFiles        []*SourceContext
	classes             []Class
	TagName             string            `json:"tag_name"`
	FuncSignatureStr    string            `json:"func_signature"`
	FuncHeadStr         string            `json:"func_head_str"`
	CtxNameStr          string            `json:"ctx_name"`
	CtxTypeStr          string            `json:"ctx_type"`
	RoutePartyName      string            `json:"route_party_name"`
	RequiredParamFormat string            `json:"required_param_format"`
	OptionalParamFormat string            `json:"optional_param_format"`
	ReadFormat          string            `json:"read_format"`
	ReadBodyFormat      string            `json:"read_body_format"`
	BadArgumentFormat   string            `json:"bad_argument_format"`
	OkFuncFormat        string            `json:"ok_func_format"`
	ErrorFuncFormat     string            `json:"err_func_format"`
	PlainTextFormat     string            `json:"plain_text_func_format"`
	PreInitObject       bool              `json:"pre_init_object"`
	Reserved            map[string]string `json:"reserved"`
	MethodMapping       map[string]string `json:"method_mapping"`
	Types               struct {
		Required map[string]ReadArgs `json:"required"`
		Optional map[string]ReadArgs `json:"optional"`
	} `json:"types"`
	ConvertNamespace string                                                    `json:"convertNS"`
	Converts         map[string]ConvertArgs                                    `json:"converts"`
	UrlStyle         string                                                    `json:"url_style"`
	ParseURL         func(rawurl string) (string, []string, map[string]string) `json:"-"`

	bodyReader string
	// readTemplate *template.Template
	bindTemplate      *template.Template
	errTemplate       *template.Template
	okTemplate        *template.Template
	plainTextTemplate *template.Template
}

func (mux *DefaultStye) Init() {
	mux.TagName = "json"
	mux.CtxNameStr = "ctx"
	mux.CtxTypeStr = "echo.Context"
	mux.FuncSignatureStr = "func(" + mux.CtxNameStr + " " + mux.CtxTypeStr + ") error "
	mux.RoutePartyName = "*echo.Group"
	mux.RequiredParamFormat = "{{.ctx}}.Param({{.name}})"
	mux.OptionalParamFormat = "{{.ctx}}.QueryParam({{.name}})"
	mux.ReadBodyFormat = "{{.ctx}}.Bind(&{{.name}})"
	mux.BadArgumentFormat = "fmt.Errorf(\"argument %%q is invalid - %%q\", \"%s\", %s, %s)"
	mux.Reserved = map[string]string{
		"*http.Request":       mux.CtxNameStr + ".Request()",
		"http.ResponseWriter": mux.CtxNameStr + ".Response().Writer",
		"context.Context":     mux.CtxNameStr + ".Request().Context()",
		"echo.Context":        mux.CtxNameStr,
		// "io.Reader":           mux.CtxNameStr + ".Request().Body",
	}
	//mux.Reserved["io.Reader"] = mux.bodyReader

	mux.OkFuncFormat = "return ctx.JSON({{.statusCode}}, {{.data}})"
	mux.ErrorFuncFormat = "ctx.Error({{.err}})\r\n     return nil"
	mux.PlainTextFormat = "return ctx.String({{.statusCode}}, {{.data}})"

}

func (mux *DefaultStye) reinit(values map[string]interface{}) {
	if mux.ParseURL == nil {
		var replace ReplaceFunc
		var canEmpty bool
		switch mux.UrlStyle {
		case "colon", "":
			replace = colonReplace
			canEmpty = true
		case "brace":
			replace = braceReplace
		default:
			log.Fatalln(errors.New("url_style '" + mux.UrlStyle + "' is invalid"))
		}

		mux.ParseURL = func(rawurl string) (string, []string, map[string]string) {
			segements, names, query := parseURL(rawurl)
			return JoinPathSegments(segements, canEmpty, replace), names, query
		}
	}

	if mux.TagName == "" {
		mux.TagName = "json"
	}

	if mux.Types.Required == nil {
		mux.Types.Required = map[string]ReadArgs{}
	}

	if mux.Types.Optional == nil {
		mux.Types.Optional = map[string]ReadArgs{}
	}

	if _, ok := mux.Types.Optional["string"]; !ok {
		mux.Types.Optional["string"] = ReadArgs{
			Name: mux.OptionalParamFormat,
		}
	}
	if _, ok := mux.Types.Required["string"]; !ok {
		mux.Types.Required["string"] = ReadArgs{
			Name: mux.RequiredParamFormat,
		}
	}

	if mux.Converts == nil {
		mux.Converts = map[string]ConvertArgs{}
	}

	if _, ok := mux.Converts["int"]; !ok {
		funcName := "strconv.Atoi({{.name}})"
		mux.Converts["int"] = ConvertArgs{Format: funcName, HasError: true}
	}

	if _, ok := mux.Converts["[]bool"]; !ok {
		funcName := mux.ConvertNamespace + "ToBoolArray({{.name}})"
		mux.Converts["[]bool"] = ConvertArgs{Format: funcName, HasError: true}
	}
	if _, ok := mux.Converts["[]int"]; !ok {
		funcName := mux.ConvertNamespace + "ToIntArray({{.name}})"
		mux.Converts["[]int"] = ConvertArgs{Format: funcName, HasError: true}
	}
	if _, ok := mux.Converts["[]int64"]; !ok {
		funcName := mux.ConvertNamespace + "ToInt64Array({{.name}})"
		mux.Converts["[]int64"] = ConvertArgs{Format: funcName, HasError: true}
	}

	if _, ok := mux.Converts["[]uint"]; !ok {
		funcName := mux.ConvertNamespace + "ToUintArray({{.name}})"
		mux.Converts["[]uint"] = ConvertArgs{Format: funcName, HasError: true}
	}
	if _, ok := mux.Converts["[]uint64"]; !ok {
		funcName := mux.ConvertNamespace + "ToUint64Array({{.name}})"
		mux.Converts["[]uint64"] = ConvertArgs{Format: funcName, HasError: true}
	}
	if _, ok := mux.Converts["[]time.Time"]; !ok {
		funcName := mux.ConvertNamespace + "ToDatetimes({{.name}})"
		mux.Converts["[]time.Time"] = ConvertArgs{Format: funcName, HasError: true}
	}

	for _, t := range []string{"int8", "int16", "int32", "int64"} {
		if _, ok := mux.Converts[t]; ok {
			continue
		}
		conv := ConvertArgs{Format: "strconv.ParseInt({{.name}}, 10, 64)", HasError: true}
		if !strings.HasSuffix(t, "64") {
			conv.NeedTransform = true
		}
		mux.Converts[t] = conv
	}
	for _, t := range []string{"uint", "uint8", "uint16", "uint32", "uint64"} {
		if _, ok := mux.Converts[t]; ok {
			continue
		}
		conv := ConvertArgs{Format: "strconv.ParseUint({{.name}}, 10, 64)", HasError: true}
		if !strings.HasSuffix(t, "64") {
			conv.NeedTransform = true
		}
		mux.Converts[t] = conv
	}
	if _, ok := mux.Converts["bool"]; !ok {
		funcName := stringWith(values, "features.boolConvert", mux.ConvertNamespace+"ToBool({{.name}})")
		mux.Converts["bool"] = ConvertArgs{Format: funcName, HasError: false}
	}
	if _, ok := mux.Converts["time.Time"]; !ok {
		funcName := stringWith(values, "features.datetimeConvert", mux.ConvertNamespace+"ToDatetime({{.name}})")
		mux.Converts["time.Time"] = ConvertArgs{Format: funcName, HasError: true}
	}
	if _, ok := mux.Converts["time.Duration"]; !ok {
		mux.Converts["time.Duration"] = ConvertArgs{Format: "time.ParseDuration({{.name}})", HasError: true}
	}
	if _, ok := mux.Converts["sql.NullBool"]; !ok {
		funcName := stringWith(values, "features.boolConvert", mux.ConvertNamespace+"ToBool({{.name}})")
		mux.Converts["sql.NullBool"] = ConvertArgs{Format: funcName, HasError: false}
	}
	if _, ok := mux.Converts["sql.NullTime"]; !ok {
		funcName := stringWith(values, "features.datetimeConvert", mux.ConvertNamespace+"ToDatetime({{.name}})")
		mux.Converts["sql.NullTime"] = ConvertArgs{Format: funcName, HasError: true}
	}
	if _, ok := mux.Converts["sql.NullInt64"]; !ok {
		mux.Converts["sql.NullInt64"] = ConvertArgs{Format: "strconv.ParseInt({{.name}}, 10, 64)", HasError: true}
	}
	if _, ok := mux.Converts["sql.NullUint64"]; !ok {
		mux.Converts["sql.NullUint64"] = ConvertArgs{Format: "strconv.ParseUint({{.name}}, 10, 64)", HasError: true}
	}

	mux.bodyReader = mux.Reserved["*http.Request"] + ".Body"
	mux.bindTemplate = template.Must(template.New("bindTemplate").Parse(mux.ReadBodyFormat))
	mux.errTemplate = template.Must(template.New("errTemplate").Parse(mux.ErrorFuncFormat))
	mux.okTemplate = template.Must(template.New("okTemplate").Parse(mux.OkFuncFormat))
	mux.plainTextTemplate = template.Must(template.New("plainTextTemplate").Parse(mux.PlainTextFormat))
}

func stringWith(values map[string]interface{}, key, defValue string) string {
	o := values[key]
	if o == nil {
		return defValue
	}
	return o.(string)
}

func intWith(values map[string]interface{}, key string, defValue int) int {
	o := values[key]
	if o == nil {
		return defValue
	}
	s := fmt.Sprint(o)
	if s == "" {
		return defValue
	}
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return defValue
}
func boolWith(values map[string]interface{}, key string, defValue bool) bool {
	o := values[key]
	if o == nil {
		return defValue
	}
	if b, ok := o.(bool); ok {
		return b
	}
	s := fmt.Sprint(o)
	if s == "" {
		return defValue
	}
	s = strings.ToLower(s)
	return s == "true" || s == "on" || s == "yes" || s == "enabled"
}

func (mux *DefaultStye) RouteParty() string {
	return mux.RoutePartyName
}

func (mux *DefaultStye) FuncSignature() string {
	return mux.FuncSignatureStr
}

func (mux *DefaultStye) CtxName() string {
	return mux.CtxNameStr
}

func (mux *DefaultStye) CtxType() string {
	return mux.CtxTypeStr
}

func (mux *DefaultStye) IsSkipped(method Method) SkippedResult {
	anno := getAnnotation(method, true)
	res := SkippedResult{
		IsSkipped: anno == nil,
	}
	if res.IsSkipped {
		res.Message = "annotation is missing"
	}
	return res
}

func (mux *DefaultStye) ReadRequired(param Param, typeName, ctxName, paramName string) string {
	format := mux.RequiredParamFormat
	if args, ok := mux.Types.Required[typeName]; ok {
		if args.Format != "" {
			format = args.Format
		}
		if len(args.Args) > 0 {
			paramName = paramName + "," + strings.Join(args.Args, ",")
		}
	}

	return renderString(format, map[string]interface{}{"ctx": ctxName,
		"name": paramName,
	})
}

func (mux *DefaultStye) ReadOptional(param Param, typeName, ctxName, paramName string) string {
	if strings.HasPrefix(typeName, "[]") {
		typeName = "[]string"
	}
	format := mux.OptionalParamFormat
	if args, ok := mux.Types.Optional[typeName]; ok {
		if args.Format != "" {
			format = args.Format
		}
		if len(args.Args) > 0 {
			paramName = paramName + "," + strings.Join(args.Args, ",")
		}
	}

	return renderString(format, map[string]interface{}{"ctx": ctxName,
		"name": paramName,
	})
}

func (mux *DefaultStye) TypeConvert(param Param, typeName, ctxName, paramName string) string {
	var sb strings.Builder

	format, ok := mux.Converts[typeName]
	if !ok {
		underlying := param.Method.Ctx.GetType(typeName)
		if underlying == nil {
			debug.PrintStack()
			log.Fatalln(param.Method.Ctx.PostionFor(param.Method.Node.Pos()), ": 1argument '"+param.Name.Name+"' is unsupported type -", typeName)
		}

		if typePrint(underlying.Type) == "string" {
			return paramName
		}
		format, ok = mux.Converts[typePrint(underlying.Type)]
		if !ok {
			log.Fatalln(param.Method.Ctx.PostionFor(param.Method.Node.Pos()), ": 2argument '"+param.Name.Name+"' is unsupported type -", typeName)
		}

		format.NeedTransform = true
	}

	tpl := template.Must(template.New("convertTemplate").Parse(format.Format))
	renderText(tpl, &sb, map[string]interface{}{
		"ctx":  ctxName,
		"name": paramName,
	})
	return sb.String()
}

func (mux *DefaultStye) ReadBody(ctxName, paramName string) string {
	var sb strings.Builder
	renderText(mux.bindTemplate, &sb, map[string]interface{}{"ctx": ctxName, "name": paramName})
	return sb.String()
}

func (mux *DefaultStye) GetPath(method Method) string {
	anno := getAnnotation(method, false)

	rawurl := anno.Attributes["path"]
	//if rawurl == "" {
	//	log.Fatalln(errors.New(method.Ctx.PostionFor(method.Node.Pos()).String() + ": path(in annotation) of method '" + method.Clazz.Name.Name + ":" + method.Name.Name + "' is missing"))
	//}
	pa, _, _ := mux.ParseURL(rawurl)
	return pa
}

type ServerParam struct {
	Param

	IsErrorDefined bool
	IsSkipUse      bool
	InBody         bool
	ParamName      string
	InitName       string

	IsAnonymous   bool
	SkipDeclare   bool
	IsArray       bool
	ArgNamePrefix string
	ArgName       string

	InitString string
}

func (param ServerParam) ToJSONName() string {
	anno := getAnnotation(*param.Method, false)
	autoUnderscore := anno.Attributes["auto_underscore"]

	if strings.ToLower(autoUnderscore) == "false" {
		return param.Name.Name
	}
	return Underscore(param.Name.Name)
}

type ServerMethod struct {
	ParamList      []ServerParam
	IsErrorDefined bool
	IsPlainText    bool
}

func (mux *DefaultStye) ToBindString(method Method, results []ServerParam) string {
	has := false
	for idx := range results {
		if results[idx].InBody {
			has = true
			break
		}
	}

	if !has {
		return ""
	}

	funcs := template.FuncMap{
		"badArgument": func(paramName, valueName, errName string) string {
			return mux.BadArgumentFunc(method, fmt.Sprintf(mux.BadArgumentFormat, paramName, valueName, errName))
		},
		"readBody": func(ctxName, paramName string) string {
			return mux.ReadBody(ctxName, paramName)
		},
	}

	bindTxt := template.Must(template.New("bindTxt").Funcs(Funcs).Funcs(funcs).Parse(`
			var bindArgs struct {
				{{- range $param := .params}}
					{{- if $param.InBody }}
  						{{goify $param.Param.Name.Name true}} {{typePrint $param.Param.Typ}} ` + "`json:\"{{$param.ToJSONName}},omitempty\"`" + `
  					{{- end}}
				{{- end}}
			}
			if err := {{readBody .ctx "bindArgs"}}; err != nil {
				{{badArgument "bindArgs" "\"body\"" "err"}}
			}
		`))

	// serverParam.InBody = true
	// serverParam.InitString = ""
	// serverParam.ParamName = "bindArgs." + Goify(serverParam.Param.Name.Name, true)
	var sb strings.Builder
	renderText(bindTxt, &sb, map[string]interface{}{
		"ctx":    mux.CtxName(),
		"params": results,
	})
	return strings.TrimSpace(sb.String())
}

func (mux *DefaultStye) ToParamList(method Method) ServerMethod {
	defer func() {
		if o := recover(); o != nil {
			fmt.Println(o)

			debug.PrintStack()
			panic(o)
		}
	}()
	var genCtx context
	var results []ServerParam

	ann := getAnnotation(method, false)
	methodStr := strings.ToUpper(strings.TrimPrefix(ann.Name, "http."))
	hasBody := methodStr == "PUT" || methodStr == "POST" || methodStr == "PATCH"

	var optParamIdx = -1
	var optArgumentIdx = -1

	for idx := range method.Params.List {
		old := genCtx.IsNeedQuery()

		if !hasBody && typePrint(method.Params.List[idx].Typ) == "map[string]string" {
			if optParamIdx < 0 {
				optParamIdx = idx

				if !genCtx.IsNeedQuery() {
					if mux.FuncHeadStr != "" {
						results = append(results, ServerParam{
							IsSkipUse:  true,
							ArgName:    "<noArgname>",
							InitString: mux.FuncHeadStr,
						})
						genCtx.SetNeedQuery()
					}
				}

				results = append(results, ServerParam{
					Param:     method.Params.List[idx],
					IsSkipUse: false,
					ArgName:   "<noArgname>",
					ParamName: method.Params.List[idx].Name.Name,
				})
				optArgumentIdx = len(results) - 1

			} else {
				err := errors.New(method.Ctx.PostionFor(method.Node.Pos()).String() +
					": param1 '"+method.Params.List[idx].Name.Name+"' of '" + method.Clazz.Name.Name + ":" + method.Name.Name + "' is invalid")
				log.Fatalln(err)
			}
			continue
		}

		segments := mux.ToParam(&genCtx, method, method.Params.List[idx], hasBody)
		if !old && genCtx.IsNeedQuery() {
			if mux.FuncHeadStr != "" {
				results = append(results, ServerParam{
					IsSkipUse:  true,
					ArgName:    "<noArgname>",
					InitString: mux.FuncHeadStr,
				})
			}
		}
		results = append(results, segments...)
	}

	if optParamIdx >= 0 {
		param := method.Params.List[optParamIdx]

		var queryParamName = mux.Reserved["url.Values"]
		if mux.FuncHeadStr != "" {
			queryParamName = "queryParams"
		}

		var sb strings.Builder

		for i := range results {
			if !results[i].InBody && results[i].ArgName != "<noArgname>" {
				if sb.Len() > 0 {
					sb.WriteString(" || \r\n")
				}
				sb.WriteString("key == \"")
				if results[i].ArgNamePrefix != "" {
					sb.WriteString(results[i].ArgNamePrefix)
				} else {
					sb.WriteString(results[i].ArgName)
				}
				sb.WriteString("\" ")
			}
		}

		var continueStr = ""
		if sb.Len() > 0 {
			continueStr = `
						if ` + sb.String() + ` {
							continue
						}
						`
		}

		results[optArgumentIdx].InitString = `var ` + param.Name.Name + ` = map[string]string{}
					for key, values := range ` + queryParamName + `{` + continueStr + `
						` + param.Name.Name + `[key] = values[len(values)-1]
					}					
					`
	}

	_, hasData := ann.Attributes["data"]
	if hasData {
		for idx := range results {
			if results[idx].InBody {
				err := errors.New(method.Ctx.PostionFor(method.Node.Pos()).String() +
					": param '"+results[idx].Name.Name+"' of '" + method.Clazz.Name.Name + ":" + method.Name.Name + "' is invalid")
				log.Fatalln(err)
			}
		}
	}

	isPlainText := ann.Attributes["content_type"] == "text"
	if isPlainText {
		isText := false
		if len(method.Results.List) == 2 {
			// if typePrint(method.Results.List[0].Typ) != "context.Context" {
			// 	panic("content_type is mismatch - '" + typePrint(method.Results.List[0].Typ) + "' in the '" + method.Name.Name + "'")
			// }
			s := typePrint(method.Results.List[0].Typ)
			isText = s == "string"
		} else if len(method.Results.List) == 1 {
			s := typePrint(method.Results.List[0].Typ)
			isText = s == "string"
		}

		if !isText {
			panic("content_type is mismatch - '" + typePrint(method.Results.List[0].Typ) + "' in the '" + method.Name.Name + "'")
		}
	}

	return ServerMethod{results, genCtx.IsErrorDefined(), isPlainText}
}

func (mux *DefaultStye) ToParam(c *context, method Method, param Param, isEdit bool) []ServerParam {

	typeStr := typePrint(param.Typ)
	if strings.HasPrefix(typeStr, "...") {
		typeStr = "[]" + strings.TrimPrefix(typeStr, "...")
	}
	elmType := strings.TrimPrefix(typeStr, "*")
	hasStar := typeStr != elmType

	funcs := template.FuncMap{
		"badArgument": func(paramName, valueName, errName string) string {
			return mux.BadArgumentFunc(method, fmt.Sprintf(mux.BadArgumentFormat, paramName, valueName, errName))
		},
		"readBody": func(param Param, ctxName string, paramName string) string {
			return mux.ReadBody(ctxName, paramName)
		},
		"readOptional": func(param Param, ctxName, elmType, paramName string) string {
			c.SetNeedQuery()
			return mux.ReadOptional(param, elmType, ctxName, paramName)
		},
		"readRequired": func(param Param, ctxName, elmType, paramName string) string {
			return mux.ReadRequired(param, elmType, ctxName, paramName)
		},
		"convert": func(param Param, ctxName, elmType, paramName string) string {
			return mux.TypeConvert(param, elmType, ctxName, paramName)
		},
		"concat": func(args ...string) string {
			var sb strings.Builder
			for _, s := range args {
				sb.WriteString(s)
			}
			return sb.String()
		},
	}

	serverParam := ServerParam{
		Param:     param,
		ParamName: param.Name.Name,

		IsErrorDefined: c.IsErrorDefined(),
	}

	var paramName = param.Name.Name
	var name = serverParam.ParamName
	if name == "result" {
		name = "result_"
		serverParam.ParamName = "result_"
	}

	anno := getAnnotation(*param.Method, false)
	inBody := anno.Attributes["data"] == param.Name.Name
	if inBody {
		if typeStr == "io.Reader" {
			serverParam.ParamName = mux.bodyReader
			return []ServerParam{serverParam}
		}

		if hasStar {
			serverParam.ParamName = "&" + serverParam.ParamName
		}

		if typeStr == "[]byte" {
			serverParam.ParamName = paramName + ".Bytes()"
		} else if typeStr == "string" {
			serverParam.ParamName = paramName + ".String()"
		}

		bindTxt := template.Must(template.New("bindTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- if eq .type "[]byte"}}
      {{- if eq .rawType "*[]byte"}}
      	var {{.name}}Buffer bytes.Buffer
    		if _, err := io.Copy(&{{.name}}Buffer, {{.reader}}); err != nil {
    			{{badArgument .name "\"body\"" "err"}}
    		}
        var {{.name}} = {{.name}}Buffer.Bytes()
      {{- else}}
    		var {{.name}} bytes.Buffer
    		if _, err := io.Copy(&{{.name}}, {{.reader}}); err != nil {
    			{{badArgument .name "\"body\"" "err"}}
    		}
      {{- end}}
		{{- else if eq .type "string"}}    
      {{- if eq .rawType "*string"}}
      	var {{.name}}Builder strings.Builder
    		if _, err := io.Copy(&{{.name}}Builder, {{.reader}}); err != nil {
    			{{badArgument .name "\"body\"" "err"}}
    		}
        var {{.name}} = {{.name}}Builder.String()
      {{- else}}
      	var {{.name}} strings.Builder
    		if _, err := io.Copy(&{{.name}}, {{.reader}}); err != nil {
    			{{badArgument .name "\"body\"" "err"}}
    		}
      {{- end}}
		{{- else}}
		{{- if .isMapType }}
		var {{.name}} = {{.type}}{}
		{{- else}}
		var {{.name}} {{.type}}
		{{- end}}
		if err := {{readBody .param .ctx .name}}; err != nil {
			{{badArgument .name "\"body\"" "err"}}
		}
		{{- end}}
		`))

		if dataType := anno.Attributes["dataType"]; dataType != "" {
			elmType = dataType
		} else if dataType := anno.Attributes["datatype"]; dataType != "" {
			elmType = dataType
		}

		renderArgs := map[string]interface{}{
			"g":         c,
			"ctx":       mux.CtxName(),
			"rawType":   typeStr,
			"type":      elmType,
			"isMapType": strings.HasPrefix(elmType, "map["),
			"name":      name,
			"rname":     paramName,
			"param":     param,
			"reader":    mux.bodyReader,
		}

		var sb strings.Builder
		renderText(bindTxt, &sb, renderArgs)
		serverParam.InitString = strings.TrimSpace(sb.String())
		return []ServerParam{serverParam}
	} else if s, ok := mux.Reserved[typeStr]; ok {
		serverParam.ParamName = s
		return []ServerParam{serverParam}
	} else if s, ok := mux.Reserved[elmType]; ok {
		serverParam.ParamName = s
		return []ServerParam{serverParam}
	}

	_, pathNames, queryNames := mux.ParseURL(anno.Attributes["path"])

	var optional = true

	isPath := false
	for _, pa := range pathNames {
		if pa == param.Name.Name {
			isPath = true
			optional = false
			break
		}

		if pa == param.Name.Name {
			isPath = true
			optional = false
			break
		}
	}
	if !isPath {
		optional = true

		if s, ok := queryNames[param.Name.Name]; ok && s != "" {
			paramName = s
		} else if isEdit {

			serverParam.InBody = true
			serverParam.InitString = ""
			serverParam.ParamName = "bindArgs." + Goify(serverParam.Param.Name.Name, true)
			return []ServerParam{serverParam}
		} else {
			paramName = Underscore(paramName)
		}
	} else if hasStar {
		if typeStr == "*string" {
			serverParam.ParamName = "&" + name
		} else if mux.Converts != nil {
			convertArgs, ok := mux.Converts[strings.TrimPrefix(typeStr, "*")]
			if ok {
				if !convertArgs.HasError {
					serverParam.ParamName = "&" + name
				}
			}
		}
	}

	renderArgs := map[string]interface{}{
		"g":             c,
		"ctx":           mux.CtxName(),
		"type":          elmType,
		"name":          name,
		"rname":         paramName,
		"param":         param,
		"initRootValue": "",
		"isArray":       IsArrayType(param.Typ) || IsSliceType(param.Typ) || IsEllipsisType(param.Typ),
	}

	getStructType := func(param Param) *Class {
		var stType *Class
		if starType, ok := param.Typ.(*ast.StarExpr); ok {
			if identType, ok := starType.X.(*ast.Ident); ok {
				stType = method.Ctx.GetClass(identType.Name)
				if stType == nil {
					for _, ctx := range mux.includeFiles {
							stType = ctx.GetClass(identType.Name)	
					}
					// if stType == nil && !isPrimaryType(identType.Name) {
					// 		panic(identType.Name + " isnot found-1")
					// }
				}
			} else if selectorExpr, ok := starType.X.(*ast.SelectorExpr); ok {
				for _, ctx := range mux.includeFiles {
					if ctx.Pkg.Name == fmt.Sprint(selectorExpr.X) {
						stType = ctx.GetClass(selectorExpr.Sel.Name)
					}

					// if stType == nil && !isPrimaryType(identType.Name) {
					// 		panic(selectorExpr.Sel.Name + " isnot found-2")
					// 	}
				}
			}
		} else if identType, ok := param.Typ.(*ast.Ident); ok {
			stType = method.Ctx.GetClass(identType.Name)
		} else if selectorExpr, ok := param.Typ.(*ast.SelectorExpr); ok {
			for _, ctx := range mux.includeFiles {
				if ctx.Pkg.Name == fmt.Sprint(selectorExpr.X) {
					stType = ctx.GetClass(selectorExpr.Sel.Name)
					// if stType == nil && !isPrimaryType(identType.Name) {
					// 		panic(selectorExpr.Sel.Name + " isnot found-3")
					// 	}
				}
			}
		}

		if stType != nil && stType.AliasName != nil {
			found := false
			for _, ctx := range mux.includeFiles {
				if ctx.Pkg.Name == fmt.Sprint(stType.AliasName.X) {
					stType = ctx.GetClass(stType.AliasName.Sel.Name)
					found = true
				}
			}

			if !found {
				panic(fmt.Sprint(stType.AliasName.X) + "." + stType.AliasName.Sel.Name + " isnot found-4")
			}
		}
		return stType
	}

	stType := getStructType(param)
	if stType != nil {
		if isPath {
			err := errors.New(strconv.Itoa(int(method.Node.Pos())) + ": argument '" + param.Name.Name + "' of method '" + method.Clazz.Name.Name + ":" + method.Name.Name + "' is invalid")
			log.Fatalln(err)
			panic(err)
		}

		p1 := serverParam
		p1.InitString = "var " + name + " " + typeStr
		if mux.PreInitObject {
			p1.InitString = "var " + name + " " + elmType
			if hasStar {
				p1.ParamName = "&" + name
			}
		}

		var paramNamePrefix = Underscore(param.Name.Name) + "."
		if s, ok := queryNames[param.Name.Name]; ok {
			if s == "" || s == "<none>" {
				paramNamePrefix = ""
			} else {
				paramNamePrefix = s + "."
			}
		}

		c.parentInited = false
		p1.IsArray = false
		p1.ArgName = "<noArgname>"

		serverParams := []ServerParam{p1}

		var addStructFields func(string, *Class, ServerParam, []int, []ServerParam)
		addStructFields = func(paramNamePrefix string, stType *Class, serverParam ServerParam, parentIndexs []int, parents []ServerParam) {
			for fieldIdx, field := range stType.Fields {
				tag := field.GetTag("gogen")
				if tag == "ignore" {
					continue
				}
				p2 := serverParam
				p2.IsSkipUse = true
				p2.Name = &ast.Ident{}
				*p2.Name = *serverParam.Name
				p2.Param.Typ = field.Typ

				if field.Name == nil {
					typeName := typePrint(field.Typ)
					if IsPtrType(field.Typ) {
						typeName = typePrint(ElemType(field.Typ))
					}
					idx := strings.Index(typeName, ".")
					if idx >= 0 {
						typeName = typeName[idx+1:]
					}
					p2.Param.Name.Name = serverParam.Name.Name
					p2.InitName = serverParam.Name.Name + "." + typeName
					p2.IsAnonymous = true
				} else {
					p2.Param.Name.Name = serverParam.Name.Name + "." + field.Name.Name
					p2.InitName = serverParam.Name.Name + "." + field.Name.Name

					p2.IsAnonymous = false
				}

				paramName2 := strings.TrimSuffix(paramNamePrefix, ".")
				if field.Name != nil {
					paramName2 = paramNamePrefix + Underscore(field.Name.Name)
				}
				if field.Tag != nil {
					tagValue := field.GetTag(mux.TagName)
					if tagValue != "" {
						ss := strings.Split(tagValue, ",")
						if len(ss) > 0 && ss[0] != "" {
							paramName2 = paramNamePrefix + ss[0]
						}
					}
				}

				if fieldIdx > 0 && len(parents) > 0 {
					for idx := len(parents) - 1; idx >= 0; idx-- {
						if parents[idx].IsAnonymous {
							parentIndexs[idx] = fieldIdx
						}
					}
				}

				stType := getStructType(p2.Param)

				if stType != nil {
					if paramName2 != "" && !strings.HasPrefix(paramName2, ".") {
						paramName2 = paramName2 + "."
					}
					newParentIndexs := append(parentIndexs, 0)
					newParents := append(parents, p2)

					// fmt.Println("====")
					// fmt.Println(method.Name, paramName2, newParentIndexs)
					// for idx := range newParents {
					// 	fmt.Println(newParents[idx].InitName)
					// }

					addStructFields(paramName2, stType, p2, newParentIndexs, newParents)
					continue
				}

				if typePrint(field.Typ) == "url.Values" {
					s := queryNames[serverParam.Name.Name]
					if field.Name != nil || s != "<none>" {
						p2.SkipDeclare = true
						// p2.TypeStr = strings.TrimPrefix(typePrint(p2.Param.Typ), "*")
						p2.IsArray = false
						p2.ArgName = paramName2
						p2.ArgNamePrefix = paramNamePrefix

						var queryParamName = mux.Reserved["url.Values"]
						if mux.FuncHeadStr != "" {
							queryParamName = "queryParams"
						}

						var initRootValue string
						for idx, parent := range parents {
							if idx == 0 && mux.PreInitObject {
								continue
							}
							if IsPtrType(parent.Param.Typ) {
								if parentIndexs[idx] == 0 && fieldIdx == 0 {
									initRootValue += parent.InitName + " = &" + typePrint(ElemType(parent.Param.Typ)) + "{}\r\n"
								} else {
									initRootValue += `if ` + parent.InitName + ` == nil {
								` + parent.InitName + ` = &` + typePrint(ElemType(parent.Param.Typ)) + `{}
							}
							`
								}
							}
						}
						// if IsPtrType(serverParam.Param.Typ) {
						// 	initRootValue += `if ` + serverParam.InitName + ` == nil {
						// 		` + serverParam.InitName + ` = &` + typePrint(ElemType(serverParam.Param.Typ)) + `{}
						// 	}
						// 	`
						// }
						if IsPtrType(p2.Param.Typ) {
							initRootValue += `if ` + p2.InitName + ` == nil {
								` + p2.InitName + ` = &` + typePrint(ElemType(p2.Param.Typ)) + `{}
							}
							`
						}

						p2.InitString = `for key, values := range ` + queryParamName + `{
						if strings.HasPrefix(key, "` + paramName2 + `.") {
							` + initRootValue + `if ` + p2.InitName + ` == nil {
							 ` + p2.InitName + ` = url.Values{}
							}

						` + p2.InitName + `[strings.TrimPrefix(key, "` + paramName2 + `.") ] = values
						}
					}
					`
						serverParams = append(serverParams, p2)
						continue
					}
				}

				reservedStr, ok := mux.Reserved[typePrint(field.Typ)]
				if !ok {
					reservedStr, ok = mux.Reserved[strings.TrimPrefix(typePrint(field.Typ), "*")]
				}
				if ok {
					p2.ParamName = reservedStr

					var initRootValue string
					for idx, parent := range parents {
						if idx == 0 && mux.PreInitObject {
							continue
						}
						if IsPtrType(parent.Param.Typ) {
							if parentIndexs[idx] == 0 && fieldIdx == 0 {
								initRootValue += parent.InitName + " = &" + typePrint(ElemType(parent.Param.Typ)) + "{}\r\n"
							} else {
								initRootValue += `if ` + parent.InitName + ` == nil {
								` + parent.InitName + ` = &` + typePrint(ElemType(parent.Param.Typ)) + `{}
							}
							`
							}
						}
					}

					// if !mux.PreInitObject && IsPtrType(serverParam.Typ) && !c.IsParentInited() {
					// 	if fieldIdx == 0 {
					// 		initRootValue += name + " = &" + elmType + "{}\r\n"
					// 	} else {
					// 		initRootValue += "if " + name + " == nil {\r\n  " + name + " = &" + elmType + "{}\r\n}\r\n"
					// 	}
					// }

					p2.InitString = "\r\n" + strings.TrimSpace(initRootValue) + "\r\n" + p2.InitName + " = " + reservedStr

					p2.SkipDeclare = true
					p2.IsArray = false
					p2.ArgName = p2.Param.Name.Name

					serverParams = append(serverParams, p2)
					continue
				}

				if typePrint(field.Typ) == "map[string]string" {
					p2.SkipDeclare = true
					// p2.TypeStr = strings.TrimPrefix(typePrint(p2.Param.Typ), "*")
					p2.IsArray = false
					p2.ArgName = paramName2

					var queryParamName = mux.Reserved["url.Values"]
					if mux.FuncHeadStr != "" {
						queryParamName = "queryParams"
					}

					var initRootValue string
					for idx, parent := range parents {
						if idx == 0 && mux.PreInitObject {
							continue
						}
						if IsPtrType(parent.Param.Typ) {
							if parentIndexs[idx] == 0 && fieldIdx == 0 {
								initRootValue += parent.InitName + " = &" + typePrint(ElemType(parent.Param.Typ)) + "{}\r\n"
							} else {
								initRootValue += `if ` + parent.InitName + ` == nil {
								` + parent.InitName + ` = &` + typePrint(ElemType(parent.Param.Typ)) + `{}
							}
							`
							}
						}
					}
					// if IsPtrType(serverParam.Param.Typ) {
					// 	initRootValue += `if ` + serverParam.InitName + ` == nil {
					// 			` + serverParam.InitName + ` = &` + typePrint(ElemType(serverParam.Param.Typ)) + `{}
					// 		}
					// 		`
					// }
					if IsPtrType(p2.Param.Typ) {
						initRootValue += `if ` + p2.InitName + ` == nil {
								` + p2.InitName + ` = &` + typePrint(ElemType(p2.Param.Typ)) + `{}
							}
							`
					}

					p2.InitString = `for key, values := range ` + queryParamName + `{
						if strings.HasPrefix(key, "` + paramName2 + `.") {
							` + initRootValue + `if ` + p2.InitName + ` == nil {
							 ` + p2.InitName + ` = map[string]string{}
							}
						` + p2.Name.Name + `[strings.TrimPrefix(key, "` + paramName2 + `.") ] = values[len(values)-1]
						}
					}
					`
					serverParams = append(serverParams, p2)
					continue
				}

				var initRootValue string
				for idx, parent := range parents {
					if idx == 0 && mux.PreInitObject {
						continue
					}
					if IsPtrType(parent.Param.Typ) {
						if parentIndexs[idx] == 0 && fieldIdx == 0 {
							initRootValue += parent.InitName + " = &" + typePrint(ElemType(parent.Param.Typ)) + "{}\r\n"
						} else {
							initRootValue += `if ` + parent.InitName + ` == nil {
								` + parent.InitName + ` = &` + typePrint(ElemType(parent.Param.Typ)) + `{}
							}
							`
						}
					}
				}

				// if !mux.PreInitObject && IsPtrType(serverParam.Typ) && !c.IsParentInited() {
				// 	if fieldIdx == 0 {
				// 		initRootValue += serverParam.InitName + " = &" + elmType + "{}\r\n"
				// 	} else {
				// 		initRootValue += "if " + serverParam.InitName + " == nil {\r\n  " + serverParam.InitName + " = &" + typePrint(ElemType(serverParam.Param.Typ)) + "{}\r\n}\r\n"
				// 	}
				// }

				renderArgs["param"] = p2.Param
				renderArgs["skipDeclare"] = true
				if strings.TrimSpace(initRootValue) == "" {
					renderArgs["initRootValue"] = ""
				} else {
					renderArgs["initRootValue"] = "\r\n" + strings.TrimSpace(initRootValue)
				}
				renderArgs["type"] = strings.TrimPrefix(typePrint(p2.Param.Typ), "*")
				renderArgs["name"] = p2.Param.Name.Name
				renderArgs["rname"] = paramName2
				renderArgs["isArray"] = IsArrayType(p2.Param.Typ) || IsSliceType(p2.Param.Typ) || IsEllipsisType(serverParam.Typ)

				p2.SkipDeclare = true
				// p2.TypeStr = strings.TrimPrefix(typePrint(p2.Param.Typ), "*")
				p2.IsArray = true
				p2.ArgName = paramName2
				p2.InitString = strings.TrimSpace(mux.initString(c, method, p2.Param, funcs, renderArgs, optional))
				serverParams = append(serverParams, p2)
			}
		}

		serverParam.InitName = serverParam.Param.Name.Name
		serverParam.IsAnonymous = true

		addStructFields(paramNamePrefix, stType, serverParam, []int{0}, []ServerParam{serverParam})
		return serverParams
	}

	renderArgs["skipDeclare"] = false

	serverParam.SkipDeclare = false
	serverParam.IsArray = false
	serverParam.ArgName = paramName
	serverParam.InitString = strings.TrimSpace(mux.initString(c, method, param, funcs, renderArgs, optional))
	return []ServerParam{serverParam}
}

func (mux *DefaultStye) initString(c *context, method Method, param Param, funcs template.FuncMap, renderArgs map[string]interface{}, optional bool) string {
	typeStr := typePrint(param.Typ)
	if strings.HasPrefix(typeStr, "...") {
		typeStr = "[]" + strings.TrimPrefix(typeStr, "...")
	}
	elmType := strings.TrimPrefix(typeStr, "*")
	hasStar := typeStr != elmType

	var immediate bool
	if optional {
		_, immediate = mux.Types.Optional[elmType]
		//fmt.Println(elmType, mux.Types.Optional)
	} else {
		_, immediate = mux.Types.Required[elmType]
		//fmt.Println(elmType, mux.Types.Required)
	}

	var sb strings.Builder
	if immediate {
		if !hasStar {
			requiredTxt := template.Must(template.New("requiredTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- .initRootValue}}{{$xxx := .g.SetParentInited}}
		{{if .skipDeclare | not}}var {{end}}{{.name}} = {{readRequired .param .ctx .type .rname}}
		`))

			optionalTxt := template.Must(template.New("optionalTxt").Funcs(Funcs).Funcs(funcs).Parse(`
      {{- if .initRootValue}}
      
  		    {{- if .skipDeclare | not}}var {{.name}} .type {{end -}}
          
          {{- if .isArray}} 
          if ss := {{readOptional .param .ctx .type .rname}}; len(ss) != 0 {
      		  {{- .initRootValue}}
            {{.name}} = ss
          } else if ss := {{readOptional .param .ctx .type (concat .rname "[]") }}; len(ss) != 0 {
      		  {{- .initRootValue}}
            {{.name}} = ss
          }
          {{- else}}
          if s := {{readOptional .param .ctx .type .rname}}; s != "" {
      		  {{- .initRootValue}}
      		    {{- if isNull .param.Typ}}
      		      {{.name}}.Valid = true
      		      {{.name}}.String = s
      		    {{- else}}
                {{.name}} = s
              {{- end}}
          }
          {{- end}}
          
      {{- else}}
  		  {{if .skipDeclare | not}}var {{end}}{{.name}} = {{readOptional .param .ctx .type .rname}}
      {{- end}}
		`))

			if optional {
				renderText(optionalTxt, &sb, renderArgs)
			} else {
				renderText(requiredTxt, &sb, renderArgs)
			}
		} else {
			requiredTxt := template.Must(template.New("requiredTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- .initRootValue}}
		{{if .skipDeclare | not}}var {{end}}{{.name}} = {{readRequired .param .ctx .type .rname}}
		`))

			optionalTxt := template.Must(template.New("optionalTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- if .skipDeclare | not}}var {{.name}} *{{.type}}{{end -}}
    
    {{- if .isArray}} 
    if ss := {{readOptional .param .ctx .type .rname}}; len(ss) != 0 {
			{{- .initRootValue}}
			{{.name}} = &ss
		}
    {{- else}}
		if s := {{readOptional .param .ctx .type .rname}}; s != "" {
			{{- .initRootValue}}
			{{.name}} = &s
		}
    {{- end}}
		`))

			if optional {
				renderText(optionalTxt, &sb, renderArgs)
			} else {
				renderText(requiredTxt, &sb, renderArgs)
			}
		}
	} else {
		requiredTxt := template.Must(template.New("requiredTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- $s := readRequired .param .ctx .type .rname }}
		{{- if not .hasConvertError}}

			{{- .initRootValue}}
			{{- if .needTransform}}
			{{- if .skipDeclare | not}}var {{end}}{{.name}} = {{.type}}({{convert .param .ctx .type $s}})
			{{- else}}
			{{- if .skipDeclare | not}}var {{end}}{{.name}} = {{convert .param .ctx .type $s}}
			{{- end}}

		{{- else}}

			{{- if not .needTransform -}}
				{{- .initRootValue}}
				{{goify .name false}}, err := {{convert .param .ctx .type $s}}
				if  err != nil {
					{{badArgument .rname $s "err"}}
				}
				{{- $xxx := ( .g.SetErrorDefined ) -}}
			{{- else -}}

				{{- if .skipDeclare | not}}var {{.name}} {{.type}}{{end}}
				if {{goify .name false}}Value, err := {{convert .param .ctx .type $s}}; err != nil {
					{{badArgument .rname $s "err"}}
				} else {
					{{- .initRootValue}}
					{{.name}} = {{.type}}({{goify .name false}}Value)
				}
			{{- end -}}
			
		{{- end}}
		`))

		optionalTxt := template.Must(template.New("optionalTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- $s := readOptional .param .ctx .type .rname }}
		{{- $ss := readOptional .param .ctx .type (concat .rname "[]") }}
		{{- if .skipDeclare | not}}var {{.name}} {{.type}}
		{{end -}}
    
    {{- $tmp := "s"}}
    {{- $tmpIs := "s != \"\""}}
    {{- if .isArray}}
      {{- $tmp = "ss"}}
      {{- $tmpIs = "len(ss) != 0"}}
    {{- end -}}

		if {{$tmp}} := {{ $s }}; {{$tmpIs}} {
		{{- if not .hasConvertError}}
			{{- .initRootValue}}
			
			{{- $suffix := ""}}
			{{- if isNull .param.Typ}}
			  {{.name}}.Valid = true
				{{- if eq .type "sql.NullBool"}}	
					{{.name}}.Bool = {{convert .param .ctx "bool" $tmp}}
				{{- else if eq .type "sql.NullTime"}}	
					{{.name}}.Time = {{convert .param .ctx "time.Time" $tmp}}
				{{- else if eq .type "sql.NullInt64"}}	
					{{.name}}.Int64 = {{convert .param .ctx "int64" $tmp}}
				{{- $suffix = ".Int64"}}
				{{- else if eq .type "sql.NullUint64"}}
					{{.name}}.Uint64 = {{convert .param .ctx "uint64" $tmp}}
				{{- else if eq .type "sql.NullString"}}	
					{{.name}}.String = {{$tmp}}
				{{- end}}
			{{- else}}

				{{- if .needTransform}}
				{{.name}}{{$suffix}} = {{.type}}({{convert .param .ctx .type $tmp}})
				{{- else}}
				{{.name}}{{$suffix}} = {{convert .param .ctx .type $tmp}}
				{{- end}}

      {{- end}}
            
		{{- else}}
			{{goify .name false}}Value, err := {{convert .param .ctx .type $tmp}}
			if err != nil {
				{{badArgument .rname $tmp "err"}}
			}
			{{- .initRootValue}}
			
			{{- $suffix := ""}}
			{{- if isNull .param.Typ}}
				{{.name}}.Valid = true
				{{- if eq .type "sql.NullBool"}}	
				{{- $suffix = ".Bool"}}
				{{- else if eq .type "sql.NullTime"}}	
				{{- $suffix = ".Time"}}
				{{- else if eq .type "sql.NullInt64"}}	
				{{- $suffix = ".Int64"}}
				{{- else if eq .type "sql.NullUint64"}}	
				{{- $suffix = ".Uint64"}}
				{{- else if eq .type "sql.NullString"}}	
				{{- $suffix = ".String"}}
				{{- end}}
      {{- end}}
            
			{{- if .needTransform}}
			{{.name}}{{$suffix}} = {{.type}}({{goify .name false}}Value)
			{{- else}}
			{{.name}}{{$suffix}} = {{goify .name false}}Value
			{{- end}}
		{{- end}}

    {{- if .isArray}}
    } else if {{$tmp}} := {{ $ss }}; {{$tmpIs}} {
			{{goify .name false}}Value, err := {{convert .param .ctx .type $tmp}}
			if err != nil {
				{{badArgument .rname $tmp "err"}}
			}
			{{- .initRootValue}}
    	{{- if .needTransform}}
			{{.name}} = {{.type}}({{goify .name false}}Value)
			{{- else}}
			{{.name}} = {{goify .name false}}Value
			{{- end}}
    {{- end}}
		}
		`))

		requiredTxtWithStar := template.Must(template.New("requiredTxtWithStar").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- $s := readRequired .param .ctx .type .rname }}
		{{- if not .hasConvertError}}
			{{- if .needTransform}}
			{{- if .skipDeclare | not}}var {{end}}{{.name}} = {{.type}}({{convert .param .ctx .type $s}})
			{{- else}}
			{{- if .skipDeclare | not}}var {{end}}{{.name}} = {{convert .param .ctx .type $s}}
			{{- end}}
		{{- else}}
		{{- if .skipDeclare | not}}var {{.name}} *{{.type}}{{end}}
		if {{goify .name false}}Value, err := {{convert .param .ctx .type $s}}; err != nil {
			{{badArgument .rname $s "err"}}
		} else {
			{{- .initRootValue}}
			{{- if .needTransform}}
			{{.name}} = new({{.type}})
			*{{.name}} = {{.type}}({{goify .name false}}Value)
			{{- else}}
			{{.name}} = &{{goify .name false}}Value
			{{- end}}
		}
		{{- end}}
		`))

		optionalTxtWithStar := template.Must(template.New("optionalTxtWithStar").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- if .skipDeclare | not}}var {{.name}} *{{.type}}{{end}}
		if s := {{readOptional .param .ctx .type .rname}}; s != "" {
		{{- if not .hasConvertError}}
			{{- if .needTransform}}
			var {{goify .name false}}Value = {{.type}}({{convert .param .ctx .type "s"}})
			{{- else}}
			var {{goify .name false}}Value = {{convert .param .ctx .type "s"}}
			{{- end}}
			{{- .initRootValue}}
			{{.name}} = &{{goify .name false}}Value
		{{- else}}
			{{goify .name false}}Value, err := {{convert .param .ctx .type "s"}}
			if err != nil {
				{{badArgument .rname "s" "err"}}
			}
			{{- .initRootValue}}
			{{- if .needTransform}}
			{{.name}} = new({{.type}})
			*{{.name}} = {{.type}}({{goify .name false}}Value)
			{{- else}}
			{{.name}} = &{{goify .name false}}Value
			{{- end}}
		{{- end}}
		}
		`))

		convertArgs, ok := mux.Converts[elmType]
		if !ok {
			underlying := method.Ctx.GetType(elmType)
			if underlying != nil {
				if typePrint(underlying.Type) == "string" {
					convertArgs.NeedTransform = true
				} else {
					convertArgs, ok = mux.Converts[typePrint(underlying.Type)]
					if !ok {
						log.Fatalln(param.Method.Ctx.PostionFor(param.Method.Node.Pos()), ": 4argument '"+param.Name.Name+"' is unsupported type -", typeStr, fmt.Sprintf("%T", param.Typ), fmt.Sprintf("%T", underlying.Type))
					}

					convertArgs.NeedTransform = true
				}

			} else {

				selectorExpr, ok := param.Typ.(*ast.SelectorExpr)
				if !ok {
					log.Fatalln(param.Method.Ctx.PostionFor(param.Method.Node.Pos()), ": 3argument '"+param.Name.Name+"' is unsupported type -", typeStr, elmType, fmt.Sprintf("%T", param.Typ))
				}

				pkgName := typePrint(selectorExpr.X)
				isSysPkg := false
				for _, nm := range []string{
					"time",
					"net",
					"sql",
					"null",
				} {
					// fmt.Println(pkgName == nm, pkgName, nm)
					if pkgName == nm {
						isSysPkg = true
						break
					}
				}
				if !isSysPkg {
					err := errors.New(strconv.Itoa(int(method.Node.Pos())) + ": argument '" + param.Name.Name +
						"' of method '" + method.Clazz.Name.Name + ":" + method.Name.Name + "' is unsupported, '" +
						typePrint(param.Typ) + "' is in another package")
					log.Fatalln(err)
				}

				// if nm == "sql" || nm == "null" {
				// 	switch typePrint(selectorExpr.Sel) {
				// 	case "NullString":
				// 		convertArgs, ok = mux.Converts["string"]
				// 		if !ok {
				// 			log.Fatalln(param.Method.Ctx.PostionFor(param.Method.Node.Pos()), ": 5argument '"+param.Name.Name+"' is unsupported type -", typeStr, fmt.Sprintf("%T", param.Typ), fmt.Sprintf("%T", underlying.Type))
				// 		}
				// 		renderArgs["isNull"] = true
				// 		renderArgs["FieldName"] = "String"
				// 	case "NullInt64":
				// 		convertArgs, ok = mux.Converts["int64"]
				// 		if !ok {
				// 			log.Fatalln(param.Method.Ctx.PostionFor(param.Method.Node.Pos()), ": 5argument '"+param.Name.Name+"' is unsupported type -", typeStr, fmt.Sprintf("%T", param.Typ), fmt.Sprintf("%T", underlying.Type))
				// 		}
				// 		renderArgs["isNull"] = true
				// 		renderArgs["FieldName"] = "Int64"
				// 	case "NullBool":
				// 		convertArgs, ok = mux.Converts["bool"]
				// 		if !ok {
				// 			log.Fatalln(param.Method.Ctx.PostionFor(param.Method.Node.Pos()), ": 5argument '"+param.Name.Name+"' is unsupported type -", typeStr, fmt.Sprintf("%T", param.Typ), fmt.Sprintf("%T", underlying.Type))
				// 		}
				// 		renderArgs["isNull"] = true
				// 		renderArgs["FieldName"] = "Bool"
				// 	default:
				// 	}
				// }
			}
		}

		renderArgs["needTransform"] = convertArgs.NeedTransform
		renderArgs["hasConvertError"] = convertArgs.HasError

		if !hasStar {
			if optional {
				renderText(optionalTxt, &sb, renderArgs)
			} else {
				renderText(requiredTxt, &sb, renderArgs)
			}
		} else {
			if optional {
				renderText(optionalTxtWithStar, &sb, renderArgs)
			} else {
				renderText(requiredTxtWithStar, &sb, renderArgs)
			}
		}
	}

	return sb.String()
}

func (mux *DefaultStye) RouteFunc(method Method) string {
	ann := getAnnotation(method, false)
	name := strings.ToUpper(strings.TrimPrefix(ann.Name, "http."))
	if mux.MethodMapping != nil {
		methodName := mux.MethodMapping[name]
		if methodName != "" {
			name = methodName
		}
	}
	return name
}

func (mux *DefaultStye) BadArgumentFunc(method Method, err string, args ...string) string {
	return mux.ErrorFunc(method, true, "http.StatusBadRequest", err, args...)
}

func (mux *DefaultStye) ErrorFunc(method Method, hasRealErrorCode bool, errCode, err string, addArgs ...string) string {
	var sb strings.Builder
	renderText(mux.errTemplate, &sb, map[string]interface{}{
		"hasRealErrorCode": hasRealErrorCode,
		"errCode":          errCode,
		"err":              err,
		"addArgs":          addArgs,
	})
	return sb.String()
}

func (mux *DefaultStye) PlainTextFunc(method Method, args ...string) string {
	ann := getAnnotation(method, false)

	noreturn := ann.Attributes["noreturn"] == "true"

	okCode := "http.StatusOK"
	methodName := strings.ToUpper(strings.TrimPrefix(ann.Name, "http."))
	switch methodName {
	case "POST":
		okCode = "http.StatusCreated"
	case "PUT":
		okCode = "http.StatusAccepted"
	}

	params := map[string]interface{}{
		"method":     methodName,
		"statusCode": okCode,
		"data":       strings.Join(args, ","),
		"noreturn":   noreturn,
	}
	if statusCode := ann.Attributes["status_code"]; statusCode != "" {
		okCode = statusCode
		params["withCode"] = statusCode  /// 仅为 loong 使用
	}

	var sb strings.Builder
	renderText(mux.plainTextTemplate, &sb, params)
	return sb.String()
}

func (mux *DefaultStye) OkFunc(method Method, args ...string) string {
	ann := getAnnotation(method, false)

	noreturn := ann.Attributes["noreturn"] == "true"

	okCode := "http.StatusOK"
	methodName := strings.ToUpper(strings.TrimPrefix(ann.Name, "http."))
	switch methodName {
	case "POST":
		okCode = "http.StatusCreated"
	case "PUT":
		okCode = "http.StatusAccepted"
	}

	var sb strings.Builder
	renderText(mux.okTemplate, &sb, map[string]interface{}{
		"method":     methodName,
		"statusCode": okCode,
		"data":       strings.Join(args, ","),
		"noreturn":   noreturn,
	})
	return sb.String()
}

func NewEchoStye() *DefaultStye {
	mux := &DefaultStye{}
	mux.Init()
	return mux
}

func NewEchoStyeFromFile(filename string) (*DefaultStye, error) {
	return readStyleConfig(filename)
}

var _ MuxStye = &DefaultStye{}
