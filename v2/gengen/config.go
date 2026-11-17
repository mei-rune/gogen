package gengen

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
	"golang.org/x/text/transform"
)

// FileConfig 是 gogen 的配置文件。
//
// 配置文件使用 properties 格式，在命令行中用 -config 指定。
// 在 go generate 中可以用 GOGEN_ROOT_DIRECTORY 环境变量指定项目根目录，
// -config 的相对路径会相对该目录解析；未指定 -config 时会自动查找
// <GOGEN_ROOT_DIRECTORY>/gogen.properties：
//
//	plugin=echo
//	httpCodeWith=errors.GetHttpCode
//	imports=example.com/project/errors,example.com/project/model
//
// 空行以及以 # 或 ! 开头的行会被忽略，键和值可以用 = 或 : 分隔。
// 配置项的优先级为：
//
//	命令行参数 > GOGEN_ 环境变量 > 配置文件 > 默认值
//
// 这样在脚本中仍可以用环境变量临时覆盖配置文件的某一项。
//
// server 和 client 子命令读取各自相关的字段，其余字段会被忽略，
// 因此同一个配置文件可以同时用于两个子命令。
type FileConfig struct {
	// 通用配置
	Imports           []string
	Errors            string
	ConvertNS         string
	ConvertParamTypes string
	ContextGetter     string
	EchoVersion       string
	SkipCompare       *bool

	// server 配置
	Plugin             string
	ServerExt          string
	BuildTag           string
	HttpCodeWith       string
	BadArgument        string
	ToEncodedError     string
	EnableResultWrap   *bool
	OkResult           string
	ErrorResult        string
	CustomReturn       string
	OutputHttpCodeWith *bool

	// client 配置
	ClientExt    string
	Tag          string
	Field        string
	Resty        string
	Context      string
	NewRequest   string
	FreeRequest  string
	TimeFormat   string
	HasWrapper   *bool
	WrapperType  string
	WrapperData  string
	WrapperError string
}

// SplitList 把逗号分隔的字符串拆成列表，并去掉空项。
func SplitList(s string) []string {
	var list []string
	for _, item := range strings.Split(s, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			list = append(list, item)
		}
	}
	return list
}

type property struct {
	key   string
	value string
	line  int
}

// defaultConfigFilenames 是在 GOGEN_ROOT_DIRECTORY 中自动查找的配置文件名。
// 未指定 -config 时，找到第一个存在的文件就用它。
var defaultConfigFilenames = []string{"gogen.properties", "gogen_default.properties", ".gogen.properties"}

// resolveConfigFile 解析配置文件的路径。
//
//   - -config 是绝对路径时直接使用；
//   - -config 是相对路径且设置了 GOGEN_ROOT_DIRECTORY 时，相对该目录解析；
//   - -config 省略 .properties 后缀时，会自动尝试加上该后缀；
//   - 未指定 -config 时，使用 GOGEN_CONFIG 环境变量指定的文件名；
//   - 两者都没有且设置了 GOGEN_ROOT_DIRECTORY 时，在该目录下查找
//     gogen.properties、gogen_default.properties 或 .gogen.properties。
//
// 这样在 go generate 中就不必为每个子目录写不同的相对路径。
func resolveConfigFile(filename string) (string, error) {
	root := os.Getenv("GOGEN_ROOT_DIRECTORY")

	if filename == "" {
		filename = os.Getenv("GOGEN_CONFIG")
	}

	if filename == "" {
		if root == "" {
			return "", nil
		}

		for _, name := range defaultConfigFilenames {
			candidate := filepath.Join(root, name)
			if isFile(candidate) {
				return candidate, nil
			}
		}
		return "", nil
	}

	var resolved string
	if filepath.IsAbs(filename) {
		resolved = filename
	} else if root != "" {
		resolved = filepath.Join(root, filename)
	} else {
		resolved = filename
	}
	return withConfigExt(resolved), nil
}

// withConfigExt 允许 -config=gogen_default 这种省略 .properties 后缀的写法。
func withConfigExt(filename string) string {
	if filepath.Ext(filename) != "" || isFile(filename) {
		return filename
	}
	if withExt := filename + ".properties"; isFile(withExt) {
		return withExt
	}
	return filename
}

func isFile(filename string) bool {
	st, err := os.Stat(filename)
	return err == nil && !st.IsDir()
}

// LoadConfigFile 读取并解析 properties 格式的配置文件。
// 支持 UTF-8（带或不带 BOM）、UTF-16 LE/BE、UTF-32 LE/BE，
// 会根据 BOM 自动识别编码。
func LoadConfigFile(filename string) (*FileConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", filename, err)
	}

	content, err := decodeConfigBytes(data)
	if err != nil {
		return nil, fmt.Errorf("decode config file %q: %w", filename, err)
	}

	props, err := parseProperties(content)
	if err != nil {
		return nil, fmt.Errorf("parse config file %q: %w", filename, err)
	}

	var cfg FileConfig
	if err := cfg.load(props); err != nil {
		return nil, fmt.Errorf("parse config file %q: %w", filename, err)
	}
	return &cfg, nil
}

// BOM 定义，注意 UTF-32 LE/BE 必须在 UTF-16 LE/BE 之前判断。
var (
	utf32BEBOM = []byte{0x00, 0x00, 0xFE, 0xFF}
	utf32LEBOM = []byte{0xFF, 0xFE, 0x00, 0x00}
	utf8BOM    = []byte{0xEF, 0xBB, 0xBF}
	utf16BEBOM = []byte{0xFE, 0xFF}
	utf16LEBOM = []byte{0xFF, 0xFE}
)

// decodeConfigBytes 根据 BOM 自动识别并解码 UTF-8/UTF-16/UTF-32。
// 没有 BOM 时按 UTF-8 处理，与 properties 的常见用法一致。
func decodeConfigBytes(data []byte) (string, error) {
	switch {
	case bytes.HasPrefix(data, utf32BEBOM):
		return decodeWith(data, utf32.UTF32(utf32.BigEndian, utf32.UseBOM))
	case bytes.HasPrefix(data, utf32LEBOM):
		return decodeWith(data, utf32.UTF32(utf32.LittleEndian, utf32.UseBOM))
	case bytes.HasPrefix(data, utf8BOM):
		return string(data[len(utf8BOM):]), nil
	case bytes.HasPrefix(data, utf16BEBOM):
		return decodeWith(data, unicode.UTF16(unicode.BigEndian, unicode.UseBOM))
	case bytes.HasPrefix(data, utf16LEBOM):
		return decodeWith(data, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM))
	default:
		return string(data), nil
	}
}

func decodeWith(data []byte, enc encoding.Encoding) (string, error) {
	decoded, _, err := transform.Bytes(enc.NewDecoder(), data)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(string(decoded), "\ufeff"), nil
}

func parseProperties(content string) ([]property, error) {
	var props []property

	lines := strings.Split(content, "\n")
	for idx, raw := range lines {
		lineNo := idx + 1
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		sep := findSeparator(line)
		if sep < 0 {
			return nil, fmt.Errorf("line %d: expected key=value", lineNo)
		}

		key := strings.TrimSpace(line[:sep])
		if key == "" {
			return nil, fmt.Errorf("line %d: key is empty", lineNo)
		}
		value := unescapeProperty(strings.TrimSpace(line[sep+1:]))
		props = append(props, property{key: key, value: value, line: lineNo})
	}
	return props, nil
}

func findSeparator(line string) int {
	escaped := false
	for idx, r := range line {
		if escaped {
			escaped = false
			continue
		}
		switch r {
		case '\\':
			escaped = true
		case '=', ':':
			return idx
		}
	}
	return -1
}

func unescapeProperty(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}

	var out strings.Builder
	escaped := false
	for _, r := range s {
		if !escaped {
			if r == '\\' {
				escaped = true
				continue
			}
			out.WriteRune(r)
			continue
		}

		escaped = false
		switch r {
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 't':
			out.WriteByte('\t')
		default:
			out.WriteRune(r)
		}
	}
	if escaped {
		out.WriteByte('\\')
	}
	return out.String()
}

func (cfg *FileConfig) load(props []property) error {
	for _, prop := range props {
		switch prop.key {
		case "imports":
			cfg.Imports = SplitList(prop.value)
		case "errors":
			cfg.Errors = prop.value
		case "convert_ns":
			cfg.ConvertNS = prop.value
		case "convert_param_types":
			cfg.ConvertParamTypes = prop.value
		case "contextGetter":
			cfg.ContextGetter = prop.value
		case "echoVersion":
			cfg.EchoVersion = prop.value
		case "skip_compare", "skipCompare":
			value, err := parsePropertyBool(prop)
			if err != nil {
				return err
			}
			cfg.SkipCompare = &value

		case "plugin":
			cfg.Plugin = prop.value
		case "server_ext":
			cfg.ServerExt = prop.value
		case "build_tag":
			cfg.BuildTag = prop.value
		case "httpCodeWith":
			cfg.HttpCodeWith = prop.value
		case "badArgument":
			cfg.BadArgument = prop.value
		case "toEncodedError":
			cfg.ToEncodedError = prop.value
		case "enableResultWrap":
			value, err := parsePropertyBool(prop)
			if err != nil {
				return err
			}
			cfg.EnableResultWrap = &value
		case "okResult":
			cfg.OkResult = prop.value
		case "errorResult":
			cfg.ErrorResult = prop.value
		case "customReturn":
			cfg.CustomReturn = prop.value
		case "outputHttpCodeWith":
			value, err := parsePropertyBool(prop)
			if err != nil {
				return err
			}
			cfg.OutputHttpCodeWith = &value

		case "client_ext":
			cfg.ClientExt = prop.value
		case "tag":
			cfg.Tag = prop.value
		case "field":
			cfg.Field = prop.value
		case "resty":
			cfg.Resty = prop.value
		case "context":
			cfg.Context = prop.value
		case "new-request":
			cfg.NewRequest = prop.value
		case "free-request":
			cfg.FreeRequest = prop.value
		case "timeFormat":
			cfg.TimeFormat = prop.value
		case "has-wrapper":
			value, err := parsePropertyBool(prop)
			if err != nil {
				return err
			}
			cfg.HasWrapper = &value
		case "wrapper-type":
			cfg.WrapperType = prop.value
		case "wrapper-data":
			cfg.WrapperData = prop.value
		case "wrapper-error":
			cfg.WrapperError = prop.value

		default:
			return fmt.Errorf("line %d: unknown config key %q", prop.line, prop.key)
		}
	}
	return nil
}

func parsePropertyBool(prop property) (bool, error) {
	value, err := strconv.ParseBool(prop.value)
	if err != nil {
		return false, fmt.Errorf("line %d: %s expects true or false, got %q", prop.line, prop.key, prop.value)
	}
	return value, nil
}

// flagSet 判断某个命令行参数是否被显式指定过。
func flagSet(fs *flag.FlagSet, name string) bool {
	if fs == nil {
		return false
	}

	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// configEnvNames 是配置项对应的 GOGEN_ 环境变量。
// 环境变量显式设置时优先于配置文件，便于在脚本中做临时覆盖。
var configEnvNames = map[string]string{
	"plugin":              "GOGEN_PLUGIN",
	"httpCodeWith":        "GOGEN_HTTPCODEWITH",
	"badArgument":         "GOGEN_BADARGUMENT",
	"toEncodedError":      "GOGEN_TOJSONERROR",
	"enableResultWrap":    "GOGEN_ENABLE_RESULT_WRAP",
	"okResult":            "GOGEN_OK_RESULT",
	"errorResult":         "GOGEN_ERROR_RESULT",
	"customReturn":        "GOGEN_CUSTOM_RETURN_FUNC",
	"convert_ns":          "GOGEN_CONVERT_NS",
	"convert_param_types": "GOGEN_CONVERT_PARAM_TYPES",
	"imports":             "GOGEN_IMPORTS",
	"errors":              "GOGEN_ERRORS",
	"contextGetter":       "GOGEN_CONTEXT_GETTER",
	"echoVersion":         "GOGEN_ECHO_VERSION",
	"skip_compare":        "GOGEN_SKIP_COMPARE",
}

func envWasSet(name string) bool {
	envName := configEnvNames[name]
	return envName != "" && os.Getenv(envName) != ""
}

// applyString 在命令行和环境变量都没有指定 name 时，用配置文件中的值覆盖 dst。
func applyString(fs *flag.FlagSet, name string, dst *string, value string) {
	if value != "" && !flagSet(fs, name) && !envWasSet(name) {
		*dst = value
	}
}

// applyBool 在命令行和环境变量都没有指定 name 时，用配置文件中的值覆盖 dst。
func applyBool(fs *flag.FlagSet, name string, dst *bool, value *bool) {
	if value != nil && !flagSet(fs, name) && !envWasSet(name) {
		*dst = *value
	}
}
