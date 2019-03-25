package gengen

import (
	"errors"
	"fmt"
	"go/ast"
	"log"
	"strconv"
	"strings"
	"text/template"
)

type ReadArgs struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

type ConvertArgs struct {
	Format        string `json:"format"`
	NeedTransform bool   `json:"needTransform"`
	HasError      bool   `json:"hasError"`
}

type DefaultStye struct {
	classes           []Class
	FuncSignatureStr  string            `json:"func_signature"`
	CtxNameStr        string            `json:"ctx_name"`
	CtxTypeStr        string            `json:"ctx_type"`
	RoutePartyName    string            `json:"route_party_name"`
	PathParam         string            `json:"path_param_format"`
	QueryParam        string            `json:"query_param_format"`
	ReadFormat        string            `json:"read_format"`
	ReadBodyFormat    string            `json:"read_body_format"`
	BadArgumentFormat string            `json:"bad_argument_format"`
	OkFuncFormat      string            `json:"ok_func_format"`
	ErrorFuncFormat   string            `json:"err_func_format"`
	Reserved          map[string]string `json:"reserved"`
	MethodMapping     map[string]string `json:"method_mapping"`
	Types             struct {
		Required map[string]ReadArgs `json:"required"`
		Optional map[string]ReadArgs `json:"optional"`
	} `json:"types"`
	Converts map[string]ConvertArgs                                    `json:"converts"`
	UrlStyle string                                                    `json:"url_style"`
	ParseURL func(rawurl string) (string, []string, map[string]string) `json:"-"`

	bodyReader   string
	readTemplate *template.Template
	bindTemplate *template.Template
	errTemplate  *template.Template
	okTemplate   *template.Template
}

func (mux *DefaultStye) Init() {
	mux.CtxNameStr = "ctx"
	mux.CtxTypeStr = "echo.Context"
	mux.FuncSignatureStr = "func(" + mux.CtxNameStr + " " + mux.CtxTypeStr + ") error "
	mux.RoutePartyName = "*echo.Group"
	mux.PathParam = "Param"
	mux.QueryParam = "QueryParam"
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

	mux.ReadFormat = `{{.ctx}}.{{.readMethodName}}("{{.name}}")`
	mux.OkFuncFormat = "return ctx.JSON({{.statusCode}}, {{.data}})"
	mux.ErrorFuncFormat = "ctx.Error({{.err}})\r\n     return nil"
}

func (mux *DefaultStye) reinit(values map[string]interface{}) {
	if mux.ParseURL == nil {
		var replace ReplaceFunc
		switch mux.UrlStyle {
		case "colon", "":
			replace = colonReplace
		case "brace":
			replace = braceReplace
		default:
			log.Fatalln(errors.New("url_style '" + mux.UrlStyle + "' is invalid"))
		}

		mux.ParseURL = func(rawurl string) (string, []string, map[string]string) {
			segements, names, query := parseURL(rawurl)
			return JoinPathSegments(segements, replace), names, query
		}
	}

	if mux.Types.Required == nil {
		mux.Types.Required = map[string]ReadArgs{}
	}

	if mux.Types.Optional == nil {
		mux.Types.Optional = map[string]ReadArgs{}
	}

	if _, ok := mux.Types.Optional["string"]; !ok {
		mux.Types.Optional["string"] = ReadArgs{
			Name: mux.QueryParam,
		}
	}
	if _, ok := mux.Types.Required["string"]; !ok {
		mux.Types.Required["string"] = ReadArgs{
			Name: mux.PathParam,
		}
	}

	if mux.Converts == nil {
		mux.Converts = map[string]ConvertArgs{}
	}
	for _, t := range []string{"int", "int8", "int16", "int32", "int64"} {
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
		funcName := stringWith(values, "features.boolConvert", "toBool({{.name}})")
		mux.Converts["bool"] = ConvertArgs{Format: funcName, HasError: false}
	}
	if _, ok := mux.Converts["time.Time"]; !ok {
		funcName := stringWith(values, "features.datetimeConvert", "toDatetime({{.name}})")
		mux.Converts["time.Time"] = ConvertArgs{Format: funcName, HasError: true}
	}

	mux.bodyReader = mux.Reserved["*http.Request"] + ".Body"
	mux.readTemplate = template.Must(template.New("readTemplate").Parse(mux.ReadFormat))
	mux.bindTemplate = template.Must(template.New("bindTemplate").Parse(mux.ReadBodyFormat))
	mux.errTemplate = template.Must(template.New("errTemplate").Parse(mux.ErrorFuncFormat))
	mux.okTemplate = template.Must(template.New("okTemplate").Parse(mux.OkFuncFormat))
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
	var sb strings.Builder
	readMethodName := mux.PathParam
	if args, ok := mux.Types.Required[typeName]; ok {
		readMethodName = args.Name
		if len(args.Args) > 0 {
			paramName = paramName + "," + strings.Join(args.Args, ",")
		}
	}
	renderText(mux.readTemplate, &sb, map[string]interface{}{"ctx": ctxName, "name": paramName, "readMethodName": readMethodName})
	return sb.String()
}

func (mux *DefaultStye) ReadOptional(param Param, typeName, ctxName, paramName string) string {
	var sb strings.Builder
	readMethodName := mux.QueryParam
	if args, ok := mux.Types.Optional[typeName]; ok {
		readMethodName = args.Name
		if len(args.Args) > 0 {
			paramName = paramName + "," + strings.Join(args.Args, ",")
		}
	}
	renderText(mux.readTemplate, &sb, map[string]interface{}{"ctx": ctxName, "name": paramName, "readMethodName": readMethodName})
	return sb.String()
}

func (mux *DefaultStye) TypeConvert(param Param, typeName, ctxName, paramName string) string {
	var sb strings.Builder

	format, ok := mux.Converts[typeName]
	if !ok {
		log.Fatalln(fmt.Errorf("%d: unsupport type - %s", param.Method.Node.Pos(), typeName))
	}

	tpl := template.Must(template.New("convertTemplate").Parse(format.Format))
	renderText(tpl, &sb, map[string]interface{}{
		"ctx":  ctxName,
		"name": paramName,
	})
	return sb.String()
}

func (mux *DefaultStye) ReadBody(param Param, ctxName, paramName string) string {
	var sb strings.Builder
	renderText(mux.bindTemplate, &sb, map[string]interface{}{"ctx": ctxName, "name": paramName})
	return sb.String()
}

func (mux *DefaultStye) GetPath(method Method) string {
	anno := getAnnotation(method, false)

	rawurl := anno.Attributes["path"]
	if rawurl == "" {
		log.Fatalln(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": path(in annotation) of method '" + method.Clazz.Name.Name + ":" + method.Name.Name + "' is missing"))
	}
	pa, _, _ := mux.ParseURL(rawurl)
	return pa
}

type ServerParam struct {
	Param

	IsSkipUse  bool
	InBody     bool
	ParamName  string
	InitString string
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
			var sb strings.Builder
			renderText(mux.bindTemplate, &sb, map[string]interface{}{"ctx": ctxName, "name": paramName})
			return sb.String()
		},
	}

	bindTxt := template.Must(template.New("bindTxt").Funcs(Funcs).Funcs(funcs).Parse(`
			var bindArgs struct {
				{{- range $param := .params}}
				{{goify $param.Param.Name.Name true}} {{typePrint $param.Param.Typ}} ` + "`json:\"{{ $param.Param.Name.Name}},omitempty\"`" + `
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

func (mux *DefaultStye) ToParamList(method Method) []ServerParam {
	var results []ServerParam

	ann := getAnnotation(method, false)
	methodStr := strings.ToUpper(strings.TrimPrefix(ann.Name, "http."))
	isEdit := methodStr == "PUT" || methodStr == "POST"

	for idx := range method.Params.List {
		results = append(results, mux.ToParam(method, method.Params.List[idx], isEdit)...)
	}

	return results
}

func (mux *DefaultStye) ToParam(method Method, param Param, isEdit bool) []ServerParam {
	typeStr := typePrint(param.Typ)
	elmType := strings.TrimPrefix(typeStr, "*")
	hasStar := typeStr != elmType

	funcs := template.FuncMap{
		"badArgument": func(paramName, valueName, errName string) string {
			return mux.BadArgumentFunc(method, fmt.Sprintf(mux.BadArgumentFormat, paramName, valueName, errName))
		},
		"readBody": func(param Param, ctxName string, paramName string) string {
			return mux.ReadBody(param, ctxName, paramName)
		},
		"readOptional": func(param Param, ctxName, elmType, paramName string) string {
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
	}

	var readParam = mux.PathParam
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

		bindTxt := template.Must(template.New("bindTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		var {{.name}} {{.type}}
		if err := {{readBody .param .ctx .name}}; err != nil {
			{{badArgument .name "\"<no value>\"" "err"}}
		}
		`))

		renderArgs := map[string]interface{}{
			"ctx":       mux.CtxName(),
			"type":      elmType,
			"name":      name,
			"rname":     paramName,
			"param":     param,
			"readParam": mux.ReadBody,
		}

		var sb strings.Builder
		renderText(bindTxt, &sb, renderArgs)
		serverParam.InitString = strings.TrimSpace(sb.String())
		return []ServerParam{serverParam}
	} else if s, ok := mux.Reserved[typeStr]; ok {
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
	}
	if !isPath {
		optional = true
		readParam = mux.QueryParam
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

	// 	type ServerParam struct {
	// 	Param

	// 	IsSkipUse  bool
	// 	InBody     bool
	// 	ParamName  string
	// 	InitString string
	// }

	if ok, startType, endType := IsRange(mux.classes, param.Typ); ok {
		if isPath {
			err := errors.New(strconv.Itoa(int(method.Node.Pos())) + ": argument '" + param.Name.Name + "' of method '" + method.Clazz.Name.Name + ":" + method.Name.Name + "' is invalid")
			log.Fatalln(err)
			panic(err)
		}

		p1 := serverParam
		p1.InitString = "var " + name + " " + typeStr

		p2 := serverParam
		p2.IsSkipUse = true
		p2.Name = &ast.Ident{}
		*p2.Name = *param.Name
		p2.Param.Name.Name = name + ".Start"
		p2.Param.Typ = startType
		paramName2 := paramName + ".start"

		var initRootValue string
		if hasStar {
			initRootValue = "\r\n  " + name + " = &" + elmType + "{}"
		}
		p2.InitString = strings.TrimSpace(mux.initString(method, p2.Param, funcs, true, optional,
			typePrint(startType), strings.TrimPrefix(typePrint(startType), "*"), name+".Start", paramName2, readParam, initRootValue))

		p3 := serverParam
		p3.IsSkipUse = true
		p3.Name = &ast.Ident{}
		*p3.Name = *param.Name
		p3.Param.Name.Name = name + ".End"
		p3.Param.Typ = endType
		paramName3 := paramName + ".end"

		if hasStar {
			initRootValue = "\r\nif " + name + " == nil {\r\n  " + name + " = &" + elmType + "{}\r\n}"
		}
		p3.InitString = strings.TrimSpace(mux.initString(method, p3.Param, funcs, true, optional,
			typePrint(endType), strings.TrimPrefix(typePrint(endType), "*"), name+".End", paramName3, readParam, initRootValue))

		return []ServerParam{p1, p2, p3}
	}

	serverParam.InitString = strings.TrimSpace(mux.initString(method, param, funcs, false, optional, typeStr, elmType, name, paramName, readParam, ""))
	return []ServerParam{serverParam}
}

func (mux *DefaultStye) initString(method Method, param Param, funcs template.FuncMap, skipDeclare, optional bool, typeStr, elmType, name, paramName, readParam, initRootValue string) string {
	var hasStar = typeStr != elmType

	var immediate bool
	if optional {
		_, immediate = mux.Types.Optional[elmType]
	} else {
		_, immediate = mux.Types.Required[elmType]
	}

	var sb strings.Builder
	if immediate {
		if !hasStar {
			requiredTxt := template.Must(template.New("requiredTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- .initRootValue}}
		{{- if .skipDeclare | not}}var {{end}}{{.name}} = {{readRequired .param .ctx .type .rname}}
		`))

			optionalTxt := template.Must(template.New("optionalTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- .initRootValue}}
		{{- if .skipDeclare | not}}var {{end}}{{.name}} = {{readOptional .param .ctx .type .rname}}
		`))

			renderArgs := map[string]interface{}{
				"skipDeclare":   skipDeclare,
				"ctx":           mux.CtxName(),
				"type":          elmType,
				"name":          name,
				"param":         param,
				"rname":         paramName,
				"readParam":     readParam,
				"initRootValue": initRootValue,
			}

			if optional {
				renderText(optionalTxt, &sb, renderArgs)
			} else {
				renderText(requiredTxt, &sb, renderArgs)
			}
		} else {
			requiredTxt := template.Must(template.New("requiredTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- .initRootValue}}
		{{- if .skipDeclare | not}}var {{end}}{{.name}} = {{readRequired .param .ctx .type .rname}}
		`))

			optionalTxt := template.Must(template.New("optionalTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- if .skipDeclare | not}}var {{.name}} *{{.type}}{{end}}
		if s := {{readOptional .param .ctx .type .rname}}; s != "" {
			{{- .initRootValue}}
			{{.name}} = &s
		}
		`))

			renderArgs := map[string]interface{}{
				"skipDeclare":   skipDeclare,
				"ctx":           mux.CtxName(),
				"type":          elmType,
				"name":          name,
				"rname":         paramName,
				"readParam":     readParam,
				"param":         param,
				"initRootValue": initRootValue,
			}

			if optional {
				renderText(optionalTxt, &sb, renderArgs)
			} else {
				renderText(requiredTxt, &sb, renderArgs)
			}
		}
	} else {
		requiredTxt := template.Must(template.New("requiredTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- $s := concat .ctx "." .readParam "(\"" .rname "\")"}}
		{{- if not .hasConvertError}}
			{{- .initRootValue}}
			{{- if .needTransform}}
			{{- if .skipDeclare | not}}var {{end}}{{.name}} = {{.type}}({{convert .param .ctx .type $s}})
			{{- else}}
			{{- if .skipDeclare | not}}var {{end}}{{.name}} = {{convert .param .ctx .type $s}}
			{{- end}}
		{{- else}}
		{{- if .skipDeclare | not}}var {{.name}} {{.type}}{{end}}
		if {{goify .name false}}Value, err := {{convert .param .ctx .type $s}}; err != nil {
			s := {{$s}}
			{{badArgument .rname "s" "err"}}
		} else {
			{{- .initRootValue}}
			{{- if .needTransform}}
			{{.name}} = {{.type}}({{goify .name false}}Value)
			{{- else}}
			{{.name}} = {{goify .name false}}Value
			{{- end}}
		}
		{{- end}}
		`))

		optionalTxt := template.Must(template.New("optionalTxt").Funcs(Funcs).Funcs(funcs).Parse(`
		{{- if .skipDeclare | not}}var {{.name}} {{.type}}{{end}}
		if s := {{.ctx}}.{{.readParam}}("{{.rname}}"); s != "" {
		{{- if not .hasConvertError}}
			{{- .initRootValue}}
			{{- if .needTransform}}
			{{.name}} = {{.type}}({{convert .param .ctx .type "s"}})
			{{- else}}
			{{.name}} = {{convert .param .ctx .type "s"}}
			{{- end}}
		{{- else}}
			{{goify .name false}}Value, err := {{convert .param .ctx .type "s"}}
			if err != nil {
				{{badArgument .rname "s" "err"}}
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
		{{- $s := concat .ctx "." .readParam "(\"" .rname "\")"}}
		{{- if not .hasConvertError}}
			{{- if .needTransform}}
			{{- if .skipDeclare | not}}var {{end}}{{.name}} = {{.type}}({{convert .param .ctx .type $s}})
			{{- else}}
			{{- if .skipDeclare | not}}var {{end}}{{.name}} = {{convert .param .ctx .type $s}}
			{{- end}}
		{{- else}}
		{{- if .skipDeclare | not}}var {{.name}} *{{.type}}{{end}}
		if {{goify .name false}}Value, err := {{convert .param .ctx .type $s}}; err != nil {
			s := {{$s}}
			{{badArgument .rname "s" "err"}}
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
		if s := {{.ctx}}.{{.readParam}}("{{.rname}}"); s != "" {
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
			log.Fatalln(param.Method.Ctx.PostionFor(param.Method.Node.Pos()), ": argument '"+param.Name.Name+"' is unsupported type -", typeStr)
		}

		renderArgs := map[string]interface{}{
			"skipDeclare": skipDeclare,
			"ctx":         mux.CtxName(),
			"type":        elmType,
			"name":        name,
			"rname":       paramName,
			"param":       param,
			//"conv":          conv,
			"readParam":       readParam,
			"needTransform":   convertArgs.NeedTransform,
			"hasConvertError": convertArgs.HasError,
			"initRootValue":   initRootValue,
		}

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

func (mux *DefaultStye) OkFunc(method Method, args ...string) string {
	ann := getAnnotation(method, false)

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
