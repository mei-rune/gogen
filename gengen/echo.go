package gengen

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

func NewEchoStye() *DefaultStye {
	mux := &DefaultStye{}
	mux.Init()
	return mux
}

func NewEchoStyeFromFile(filename string) (*DefaultStye, error) {
	return readStyleConfig(filename)
}

var _ MuxStye = &DefaultStye{}

type DefaultStye struct {
	FuncSignatureStr  string            `json:"func_signature"`
	CtxNameStr        string            `json:"ctx_name"`
	CtxTypeStr        string            `json:"ctx_type"`
	RoutePartyName    string            `json:"route_party_name"`
	PathParam         string            `json:"path_param_format"`
	QueryParam        string            `json:"query_param_format"`
	ReadBody          string            `json:"read_body_format"`
	BadArgumentFormat string            `json:"bad_argument_format"`
	Reserved          map[string]string `json:"bad_argument_format"`
	bodyReader        string
	ParseURL          func(rawurl string) (string, []string, map[string]string) `json:"-"`
}

func (mux *DefaultStye) Init() {
	mux.CtxNameStr = "ctx"
	mux.CtxTypeStr = "echo.Context"
	mux.FuncSignatureStr = "func(" + mux.CtxNameStr + " " + mux.CtxTypeStr + ") error "
	mux.RoutePartyName = "*echo.Group"
	mux.PathParam = "Param"
	mux.QueryParam = "QueryParam"
	mux.ReadBody = "Bind"
	mux.BadArgumentFormat = "fmt.Errorf(\"argument %%q is invalid - %%q\", %s, %s, %s)"
	mux.Reserved = map[string]string{
		"*http.Request":       mux.CtxNameStr + ".Request()",
		"http.ResponseWriter": mux.CtxNameStr + ".Response().Writer",
		"context.Context":     mux.CtxNameStr + ".Request().Context()",
		"echo.Context":        mux.CtxNameStr,
		// "io.Reader":           mux.CtxNameStr + ".Request().Body",
	}
	mux.bodyReader = mux.CtxNameStr + ".Request().Body"
	//mux.Reserved["io.Reader"] = mux.bodyReader

	if mux.ParseURL == nil {
		mux.ParseURL = parseURL
	}
}

func (mux *DefaultStye) reinit(values map[string]interface{}) {
	mux.bodyReader = mux.CtxNameStr + ".Request().Body"

	if mux.ParseURL == nil {
		mux.ParseURL = parseURL
	}
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

func (mux *DefaultStye) FuncSignature() string {
	return mux.FuncSignatureStr
}

func (mux *DefaultStye) CtxName() string {
	return mux.CtxNameStr
}

func (mux *DefaultStye) CtxType() string {
	return mux.CtxTypeStr
}

func (mux *DefaultStye) IsReserved(param Param) bool {
	typeStr := typePrint(param.Typ)
	_, ok := mux.Reserved[typeStr]
	return ok
}

func (mux *DefaultStye) ToReserved(param Param) string {
	typeStr := typePrint(param.Typ)
	s := mux.Reserved[typeStr]
	return s
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

	if typeStr == "*string" {
		_, pathNames, _ := mux.ParseURL(anno.Attributes["path"])
		for _, name := range pathNames {
			if name == param.Name.Name {
				return "&" + name
			}
		}
	}

	return name
}

func (mux *DefaultStye) InitParam(param Param) string {
	typeStr := typePrint(param.Typ)

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
	}

	var sb strings.Builder
	if inBody {

		bindTxt := template.Must(template.New("bindTxt").Funcs(funcs).Parse(`
		var {{.name}} {{.type}}
		if err := {{.ctx}}.{{.readParam}}(&{{.name}}); err != nil {
			{{badArgument .name "\"<no value>\"" "err"}}
		}
		`))

		renderArgs := map[string]interface{}{
			"ctx":       mux.CtxName(),
			"type":      strings.TrimPrefix(typeStr, "*"),
			"name":      name,
			"rname":     paramName,
			"readParam": mux.ReadBody,
		}

		renderText(bindTxt, &sb, renderArgs)
		return strings.TrimSpace(sb.String())
	}

	switch typeStr {
	case "string":
		requiredTxt := template.Must(template.New("requiredTxt").Funcs(funcs).Parse(`
		var {{.name}} = {{.ctx}}.{{.readParam}}("{{.rname}}")
		`))

		renderArgs := map[string]interface{}{
			"ctx":       mux.CtxName(),
			"type":      strings.TrimPrefix(typeStr, "*"),
			"name":      name,
			"rname":     paramName,
			"readParam": readParam,
		}

		renderText(requiredTxt, &sb, renderArgs)
	case "*string":
		requiredTxt := template.Must(template.New("requiredTxt").Funcs(funcs).Parse(`
		var {{.name}} = {{.ctx}}.{{.readParam}}("{{.rname}}")
		`))

		optionalTxt := template.Must(template.New("optionalTxt").Funcs(funcs).Parse(`
		var {{.name}} *{{.type}}
		if s := {{.ctx}}.{{.readParam}}("{{.rname}}"); s != "" {
			{{.name}} = &s
		}
		`))

		renderArgs := map[string]interface{}{
			"ctx":       mux.CtxName(),
			"type":      strings.TrimPrefix(typeStr, "*"),
			"name":      name,
			"rname":     paramName,
			"readParam": readParam,
		}

		if optional {
			renderText(optionalTxt, &sb, renderArgs)
		} else {
			renderText(requiredTxt, &sb, renderArgs)
		}
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		conv := "strconv.ParseInt"
		if strings.HasPrefix(typeStr, "u") {
			conv = "strconv.ParseUint"
		}

		requiredTxt := template.Must(template.New("requiredTxt").Funcs(funcs).Parse(`
		var {{.name}} {{.type}}			
		if v64, err := {{.conv}}({{.ctx}}.{{.readParam}}("{{.rname}}"), 10, 64); err != nil {
			s := {{.ctx}}.{{.readParam}}("{{.rname}}")
			{{badArgument .rname "s" "err"}}
		} else {
			{{- if .needTransform}}
			{{.name}} = {{.type}}(v64)
			{{- else}}
			{{.name}} = v64
			{{- end}}
		}
		`))

		optionalTxt := template.Must(template.New("optionalTxt").Funcs(funcs).Parse(`
		var {{.name}} {{.type}}
		if s := {{.ctx}}.{{.readParam}}("{{.rname}}"); s != "" {
			v64, err := {{.conv}}(s, 10, 64)
			if err != nil {
				{{badArgument .rname "s" "err"}}
			}
			{{- if .needTransform}}
			{{.name}} = {{.type}}(v64)
			{{- else}}
			{{.name}} = v64
			{{- end}}
		}
		`))

		renderArgs := map[string]interface{}{
			"ctx":           mux.CtxName(),
			"type":          strings.TrimPrefix(typeStr, "*"),
			"name":          name,
			"rname":         paramName,
			"conv":          conv,
			"readParam":     readParam,
			"needTransform": !strings.HasSuffix(typeStr, "64"),
		}

		if optional {
			renderText(optionalTxt, &sb, renderArgs)
		} else {
			renderText(requiredTxt, &sb, renderArgs)
		}
	case "*int", "*int8", "*int16", "*int32", "*int64", "*uint", "*uint8", "*uint16", "*uint32", "*uint64":
		conv := "strconv.ParseInt"
		if strings.HasPrefix(typeStr, "u") {
			conv = "strconv.ParseUint"
		}

		requiredTxt := template.Must(template.New("requiredTxt").Funcs(funcs).Parse(`
		var {{.name}} *{{.type}}			
		if v64, err := {{.conv}}({{.ctx}}.{{.readParam}}("{{.rname}}"), 10, 64); err != nil {
			s := {{.ctx}}.{{.readParam}}("{{.rname}}")
			{{badArgument .rname "s" "err"}}
		} else {
			{{.name}} = new({{.type}})
			{{- if .needTransform}}
			*{{.name}} = {{.type}}(v64)
			{{- else}}
			*{{.name}} = v64
			{{- end}}
		}
		`))

		optionalTxt := template.Must(template.New("optionalTxt").Funcs(funcs).Parse(`
		var {{.name}} *{{.type}}
		if s := {{.ctx}}.{{.readParam}}("{{.rname}}"); s != "" {
			v64, err := {{.conv}}(s, 10, 64)
			if err != nil {
				{{badArgument .rname "s" "err"}}
			}
			{{.name}} = new({{.type}})
			{{- if .needTransform}}
			*{{.name}} = {{.type}}(v64)
			{{- else}}
			*{{.name}} = v64
			{{- end}}
		}
		`))

		renderArgs := map[string]interface{}{
			"ctx":           mux.CtxName(),
			"type":          strings.TrimPrefix(typeStr, "*"),
			"name":          name,
			"rname":         paramName,
			"conv":          conv,
			"readParam":     readParam,
			"needTransform": !strings.HasSuffix(typeStr, "64"),
		}

		if optional {
			renderText(optionalTxt, &sb, renderArgs)
		} else {
			renderText(requiredTxt, &sb, renderArgs)
		}
	default:
		log.Fatalln(param.Method.Node.Pos(), ": argument '"+param.Name.Name+"' is unsupported type -", typeStr)
	}

	// ann := mux.GetAnnotation(method)
	// if ann == nil {
	//  log.Fatalln(errors.New(strconv.FormatInt(method.Node.Pos(), 10) + ": Annotation of method '" + method.Itf.Name + ":" + method.Name + "' is missing"))
	// }
	// return strings.TrimPrefix(ann.Name, "http.")
	return strings.TrimSpace(sb.String())
}

func (mux *DefaultStye) RouteFunc(method Method) string {
	ann := mux.GetAnnotation(method, false)
	return strings.TrimPrefix(ann.Name, "http.")
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

func (mux *DefaultStye) BadArgumentFunc(method Method, args ...string) string {
	return mux.ErrorFunc(method, args...)
}

func (mux *DefaultStye) ErrorFunc(method Method, args ...string) string {
	return "" + mux.CtxName() + ".Error(" + strings.Join(args, ",") + ")\r\n    return nil"
}

func (mux *DefaultStye) OkFunc(method Method, args ...string) string {
	return "return " + mux.CtxName() + ".JSON(" + mux.okCode(method) + ", " + strings.Join(args, ",") + ")"
}

func (mux *DefaultStye) okCode(method Method) string {
	ann := mux.GetAnnotation(method, false)
	switch ann.Name {
	case "http.POST":
		return "http.StatusCreated"
	case "http.PUT":
		return "http.StatusAccepted"
	}
	return "http.StatusOK"
}

func renderText(txt *template.Template, out io.Writer, renderArgs interface{}) {
	err := txt.Execute(out, renderArgs)
	if err != nil {
		log.Fatalln(err)
	}
}
