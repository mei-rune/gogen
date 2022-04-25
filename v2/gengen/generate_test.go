package gengen

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/aryann/difflib"
)

func getGogen() string {
	for _, pa := range filepath.SplitList(os.Getenv("GOPATH")) {
		dir := filepath.Join(pa, "src/github.com/runner-mei/gogen/v2")
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			return dir
		}
	}
	return ""
}

func TestGenerate(t *testing.T) {
	wd := getGogen()

	t.Run("gingen", func(t *testing.T) {
		for _, name := range []string{"casetest", "test"} {
			t.Log("=====================", name)
			os.Remove(filepath.Join(wd, "gentest", name+".gin-gen.go"))
			// fmt.Println(filepath.Join(wd, "gentest", name+".gobatis.go"))

			var gen = &ServerGenerator{
				plugin:   "gin",
				ext:      ".gin-gen.go",
				buildTag: "gin",
				// includes: filepath.Join(wd, "gentest", "models", "requests.go"),
			}
			if err := gen.Run([]string{filepath.Join(wd, "gentest", name+".go")}); err != nil {
				fmt.Println(err)
				t.Error(err)
				continue
			}

			actual := readFile(filepath.Join(wd, "gentest", name+".gin-gen.go"))
			excepted := readFile(filepath.Join(wd, "gentest", name+".gin-gen.txt"))
			if !reflect.DeepEqual(actual, excepted) {
				results := difflib.Diff(excepted, actual)
				for _, result := range results {
					if result.Delta == difflib.Common {
						continue
					}
					t.Error(result)
				}
			}
		}
	})

	t.Run("chi", func(t *testing.T) {
		for _, name := range []string{"casetest", "test"} {
			t.Log("=====================", name)
			os.Remove(filepath.Join(wd, "gentest", name+".chi-gen.go"))
			// fmt.Println(filepath.Join(wd, "gentest", name+".gobatis.go"))

			var gen = &ServerGenerator{
				plugin:   "chi",
				ext:      ".chi-gen.go",
				buildTag: "chi",
				// includes: filepath.Join(wd, "gentest", "models", "requests.go"),
			}
			if err := gen.Run([]string{filepath.Join(wd, "gentest", name+".go")}); err != nil {
				fmt.Println(err)
				t.Error(err)
				continue
			}

			actual := readFile(filepath.Join(wd, "gentest", name+".chi-gen.go"))
			excepted := readFile(filepath.Join(wd, "gentest", name+".chi-gen.txt"))
			if !reflect.DeepEqual(actual, excepted) {
				results := difflib.Diff(excepted, actual)
				for _, result := range results {
					if result.Delta == difflib.Common {
						continue
					}
					t.Error(result)
				}
			}
		}
	})

	t.Run("echo", func(t *testing.T) {
		for _, name := range []string{"casetest", "test"} {
			t.Log("=====================", name)
			os.Remove(filepath.Join(wd, "gentest", name+".echo-gen.go"))

			var gen = &ServerGenerator{
				plugin:   "echo",
				ext:      ".echo-gen.go",
				buildTag: "echo",
				// includes: filepath.Join(wd, "gentest", "models", "requests.go"),
			}
			if err := gen.Run([]string{filepath.Join(wd, "gentest", name+".go")}); err != nil {
				fmt.Println(err)
				t.Error(err)
				continue
			}

			actual := readFile(filepath.Join(wd, "gentest", name+".echo-gen.go"))
			excepted := readFile(filepath.Join(wd, "gentest", name+".echo-gen.txt"))
			if !reflect.DeepEqual(actual, excepted) {
				results := difflib.Diff(excepted, actual)
				for _, result := range results {
					if result.Delta == difflib.Common {
						continue
					}
					t.Error(result)
				}
			}
		}
	})

	t.Run("iris", func(t *testing.T) {
		for _, name := range []string{"casetest", "test"} {
			t.Log("=====================", name)
			os.Remove(filepath.Join(wd, "gentest", name+".iris-gen.go"))
			// fmt.Println(filepath.Join(wd, "gentest", name+".gobatis.go"))

			var gen = &ServerGenerator{
				plugin:   "iris",
				ext:      ".iris-gen.go",
				buildTag: "iris",
				// includes: filepath.Join(wd, "gentest", "models", "requests.go"),
			}
			if err := gen.Run([]string{filepath.Join(wd, "gentest", name+".go")}); err != nil {
				fmt.Println(err)
				t.Error(err)
				continue
			}

			actual := readFile(filepath.Join(wd, "gentest", name+".iris-gen.go"))
			excepted := readFile(filepath.Join(wd, "gentest", name+".iris-gen.txt"))
			if !reflect.DeepEqual(actual, excepted) {
				results := difflib.Diff(excepted, actual)
				for _, result := range results {
					if result.Delta == difflib.Common {
						continue
					}
					t.Error(result)
				}
			}
		}
	})

	t.Run("loong", func(t *testing.T) {
		for _, name := range []string{"casetest", "test"} {
			t.Log("=====================", name)
			os.Remove(filepath.Join(wd, "gentest", name+".loong-gen.go"))
			// fmt.Println(filepath.Join(wd, "gentest", name+".gobatis.go"))

			var gen = &ServerGenerator{
				plugin:   "loong",
				ext:      ".loong-gen.go",
				buildTag: "loong",
				// includes: filepath.Join(wd, "gentest", "models", "requests.go"),
			}
			if err := gen.Run([]string{filepath.Join(wd, "gentest", name+".go")}); err != nil {
				fmt.Println(err)
				t.Error(err)
				continue
			}

			actual := readFile(filepath.Join(wd, "gentest", name+".loong-gen.go"))
			excepted := readFile(filepath.Join(wd, "gentest", name+".loong-gen.txt"))
			if !reflect.DeepEqual(actual, excepted) {
				results := difflib.Diff(excepted, actual)
				for _, result := range results {
					if result.Delta == difflib.Common {
						continue
					}
					t.Error(result)
				}
			}
		}
	})

	// t.Run("beegen", func(t *testing.T) {
	// 	for _, name := range []string{"test"} {
	// 		t.Log("=====================", name)
	// 		os.Remove(filepath.Join(wd, "gentest", name+".beegen.go"))
	// 		// fmt.Println(filepath.Join(wd, "gentest", name+".gobatis.go"))

	// 		var gen = WebServerGenerator{
	// 			GeneratorBase: GeneratorBase{
	// 				ext:      ".beegen.go",
	// 				buildTag: "beego",
	// 				includes: filepath.Join(wd, "gentest", "models", "requests.go"),
	// 			},
	// 			config: "@beego",
	// 		}
	// 		if err := gen.Run([]string{filepath.Join(wd, "gentest", name+".go")}); err != nil {
	// 			fmt.Println(err)
	// 			t.Error(err)
	// 			continue
	// 		}

	// 		actual := readFile(filepath.Join(wd, "gentest", name+".beegen.go"))
	// 		excepted := readFile(filepath.Join(wd, "gentest", name+".beegen.txt"))
	// 		if !reflect.DeepEqual(actual, excepted) {
	// 			results := difflib.Diff(excepted, actual)
	// 			for _, result := range results {
	// 				if result.Delta == difflib.Common {
	// 					continue
	// 				}
	// 				t.Error(result)
	// 			}
	// 		}
	// 	}
	// })

	t.Run("client", func(t *testing.T) {
		for _, name := range []string{"casetest", "test"} {
			t.Log("=====================", name)
			os.Remove(filepath.Join(wd, "gentest", name+".client-gen.go"))
			// fmt.Println(filepath.Join(wd, "gentest", name+".client-gen.go"))

			var gen = ClientGenerator{
				ext: ".client-gen.go",
			}
			gen.Flags(flag.NewFlagSet("", flag.PanicOnError)).Parse([]string{})
			gen.config.HasWrapper = false
			// gen.includes = filepath.Join(wd, "gentest", "models", "requests.go")

			if err := gen.Run([]string{filepath.Join(wd, "gentest", name+".go")}); err != nil {
				fmt.Println(err)
				t.Error(err)
				continue
			}

			actual := readFile(filepath.Join(wd, "gentest", name+".client-gen.go"))
			excepted := readFile(filepath.Join(wd, "gentest", name+".client-gen.txt"))
			if !reflect.DeepEqual(actual, excepted) {
				results := difflib.Diff(excepted, actual)
				for _, result := range results {
					if result.Delta == difflib.Common {
						continue
					}
					t.Error(result)
				}
			}
		}
	})
}

func readFile(pa string) []string {
	bs, err := ioutil.ReadFile(pa)
	if err != nil {
		panic(err)
	}

	return splitLines(bs)
}

func splitLines(txt []byte) []string {
	//r := bufio.NewReader(strings.NewReader(s))
	s := bufio.NewScanner(bytes.NewReader(txt))
	var ss []string
	for s.Scan() {
		ss = append(ss, s.Text())
	}
	return ss
}
