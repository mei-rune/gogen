package gengen

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
	"golang.org/x/text/transform"
)

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	filename := filepath.Join(t.TempDir(), "gogen.properties")
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return filename
}

func TestLoadConfigFile(t *testing.T) {
	filename := writeConfigFile(t, `
# 注释
! 另一种注释
plugin = echo
httpCodeWith: errors.GetHttpCode
enableResultWrap: true
imports = example.com/a, example.com/b
contextGetter = ctx.Value(\"user\")
`)

	cfg, err := LoadConfigFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Plugin != "echo" {
		t.Errorf("plugin = %q, want %q", cfg.Plugin, "echo")
	}
	if cfg.HttpCodeWith != "errors.GetHttpCode" {
		t.Errorf("httpCodeWith = %q", cfg.HttpCodeWith)
	}
	if cfg.EnableResultWrap == nil || !*cfg.EnableResultWrap {
		t.Errorf("enableResultWrap = %v, want true", cfg.EnableResultWrap)
	}
	if len(cfg.Imports) != 2 || cfg.Imports[0] != "example.com/a" || cfg.Imports[1] != "example.com/b" {
		t.Errorf("imports = %#v", cfg.Imports)
	}
	if cfg.ContextGetter != `ctx.Value("user")` {
		t.Errorf("contextGetter = %q", cfg.ContextGetter)
	}
}

func TestLoadConfigFileBOMs(t *testing.T) {
	content := "plugin=echo\nenableResultWrap=true\n"

	for _, test := range []struct {
		name string
		enc  encoding.Encoding
	}{
		{name: "UTF-8", enc: unicode.UTF8BOM},
		{name: "UTF-16LE", enc: unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)},
		{name: "UTF-16BE", enc: unicode.UTF16(unicode.BigEndian, unicode.UseBOM)},
		{name: "UTF-32LE", enc: utf32.UTF32(utf32.LittleEndian, utf32.UseBOM)},
		{name: "UTF-32BE", enc: utf32.UTF32(utf32.BigEndian, utf32.UseBOM)},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, _, err := transform.Bytes(test.enc.NewEncoder(), []byte(content))
			if err != nil {
				t.Fatal(err)
			}

			filename := filepath.Join(t.TempDir(), "gogen.properties")
			if err := os.WriteFile(filename, data, 0644); err != nil {
				t.Fatal(err)
			}

			cfg, err := LoadConfigFile(filename)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Plugin != "echo" {
				t.Errorf("plugin = %q, want %q", cfg.Plugin, "echo")
			}
			if cfg.EnableResultWrap == nil || !*cfg.EnableResultWrap {
				t.Errorf("enableResultWrap = %v, want true", cfg.EnableResultWrap)
			}
		})
	}
}

func TestLoadConfigFileErrors(t *testing.T) {
	for _, test := range []struct {
		name    string
		content string
	}{
		{name: "missing separator", content: "plugin\n"},
		{name: "empty key", content: "= echo\n"},
		{name: "unknown key", content: "unknown-key = 1\n"},
		{name: "invalid bool", content: "enableResultWrap = yes\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := LoadConfigFile(writeConfigFile(t, test.content))
			if err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestResolveConfigFile(t *testing.T) {
	root := t.TempDir()
	defaultConfig := filepath.Join(root, "gogen.properties")
	if err := os.WriteFile(defaultConfig, []byte("plugin=echo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOGEN_ROOT_DIRECTORY", root)

	got, err := resolveConfigFile("conf/gogen.properties")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "conf", "gogen.properties"); got != want {
		t.Errorf("resolveConfigFile() = %q, want %q", got, want)
	}

	abs := filepath.Join(root, "abs.properties")
	got, err = resolveConfigFile(abs)
	if err != nil {
		t.Fatal(err)
	}
	if got != abs {
		t.Errorf("resolveConfigFile(%q) = %q, want %q", abs, got, abs)
	}

	got, err = resolveConfigFile("")
	if err != nil {
		t.Fatal(err)
	}
	if got != defaultConfig {
		t.Errorf("resolveConfigFile(\"\") = %q, want %q", got, defaultConfig)
	}

	customConfig := filepath.Join(root, "conf", "custom.properties")
	if err := os.MkdirAll(filepath.Dir(customConfig), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(customConfig, []byte("plugin=echo\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// 省略 .properties 后缀时也应能解析。
	got, err = resolveConfigFile("conf/custom")
	if err != nil {
		t.Fatal(err)
	}
	if got != customConfig {
		t.Errorf("resolveConfigFile(conf/custom) = %q, want %q", got, customConfig)
	}

	// 未指定 -config 时使用 GOGEN_CONFIG。
	t.Setenv("GOGEN_CONFIG", "conf/custom")
	got, err = resolveConfigFile("")
	if err != nil {
		t.Fatal(err)
	}
	if got != customConfig {
		t.Errorf("resolveConfigFile(\"\") = %q, want %q", got, customConfig)
	}
}

func TestEnvOverridesConfigFile(t *testing.T) {
	t.Setenv("GOGEN_PLUGIN", "gin")
	t.Setenv("GOGEN_HTTPCODEWITH", "env.GetHttpCode")

	filename := writeConfigFile(t, `
plugin = echo
httpCodeWith = config.GetHttpCode
`)

	var gen ServerGenerator
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	gen.Flags(fs)
	if err := fs.Parse([]string{"-config", filename}); err != nil {
		t.Fatal(err)
	}
	if err := gen.applyConfigFile(); err != nil {
		t.Fatal(err)
	}

	if gen.plugin != "gin" {
		t.Errorf("plugin = %q, want %q", gen.plugin, "gin")
	}
	if gen.cfg.HttpCodeWith != "env.GetHttpCode" {
		t.Errorf("httpCodeWith = %q", gen.cfg.HttpCodeWith)
	}
}

func TestResolveConfigFileWithoutRootDirectory(t *testing.T) {
	t.Setenv("GOGEN_ROOT_DIRECTORY", "")
	t.Setenv("GOGEN_CONFIG", "")

	got, err := resolveConfigFile("gogen.properties")
	if err != nil {
		t.Fatal(err)
	}
	if got != "gogen.properties" {
		t.Errorf("resolveConfigFile() = %q, want %q", got, "gogen.properties")
	}

	got, err = resolveConfigFile("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("resolveConfigFile(\"\") = %q, want empty", got)
	}
}

func TestServerConfigFileFromRootDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "gogen.properties"), []byte("plugin=echo\nbadArgument=errors.NewBadArgument\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOGEN_ROOT_DIRECTORY", root)
	t.Setenv("GOGEN_PLUGIN", "")

	var gen ServerGenerator
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	gen.Flags(fs)
	if err := fs.Parse(nil); err != nil {
		t.Fatal(err)
	}
	if err := gen.applyConfigFile(); err != nil {
		t.Fatal(err)
	}

	if gen.plugin != "echo" {
		t.Errorf("plugin = %q, want %q", gen.plugin, "echo")
	}
	if gen.cfg.NewBadArgument != "errors.NewBadArgument" {
		t.Errorf("badArgument = %q", gen.cfg.NewBadArgument)
	}
}

func TestServerConfigFileFromRootDirectoryDefaultName(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "gogen_default.properties"), []byte("plugin=echo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOGEN_ROOT_DIRECTORY", root)
	t.Setenv("GOGEN_PLUGIN", "")

	var gen ServerGenerator
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	gen.Flags(fs)
	if err := fs.Parse(nil); err != nil {
		t.Fatal(err)
	}
	if err := gen.applyConfigFile(); err != nil {
		t.Fatal(err)
	}

	if gen.plugin != "echo" {
		t.Errorf("plugin = %q, want %q", gen.plugin, "echo")
	}
}

func TestServerConfigFilePrecedence(t *testing.T) {
	t.Setenv("GOGEN_PLUGIN", "env-plugin")
	t.Setenv("GOGEN_HTTPCODEWITH", "")

	filename := writeConfigFile(t, `
plugin = echo
server_ext = .server-gen.go
client_ext = .client-gen.go
httpCodeWith = errors.GetHttpCode
enableResultWrap = true
imports = example.com/a, example.com/b
`)

	var gen ServerGenerator
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	gen.Flags(fs)
	if err := fs.Parse([]string{"-config", filename, "-plugin=gin"}); err != nil {
		t.Fatal(err)
	}
	if err := gen.applyConfigFile(); err != nil {
		t.Fatal(err)
	}

	// 命令行显式指定的 plugin 优先级高于配置文件。
	if gen.plugin != "gin" {
		t.Errorf("plugin = %q, want %q", gen.plugin, "gin")
	}
	if gen.cfg.HttpCodeWith != "errors.GetHttpCode" {
		t.Errorf("httpCodeWith = %q", gen.cfg.HttpCodeWith)
	}
	if !gen.enableResultWrap {
		t.Errorf("enableResultWrap = false, want true")
	}
	if gen.ext != ".server-gen.go" {
		t.Errorf("ext = %q", gen.ext)
	}
	if gen.importList != "example.com/a,example.com/b" {
		t.Errorf("importList = %q", gen.importList)
	}
}

func TestClientConfigFilePrecedence(t *testing.T) {
	t.Setenv("GOGEN_CONVERT_NS", "")

	filename := writeConfigFile(t, `
server_ext = .server-gen.go
client_ext = .xx-gen.go
tag = yaml
convert_ns = conv.
has-wrapper = true
imports = example.com/a, example.com/b
`)

	var gen ClientGenerator
	fs := flag.NewFlagSet("client", flag.ContinueOnError)
	gen.Flags(fs)
	if err := fs.Parse([]string{"-config", filename, "-tag=json"}); err != nil {
		t.Fatal(err)
	}
	if err := gen.applyConfigFile(); err != nil {
		t.Fatal(err)
	}

	if gen.ext != ".xx-gen.go" {
		t.Errorf("ext = %q", gen.ext)
	}
	if gen.config.TagName != "json" {
		t.Errorf("tag = %q, want %q", gen.config.TagName, "json")
	}
	if gen.config.ConvertNS != "conv." {
		t.Errorf("convert_ns = %q", gen.config.ConvertNS)
	}
	if !gen.config.HasWrapper {
		t.Errorf("has-wrapper = false, want true")
	}
	if gen.importList != "example.com/a,example.com/b" {
		t.Errorf("importList = %q", gen.importList)
	}
}
