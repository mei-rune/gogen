package gengen

import (
	"errors"
	"flag"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

type Generator struct {
	plugin   string
	ext      string
	buildTag string

	enableHttpCodeWith bool
	convertNamespace   string
}

func (cmd *Generator) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.ext, "ext", ".gogen.go", "文件后缀名")
	fs.StringVar(&cmd.buildTag, "build_tag", "", "生成 go build tag")

	fs.StringVar(&cmd.plugin, "plugin", "", "")
	fs.BoolVar(&cmd.enableHttpCodeWith, "httpCodeWith", false, "生成 enableHttpCodeWith 函数")
	fs.StringVar(&cmd.convertNamespace, "convert_ns", "", "转换函数的前缀")
	return fs
}

func (cmd *Generator) Run(args []string) error {
	plugin, err := createPlugin(cmd.plugin)
	if err != nil {
		return err
	}
	for _, filename := range args {
		file, err := ParseFile(nil, filename)
		if err != nil {
			return err
		}
		targetFile := strings.TrimSuffix(filename, ".go") + cmd.ext
		out, err := os.Create(targetFile)
		if err != nil {
			return err
		}

		err = cmd.genHeader(plugin, out, file)
		if err != nil {
			return err
		}

		err = cmd.genInitFunc(plugin, out, file)
		if err != nil {
			return err
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

func (cmd *Generator) genHeader(cfg Plugin, out io.Writer, file *astutil.File) error {
	if cmd.buildTag != "" {
		io.WriteString(out, "// +build ")
		io.WriteString(out, cmd.buildTag)
		io.WriteString(out, "\r\n")
		io.WriteString(out, "\r\n")
	}
	io.WriteString(out, "// Please don't edit this file!\r\npackage ")
	io.WriteString(out, file.Pkg.Name)
	io.WriteString(out, "\r\n\r\nimport (")

	found := false
	for pa, alias := range cfg.Imports() {
		if alias == "errors" || strings.HasSuffix(pa, "/errors") {
			found = true
			break
		}
	}
	for _, pa := range file.Imports {
		if (pa.Name != nil && astutil.ToString(pa.Name) == "errors") ||
			strings.HasSuffix(strings.Trim(astutil.ToString(pa.Path), "\""), "/errors") {
			found = true
			break
		}
	}
	if !found {
		io.WriteString(out, "\r\n\t\"errors\"")
	}
	for _, pa := range file.Imports {
		io.WriteString(out, "\r\n\t")

		if pa.Name != nil {
			io.WriteString(out, astutil.ToString(pa.Name))
			io.WriteString(out, " ")
		}
		io.WriteString(out, pa.Path.Value)
	}
	for pa, alias := range cfg.Imports() {
		io.WriteString(out, "\r\n\t")
		if alias != "" {
			io.WriteString(out, alias)
			io.WriteString(out, " ")
		}
		io.WriteString(out, "\""+pa+"\"")
	}
	io.WriteString(out, "\r\n)\r\n")
	return nil
}

func (cmd *Generator) genInitFunc(cfg Plugin, out io.Writer, file *astutil.File) error {
	for _, ts := range file.TypeList {
		if ts.Struct == nil && ts.Interface == nil {
			continue
		}
		star := ""
		if ts.Struct != nil {
			star = "*"
		}
		methods, err := resolveMethods(ts)
		if err != nil {
			return err
		}
		if len(methods) == 0 {
			io.WriteString(out, "\r\n// "+ts.Name+" is skipped")
			continue
		}

		io.WriteString(out, "\r\n\r\nfunc Init"+ts.Name+"(mux "+cfg.PartyTypeName()+", svc "+star+ts.Name+") {")
		for _, method := range methods {
			// RenderFuncHeader 将输出： mux.Get("/allfiles", func(w http.ResponseWriter, r *http.Request) {
			switch len(method.Operation.RouterProperties) {
			case 0:
				return errors.New("RouterProperties is empty")
			case 1:
				break
			default:
				return errors.New("RouterProperties is mult choices")
			}
			err := cfg.RenderFuncHeader(out, method, method.Operation.RouterProperties[0])
			if err != nil {
				return err
			}

			err = cmd.genImplFuncBody(cfg, out, ts, method)
			if err != nil {
				return err
			}
			io.WriteString(out, "\r\n})")
		}
		io.WriteString(out, "\r\n}")
	}
	return nil
}

func (cmd *Generator) genImplFuncBody(plugin Plugin, out io.Writer, ts *astutil.TypeSpec, method *Method) error {
	/// 输出参数解析
	if len(method.Method.Params.List) > 0 {
		params, err := method.GetParams()
		if err != nil {
			return err
		}
		for idx := range params {
			err := params[idx].RenderDeclareAndInit(plugin, out, ts, method)
			if err != nil {
				return err
			}
		}
	}

	return cmd.genImplFuncInvokeAndReturn(plugin, out, ts, method)
}

func (cmd *Generator) genImplFuncInvokeAndReturn(cfg Plugin, out io.Writer, ts *astutil.TypeSpec, method *Method) error {
	io.WriteString(out, "\r\n")
	/// 输出返回参数
	if len(method.Method.Results.List) > 2 {
		for idx, result := range method.Method.Results.List {
			if idx > 0 {
				io.WriteString(out, ", ")
			}
			if result.Type().IsErrorType() {
				io.WriteString(out, "err")
			} else {
				io.WriteString(out, result.Name)
			}
		}
		io.WriteString(out, " :=")
	} else if len(method.Method.Results.List) == 1 {
		if method.Method.Results.List[0].Type().IsErrorType() {
			// if isErrorDefined {
			// 	io.WriteString(out, "err =")
			// } else {
			io.WriteString(out, "err :=")
			// }
		} else {
			io.WriteString(out, "result :=")
		}
	} else {
		io.WriteString(out, "result, err :=")
	}

	/// 输出调用
	io.WriteString(out, "svc.")
	io.WriteString(out, method.Method.Name)
	io.WriteString(out, "(")
	params, err := method.GetParams()
	if err != nil {
		return err
	}
	for idx, param := range params {
		// {{- if $param.IsSkipUse -}}
		//    {{- continue --}}
		// {{- end -}}
		if idx > 0 {
			io.WriteString(out, ", ")
		}
		io.WriteString(out, param.GoArgumentLiteral())
		if param.IsVariadic() {
			io.WriteString(out, "...")
		}
	}
	io.WriteString(out, ")")

	/// 输出返回

	if len(method.Method.Results.List) > 2 {
		io.WriteString(out, "\r\n\tif err != nil {")
		cfg.RenderReturnError(out, method, "", "err")
		io.WriteString(out, "\r\n\t}")

		io.WriteString(out, "\r\n\tresult := map[string]interface{}{")

		for _, result := range method.Method.Results.List {
			if result.Type().IsErrorType() {
				continue
			}

			io.WriteString(out, "\r\n\t\"")
			io.WriteString(out, Underscore(result.Name))
			io.WriteString(out, "\":")
			io.WriteString(out, result.Name)
			io.WriteString(out, ",")
		}
		io.WriteString(out, "\r\n\t}\r\n")
		cfg.RenderReturnOK(out, method, "", "result")
	} else if len(method.Method.Results.List) == 1 {

		arg := method.Method.Results.List[0]
		if arg.Type().IsErrorType() {
			io.WriteString(out, "\r\nif err != nil {\r\n")
			cfg.RenderReturnError(out, method, "", "err")
			io.WriteString(out, "\r\n}\r\n")
			cfg.RenderReturnOK(out, method, "", "\"OK\"")
		} else {
			// if methodParams.IsPlainText {
			//  	{{$.mux.PlainTextFunc $method "result"}}
			// } else {
			io.WriteString(out, "\r\n")
			cfg.RenderReturnOK(out, method, "", "result")
			// }
		}

	} else {
		io.WriteString(out, "\r\n\tif err != nil {\r\n")
		cfg.RenderReturnError(out, method, "", "err")
		io.WriteString(out, "\r\n\t}\r\n")

		// {{- if $methodParams.IsPlainText }}
		//   {{$.mux.PlainTextFunc $method "result"}}
		// {{- else}}
		cfg.RenderReturnOK(out, method, "", "result")
		// {{- end}}
	}

	return nil
}

func ParseFile(ctx *astutil.Context, filename string) (*astutil.File, error) {
	if ctx == nil {
		ctx = astutil.NewContext(nil)
	}

	return ctx.LoadFile(filename)
}
