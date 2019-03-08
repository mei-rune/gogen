package gengen

import (
	"errors"
	"fmt"
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
	mux.BadArgumentFormat = "fmt.Errorf(\"argument %%q is invalid - %%q\", %s, %s, %s)"
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
		mux.ParseURL = parseURL
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
	anno := mux.GetAnnotation(method, true)
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
	anno := mux.GetAnnotation(method, false)

	rawurl := anno.Attributes["path"]
	if rawurl == "" {
		log.Fatalln(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": path(in annotation) of method '" + method.Itf.Name.Name + ":" + method.Name.Name + "' is missing"))
	}
	pa, _, _ := mux.ParseURL(rawurl)
	return pa
}

func (mux *DefaultStye) UseParam(param Param) string {
	name := param.Name.Name
	if name == "result" {
		name = "result_"
	}

	typeStr := typePrint(param.Typ)
	anno := mux.GetAnnotation(*param.Method, false)
	if anno.Attributes["data"] == param.Name.Name {
		if typeStr == "io.Reader" {
			return mux.bodyReader
		}

		if strings.HasPrefix(typeStr, "*") {
			return "&" + name
		}
		return name
	}

	if s, ok := mux.Reserved[typeStr]; ok {
		return s
	}

	if strings.HasPrefix(typeStr, "*") {

		_, pathNames, _ := mux.ParseURL(anno.Attributes["path"])
		for _, name := range pathNames {
			if name == param.Name.Name {

				if typeStr == "*string" {
					return "&" + name
				}

				if mux.Converts != nil {
					convertArgs, ok := mux.Converts[strings.TrimPrefix(typeStr, "*")]
					if ok {
						if !convertArgs.HasError {
							return "&" + name
						}
					}
				}

			}
		}
	}

	return name
}

func (mux *DefaultStye) InitParam(param Param) string {
	typeStr := typePrint(param.Typ)
	elmType := strings.TrimPrefix(typeStr, "*")
	hasStar := typeStr != elmType

	anno := mux.GetAnnotation(*param.Method, false)
	inBody := anno.Attributes["data"] == param.Name.Name
	if inBody {
		if typeStr == "io.Reader" {
			return ""
		}
	} else if _, ok := mux.Reserved[typeStr]; ok {
		return ""
	}

	name := param.Name.Name
	if name == "result" {
		name = "result_"
	}

	_, pathNames, queryNames := mux.ParseURL(anno.Attributes["path"])

	var optional = true
	var readParam = mux.PathParam
	var paramName = param.Name.Name

	isPath := false
	for _, name := range pathNames {
		if name == param.Name.Name {
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
		}
	}

	funcs := template.FuncMap{
		"badArgument": func(paramName, valueName, errName string) string {
			return mux.BadArgumentFunc(*param.Method, fmt.Sprintf(mux.BadArgumentFormat, paramName, valueName, errName))
		},
		"readBody": func(ctxName, paramName string) string {
			return mux.ReadBody(param, ctxName, paramName)
		},
		"readOptional": func(ctxName, paramName string) string {
			return mux.ReadOptional(param, elmType, ctxName, paramName)
		},
		"readRequired": func(ctxName, paramName string) string {
			return mux.ReadRequired(param, elmType, ctxName, paramName)
		},
		"convert": func(ctxName, paramName string) string {
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

	var sb strings.Builder
	if inBody {
		bindTxt := template.Must(template.New("bindTxt").Funcs(funcs).Parse(`
		var {{.name}} {{.type}}
		if err := {{readBody .ctx .name}}; err != nil {
			{{badArgument .name "\"<no value>\"" "err"}}
		}
		`))

		renderArgs := map[string]interface{}{
			"ctx":       mux.CtxName(),
			"type":      elmType,
			"name":      name,
			"rname":     paramName,
			"readParam": mux.ReadBody,
		}

		renderText(bindTxt, &sb, renderArgs)
		return strings.TrimSpace(sb.String())
	}

	var immediate bool
	if optional {
		_, immediate = mux.Types.Optional[elmType]
	} else {
		_, immediate = mux.Types.Required[elmType]
	}

	if immediate {
		if !strings.HasPrefix(typeStr, "*") {
			requiredTxt := template.Must(template.New("requiredTxt").Funcs(funcs).Parse(`
		var {{.name}} = {{readRequired .ctx .rname}}
		`))

			optionalTxt := template.Must(template.New("optionalTxt").Funcs(funcs).Parse(`
		var {{.name}} = {{readOptional .ctx .rname}}
		`))

			renderArgs := map[string]interface{}{
				"ctx":       mux.CtxName(),
				"type":      elmType,
				"name":      name,
				"rname":     paramName,
				"readParam": readParam,
			}

			if optional {
				renderText(optionalTxt, &sb, renderArgs)
			} else {
				renderText(requiredTxt, &sb, renderArgs)
			}
		} else {
			requiredTxt := template.Must(template.New("requiredTxt").Funcs(funcs).Parse(`
		var {{.name}} = {{readRequired .ctx .rname}}
		`))

			optionalTxt := template.Must(template.New("optionalTxt").Funcs(funcs).Parse(`
		var {{.name}} *{{.type}}
		if s := {{readOptional .ctx .rname}}; s != "" {
			{{.name}} = &s
		}
		`))

			renderArgs := map[string]interface{}{
				"ctx":       mux.CtxName(),
				"type":      elmType,
				"name":      name,
				"rname":     paramName,
				"readParam": readParam,
			}

			if optional {
				renderText(optionalTxt, &sb, renderArgs)
			} else {
				renderText(requiredTxt, &sb, renderArgs)
			}

		}
	} else {

		requiredTxt := template.Must(template.New("requiredTxt").Funcs(funcs).Parse(`
		{{- $s := concat .ctx "." .readParam "(\"" .rname "\")"}}
		{{- if not .hasConvertError}}
			{{- if .needTransform}}
			var {{.name}} = {{.type}}({{convert .ctx $s}})
			{{- else}}
			var {{.name}} = {{convert .ctx $s}}
			{{- end}}
		{{- else}}
		var {{.name}} {{.type}}
		if {{.name}}Value, err := {{convert .ctx $s}}; err != nil {
			s := {{$s}}
			{{badArgument .rname "s" "err"}}
		} else {
			{{- if .needTransform}}
			{{.name}} = {{.type}}({{.name}}Value)
			{{- else}}
			{{.name}} = {{.name}}Value
			{{- end}}
		}
		{{- end}}
		`))

		optionalTxt := template.Must(template.New("optionalTxt").Funcs(funcs).Parse(`
		var {{.name}} {{.type}}
		if s := {{.ctx}}.{{.readParam}}("{{.rname}}"); s != "" {
		{{- if not .hasConvertError}}
			{{- if .needTransform}}
			{{.name}} = {{.type}}({{convert .ctx "s"}})
			{{- else}}
			{{.name}} = {{convert .ctx "s"}}
			{{- end}}
		{{- else}}
			{{.name}}Value, err := {{convert .ctx "s"}}
			if err != nil {
				{{badArgument .rname "s" "err"}}
			}
			{{- if .needTransform}}
			{{.name}} = {{.type}}({{.name}}Value)
			{{- else}}
			{{.name}} = {{.name}}Value
			{{- end}}
		{{- end}}
		}
		`))

		requiredTxtWithStar := template.Must(template.New("requiredTxtWithStar").Funcs(funcs).Parse(`
		{{- $s := concat .ctx "." .readParam "(\"" .rname "\")"}}
		{{- if not .hasConvertError}}
			{{- if .needTransform}}
			var {{.name}} = {{.type}}({{convert .ctx $s}})
			{{- else}}
			var {{.name}} = {{convert .ctx $s}}
			{{- end}}
		{{- else}}
		var {{.name}} *{{.type}}
		if {{.name}}Value, err := {{convert .ctx $s}}; err != nil {
			s := {{$s}}
			{{badArgument .rname "s" "err"}}
		} else {
			{{.name}} = new({{.type}})
			{{- if .needTransform}}
			*{{.name}} = {{.type}}({{.name}}Value)
			{{- else}}
			*{{.name}} = {{.name}}Value
			{{- end}}
		}
		{{- end}}
		`))

		optionalTxtWithStar := template.Must(template.New("optionalTxtWithStar").Funcs(funcs).Parse(`
		var {{.name}} *{{.type}}
		if s := {{.ctx}}.{{.readParam}}("{{.rname}}"); s != "" {
		{{- if not .hasConvertError}}
			{{- if .needTransform}}
			var {{.name}}Value = {{.type}}({{convert .ctx "s"}})
			{{- else}}
			var {{.name}}Value = {{convert .ctx "s"}}
			{{- end}}
			{{.name}} = &{{.name}}Value
		{{- else}}
			{{.name}}Value, err := {{convert .ctx "s"}}
			if err != nil {
				{{badArgument .rname "s" "err"}}
			}
			{{.name}} = new({{.type}})
			{{- if .needTransform}}
			*{{.name}} = {{.type}}({{.name}}Value)
			{{- else}}
			*{{.name}} = {{.name}}Value
			{{- end}}
		{{- end}}
		}
		`))

		convertArgs, ok := mux.Converts[elmType]
		if !ok {
			log.Fatalln(param.Method.Node.Pos(), ": argument '"+param.Name.Name+"' is unsupported type -", typeStr)
		}

		renderArgs := map[string]interface{}{
			"ctx":   mux.CtxName(),
			"type":  elmType,
			"name":  name,
			"rname": paramName,
			//"conv":          conv,
			"readParam":       readParam,
			"needTransform":   convertArgs.NeedTransform,
			"hasConvertError": convertArgs.HasError,
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

	return strings.TrimSpace(sb.String())
}

func (mux *DefaultStye) RouteFunc(method Method) string {
	ann := mux.GetAnnotation(method, false)
	name := strings.ToUpper(strings.TrimPrefix(ann.Name, "http."))
	if mux.MethodMapping != nil {
		methodName := mux.MethodMapping[name]
		if methodName != "" {
			name = methodName
		}
	}
	return name
}

func (mux *DefaultStye) GetAnnotation(method Method, nilIfNotExists bool) *Annotation {
	var annotation *Annotation
	for idx := range method.Annotations {
		if !strings.HasPrefix(method.Annotations[idx].Name, "http.") {
			continue
		}

		if annotation != nil {
			log.Fatalln(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": Annotation of method '" + method.Itf.Name.Name + ":" + method.Name.Name + "' is duplicated"))
		}
		annotation = &method.Annotations[idx]
	}
	if nilIfNotExists {
		return annotation
	}
	if annotation == nil {
		log.Fatalln(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": Annotation of method '" + method.Itf.Name.Name + ":" + method.Name.Name + "' is missing"))
	}
	return annotation
}

func (mux *DefaultStye) BadArgumentFunc(method Method, err string, args ...string) string {
	return mux.ErrorFunc(method, err, args...)
}

func (mux *DefaultStye) ErrorFunc(method Method, err string, addArgs ...string) string {
	var sb strings.Builder
	renderText(mux.errTemplate, &sb, map[string]interface{}{
		"err":     err,
		"addArgs": addArgs,
	})
	return sb.String()
}

func (mux *DefaultStye) OkFunc(method Method, args ...string) string {
	var sb strings.Builder
	renderText(mux.okTemplate, &sb, map[string]interface{}{
		"statusCode": mux.okCode(method),
		"data":       strings.Join(args, ","),
	})
	return sb.String()
}

func (mux *DefaultStye) okCode(method Method) string {
	ann := mux.GetAnnotation(method, false)
	switch strings.ToUpper(strings.TrimPrefix(ann.Name, "http.")) {
	case "POST":
		return "http.StatusCreated"
	case "PUT":
		return "http.StatusAccepted"
	}
	return "http.StatusOK"
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
