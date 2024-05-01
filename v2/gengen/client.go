package gengen

import (
	"errors"
	"flag"
	"go/ast"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
	"github.com/swaggo/swag"
)

type ClientGenerator struct {
	ext      string
	buildTag string

	config ClientConfig
}

func (cmd *ClientGenerator) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.ext, "ext", ".client-gen.go", "文件后缀名")
	fs.StringVar(&cmd.buildTag, "build_tag", "", "生成 go build tag")

	fs.StringVar(&cmd.config.TagName, "tag", "json", "")
	fs.StringVar(&cmd.config.RestyField, "field", "Proxy", "")
	fs.StringVar(&cmd.config.RestyName, "resty", "*resty.Proxy", "")
	fs.StringVar(&cmd.config.ContextClassName, "context", "context.Context", "")
	fs.StringVar(&cmd.config.newRequest, "new-request", "resty.NewRequest({{.proxy}},{{.url}})", "")
	fs.StringVar(&cmd.config.releaseRequest, "free-request", "resty.ReleaseRequest({{.proxy}},{{.request}})", "")

	fs.StringVar(&cmd.config.ConvertNS, "convert_ns", "", "")
	fs.StringVar(&cmd.config.TimeFormat, "timeFormat", "client.Proxy.TimeFormat", "")

	fs.BoolVar(&cmd.config.HasWrapper, "has-wrapper", false, "")
	fs.StringVar(&cmd.config.WrapperType, "wrapper-type", "loong.Result", "")
	fs.StringVar(&cmd.config.WrapperData, "wrapper-data", "Data", "")
	fs.StringVar(&cmd.config.WrapperError, "wrapper-error", "Error", "")
	return fs
}

func (cmd *ClientGenerator) Run(args []string) error {
	swaggerParser := swag.New()
	swaggerParser.GoGenEnabled = true
	swaggerParser.ParseVendor = true
	swaggerParser.ParseDependency = true
	swaggerParser.ParseInternal = true

	var files []*astutil.File
	for _, filename := range args {
		file, err := ParseFile(nil, filename)
		if err != nil {
			return err
		}
		err = swaggerParser.Packages().CollectAstFile(file.Package.ImportPath, file.Filename, file.AstFile)
		if err != nil {
			return errors.New("collect astFile: " + err.Error())
		}
		files = append(files, file)
	}
	_, err := swaggerParser.Packages().ParseTypes()
	if err != nil {
		return errors.New("parse types: " + err.Error())
	}

	for idx, file := range files {
		filename := args[idx]

		targetFile := strings.TrimSuffix(filename, ".go") + cmd.ext
		out, err := os.Create(targetFile)
		if err != nil {
			return err
		}

		err = cmd.genHeader(out, swaggerParser, file)
		if err != nil {
			return err
		}

		for _, ts := range file.TypeList {
			if ts.Interface == nil && ts.Struct == nil {
				continue
			}

			err = cmd.genInterfaceImpl(out, swaggerParser, ts)
			if err != nil {
				return err
			}
		}

		err = out.Close()
		if err != nil {
			return err
		}

		exec.Command("goimports", "-w", targetFile).Run()
		exec.Command("goimports", "-w", targetFile).Run()
	}
	return nil
}

func (cmd *ClientGenerator) genHeader(out io.Writer, swaggerParser *swag.Parser, file *astutil.File) error {
	if cmd.buildTag != "" {
		io.WriteString(out, "//go:build ")
		io.WriteString(out, cmd.buildTag)
		io.WriteString(out, "\r\n")
		io.WriteString(out, "// +build ")
		io.WriteString(out, cmd.buildTag)
		io.WriteString(out, "\r\n")
		io.WriteString(out, "\r\n")
	}
	io.WriteString(out, "// Please don't edit this file!\r\npackage ")
	io.WriteString(out, file.Pkg.Name)
	io.WriteString(out, "\r\n\r\nimport (")

	found := false
	// for pa, alias := range cfg.Imports() {
	// 	if alias == "errors" || strings.HasSuffix(pa, "/errors") {
	// 		found = true
	// 		break
	// 	}
	// }
	for _, pa := range file.Imports {
		if (pa.Name != nil && astutil.ToString(pa.Name) == "errors") ||
			strings.HasSuffix(strings.Trim(astutil.ToString(pa.Path), "\""), "/errors") {
			found = true
			break
		}
	}
	if !found {
		if s := os.Getenv("GOGEN_ERRORS"); s != "" {
			io.WriteString(out, "\r\n\t\""+s+"\"")
		} else {
			io.WriteString(out, "\r\n\t\"errors\"")
		}
	}
	for _, pa := range file.Imports {
		io.WriteString(out, "\r\n\t")

		if pa.Name != nil && pa.Name.Name != "_" {
			io.WriteString(out, astutil.ToString(pa.Name))
			io.WriteString(out, " ")
		}
		io.WriteString(out, pa.Path.Value)
	}

	io.WriteString(out, "\r\n\t")
	io.WriteString(out, `"github.com/runner-mei/loong"`)
	io.WriteString(out, "\r\n\t")
	io.WriteString(out, `"github.com/runner-mei/resty"`)

	if s := os.Getenv("GOGEN_IMPORTS"); s != "" {
		for _, pa := range strings.Split(s, ",") {
			pa = strings.TrimSpace(pa)
			io.WriteString(out, "\r\n\t")
			if strings.HasSuffix(pa, "\"") {
				io.WriteString(out, pa)
			} else {
				io.WriteString(out, "\""+pa+"\"")
			}
		}
	}

	// for pa, alias := range cfg.Imports() {
	// 	io.WriteString(out, "\r\n\t")
	// 	if alias != "" {
	// 		io.WriteString(out, alias)
	// 		io.WriteString(out, " ")
	// 	}
	// 	io.WriteString(out, "\""+pa+"\"")
	// }
	io.WriteString(out, "\r\n)\r\n")
	return nil
}

func getClassName(doc *ast.CommentGroup) (name string, reference bool, ok bool) {
	if doc == nil {
		return "", false, false
	}

	for _, comment := range doc.List {
		line := strings.TrimLeft(comment.Text, "/")
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "@http.Client":
			ok = true

			for _, key := range fields[1:] {
				ss := strings.SplitN(key, "=", 2)
				switch ss[0] {
				case "name":
					name = strings.Trim(ss[1], "\"")
				case "reference":
					reference = strings.ToLower(strings.Trim(ss[1], "\"")) == "true"
				}
			}
			return
		}
	}

	return "", false, false
}

func (cmd *ClientGenerator) genInterfaceImpl(out io.Writer, swaggerParser *swag.Parser, ts *astutil.TypeSpec) error {
	var optionalRoutePrefix string
	var ignore bool

	if doc := ts.Doc(); doc != nil {
		for _, comment := range doc.List {
			line := strings.TrimSpace(strings.TrimLeft(comment.Text, "/"))

			if strings.HasPrefix(line, "@gogen.optional_route_prefix") {
				optionalRoutePrefix = strings.TrimSpace(strings.TrimPrefix(line, "@gogen.optional_route_prefix"))
			} else if strings.HasPrefix(line, "@gogen.ignore") {
				ignore = true
			}
		}
	}
	if ignore {
		io.WriteString(out, "\r\n// "+ts.Name+" is skipped")
		return nil
	}

	className := ts.Name + "Client"
	recvClassName := className
	// @http.Client name="TestClient" reference="true"
	clientName, reference, ok := getClassName(ts.Node.Doc)
	if ok {
		if clientName != "" {
			className = clientName
			recvClassName = clientName
		}
		if reference {
			recvClassName = "*" + recvClassName
		}
	}

	methods, err := resolveMethods(swaggerParser, ts)
	if err != nil {
		return err
	}
	count := 0
	for idx := range methods {
		if len(methods[idx].Operation.RouterProperties) == 0 {
			continue
		}
		count++
	}

	if count == 0 {
		io.WriteString(out, "\r\n// ")
		io.WriteString(out, ts.Name)
		io.WriteString(out, "is skipped")
		return nil
	}

	io.WriteString(out, "\r\n\r\ntype ")
	io.WriteString(out, className+" struct {")
	io.WriteString(out, "\r\n\t"+cmd.config.RestyField+" "+cmd.config.RestyName)

	if optionalRoutePrefix != "" {
		io.WriteString(out, "\r\n  NoRoutePrefix bool")

		// close struct
		io.WriteString(out, "\r\n}\r\n")

		io.WriteString(out, "func (client *"+className+") SetRoutePrefix(enable bool) {\r\n")
		io.WriteString(out, "	client.NoRoutePrefix = !enable\r\n")
		io.WriteString(out, "}\r\n\r\n")

		io.WriteString(out, "func (client "+className+") routePrefix() string {\r\n")
		io.WriteString(out, "	 if client.NoRoutePrefix {\r\n")
		io.WriteString(out, "	   return \"\"\r\n")
		io.WriteString(out, "	 }\r\n")
		io.WriteString(out, "  return \""+optionalRoutePrefix+"\"\r\n")
		io.WriteString(out, "}\r\n")
	} else {
		// close struct
		io.WriteString(out, "\r\n}\r\n")
	}

	for idx := range methods {
		if len(methods[idx].Operation.RouterProperties) == 0 {
			io.WriteString(out, "\r\n// "+methods[idx].Method.Name+": annotation is missing")
			continue
		}
		err := cmd.genInterfaceMethod(out, recvClassName, methods[idx], optionalRoutePrefix)
		if err != nil {
			return err
		}
	}
	return nil
}

func getResultCount(method *Method) int {
	resultCount := 0
	for _, result := range method.Method.Results.List {
		if result.Type().IsErrorType() {
			continue
		}
		resultCount++
	}
	return resultCount
}

func getResultName(method *Method) string {
	resultName := "result"
	isNameExist := func(name string) bool {
		for idx := range method.Method.Params.List {
			if method.Method.Params.List[idx].Name == name {
				return true
			}
		}

		for idx := range method.Method.Results.List {
			if method.Method.Results.List[idx].Name == name {
				return true
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

func formatParamName(name string) string {
	if name == "request" {
		return "_request"
	}
	return name
}

func (cmd *ClientGenerator) genInterfaceMethodSignature(out io.Writer, recvClassName string, method *Method) error {
	io.WriteString(out, "\r\n\r\nfunc (client "+recvClassName+") "+method.Method.Name+
		"(ctx "+cmd.config.ContextClassName)

	for _, param := range method.Method.Params.List {
		if param.Type().IsContextType() {
			continue
		}

		io.WriteString(out, ", "+formatParamName(param.Name)+" ")
		if param.IsVariadic {
			io.WriteString(out, "...")
		}
		io.WriteString(out, param.Type().ToLiteral())
	}
	io.WriteString(out, ") (")

	for _, result := range method.Method.Results.List {
		if result.Type().IsErrorType() {
			continue
		}

		io.WriteString(out, result.Type().ToLiteral())
		io.WriteString(out, ",")
	}
	io.WriteString(out, " error) {")

	return nil
}

func (cmd *ClientGenerator) genInterfaceMethodReturnVars(out io.Writer, recvClassName string, method *Method) error {
	resultName := getResultName(method)
	resultCount := getResultCount(method)

	switch resultCount {
	case 0:
		break
	case 1:

		io.WriteString(out, "\r\n\tvar ")
		io.WriteString(out, resultName)
		io.WriteString(out, " ")
		io.WriteString(out, strings.TrimPrefix(method.Method.Results.List[0].Type().ToLiteral(), "*"))

	default:

		io.WriteString(out, "\r\n\tvar ")
		io.WriteString(out, getResultName(method))
		io.WriteString(out, " struct {")

		for _, result := range method.Method.Results.List {
			if result.Type().IsErrorType() {
				continue
			}

			io.WriteString(out, "\r\n\tE"+result.Name+" ")
			// io.WriteString(out, "\r\n\t"+toUpperFirst(result.Name)+" ")
			io.WriteString(out, result.Type().ToLiteral()+" `json:\""+Underscore(result.Name)+"\"`")
		}
		io.WriteString(out, "\r\n}")
	}

	if cmd.config.HasWrapper {
		if resultCount > 0 {
			io.WriteString(out, "\r\n\tvar ")
			io.WriteString(out, getResultName(method))
			io.WriteString(out, "Wrap ")
			io.WriteString(out, cmd.config.WrapperType)

			io.WriteString(out, "\r\n"+getResultName(method))
			io.WriteString(out, "Wrap.Data = &")
			io.WriteString(out, getResultName(method))
		}
	}

	if resultCount > 0 {
		io.WriteString(out, "\r\n")
	}

	return nil
}

func (cmd *ClientGenerator) genInterfaceMethod(out io.Writer, recvClassName string, method *Method, optionalRoutePrefix string) error {
	if err := cmd.genInterfaceMethodSignature(out, recvClassName, method); err != nil {
		return err
	}

	if err := cmd.genInterfaceMethodReturnVars(out, recvClassName, method); err != nil {
		return err
	}

	io.WriteString(out, "\r\n\trequest := ")
	io.WriteString(out, cmd.config.NewRequest("client."+cmd.config.RestyField, cmd.config.GetPath(optionalRoutePrefix, method)))

	needAssignment := false
	var inBody []astutil.Param
	var inParameters []spec.Parameter
	for _, param := range method.Method.Params.List {
		if param.Type().IsContextType() {
			continue
		}

		if typeStr := param.Type().ToLiteral(); typeStr == "*http.Request" ||
			typeStr == "http.ResponseWriter" {
			continue
		}

		switch param.Type().ToLiteral() {
		case "map[string]string":
			webPrefix := toSnakeCase(param.Name)

			st := searchStructParam(method.Operation, param.Name)
			if st == nil {
				foundIndex := searchParam(method.Operation, param.Name)
				if foundIndex >= 0 {
					st = &method.Operation.Parameters[foundIndex]
				}
			}
			if st != nil {
				if st.In != "query" {
					inBody = append(inBody, param)
					inParameters = append(inParameters, *st)
					continue
				}

				if isExtendInline(st) {
					webPrefix = ""
				} else {
					webPrefix = st.Name
				}
			}

			if needAssignment {
				io.WriteString(out, "\r\nrequest = request.")
			} else {
				io.WriteString(out, ".\r\n")
			}
			needAssignment = false

			if webPrefix == "" {
				io.WriteString(out, "SetParamValues("+formatParamName(param.Name)+")")
			} else {
				io.WriteString(out, "SetParamValuesWithPrefix(\""+webPrefix+".\", "+formatParamName(param.Name)+")")
			}
			continue
		case "url.Values":
			webPrefix := toSnakeCase(param.Name)

			st := searchStructParam(method.Operation, param.Name)
			if st == nil {
				foundIndex := searchParam(method.Operation, param.Name)
				if foundIndex >= 0 {
					st = &method.Operation.Parameters[foundIndex]
				}
			}
			if st != nil {
				if st.In != "query" {
					inBody = append(inBody, param)
					inParameters = append(inParameters, *st)
					continue
				}

				if isExtendInline(st) {
					webPrefix = ""
				} else {
					webPrefix = st.Name
				}
			}

			if needAssignment {
				io.WriteString(out, "\r\nrequest = request.")
			} else {
				io.WriteString(out, ".\r\n")
			}
			needAssignment = false

			if webPrefix == "" {
				io.WriteString(out, "SetParams("+formatParamName(param.Name)+")")
			} else {
				io.WriteString(out, "SetParamsWithPrefix(\""+webPrefix+".\", "+formatParamName(param.Name)+")")
			}
			continue
		}

		foundIndex := searchParam(method.Operation, param.Name)
		if foundIndex >= 0 {
			option := method.Operation.Parameters[foundIndex]

			if option.In == "path" {
				continue
			}

			if option.In != "query" {
				inBody = append(inBody, param)
				inParameters = append(inParameters, option)
				continue
			}

			param.Name = formatParamName(param.Name)
			err := cmd.genInterfaceMethodParam(out, method, &param, &option, &needAssignment)
			if err != nil {
				return err
			}
			continue
		}

		parent := searchStructParam(method.Operation, param.Name)
		if parent != nil {
			if parent.In != "query" {
				inBody = append(inBody, param)
				inParameters = append(inParameters, *parent)
				continue
			}

			webPrefix := parent.Name
			if isExtendInline(parent) {
				webPrefix = ""
			}
			param.Name = formatParamName(param.Name)

			isPtrType := param.Type().IsPtrType()
			if isPtrType {
				needAssignment = true
				io.WriteString(out, "\r\nif "+param.Name+" != nil {")
			}

			err := cmd.genInterfaceMethodStructParam(out, method, &param, webPrefix, &needAssignment)
			if err != nil {
				return err
			}

			if isPtrType {
				needAssignment = true
				io.WriteString(out, "\r\n}")
			}
			continue
		}

		return errors.New("'" + param.Name + "' is unsupported type - '" + param.Type().ToLiteral() + "'")
	}

	if len(inBody) > 0 {
		if needAssignment {
			io.WriteString(out, "\r\nrequest = request.")
		} else {
			io.WriteString(out, ".\r\n")
		}

		typeStr := inBody[0].Type().ToLiteral()

		if len(inBody) == 1 && (isExtendEntire(&inParameters[0]) || typeStr == "io.Reader") {
			io.WriteString(out, "SetBody("+formatParamName(inBody[0].Name)+")")
		} else {
			io.WriteString(out, "SetBody(map[string]interface{}{")
			for idx, param := range inBody {
				io.WriteString(out, "\r\n\t\""+inParameters[idx].Name+"\": "+formatParamName(param.Name)+",")
			}
			io.WriteString(out, "\r\n})")
		}
	}

	resultCount := getResultCount(method)
	if resultCount > 0 {
		if needAssignment {
			io.WriteString(out, "\r\nrequest = request.")
		} else {
			io.WriteString(out, ".\r\n")
		}

		resultName := getResultName(method)
		if cmd.config.HasWrapper {
			io.WriteString(out, "Result(&"+resultName+"Wrap)")
		} else {
			io.WriteString(out, "Result(&"+resultName+")")
		}
	}
	return cmd.genInterfaceMethodInvokeAndReturn(out, recvClassName, method)
}

func (cmd *ClientGenerator) genInterfaceMethodInvokeAndReturn(out io.Writer, recvClassName string, method *Method) error {
	resultCount := getResultCount(method)

	if resultCount == 0 /* && !cmd.config.HasWrapper */ {
		io.WriteString(out, "\r\n")
		io.WriteString(out, "\r\ndefer "+cmd.config.ReleaseRequest("client."+cmd.config.RestyField, "request"))
		io.WriteString(out, "\r\nreturn ")
	} else {
		io.WriteString(out, "\r\n\r\nerr := ")
	}
	io.WriteString(out, "request."+cmd.config.RouteFunc(method)+"(ctx)")

	resultName := getResultName(method)
	if resultCount == 0 {
		// if cmd.config.HasWrapper {
		// 	io.WriteString(out, "\r\n\tif err != nil {")
		// 	io.WriteString(out, "\r\n\t\treturn err")
		// 	io.WriteString(out, "\r\n\t}")
		// 	io.WriteString(out, "\r\n\tif !"+resultName+"Wrap.Success {")
		// 	io.WriteString(out, "\r\n\t\tif "+resultName+"Wrap.Error == nil {")
		// 	io.WriteString(out, "\r\n\t\t\treturn errors.New(\"error is nil\")")
		// 	io.WriteString(out, "\r\n\t\t}")
		// 	io.WriteString(out, "\r\n\t\treturn "+resultName+"Wrap.Error")
		// 	io.WriteString(out, "\r\n\t}")
		// 	io.WriteString(out, "\r\n\treturn nil")
		// }
	} else if resultCount == 1 {
		io.WriteString(out, "\r\n\t"+cmd.config.ReleaseRequest("client."+cmd.config.RestyField, "request"))

		zeroValueStr := zeroValueLiteral(method.Method.Results.List[0].Type())
		if cmd.config.HasWrapper {
			io.WriteString(out, "\r\n\tif err != nil {")
			io.WriteString(out, "\r\n\t\treturn "+zeroValueStr+", err")
			io.WriteString(out, "\r\n\t}")
			io.WriteString(out, "\r\n\tif !"+resultName+"Wrap.Success {")
			io.WriteString(out, "\r\n\t\tif "+resultName+"Wrap.Error == nil {")
			io.WriteString(out, "\r\n\t\t\treturn  "+zeroValueStr+", errors.New(\"error is nil\")")
			io.WriteString(out, "\r\n\t\t}")
			io.WriteString(out, "\r\n\t\treturn "+zeroValueStr+", "+resultName+"Wrap.Error")
			io.WriteString(out, "\r\n\t}")
			io.WriteString(out, "\r\n\treturn "+resultName+", nil")
		} else {
			if method.Method.Results.List[0].Type().IsPtrType() {
				io.WriteString(out, "\r\n\treturn &"+resultName+", err")
			} else {
				io.WriteString(out, "\r\n\treturn "+resultName+", err")
			}
		}
	} else {
		io.WriteString(out, "\r\n\t"+cmd.config.ReleaseRequest("client."+cmd.config.RestyField, "request"))

		io.WriteString(out, "\r\n")

		var sb strings.Builder
		for _, result := range method.Method.Results.List {
			if result.Type().IsErrorType() {
				continue
			}

			io.WriteString(&sb, zeroValueLiteral(result.Type()))
			io.WriteString(&sb, ", ")
		}

		io.WriteString(out, "\r\n\tif err != nil {")
		io.WriteString(out, "\r\n\t\treturn "+sb.String()+" err")
		io.WriteString(out, "\r\n\t}")

		if cmd.config.HasWrapper {
			io.WriteString(out, "\r\n\tif !"+resultName+"Wrap.Success {")
			io.WriteString(out, "\r\n\t\tif "+resultName+"Wrap.Error == nil {")
			io.WriteString(out, "\r\n\t\t\treturn  "+sb.String()+"errors.New(\"error is nil\")")
			io.WriteString(out, "\r\n\t\t}")
			io.WriteString(out, "\r\n\t\treturn "+sb.String()+resultName+"Wrap.Error")
			io.WriteString(out, "\r\n\t}")
			// io.WriteString(out, "\r\n\treturn "+sb.String()+"nil")
		}

		io.WriteString(out, "\r\n\treturn ")

		for _, result := range method.Method.Results.List {
			if result.Type().IsErrorType() {
				continue
			}

			if result.Type().IsPtrType() {
				io.WriteString(out, "&")
			}
			io.WriteString(out, resultName)
			io.WriteString(out, ".E"+result.Name)
			io.WriteString(out, ", ")
		}

		io.WriteString(out, "nil")
	}

	io.WriteString(out, "\r\n}")
	return nil
}

func (cmd *ClientGenerator) genInterfaceMethodStructParam(out io.Writer, method *Method, param *astutil.Param, webPrefix string, needAssignment *bool) error {
	typ := param.Type()
	if typ.IsPtrType() {
		typ = typ.PtrElemType()
	}

	ts, err := typ.ToTypeSpec(true)
	if err != nil {
		return errors.New("param '" + param.Name + "' of '" +
			method.FullName() +
			"' cannot convert to type spec: " + err.Error())
	}

	var fields = ts.Fields()
	for _, f := range ts.Struct.Embedded {
		fields = append(fields, f)
	}

	for idx := range fields {
		var s, _ = getTagValue(&fields[idx], "swaggerignore")
		if strings.ToLower(s) == "true" {
			continue
		}

		hasOmitEmpty := false
		webParamName := webPrefix
		goFieldName := param.Name
		var jsonName string
		if !fields[idx].IsAnonymous {
			jsonName, _ = getTagValue(&fields[idx], "json")
			if jsonName == "" {
				jsonName = toSnakeCase(fields[idx].Name)
			} else {
				ss := strings.Split(jsonName, ",")
				if len(ss) >= 1 && ss[0] != "" {
					jsonName = ss[0]
				}
				for _, s := range ss {
					if s == "omitempty" {
						hasOmitEmpty = true
					}
				}
			}
			if webPrefix == "" {
				webParamName = jsonName
			} else {
				webParamName = webPrefix + "." + jsonName
			}
		}

		if goFieldName == "" {
			goFieldName = fields[idx].Name
		} else {
			goFieldName = param.Name + "." + fields[idx].Name
		}

		switch fields[idx].Type().ToLiteral() {
		case "map[string]string":
			if *needAssignment {
				io.WriteString(out, "\r\nrequest = request.")
			} else {
				io.WriteString(out, ".\r\n")
			}
			if webParamName == "" {
				io.WriteString(out, "SetParamValues("+goFieldName+")")
			} else {
				io.WriteString(out, "SetParamValuesWithPrefix(\""+webParamName+".\", "+goFieldName+")")
			}
			*needAssignment = false
			continue
		case "url.Values":
			if *needAssignment {
				io.WriteString(out, "\r\nrequest = request.")
			} else {
				io.WriteString(out, ".\r\n")
			}
			if webParamName == "" {
				io.WriteString(out, "SetParams("+goFieldName+")")
			} else {
				io.WriteString(out, "SetParamsWithPrefix(\""+webParamName+".\", "+goFieldName+")")
			}
			*needAssignment = false
			continue
		}

		fieldType := fields[idx].Type()
		isPtrType := false
		if t := fieldType.PtrElemType(); t.IsValid() {
			isPtrType = true
			fieldType = t
		}
		if fieldType.IsStructType() &&
			!fieldType.IsSqlNullableType() &&
			!isBultinType(fieldType.ToLiteral()) {
			subparam := *param
			subparam.ExprFile = fieldType.File
			subparam.Expr = fieldType.Expr

			if isPtrType {
				*needAssignment = true
				io.WriteString(out, "\r\n\tif "+param.Name+"."+fields[idx].Name+" != nil {")
			}

			if fields[idx].IsAnonymous {
				err = cmd.genInterfaceMethodStructParam(out, method, &subparam, webParamName, needAssignment)
			} else {
				subparam.Name = param.Name + "." + fields[idx].Name
				err = cmd.genInterfaceMethodStructParam(out, method, &subparam, webParamName, needAssignment)
			}
			if err != nil {
				return err
			}

			if isPtrType {
				io.WriteString(out, "\r\n\t}")
				*needAssignment = true
			}

			continue
		}

		optidx := searchStructFieldParam(method.Operation, param.Name, &fields[idx])
		if optidx < 0 {
			return errors.New("param '" + param.Name + "." + fields[idx].Name +
				"' of '" + method.FullName() +
				"' not found in the swagger1 annotations")
		}

		subparam := *param
		subparam.Name = param.Name + "." + fields[idx].Name
		subparam.ExprFile = fields[idx].Clazz.File
		subparam.Expr = fields[idx].Expr

		option := &method.Operation.Parameters[optidx]
		if option.In == "query" && hasOmitEmpty {
			option.Required = false
		} else {
			option.Required = true
		}
		err := cmd.genInterfaceMethodParam(out, method, &subparam, option, needAssignment)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cmd *ClientGenerator) genInterfaceMethodParam(out io.Writer, method *Method, param *astutil.Param, option *spec.Parameter, needAssignment *bool) error {
	typeName := param.Type().ToLiteral()

	if strings.HasPrefix(typeName, "*http.Request") {
		return nil
	}
	if strings.HasPrefix(typeName, "*") {
		io.WriteString(out, "\r\nif "+param.Name+" != nil {")
		io.WriteString(out, "\r\n\trequest = request.SetParam(\""+option.Name+"\", "+convertToStringLiteral(param, "", cmd.config.ConvertNS, cmd.config.TimeFormat)+")")
		io.WriteString(out, "\r\n}")
		*needAssignment = true
	} else {
		if param.Type().IsSliceType() || param.IsVariadic {
			isStr := param.Type().IsStringType(true)
			if !isStr {
				if param.Type().IsSliceType() {
					typ := param.Type().SliceElemType()
					isStr = typ.IsStringType(true)
				}
			}

			if isStr {
				if *needAssignment {
					io.WriteString(out, "\r\nrequest = request.")
				} else {
					io.WriteString(out, ".\r\n")
				}
				io.WriteString(out, "SetParamArray(\""+option.Name+"\", "+param.Name+")")
				*needAssignment = false
			} else {
				io.WriteString(out, "\r\nfor idx := range "+param.Name+" {")
				io.WriteString(out, "\r\n  request = request.AddParam(\""+option.Name+"\", "+convertToStringLiteral(param, "[idx]", cmd.config.ConvertNS, cmd.config.TimeFormat)+")")
				io.WriteString(out, "\r\n}")
				*needAssignment = true
			}
		} else if param.Type().IsSqlNullableType() {
			io.WriteString(out, "\r\nif "+param.Name+".Valid {")
			io.WriteString(out, "\r\n  request = request.SetParam(\""+option.Name+"\", "+convertToStringLiteral(param, "", cmd.config.ConvertNS, cmd.config.TimeFormat)+")")
			io.WriteString(out, "\r\n}")
			*needAssignment = true
		} else if !option.Required {
			if param.Type().ToLiteral() == "time.Time" {
				io.WriteString(out, "\r\nif !"+param.Name+".IsZero() {")
			} else if param.Type().ToLiteral() == "bool" {
				io.WriteString(out, "\r\nif "+param.Name+" {")
			} else {
				io.WriteString(out, "\r\nif "+param.Name+" != ")
				io.WriteString(out, zeroValueLiteral(param.Type()))
				io.WriteString(out, " {")
			}
			io.WriteString(out, "\r\n\trequest = request.SetParam(\""+option.Name+"\", "+convertToStringLiteral(param, "", cmd.config.ConvertNS, cmd.config.TimeFormat)+")")
			io.WriteString(out, "\r\n}")
			*needAssignment = true
		} else {
			if *needAssignment {
				io.WriteString(out, "\r\nrequest = request.")
			} else {
				io.WriteString(out, ".\r\n")
			}

			io.WriteString(out, "SetParam(\""+option.Name+"\", "+convertToStringLiteral(param, "", cmd.config.ConvertNS, cmd.config.TimeFormat)+")")
			*needAssignment = false
		}
	}

	return nil
}

type ClientConfig struct {
	TagName          string
	RestyName        string
	RestyField       string
	ContextClassName string

	ConvertNS  string
	TimeFormat string

	HasWrapper   bool
	WrapperType  string
	WrapperData  string
	WrapperError string

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

func (c *ClientConfig) ResultName(method Method) string {
	resultName := "result"
	isNameExist := func(name string) bool {
		for idx := range method.Method.Params.List {
			if method.Method.Params.List[idx].Name == name {
				return true
			}
		}

		for idx := range method.Method.Results.List {
			if method.Method.Results.List[idx].Name == name {
				return true
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

func (c *ClientConfig) RouteFunc(method *Method) string {
	return method.Operation.RouterProperties[0].HTTPMethod
}

func (c *ClientConfig) GetPath(optionalRoutePrefix string, method *Method) string {
	rawurl := method.Operation.RouterProperties[0].Path
	var replace = ReplaceFunc(func(segement PathSegement) string {
		for idx := range method.Method.Params.List {
			if strings.EqualFold(method.Method.Params.List[idx].Name, segement.Value) ||
				strings.EqualFold(toSnakeCase(method.Method.Params.List[idx].Name), segement.Value) {
				return "\" + " + convertToStringLiteral(&method.Method.Params.List[idx], "", c.ConvertNS, c.TimeFormat) + " + \""
			}
		}
		err := errors.New(method.Method.Clazz.File.PostionFor(method.Method.Node.Pos()).String() + ": param.Typ '" + segement.Value + "' isnot found")
		log.Fatalln(err)
		panic(err)
	})
	segements, _ := parseURL(rawurl)

	urlPath := JoinPathSegments(segements, false, replace)

	if optionalRoutePrefix == "" {
		return strings.TrimSuffix("\""+urlPath+"\"", "+ \"\"")
	}

	urlPath = strings.TrimPrefix(urlPath, optionalRoutePrefix)
	return strings.TrimSuffix("client.routePrefix() + \""+urlPath+"\"", "+ \"\"")
}

func convertToStringLiteral(param *astutil.Param, index, convertNS, timeFormat string) string {
	name := param.Name

	typ := param.Type()
	var typeStr = typ.ToLiteral()
	if strings.HasPrefix(typeStr, "[]") || param.IsVariadic {
		typ = typ.GetElemType(false)
		typeStr = strings.TrimPrefix(typeStr, "[]")
		if index == "" {
			return convertNS + typeStr + "ArrayToString(" + name + ")"
		}
		name = param.Name + index
	}
	isFirst := true
	needWrap := false

retry:
	switch typeStr {
	case "string":
		return name
	case "*string":
		return "*" + name
	case "int", "int8", "int16", "int32":
		return "strconv.FormatInt(int64(" + name + "), 10)"
	case "*int", "*int8", "*int16", "*int32":
		return "strconv.FormatInt(int64(*" + name + "), 10)"
	case "int64":
		if needWrap {
			return "strconv.FormatInt(int64(" + name + "), 10)"
		}
		return "strconv.FormatInt(" + name + ", 10)"
	case "*int64":
		if needWrap {
			return "strconv.FormatInt(int64(*" + name + "), 10)"
		}
		return "strconv.FormatInt(*" + name + ", 10)"
	case "uint", "uint8", "uint16", "uint32":
		return "strconv.FormatUint(uint64(" + name + "), 10)"
	case "*uint", "*uint8", "*uint16", "*uint32":
		return "strconv.FormatUint(uint64(*" + name + "), 10)"
	case "uint64":
		if needWrap {
			return "strconv.FormatUint(uint64(" + name + "), 10)"
		}
		return "strconv.FormatUint(" + param.Name + ", 10)"
	case "*uint64":
		if needWrap {
			return "strconv.FormatUint(uint64(*" + name + "), 10)"
		}
		return "strconv.FormatUint(*" + name + ", 10)"
	case "bool":
		return convertNS + "BoolToString(" + name + ")"
	case "*bool":
		return convertNS + "BoolToString(*" + name + ")"
	case "time.Time", "*time.Time":
		return name + ".Format(" + timeFormat + ")"
	case "time.Duration", "*time.Duration":
		return name + ".String()"
	case "net.IP", "*net.IP":
		return name + ".String()"
	case "sql.NullTime":
		return name + ".Time.Format(" + timeFormat + ")"
	case "sql.NullBool":
		return convertNS + "BoolToString(" + name + ".Bool)"
	case "sql.NullInt64":
		return "strconv.FormatInt(" + name + ".Int64, 10)"
	case "sql.NullUint64":
		return "strconv.FormatUint(" + name + ".Uint64, 10)"
	case "sql.NullString":
		return name + ".String"
	default:

		underlying := typ.GetUnderlyingType()
		if underlying.IsValid() {
			if isFirst {
				isFirst = false
				needWrap = true
				typeStr = underlying.ToLiteral()

				if typeStr != "string" {
					goto retry
				}
			}
		}

		return "fmt.Sprint(" + name + ")"

		// err := errors.New(param.Method.Ctx.PostionFor(param.Method.Node.Pos()).String() + ": path param '" + param.Name.Name + "' of '" + param.Method.Name.Name + "' is unsupport type - " + typ)

		// log.Fatalln(err)
		// panic(err)
	}
}
