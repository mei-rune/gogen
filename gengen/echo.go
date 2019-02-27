package gengen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func NewEchoStye() *EchoStye {
	mux := &EchoStye{}
	mux.Init()
	return mux
}

var _ MuxStye = &EchoStye{}

type EchoStye struct {
	RoutePartyName    string
	PathParam         string
	QueryParam        string
	ReadBody          string
	BadArgumentFormat string

	ParseURL func(rawurl string) (string, []string, map[string]string)
}

func (mux *EchoStye) Init() {
	mux.RoutePartyName = "*echo.Group"
	mux.PathParam = "Param"
	mux.QueryParam = "QueryParam"
	mux.ReadBody = "Bind"
	mux.BadArgumentFormat = "errors.New(\"argument %%q is invalid - %%q\", %s, %s, %s)"
}

func (mux *EchoStye) CtxName() string {
	return `ctx`
}

func (mux *EchoStye) CtxType() string {
	return `ctx`
}

func (mux *EchoStye) IsSkipped(method Method) SkippedResult {
	anno := mux.GetAnnotation(method)
	res := SkippedResult{
		IsSkipped: anno == nil,
	}
	if res.IsSkipped {
		res.Message = "annotation is missing"
	}
	return res
}

func (mux *EchoStye) GetPath(method Method) string {
	anno := mux.GetAnnotation(method)
	if anno == nil {
		panic(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": Annotation of method '" + method.Itf.Name.Name + ":" + method.Name.Name + "' is missing"))
	}

	rawurl := anno.Attributes["path"]
	if rawurl == "" {
		panic(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": path(in annotation) of method '" + method.Itf.Name.Name + ":" + method.Name.Name + "' is missing"))
	}

	if mux.ParseURL == nil {
		mux.ParseURL = parseURL
	}
	pa, _, _ := mux.ParseURL(rawurl)
	return pa
}

func (mux *EchoStye) ReadParam(param Param, name string) string {
	typeStr := typePrint(param.Typ)

	anno := mux.GetAnnotation(*param.Method)
	if anno == nil {
		fmt.Println(param.Method.Annotations)
		panic(errors.New(strconv.Itoa(int(param.Method.Node.Pos())) + ": Annotation of method '" + param.Method.Itf.Name.Name + ":" + param.Method.Name.Name + "' is missing"))
	}

	_, pathNames, queryNames := parseURL(anno.Attributes["path"])

	var optional = true
	var readParam = mux.PathParam
	var paramName = param.Name.Name

	isPath := false
	for _, name := range pathNames {
		if name == param.Name.Name {
			isPath = true
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

	var sb strings.Builder
	sb.WriteString(" ")

	if anno.Attributes["data"] == param.Name.Name {
		sb.WriteString(typeStr)

		sb.WriteString("\r\n    if err := ")
		sb.WriteString(mux.CtxName())
		sb.WriteString(".")
		sb.WriteString(mux.ReadBody)
		sb.WriteString("(&")
		sb.WriteString(param.Name.Name)
		sb.WriteString("); err != nil {")
		sb.WriteString("\r\n          ")
		sb.WriteString(mux.BadArgumentFunc(*param.Method, fmt.Sprintf(mux.BadArgumentFormat, param.Name.Name, "s", "err")))
		sb.WriteString("\r\n        }")
		sb.WriteString("\r\n        ")
		return sb.String()
	}

	switch typeStr {
	case "string":
		sb.WriteString(" = ")
		sb.WriteString(mux.CtxName())
		sb.WriteString(".")
		sb.WriteString(readParam)
		sb.WriteString("(\"")
		sb.WriteString(paramName)
		sb.WriteString("\")")

	case "*string":
		if optional {
			sb.WriteString(typeStr)
			sb.WriteString("\r\n    if s := ")
			sb.WriteString(mux.CtxName())
			sb.WriteString(".")
			sb.WriteString(readParam)
			sb.WriteString("(\"")
			sb.WriteString(paramName)
			sb.WriteString("\"); s != \"\" {\r\n      ")
			sb.WriteString(name)
			sb.WriteString(" = &s\r\n}")
		} else {
			sb.WriteString(typeStr)
			sb.WriteString("\r\n        *")
			sb.WriteString(name)
			sb.WriteString(" = ")
			sb.WriteString(mux.CtxName())
			sb.WriteString(".")
			sb.WriteString(readParam)
			sb.WriteString("(\"")
			sb.WriteString(paramName)
			sb.WriteString("\")")
		}
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		sb.WriteString(typeStr)

		if optional {
			sb.WriteString("\r\n    if s := ")
			sb.WriteString(mux.CtxName())
			sb.WriteString(".")
			sb.WriteString(readParam)
			sb.WriteString("(\"")
			sb.WriteString(paramName)
			sb.WriteString("\"); s != \"\" {\r\n      ")
		} else {
			sb.WriteString("\r\n    ")
		}

		conv := "strconv.ParseInt"
		if strings.HasPrefix(typeStr, "u") {
			conv = "strconv.ParseUint"
		}

		sb.WriteString("if v64, err := ")
		sb.WriteString(conv)
		sb.WriteString("(s, 10, 64); err != nil {")
		sb.WriteString("\r\n          ")
		sb.WriteString(mux.BadArgumentFunc(*param.Method, fmt.Sprintf(mux.BadArgumentFormat, param.Name.Name, "s", "err")))
		sb.WriteString("\r\n        }")
		sb.WriteString("\r\n        ")
		sb.WriteString(name)
		sb.WriteString(" = ")
		if strings.HasSuffix(typeStr, "64") {
			sb.WriteString("v64")
		} else {
			sb.WriteString(typeStr)
			sb.WriteString("(v64)")
		}

		if optional {
			sb.WriteString("\r\n    }")
		}

	case "*int", "*int8", "*int16", "*int32", "*int64", "*uint", "*uint8", "*uint16", "*uint32", "*uint64":
		sb.WriteString(typeStr)

		if optional {
			sb.WriteString("\r\n    if s := ")
			sb.WriteString(mux.CtxName())
			sb.WriteString(".")
			sb.WriteString(readParam)
			sb.WriteString("(\"")
			sb.WriteString(paramName)
			sb.WriteString("\"); s != \"\" {\r\n      ")
		} else {
			sb.WriteString("\r\n    ")
		}

		conv := "strconv.ParseInt"
		if strings.HasPrefix(typeStr, "u") {
			conv = "strconv.ParseUint"
		}

		sb.WriteString("if v64, err := ")
		sb.WriteString(conv)
		sb.WriteString("(")

		if optional {
			sb.WriteString("s")
		} else {
			sb.WriteString(mux.CtxName())
			sb.WriteString(".")
			sb.WriteString(readParam)
			sb.WriteString("(\"")
			sb.WriteString(param.Name.Name)
			sb.WriteString("\")")
		}
		sb.WriteString(", 10, 64); err != nil {")
		sb.WriteString("\r\n          ")
		sb.WriteString(mux.BadArgumentFunc(*param.Method, fmt.Sprintf(mux.BadArgumentFormat, param.Name.Name, "s", "err")))
		sb.WriteString("\r\n        }")
		sb.WriteString("\r\n        ")
		sb.WriteString(name)
		sb.WriteString(" = ")
		if strings.HasSuffix(typeStr, "64") {
			sb.WriteString("&v64")
		} else {
			sb.WriteString("new(")
			sb.WriteString(typeStr)
			sb.WriteString(")\r\n        *")
			sb.WriteString(name)
			sb.WriteString(" = ")
			sb.WriteString(typeStr)
			sb.WriteString("(v64)")
		}

		if optional {
			sb.WriteString("\r\n    }")
		}
	}

	// ann := mux.GetAnnotation(method)
	// if ann == nil {
	//  panic(errors.New(strconv.FormatInt(method.Node.Pos(), 10) + ": Annotation of method '" + method.Itf.Name + ":" + method.Name + "' is missing"))
	// }
	// return strings.TrimPrefix(ann.Name, "http.")
	return sb.String()
}

func (mux *EchoStye) FuncSignature() string {
	return `func(` + mux.CtxName() + ` echo.Context) error `
}

func (mux *EchoStye) RouteFunc(method Method) string {
	ann := mux.GetAnnotation(method)
	if ann == nil {
		panic(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": Annotation of method '" + method.Itf.Name.Name + ":" + method.Name.Name + "' is missing"))
	}
	return strings.TrimPrefix(ann.Name, "http.")
}

func (mux *EchoStye) GetAnnotation(method Method) *Annotation {
	var annotation *Annotation
	for idx := range method.Annotations {
		if !strings.HasPrefix(method.Annotations[idx].Name, "http.") {
			continue
		}

		if annotation != nil {
			panic(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": Annotation of method '" + method.Itf.Name.Name + ":" + method.Name.Name + "' is duplicated"))
		}
		annotation = &method.Annotations[idx]
	}
	return annotation
}

func (mux *EchoStye) BadArgumentFunc(method Method, args ...string) string {
	return mux.ErrorFunc(method, args...)
}

func (mux *EchoStye) ErrorFunc(method Method, args ...string) string {
	return "" + mux.CtxName() + ".Error(" + strings.Join(args, ",") + ")\r\n    return nil"
}

func (mux *EchoStye) OkFunc(method Method, args ...string) string {
	return "return " + mux.CtxName() + ".JSON(" + strings.Join(args, ",") + ")"
}
