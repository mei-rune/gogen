package gengen

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"

	"github.com/go-openapi/spec"
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

type Param struct {
	Param  *astutil.Param
	Option spec.Parameter
}

// GoVarName 申明时的变量名
func (param *Param) GoVarName() string {
	return param.Param.Name
}

// GoVarName 函数调用时变量，如变量名为 a, 调用可能是 &a
func (param *Param) GoArgumentLiteral() string {
	var isStringPtrType = param.IsPtrType() &&  astutil.ToString(param.PtrElemType()) == "string"
	if isStringPtrType && param.Option.In == "path"{
		return "&" + param.Param.Name
	}
	return param.Param.Name
}

func (param *Param) GoType() string {
	return astutil.ToString(param.Param.Typ)
}

// func (param *Param) IsExceptedType(underlying bool, excepted ...string) bool {
// 	typ := param.Param.Typ
// 	if astutil.IsPtrType(typ) {
// 		for _, name := range excepted {
// 			if name == "ptr" {
// 				return true
// 			}
// 		}
// 		if underlying {
// 			typ = astutil.PtrElemType(typ)
// 		}
// 	} else if astutil.IsSliceType(typ) {
// 		for _, name := range excepted {
// 			if name == "slice" {
// 				return true
// 			}
// 		}
// 		if underlying {
// 			typ = astutil.SliceElemType(typ)
// 		}
// 	}

// 	for _, name := range excepted {
// 		switch name {
// 		case "basic":
// 			if astutil.IsBasicType(typ) {
// 				return true
// 			}
// 		case "nullable":
// 			if isNullableType(astutil.ToString(typ)) {
// 				return true
// 			}
// 		case "bultin":
// 			if isBultinType(astutil.ToString(typ)) {
// 				return true
// 			}
// 		default:
// 			panic("type '" + name + "' is unknown type")
// 		}
// 	}

// 	return false
// }

// func (param *Param) IsBasicType() bool {
// 	return astutil.IsBasicType(param.ElemType())
// }

// func (param *Param) IsSimpleValue() bool {
// 	name := astutil.ToString(param.ElemType())

// 	return "net.IP" == name ||
// 	"net.HardwareAddr" == name ||
// 	"time.Time"
// }

func (param *Param) IsPtrType() bool {
	return astutil.IsPtrType(param.Param.Typ)
}

func (param *Param) IsArrayType() bool {
	return astutil.IsSliceType(param.Param.Typ)
}

func (param *Param) Type() ast.Expr {
	return param.Param.Typ
}

// func (param *Param) UnderlyingType() ast.Expr {
// 	typ := astutil.PtrElemType(param.Param.Typ)
// 	if typ == nil {
// 		typ = param.Param.Typ
// 	}
// 	elmType := astutil.SliceElemType(typ)
// 	if elmType == nil {
// 		elmType = typ
// 	}
// 	return elmType
// }

func (param *Param) SliceType() ast.Expr {
	return astutil.SliceElemType(param.Param.Typ)
}

func (param *Param) PtrElemType() ast.Expr {
	return astutil.PtrElemType(param.Param.Typ)
}

func (param *Param) IsVariadic() bool {
	return param.Param.IsVariadic
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
