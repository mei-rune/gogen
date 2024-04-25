package gengen

import (
	"errors"
	"flag"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
	"github.com/swaggo/swag"
)

type ServerGenerator struct {
	plugin   string
	ext      string
	buildTag string

	cfg                Config
	convertNamespace   string
	outputHttpCodeWith bool
}

func (cmd *ServerGenerator) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.ext, "ext", "", "文件后缀名")
	fs.StringVar(&cmd.buildTag, "build_tag", "", "生成 go build tag")

	defaultPlugin := os.Getenv("GOGEN_PLUGIN")
	fs.StringVar(&cmd.plugin, "plugin", defaultPlugin, "指定生成框架，可取值: chi, gin, echo, iris, loong")
	
	defaultHttpCodeWith := os.Getenv("GOGEN_HTTPCODEWITH")
	if defaultHttpCodeWith == "" {
		defaultHttpCodeWith = "httpCodeWith"
	}
	fs.StringVar(&cmd.cfg.HttpCodeWith, "httpCodeWith", defaultHttpCodeWith, "使用 httpCodeWith 函数")
	fs.BoolVar(&cmd.outputHttpCodeWith, "outputHttpCodeWith", false, "生成 httpCodeWith 函数")

	fs.StringVar(&cmd.convertNamespace, "convert_ns", "", "转换函数的前缀")
	return fs
}

func (cmd *ServerGenerator) Run(args []string) error {
	if cmd.plugin == "" {
		return errors.New("缺少 plugin 参数")
	}

	plugin, err := createPlugin(cmd.plugin, cmd.cfg)
	if err != nil {
		return err
	}

	if cmd.ext == "" {
		cmd.ext = "." + cmd.plugin + "-gen.go"
	}

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
	_, err = swaggerParser.Packages().ParseTypes()
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

		err = cmd.genHeader(plugin, out, swaggerParser, file)
		if err != nil {
			return err
		}

		err = cmd.genInitFunc(plugin, out, swaggerParser, file)
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

func (cmd *ServerGenerator) genHeader(cfg Plugin, out io.Writer, swaggerParser *swag.Parser, file *astutil.File) error {
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
	for pa, alias := range cfg.Imports() {
		io.WriteString(out, "\r\n\t")
		if alias != "" {
			io.WriteString(out, alias)
			io.WriteString(out, " ")
		}
		io.WriteString(out, "\""+pa+"\"")
	}
	io.WriteString(out, "\r\n)\r\n")

	if cmd.outputHttpCodeWith {
		if cmd.cfg.HttpCodeWith != "" {
			txt := strings.Replace(httpCodeWithTxt, "httpCodeWith", cmd.cfg.HttpCodeWith, -1)

			io.WriteString(out, "\r\n")
			io.WriteString(out, txt)
			io.WriteString(out, "\r\n")
		}
	}
	return nil
}

func (cmd *ServerGenerator) genInitFunc(plugin Plugin, out io.Writer, swaggerParser *swag.Parser, file *astutil.File) error {
	for _, ts := range file.TypeList {
		if ts.Struct == nil && ts.Interface == nil {
			continue
		}


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
			continue
		}


		star := ""
		if ts.Struct != nil {
			star = "*"
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
			count ++
		}
		if count == 0 {
			io.WriteString(out, "\r\n// "+ts.Name+" is skipped")
			continue
		}

		if optionalRoutePrefix != "" {
				io.WriteString(out, "\r\n\r\nfunc Init"+ts.Name+"(mux "+plugin.PartyTypeName()+", enabledPrefix bool, svc "+star+ts.Name+") {")
				if !plugin.IsPartyFluentStyle() {
						io.WriteString(out, "\r\ninitFunc := func(mux "+plugin.PartyTypeName()+") {")
				} else {
					io.WriteString(out, "\r\n\tif enabledPrefix {")
					io.WriteString(out, "\r\n\t\tmux = mux.Group(\""+optionalRoutePrefix+"\")")
					io.WriteString(out, "\r\n\t}")
				}
		} else {
				io.WriteString(out, "\r\n\r\nfunc Init"+ts.Name+"(mux "+plugin.PartyTypeName()+", svc "+star+ts.Name+") {")
		}

		for _, method := range methods {
			// RenderFuncHeader 将输出： mux.Get("/allfiles", func(w http.ResponseWriter, r *http.Request) {
			switch len(method.Operation.RouterProperties) {
			case 0:
				io.WriteString(out, "\r\n// "+method.Method.Name+": annotation is missing")
				continue
				// return errors.New(method.Method.PostionString() + ": RouterProperties is empty")
			case 1:
				break
			default:
				return errors.New(method.Method.PostionString() + ": RouterProperties is mult choices")
			}

			routeProps := method.Operation.RouterProperties[0]
			if optionalRoutePrefix != "" {
				if strings.HasPrefix(routeProps.Path, optionalRoutePrefix) {
					routeProps.Path = strings.TrimPrefix(routeProps.Path, optionalRoutePrefix)
				}
			}

			if err := checkUrlValid(method, routeProps); err != nil {
				return err
			}
			err := plugin.RenderFuncHeader(out, method, routeProps)
			if err != nil {
				return err
			}

			ctx := &GenContext{
				convertNS: cmd.convertNamespace,
				plugin:    plugin,
				out:       out,
			}
			err = method.renderImpl(ctx)
			if err != nil {
				return err
			}
			io.WriteString(out, "\r\n})")
		}


		if optionalRoutePrefix != "" && !plugin.IsPartyFluentStyle() {
			io.WriteString(out, "\r\n\t}")
			io.WriteString(out, "\r\n\tif enabledPrefix {")
			io.WriteString(out, "\r\n\t\tmux = mux.Route(\""+optionalRoutePrefix+"\", initFunc)")
			io.WriteString(out, "\r\n\t} else {")
			io.WriteString(out, "\r\n\t\tinitFunc(mux)")
			io.WriteString(out, "\r\n\t}")
		}

		io.WriteString(out, "\r\n}")
	}
	return nil
}

func ParseFile(ctx *astutil.Context, filename string) (*astutil.File, error) {
	if ctx == nil {
		ctx = astutil.NewContext(nil)
	}

	return ctx.LoadFile(filename)
}

func checkUrlValid(method *Method, routeProps swag.RouteProperties) error {
	for _, param := range method.Operation.Parameters {
		if param.In != "path" {
			continue
		}

		if !strings.Contains(routeProps.Path, "{"+param.Name+"}") {
			return errors.New(method.Method.PostionString() +  ": param '"+param.Name+"' isnot exists in the url path")
		}
	}

	for idx := range method.Operation.Parameters {
		if method.Operation.Parameters[idx].In != "query" {
			continue
		}


		oname := method.Operation.Parameters[idx].Name


		localstructargname, _ := method.Operation.Parameters[idx].Extensions.GetString("x-gogen-extend-struct")
		// localname, _ := method.Operation.Parameters[idx].Extensions.GetString("x-gogen-extend-field")

		if localstructargname != "" {
			continue
		}

		found := false
		for idx := range method.Method.Params.List {
			param := &method.Method.Params.List[idx]

			if strings.EqualFold(oname, param.Name) {
				found = true
				break
			}
		}
		if !found {
			for idx := range method.Method.Params.List {
				param := &method.Method.Params.List[idx]

				snakeParamName := toSnakeCase(param.Name)
				if strings.EqualFold(oname, snakeParamName) {
					found = true
					break
				}
			}
		}
		if !found {
			return errors.New(method.Method.PostionString() +  ": param '"+oname+"' isnot exists in the method param list")
		}
	}
	return nil
}

const httpCodeWithTxt = `func httpCodeWith(err error, statusCode ...int) int {
  if herr, ok := err.(interface{
    HTTPCode() int
    }); ok {
      return herr.HTTPCode()
    }
  if len(statusCode) > 0 {
  	return statusCode[0]
  }
  return http.StatusInternalServerError
}`
