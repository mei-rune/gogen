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
	"strings"

	"github.com/aryann/difflib"
)

func getGogen() string {
	for _, pa := range filepath.SplitList(os.Getenv("GOPATH")) {
		dir := filepath.Join(pa, "src/github.com/runner-mei/gogen/v2")
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			return dir
		}
	}

	parent, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for {
		info, err := os.Stat(filepath.Join(parent, "go.mod"))
		if err == nil && !info.IsDir() {
			break
		}
		d := filepath.Dir(parent)
		if len(d) >= len(parent) {
			return ""
		}
		parent = d
	}
	return filepath.Join(parent, "v2")
}

func TestGenerate(t *testing.T) {
	wd := getGogen()

	type TestCase struct {
		Name string
		Args []string
	}

	testCases := []TestCase {
		{
			Name: "casetest",
		},
		{
			Name: "test",
		},
		{
			Name: "errtest",
			Args: []string{
				"-httpCodeWith=errors.GetHttpCode",
				"-badArgument=errors.NewBadArgument",
				"-toEncodedError=errors.ToEncodedError", 
			},
		},
	}
	t.Run("gingen", func(t *testing.T) {
		for _, test := range testCases {
			t.Log("=====================", test.Name)
			os.Remove(filepath.Join(wd, "gentest", test.Name+".gin-gen.go"))
			// fmt.Println(filepath.Join(wd, "gentest", test.Name+".gobatis.go"))

			var gen = &ServerGenerator{}
			gen.Flags(flag.NewFlagSet("", flag.PanicOnError)).Parse(append([]string{
				"-plugin=gin",
				"-build_tag=gin",
			}, test.Args...))

			if err := gen.Run([]string{filepath.Join(wd, "gentest", test.Name+".go")}); err != nil {
				fmt.Println(err)
				t.Error(err)
				continue
			}

			actual := readFile(filepath.Join(wd, "gentest", test.Name+".gin-gen.go"))
			excepted := readFile(filepath.Join(wd, "gentest", test.Name+".gin-gen.txt"))
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
		for _, test := range testCases {
			t.Log("=====================", test.Name)
			os.Remove(filepath.Join(wd, "gentest", test.Name+".chi-gen.go"))
			// fmt.Println(filepath.Join(wd, "gentest", test.Name+".gobatis.go"))

			var gen = &ServerGenerator{}
			gen.Flags(flag.NewFlagSet("", flag.PanicOnError)).Parse(append([]string{
				"-plugin=chi",
				"-build_tag=chi",
			}, test.Args...))

			if err := gen.Run([]string{filepath.Join(wd, "gentest", test.Name+".go")}); err != nil {
				fmt.Println(err)
				t.Error(err)
				continue
			}

			actual := readFile(filepath.Join(wd, "gentest", test.Name+".chi-gen.go"))
			excepted := readFile(filepath.Join(wd, "gentest", test.Name+".chi-gen.txt"))
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
		for _, test := range testCases {
			t.Log("=====================", test.Name)
			os.Remove(filepath.Join(wd, "gentest", test.Name+".echo-gen.go"))

			var gen = &ServerGenerator{}
			gen.Flags(flag.NewFlagSet("", flag.PanicOnError)).Parse(append([]string{
				"-plugin=echo",
				"-build_tag=echo",
			}, test.Args...))

			if err := gen.Run([]string{filepath.Join(wd, "gentest", test.Name+".go")}); err != nil {
				fmt.Println(err)
				t.Error(err)
				continue
			}

			actual := readFile(filepath.Join(wd, "gentest", test.Name+".echo-gen.go"))
			excepted := readFile(filepath.Join(wd, "gentest", test.Name+".echo-gen.txt"))
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
		for _, test := range testCases {
			t.Log("=====================", test.Name)
			os.Remove(filepath.Join(wd, "gentest", test.Name+".iris-gen.go"))
			// fmt.Println(filepath.Join(wd, "gentest", test.Name+".gobatis.go"))

			var gen = &ServerGenerator{}
			gen.Flags(flag.NewFlagSet("", flag.PanicOnError)).Parse(append([]string{
				"-plugin=iris",
				"-build_tag=iris",
			}, test.Args...))

			if err := gen.Run([]string{filepath.Join(wd, "gentest", test.Name+".go")}); err != nil {
				fmt.Println(err)
				t.Error(err)
				continue
			}

			actual := readFile(filepath.Join(wd, "gentest", test.Name+".iris-gen.go"))
			excepted := readFile(filepath.Join(wd, "gentest", test.Name+".iris-gen.txt"))
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
		for _, test := range testCases {
			t.Log("=====================", test.Name)
			os.Remove(filepath.Join(wd, "gentest", test.Name+".loong-gen.go"))
			// fmt.Println(filepath.Join(wd, "gentest", test.Name+".gobatis.go"))

			var gen = &ServerGenerator{}
			gen.Flags(flag.NewFlagSet("", flag.PanicOnError)).Parse(append([]string{
				"-plugin=loong",
				"-build_tag=loong",
			}, test.Args...))

			if err := gen.Run([]string{filepath.Join(wd, "gentest", test.Name+".go")}); err != nil {
				fmt.Println(err)
				t.Error(err)
				continue
			}

			actual := readFile(filepath.Join(wd, "gentest", test.Name+".loong-gen.go"))
			excepted := readFile(filepath.Join(wd, "gentest", test.Name+".loong-gen.txt"))
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
			gen.buildTag = "!loong"
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

	t.Run("loongclient", func(t *testing.T) {
		for _, name := range []string{"test"} {
			t.Log("=====================", name)
			os.Remove(filepath.Join(wd, "gentest", name+".loongclient-gen.go"))

			var gen = ClientGenerator{
				ext: ".loongclient-gen.go",
			}
			gen.Flags(flag.NewFlagSet("", flag.PanicOnError)).Parse([]string{
				"-has-wrapper", "true",
				"-ext", ".loongclient-gen.go",
			})
			gen.config.HasWrapper = true
			gen.ext = ".loongclient-gen.go"
			gen.buildTag = "loong"

			if err := gen.Run([]string{filepath.Join(wd, "gentest", name+".go")}); err != nil {
				fmt.Println(err)
				t.Error(err)
				continue
			}

			actual := readFile(filepath.Join(wd, "gentest", name+".loongclient-gen.go"))
			excepted := readFile(filepath.Join(wd, "gentest", name+".loongclient-gen.txt"))
			if !reflect.DeepEqual(actual, excepted) {
				results := difflib.Diff(excepted, actual)
				for _, result := range results {
					if result.Delta == difflib.Common {
						continue
					}
					t.Error(result.Delta)
					t.Error(result)
				}
			}
		}
	})

	t.Run("chi", func(t *testing.T) {
		for _, test := range []struct{
			Name string
			Error string
			Code string
		} {
			{
				Name: "path fail",
				Error: "param 'trigger_id' isnot exists in the url path",
				Code: `package main

						type Test interface {
							// @Summary  按 trigger ID 获取所有的告警或历史记录等规则
							// @Param    trigger_id        path   int             true     "trigger 规则的 ID "
							// @Accept   json
							// @Produce  json
							// @Router /query14/{triggerID} [get]
							// @Success 200 {object} interface{}
							FindByTriggerID(triggerID int64) (interface{}, error)
						}`,
			},
			{
				Name: "query fail",
				Error: "param 'abc' isnot exists in the method param list",
				Code: `package main

						type Test interface {
							// @Summary  按 trigger ID 获取所有的告警或历史记录等规则
							// @Param    abc        query   int             true     "trigger 规则的 ID "
							// @Accept   json
							// @Produce  json
							// @Router /query14 [get]
							// @Success 200 {object} interface{}
							FindByTriggerID() (interface{}, error)
						}`,
			},
		} {
			t.Log("=====================", test.Name)
			os.Remove(filepath.Join(wd, "gentest", test.Name+".chi-gen.go"))
			// fmt.Println(filepath.Join(wd, "gentest", name+".gobatis.go"))

			var gen = &ServerGenerator{}
			gen.Flags(flag.NewFlagSet("", flag.PanicOnError)).Parse([]string{
				"-plugin=chi",
				"-build_tag=chi",
			})

			name := "test_error"


			filename := filepath.Join(wd, "gentest", name + ".go")
			ioutil.WriteFile(filename, []byte(test.Code), 0666)

			if err := gen.Run([]string{filename}); err != nil {
				if !strings.Contains(err.Error(), test.Error) {
					t.Error(err)
				}
				continue
			}
			t.Error("want error got ok")
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
