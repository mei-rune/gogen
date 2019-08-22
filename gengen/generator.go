package gengen

import (
	"errors"
	"flag"
	"go/ast"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type Generator interface {
	Flags(fs *flag.FlagSet) *flag.FlagSet
	Run(args []string) error
}

type GeneratorBase struct {
	ext      string
	buildTag string

	imports map[string]string
}

func (cmd *GeneratorBase) Flags(fs *flag.FlagSet) *flag.FlagSet {
	ext := cmd.ext
	if ext == "" {
		ext = ".gogen.go"
	}
	fs.StringVar(&cmd.ext, "ext", ext, "文件后缀名")
	fs.StringVar(&cmd.buildTag, "build_tag", "", "生成 go build tag")
	return fs
}

func (cmd *GeneratorBase) generateHeader(out io.Writer, file *SourceContext, cb func(out io.Writer) error) error {
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
	for pa, alias := range cmd.imports {
		if alias == "errors" || strings.HasSuffix(pa, "/errors") {
			found = true
			break
		}
	}
	if !found {
		io.WriteString(out, "\r\n\t\"errors\"")
	}
	for _, pa := range file.Imports {
		io.WriteString(out, "\r\n\t")
		io.WriteString(out, pa.Path.Value)
	}
	for pa, alias := range cmd.imports {
		io.WriteString(out, "\r\n\t")
		if alias != "" {
			io.WriteString(out, alias)
			io.WriteString(out, " ")
		}

		io.WriteString(out, "\"")
		io.WriteString(out, pa)
		io.WriteString(out, "\"")
	}

	io.WriteString(out, "\r\n)\r\n")

	if cb != nil {
		return cb(out)
	}
	return nil
}

var ErrDuplicated = errors.New("annotation is duplicated")

func findAnnotation(annotations []Annotation, name string) (*Annotation, error) {
	var annotation *Annotation
	for idx := range annotations {
		if annotations[idx].Name != name {
			continue
		}

		if annotation != nil {
			return nil, ErrDuplicated
		}
		annotation = &annotations[idx]
	}
	return annotation, nil
}

func getAnnotation(method Method, nilIfNotExists bool) *Annotation {
	var annotation *Annotation
	for idx := range method.Annotations {
		if !strings.HasPrefix(method.Annotations[idx].Name, "http.") {
			continue
		}

		if annotation != nil {
			log.Fatalln(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": Annotation of method '" + method.Clazz.Name.Name + ":" + method.Name.Name + "' is duplicated"))
		}
		annotation = &method.Annotations[idx]
	}
	if nilIfNotExists {
		return annotation
	}
	if annotation == nil {
		log.Fatalln(errors.New(strconv.Itoa(int(method.Node.Pos())) + ": Annotation of method '" + method.Clazz.Name.Name + ":" + method.Name.Name + "' is missing"))
	}
	return annotation
}

func renderText(txt *template.Template, out io.Writer, renderArgs interface{}) {
	err := txt.Execute(out, renderArgs)
	if err != nil {
		log.Fatalln(err)
	}
}

func renderString(txt string, renderArgs interface{}) string {
	var out strings.Builder
	err := template.Must(template.New("a").Funcs(Funcs).Parse(txt)).Execute(&out, renderArgs)
	if err != nil {
		log.Fatalln(err)
	}
	return out.String()
}

var Funcs = template.FuncMap{
	"concat": func(args ...string) string {
		var sb strings.Builder
		for _, s := range args {
			sb.WriteString(s)
		}
		return sb.String()
	},
	"containSubstr": strings.Contains,
	"startWith":     strings.HasPrefix,
	"endWith":       strings.HasSuffix,
	"trimPrefix":    strings.TrimPrefix,
	"trimSuffix":    strings.TrimSuffix,

	"sub": func(a, b int) int {
		return a - b
	},
	"sum": func(a, b int) int {
		return a + b
	},
	"default": func(value, defvalue interface{}) interface{} {
		if nil == value {
			return defvalue
		}
		if s, ok := value.(string); ok && "" == s {
			return defvalue
		}
		return value
	},
	"set": func(args map[string]interface{}, name string, value interface{}) string {
		args[name] = value
		return ""
	},
	"arg": func(name string, value interface{}, args map[string]interface{}) map[string]interface{} {
		args[name] = value
		return args
	},
	"last": func(objects interface{}) interface{} {
		if objects == nil {
			return nil
		}

		rv := reflect.ValueOf(objects)
		if rv.Kind() == reflect.Array {
			return rv.Index(rv.Len() - 1).Interface()
		}
		if rv.Kind() == reflect.Slice {
			return rv.Index(rv.Len() - 1).Interface()
		}
		return nil
	},

	"isLast": func(objects interface{}, idx int) bool {
		if objects == nil {
			return true
		}

		rv := reflect.ValueOf(objects)
		if rv.Kind() == reflect.Array {
			return (rv.Len() - 1) == idx
		}
		if rv.Kind() == reflect.Slice {
			return (rv.Len() - 1) == idx
		}
		return false
	},
	"typePrint":               typePrint,
	"convertToStringLiteral":  convertToStringLiteral,
	"convertToStringLiteral2": convertToStringLiteral2,
	"goify":                   Goify,
	"underscore":              Underscore,
	"zeroValue": func(typ ast.Expr) string {
		switch typ.(type) {
		case *ast.StarExpr:
			return "nil"
		case *ast.ArrayType:
			return "nil"
		case *ast.MapType:
			return "nil"
		}

		s := typePrint(typ)
		if lit, ok := zeroLits[s]; ok {
			return lit
		}
		return "0"
	},
}

var zeroLits = map[string]string{
	"time.Time": "time.Time{}",
	"string":    "\"\"",
}
