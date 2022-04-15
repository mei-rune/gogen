package gengen

import (
	"errors"
	"fmt"
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
