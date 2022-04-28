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

	enableHttpCodeWith bool
	convertNamespace   string
}

func (cmd *ServerGenerator) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.ext, "ext", "", "文件后缀名")
	fs.StringVar(&cmd.buildTag, "build_tag", "", "生成 go build tag")

	fs.StringVar(&cmd.plugin, "plugin", "", "指定生成框架，可取值: chi, gin, echo, iris, loong")
	fs.BoolVar(&cmd.enableHttpCodeWith, "httpCodeWith", false, "生成 enableHttpCodeWith 函数")
	fs.StringVar(&cmd.convertNamespace, "convert_ns", "", "转换函数的前缀")
	return fs
}

func (cmd *ServerGenerator) Run(args []string) error {
	if cmd.plugin == "" {
		return errors.New("缺少 plugin 参数")
	}
	
	plugin, err := createPlugin(cmd.plugin)
	if err != nil {
		return err
	}

	if cmd.ext == "" {
		cmd.ext = "."+cmd.plugin+"-gen.go"
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

func (cmd *ServerGenerator) genInitFunc(plugin Plugin, out io.Writer, swaggerParser *swag.Parser, file *astutil.File) error {
	for _, ts := range file.TypeList {
		if ts.Struct == nil && ts.Interface == nil {
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
		if len(methods) == 0 {
			io.WriteString(out, "\r\n// "+ts.Name+" is skipped")
			continue
		}

		io.WriteString(out, "\r\n\r\nfunc Init"+ts.Name+"(mux "+plugin.PartyTypeName()+", svc "+star+ts.Name+") {")
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
			err := plugin.RenderFuncHeader(out, method, method.Operation.RouterProperties[0])
			if err != nil {
				return err
			}

			err = method.renderImpl(plugin, out)
			if err != nil {
				return err
			}
			io.WriteString(out, "\r\n})")
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
