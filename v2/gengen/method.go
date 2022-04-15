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
}

func (method *Method) NoReturn() bool {
	// TODO: implement it
	return false
}

func (method *Method) GetParams() ([]Param, error) {
	var results []Param
	for idx := range method.Method.Params.List {
		foundIndex := -1
		for i := range method.Operation.Parameters {
			if method.Operation.Parameters[i].Name == method.Method.Params.List[idx].Name {
				foundIndex = i
				break
			}
			if strings.ToLower(method.Operation.Parameters[i].Name) == strings.ToLower(method.Method.Params.List[idx].Name) {
				foundIndex = i
				break
			}
		}

		if foundIndex < 0 {
			return nil, errors.New("param '" + method.Method.Params.List[0].Name +
				"' of '" + method.Method.Clazz.Name + "." + method.Method.Name +
				"' not found in the swagger annotation")
		}

		results = append(results, Param{
			Param:  &method.Method.Params.List[idx],
			Option: method.Operation.Parameters[foundIndex],
		})
	}

	return results, nil
}

func (method *Method) renderImpl(plugin Plugin, out io.Writer) error {
	/// 输出参数解析
	if len(method.Method.Params.List) > 0 {
		params, err := method.GetParams()
		if err != nil {
			return err
		}
		for idx := range params {
			err := params[idx].RenderDeclareAndInit(plugin, out, method)
			if err != nil {
				return err
			}
		}
	}

	return method.renderInvokeAndReturn(plugin, out)
}

func (method *Method) renderInvokeAndReturn(cfg Plugin, out io.Writer) error {
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

func resolveMethods(ts *astutil.TypeSpec) ([]*Method, error) {
	var methods []*Method
	list := ts.Methods()
	for idx, method := range list {
		var doc = method.Doc()
		if doc == nil || len(doc.List) == 0 {
			continue
		}
		operation := swag.NewOperation(nil)
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
