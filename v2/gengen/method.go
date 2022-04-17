package gengen

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
	"github.com/swaggo/swag"
)

type Method struct {
	Method    *astutil.Method
	Operation *swag.Operation

	errorDeclared bool
}

func (method *Method) SetErrorDeclared() {
	method.errorDeclared = true
}

func (method *Method) IsErrorDeclared() bool {
	return method.errorDeclared
}

func (method *Method) FullName() string {
	return method.Method.Clazz.Name + "." + method.Method.Name
}

func (method *Method) NoReturn() bool {
	// TODO: implement it
	return false
}

func (method *Method) SearchSwaggerParameter(structargname, name string) int {
	foundIndex := -1
	for idx := range method.Operation.Parameters {
		localstructargname, _ := method.Operation.Parameters[idx].Extensions.GetString("x-extend-struct")
		localname, _ := method.Operation.Parameters[idx].Extensions.GetString("x-extend-field")	

		if strings.EqualFold(structargname, localstructargname) &&
		    strings.EqualFold(name, localname) {
			
			foundIndex = idx
			break
		}
	}
	return foundIndex
}


func (method *Method) collectBodyParams(plugin Plugin, params []Param) []int {
	var result []int
	for idx := range params {
		if params[idx].Option.In == "body" || 
			params[idx].Option.In == "formData" {
				result = append(result, idx)
		}
	}
	return result
}

func (method *Method) GetParams(plugin Plugin) ([]Param, error) {
	var results []Param
	for idx := range method.Method.Params.List {
		if _, ok := plugin.TypeInContext(method.Method.Params.List[idx].Type().ToLiteral()); ok {
			results = append(results, Param{
				Method: method,
				Param:  &method.Method.Params.List[idx],
			})
			continue
		}


		foundIndex := -1
		for i := range method.Operation.Parameters {
			oname := method.Operation.Parameters[i].Name
			pname := method.Method.Params.List[idx].Name
			if strings.EqualFold(oname, pname) {
				foundIndex = i
				break
			}
		}

		if foundIndex >= 0 {
			results = append(results, Param{
				Method: method,
				Param:  &method.Method.Params.List[idx],
				Option: method.Operation.Parameters[foundIndex],
			})
			continue
		}

		if foundIndex < 0 {
			for i := range method.Operation.Parameters {
				structargname, _ := method.Operation.Parameters[i].Extensions.GetString("x-extend-struct")
				pname := method.Method.Params.List[idx].Name
				if strings.EqualFold(structargname, pname) {
					foundIndex = i
					break
				}
			}
		}

		if foundIndex < 0 {
			return nil, errors.New("param '" + method.Method.Params.List[idx].Name +
				"' of '" + method.FullName() +
				"' not found in the swagger annotations")
		}

		results = append(results, Param{
			Method: method,
			Param:  &method.Method.Params.List[idx],
		})
	}

	return results, nil
}

func (method *Method) renderImpl(plugin Plugin, out io.Writer) error {
	params, err := method.GetParams(plugin)
	if err != nil {
		return err
	}

	/// 输出参数解析
	for idx := range params {
		err := params[idx].RenderDeclareAndInit(plugin, out)
		if err != nil {
			return err
		}
	}

	/// 输出 body 参数的初始化
	list := method.collectBodyParams(plugin, params)
	if len(list) > 0 {
		err := method.renderBodyParams(plugin, out, params, list)
		if err != nil {
			return err
		}
	}


	return method.renderInvokeAndReturn(plugin, out, params)
}

func isEntire(param *Param) bool {
	ok, _ := param.Option.Extensions.GetBool("x-entire-body")
	return ok
}

func (method *Method) renderBodyParams(plugin Plugin, out io.Writer, params []Param, list []int) error {
	if len(list) == 1  && isEntire(&params[list[0]]) {
		io.WriteString(out, "\r\n\tvar bindArgs "+params[list[0]].Type().ToLiteral())
		params[list[0]].goArgumentLiteral = "bindArgs"
	} else {
		io.WriteString(out, "\r\n\tvar bindArgs struct{")
		for _, idx := range list {
			fieldName := toUpperFirst(params[idx].Param.Name)
			io.WriteString(out, "\r\n\t\t")
			io.WriteString(out, fieldName)
			io.WriteString(out, "\t")
			io.WriteString(out, params[idx].Param.Type().ToLiteral())
			io.WriteString(out, "\t`json:\"")
			if params[idx].Option.Name != "" {
				io.WriteString(out, toSnakeCase(params[idx].Option.Name))
			} else {
				io.WriteString(out, toSnakeCase(params[idx].Param.Name))
			}
			io.WriteString(out, ",omitempty\"`")

			params[idx].goArgumentLiteral = "bindArgs." + fieldName
		}
		io.WriteString(out, "\r\n\t}")
	}

	io.WriteString(out, "\r\n\tif err := ")
	io.WriteString(out, plugin.ReadBodyFunc("&bindArgs"))
	io.WriteString(out, "; err != nil {\r\n")
	txt := `NewBadArgument(err, "bindArgs", "body")`
	plugin.RenderReturnError(out, method, "http.StatusBadRequest", txt)
	io.WriteString(out, "\r\n\t}")

	return nil
}

func (method *Method) renderInvokeAndReturn(plugin Plugin, out io.Writer, params []Param) error {
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
			if method.IsErrorDeclared() {
			 	io.WriteString(out, "err =")
			} else {
				io.WriteString(out, "err :=")
			}
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
		plugin.RenderReturnError(out, method, "", "err")
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
		plugin.RenderReturnOK(out, method, "", "result")
	} else if len(method.Method.Results.List) == 1 {

		arg := method.Method.Results.List[0]
		if arg.Type().IsErrorType() {
			io.WriteString(out, "\r\nif err != nil {\r\n")
			plugin.RenderReturnError(out, method, "", "err")
			io.WriteString(out, "\r\n}\r\n")
			plugin.RenderReturnOK(out, method, "", "\"OK\"")
		} else {
			// if methodParams.IsPlainText {
			//  	{{$.mux.PlainTextFunc $method "result"}}
			// } else {
			io.WriteString(out, "\r\n")
			plugin.RenderReturnOK(out, method, "", "result")
			// }
		}

	} else {
		io.WriteString(out, "\r\n\tif err != nil {\r\n")
		plugin.RenderReturnError(out, method, "", "err")
		io.WriteString(out, "\r\n\t}\r\n")

		// {{- if $methodParams.IsPlainText }}
		//   {{$.mux.PlainTextFunc $method "result"}}
		// {{- else}}
		plugin.RenderReturnOK(out, method, "", "result")
		// {{- end}}
	}

	return nil
}

func resolveMethods(swaggerParser *swag.Parser, ts *astutil.TypeSpec) ([]*Method, error) {
	var methods []*Method
	list := ts.Methods()
	for idx, method := range list {
		var doc = method.Doc()
		if doc == nil || len(doc.List) == 0 {
			continue
		}
		operation := swag.NewOperation(swaggerParser)
		for _, comment := range doc.List {
			err := operation.ParseComment(comment.Text, ts.File.AstFile)
			if err != nil {
				return nil, fmt.Errorf("ParseComment error in file %s :%+v", ts.File.Filename, err)
			}
		}

		methods = append(methods, &Method{
			Method:    &list[idx],
			Operation: operation,
		})
	}
	return methods, nil
}
