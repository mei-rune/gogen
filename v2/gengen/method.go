package gengen

import (
	"errors"
	"fmt"
	"go/ast"
	"io"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
	"github.com/swaggo/swag"
)

type GenContext struct {
	convertNS string
	plugin    Plugin
	out       io.Writer
}

var specificParamName = "otherValues"

func resolveMethods(swaggerParser *swag.Parser, ts *astutil.TypeSpec) ([]*Method, error) {
	var methods []*Method
	list := ts.Methods()
	for idx, method := range list {
		var doc = method.Doc()

		operation := swag.NewOperation(swaggerParser)
		if doc == nil || len(doc.List) == 0 {
			methods = append(methods, &Method{
				Method:    &list[idx],
				Operation: operation,
			})
			continue
		}
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

type Method struct {
	Method    *astutil.Method
	Operation *swag.Operation

	errorDeclared      bool
	goArgumentLiterals []string
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

func searchParam(operation *swag.Operation, paramName string) int {
	for i := range operation.Parameters {
		oname := operation.Parameters[i].Name
		if strings.EqualFold(oname, paramName) {
			return i
		}
	}

	snakeParamName := toSnakeCase(paramName)
	for i := range operation.Parameters {
		oname := operation.Parameters[i].Name

		if strings.EqualFold(oname, snakeParamName) {
			return i
		}
	}
	return -1
}

func searchStructParam(operation *swag.Operation, paramName string) *spec.Parameter {
	snakeParamName := toSnakeCase(paramName)
	for key, value := range operation.Extensions {
		if !strings.HasPrefix(key, "x-gogen-param-") {
			continue
		}

		structargname := strings.TrimPrefix(key, "x-gogen-param-")
		if !strings.EqualFold(structargname, paramName) &&
			!strings.EqualFold(structargname, snakeParamName) {
			continue
		}

		param, ok := value.(spec.Parameter)
		if ok {
			return &param
		}
	}
	return nil
}

func searchStructFieldParam(operation *swag.Operation, structargname string, field *astutil.Field) int {
	name := field.Name
	jsonName := field.Name
	if s, _ := getTagValue(field, "json"); s != "" {
		jsonName = getJSONName(s)
	}

	snakeCaseStructArgName := toSnakeCase(structargname)
	for idx := range operation.Parameters {
		localstructargname, _ := operation.Parameters[idx].Extensions.GetString("x-gogen-extend-struct")
		localname, _ := operation.Parameters[idx].Extensions.GetString("x-gogen-extend-field")

		// fmt.Println("====",operation.Parameters[idx].Name, "|", localstructargname, "|", localname, "|", structargname, "|", name)
		if (strings.EqualFold(name, localname) ||
			strings.EqualFold(jsonName, localname)) &&
			(strings.EqualFold(structargname, localstructargname) ||
				strings.EqualFold(snakeCaseStructArgName, localstructargname)) {
			return idx
		}
	}

	return -1
}

type Field struct {
	*astutil.Field

	isInitialized bool
	isFirstField  bool
}

type Param struct {
	*astutil.Param
	option *spec.Parameter
	index  int

	isInitialized bool
}

type BodyParam struct {
	Param  *astutil.Param
	Option *spec.Parameter
	Index  int
}

func defaultValue(resultType string, param *Param, parents []*Field) string {
	var value interface{}
	if len(parents) == 0 {
		// FIXME: get default value
	} else {
		value = param.option.Default
	}
	if value != nil {
		if resultType == "string" {
			return "\"" + fmt.Sprint(value) + "\""
		}
		return fmt.Sprint(value)
	}
	if resultType == "string" {
		return "\"\""
	}
	return "0"
}

func (method *Method) renderImpl(ctx *GenContext) error {
	method.goArgumentLiterals = make([]string, len(method.Method.Params.List))

	var inBody []BodyParam

	for idx := range method.Method.Params.List {
		param := &method.Method.Params.List[idx]

		paramType := param.Type()

		switch paramType.ToLiteral() {
		case "map[string]string":
			st := searchStructParam(method.Operation, param.Name)
			if st == nil {
				foundIndex := searchParam(method.Operation, param.Name)
				if foundIndex >= 0 {
					st = &method.Operation.Parameters[foundIndex]
				}
			}
			if st != nil {
				if st.In == "body" || st.In == "formData" {
					method.goArgumentLiterals[idx] = ""

					inBody = append(inBody, BodyParam{
						Param:  param,
						Option: st,
						Index:  idx,
					})
					continue
				}
			}

			method.goArgumentLiterals[idx] = param.Name
			io.WriteString(ctx.out, "\r\n")
			err := method.renderMapParam(ctx,
				&Param{Param: param, option: st, index: idx},
				nil,
				"map[string]string",
				"[len(values)-1]")
			if err != nil {
				return err
			}
			continue
		case "url.Values":
			st := searchStructParam(method.Operation, param.Name)
			if st == nil {
				foundIndex := searchParam(method.Operation, param.Name)
				if foundIndex >= 0 {
					st = &method.Operation.Parameters[foundIndex]
				}
			}
			if  st != nil {
				if st.In == "body" || st.In == "formData" {
					method.goArgumentLiterals[idx] = ""

					inBody = append(inBody, BodyParam{
						Param:  param,
						Option: st,
						Index:  idx,
					})
					continue
				}
			}

			method.goArgumentLiterals[idx] = param.Name
			io.WriteString(ctx.out, "\r\n")
			err := method.renderMapParam(ctx,
				&Param{Param: param, option: st, index: idx},
				nil,
				"url.Values",
				"")
			if err != nil {
				return err
			}
			continue
		}

		if s, ok := ctx.plugin.GetSpecificTypeArgument(paramType.ToLiteral()); ok {
			method.goArgumentLiterals[idx] = s
			continue
		}

		foundIndex := searchParam(method.Operation, param.Name)
		if foundIndex >= 0 {
			if method.Operation.Parameters[foundIndex].In == "body" ||
				method.Operation.Parameters[foundIndex].In == "formData" {
				method.goArgumentLiterals[idx] = ""

				inBody = append(inBody, BodyParam{
					Param:  param,
					Option: &method.Operation.Parameters[foundIndex],
					Index:  idx,
				})
				continue
			}

			method.goArgumentLiterals[idx] = param.Name

			option := &method.Operation.Parameters[foundIndex]
			err := method.renderSimpleParam(ctx,
				&Param{Param: param, option: option, index: idx})
			if err != nil {
				return err
			}

			continue
		}

		if st := searchStructParam(method.Operation, param.Name); st != nil {
			if st.In == "body" || st.In == "formData" {
				method.goArgumentLiterals[idx] = ""

				inBody = append(inBody, BodyParam{
					Param:  param,
					Option: st,
					Index:  idx,
				})
				continue
			}
			method.goArgumentLiterals[idx] = param.Name

			io.WriteString(ctx.out, "\r\n\tvar "+param.Name+" "+param.Type().ToLiteral())
			err := method.renderStructParam(ctx,
				&Param{Param: param, option: st, index: idx}, nil)
			if err != nil {
				return err
			}

			continue
		}

		return errors.New("param '" + param.Name +
			"' of '" + method.FullName() +
			"' not found in the swagger annotations")
	}

	/// 输出 body 参数的初始化
	if len(inBody) > 0 {
		err := method.renderBodyParams(ctx, inBody)
		if err != nil {
			return err
		}
	}

	return method.renderInvokeAndReturn(ctx)
}

func (method *Method) HasQueryParam() bool {
	for idx := range method.Method.Params.List {
		param := &method.Method.Params.List[idx]
		if param.Type().IsContextType() {
			continue
		}

		var option *spec.Parameter
		foundIndex := searchParam(method.Operation, param.Name)
		if foundIndex >= 0 {
			if method.Operation.Parameters[foundIndex].In == "query" {
				return true
			}
			continue
		}
		if option == nil {
			option = searchStructParam(method.Operation, param.Name)
			if option == nil {
				continue
			}
			if option.In == "query" {
				return true
			}
			continue
		}

		switch param.Type().ToLiteral() {
		case "map[string]string":
			return true
		case "url.Values":
			return true
		}
	}
	return false
}

func (method *Method) getSiblingParamNames(expected []string) ([]SiblingName, error) {
	var names []SiblingName

	for idx := range method.Method.Params.List {
		param := &method.Method.Params.List[idx]
		if param.Type().IsContextType() {
			continue
		}

		var option *spec.Parameter
		foundIndex := searchParam(method.Operation, param.Name)
		if foundIndex >= 0 {
			option = &method.Operation.Parameters[foundIndex]
		}
		if option == nil {
			option = searchStructParam(method.Operation, param.Name)
			if option == nil {
				continue
			}
		}
		if isExtendInline(option) {
			continue
		}

		for _, name := range expected {
			if name == option.In {
				isPrefix := isPrefixForType(param.Type())

				names = append(names, SiblingName{
					Name:     option.Name,
					IsPrefix: isPrefix,
				})
				break
			}
		}
	}

	return names, nil
}

func isPrefixForType(typ astutil.Type) bool {
	if t := typ; t.IsStructType() &&
		!isExceptedType(t.ToLiteral(), bultinTypes) &&
		!t.IsSqlNullableType() {
		return true
	} else if t = typ.PtrElemType(); t.IsValid() &&
		t.IsStructType() &&
		!isExceptedType(t.ToLiteral(), bultinTypes) &&
		!t.IsSqlNullableType() {
		return true
	} else if s := typ.ToLiteral(); s == "map[string]string" ||
		s == "url.Values" {
		return true
	}
	return false
}

func getFieldSiblingNames(typ astutil.Type) ([]SiblingName, error) {
	if t := typ.PtrElemType(); t.IsValid() {
		typ = t
	}
	ts, err := typ.ToTypeSpec(true)
	if err != nil {
		return nil, errors.New("cannot convert '" + typ.ToLiteral() + "' to type spec: " + err.Error())
	}
	if ts.Struct == nil {
		return nil, errors.New("type '" + typ.ToLiteral() + "' isnot struct")
	}

	var names []SiblingName

	var fields = ts.Fields()
	for _, f := range ts.Struct.Embedded {
		fields = append(fields, f)
	}
	for idx := range fields {
		var s, _ = getTagValue(&fields[idx], "swaggerignore")
		if strings.ToLower(s) == "true" {
			continue
		}
		isPrefix := isPrefixForType(fields[idx].Type())

		s, _ = getTagValue(&fields[idx], "json")
		s = getJSONName(s)
		if s == "" {
			if fields[idx].IsAnonymous {
				t := fields[idx].Type()
				if a := t.PtrElemType(); a.IsValid() {
					t = a
				}

				if t.IsStructType() &&
					!isExceptedType(t.ToLiteral(), bultinTypes) &&
					!t.IsSqlNullableType() {
					results, err := getFieldSiblingNames(fields[idx].Type())
					if err != nil {
						return nil, err
					}

					for _, result := range results {
						found := false
						for idx := range names {
							if names[idx].Name == result.Name {
								found = true
								break
							}
						}
						if !found {
							names = append(names, result)
						}
					}
				}
				continue
			}

			s = toSnakeCase(fields[idx].Name)
		}
		if isPrefix {
			s = s + "."
		}

		found := false
		for idx := range names {
			if names[idx].Name == s {
				found = true
				break
			}
		}
		if !found {
			names = append(names, SiblingName{
				Name:     s,
				IsPrefix: isPrefix,
			})
		}
	}
	return names, nil
}

func (method *Method) renderStructParam(ctx *GenContext, param *Param, parents []*Field) error {
	var typ astutil.Type
	if len(parents) == 0 {
		typ = param.Type()
	} else {
		typ = parents[len(parents)-1].Type()
	}

	if typ.IsPtrType() {
		typ = typ.PtrElemType()
	}

	ts, err := typ.ToTypeSpec(true)
	if err != nil {
		return errors.New("param '" + GetGoVarName(param, parents, true) + "' of '" +
			method.FullName() +
			"' cannot convert to type spec: " + err.Error())
	}

	if ts.Struct == nil {
		return errors.New("param '" + GetGoVarName(param, parents, true) + "' of '" +
			method.FullName() +
			"' cannot convert to struct")
	}

	var fields = ts.Fields()
	for _, f := range ts.Struct.Embedded {
		fields = append(fields, f)
	}
	for idx := range fields {
		var s, _ = getTagValue(&fields[idx], "swaggerignore")
		if strings.ToLower(s) == "true" {
			continue
		}

		fieldType := fields[idx].Type()
		isPtrType := false
		if t := fieldType.PtrElemType(); t.IsValid() {
			isPtrType = true
			fieldType = t
		}
		isNullableType := fieldType.IsSqlNullableType()

		switch fieldType.ToLiteral() {
		case "map[string]string":
			io.WriteString(ctx.out, "\r\n")
			err := method.renderMapParam(ctx, param,
				append(parents, &Field{
					Field:        &fields[idx],
					isFirstField: idx == 0,
				}), "map[string]string", "[len(values)-1]")
			if err != nil {
				return err
			}
			continue
		case "url.Values":
			io.WriteString(ctx.out, "\r\n")
			err := method.renderMapParam(ctx, param,
				append(parents, &Field{
					Field:        &fields[idx],
					isFirstField: idx == 0,
				}), "url.Values", "")
			if err != nil {
				return err
			}
			continue
		}

		if fieldType.IsStructType() &&
			!isNullableType &&
			!isExceptedType(fieldType.ToLiteral(), bultinTypes) {
			err = method.renderStructParam(ctx, param,
				append(parents, &Field{
					Field:        &fields[idx],
					isFirstField: idx == 0,
				}))
			if err != nil {
				return err
			}
			continue
		}

		optidx := searchStructFieldParam(method.Operation, GetGoVarName(param, parents, true), &fields[idx])
		if optidx < 0 {
			return errors.New("param '" + GetGoVarName(param, parents, true) + "." + fields[idx].Name +
				"' of '" + method.FullName() +
				"' not found in the swagger1 annotations")
		}

		if isNullableType {
			if err := method.renderNullableParam(ctx, param,
				append(parents, &Field{
					Field:        &fields[idx],
					isFirstField: idx == 0,
				})); err != nil {
				return err
			}
			continue
		}

		if isPtrType {
			if err := method.renderPtrTypeParam(ctx, param,
				append(parents, &Field{
					Field:        &fields[idx],
					isFirstField: idx == 0,
				})); err != nil {
				return err
			}
			continue
		}

		if err := method.renderPrimitiveTypeParam(ctx, param,
			append(parents, &Field{
				Field:        &fields[idx],
				isFirstField: idx == 0,
			})); err != nil {
			return err
		}
	}
	return nil
}

func (method *Method) renderSimpleParam(ctx *GenContext, param *Param) error {
	typ := param.Type()
	isPtrType := false
	if t := typ.PtrElemType(); t.IsValid() {
		isPtrType = true
		typ = t
	}
	isNullableType := typ.IsSqlNullableType()

	if isNullableType {
		if err := method.renderNullableParam(ctx, param, nil); err != nil {
			return err
		}

		return nil
	}

	if isPtrType {
		if err := method.renderPtrTypeParam(ctx, param, nil); err != nil {
			return err
		}
		return nil
	}

	return method.renderPrimitiveTypeParam(ctx, param, nil)
}

func selectFunction(plugin Plugin, required, isArray bool, typeStr string) *Function {
	functions := plugin.Functions()
	for idx := range functions {
		if required != functions[idx].Required {
			continue
		}
		if isArray != functions[idx].IsArray {
			continue
		}
		if typeStr == functions[idx].ResultType {
			return &functions[idx]
		}
	}
	return nil
}

func (method *Method) renderPrimitiveTypeParam(ctx *GenContext, param *Param, fields []*Field) error {
	io.WriteString(ctx.out, "\r\n")

	required := false
	isArray := false
	var typ astutil.Type
	if len(fields) == 0 {
		typ = param.Type()
		required = param.option.In == "path"
	} else {
		typ = fields[len(fields)-1].Type()
	}

	goVarName := GetGoVarName(param, fields)

	if len(fields) == 0 {
		if param.IsVariadic {
			isArray = true

			if  typ.IsSliceType() {
				return errors.New("param '" + goVarName + "' of '" +
						method.FullName() +
						"' is unsupported type - '..."+typ.ToLiteral()+"'")
			}

			typ = astutil.Type{
				File: typ.File,
				Expr: &ast.ArrayType{
					Lbrack: param.Expr.Pos(),
					Elt:    typ.Expr,
				},
			}
		}
	}

	if !isArray {
		isArray = typ.IsSliceType()
	} 



	var elmType, underlying, elmUnderlying astutil.Type
	if isArray {
		elmType = typ.SliceElemType()
		elmUnderlying = elmType.GetUnderlyingType()

		if elmUnderlying.IsValid() {
			arrayExpr := typ.Expr.(*ast.ArrayType)
			underlying = astutil.Type{
				File: typ.File,
				Expr: &ast.ArrayType{
					Lbrack: arrayExpr.Lbrack,
					Len:    arrayExpr.Len,
					Elt:    elmUnderlying.Expr,
				},
			}
		}
	} else {
		underlying = typ.GetUnderlyingType()
	}


	var fn *Function
	if isArray {
		if elmUnderlying.IsValid() {
			fn = selectFunction(ctx.plugin, required, isArray, elmUnderlying.ToLiteral())
		} else if elmType.IsValid() {
			fn = selectFunction(ctx.plugin, required, isArray, elmType.ToLiteral())
		}
	} else {
		if underlying.IsValid() {
			fn = selectFunction(ctx.plugin, required, isArray, underlying.ToLiteral())
		} else {
			fn = selectFunction(ctx.plugin, required, isArray, typ.ToLiteral())
		}
	}
	if fn != nil {
		webParamName := GetWebParamName(param, fields)
		var valueReadText string
		if fn.WithDefault {
			valueReadText = fmt.Sprintf(fn.Format, webParamName,
				defaultValue(fn.ResultType, param, fields))
		} else {
			valueReadText = fmt.Sprintf(fn.Format, webParamName)
		}

		// 情况4, 6
		if fn.ResultError || fn.ResultBool {
			io.WriteString(ctx.out, goVarName)
			if fn.ResultBool {
				io.WriteString(ctx.out, ", ok := ")
				io.WriteString(ctx.out, valueReadText)
				io.WriteString(ctx.out, "\r\n\tif !ok {\r\n")
				renderCastError(ctx, method, webParamName, "nil", "\"\"")
				io.WriteString(ctx.out, "\r\n\t}")
			} else {
				io.WriteString(ctx.out, ", err := ")
				io.WriteString(ctx.out, valueReadText)
				io.WriteString(ctx.out, "\r\n\tif err != nil {\r\n")
				renderCastError(ctx, method, webParamName, "err", "\"\"")
				io.WriteString(ctx.out, "\r\n\t}")

				method.SetErrorDeclared()
			}

			if underlying.IsValid() {
				if len(fields) == 0 {
					method.goArgumentLiterals[param.index] = typ.ToLiteral() + "(" + goVarName + ")"
				}
			}
			return nil
		}

		var isOptional = false

		// 情况1, 2
		if len(fields) == 0 {
			io.WriteString(ctx.out, "\tvar ")
		} else {
			if !fn.Required && needParentInitialize(param, fields) {
				if fn.IsArray {
					io.WriteString(ctx.out, "\tif ss := "+valueReadText+"; len(ss) != 0 {")
				} else {
					io.WriteString(ctx.out, "\tif s := "+valueReadText+"; s != \"\" {")
				}
				if err := renderParentInit(ctx, param, fields, true); err != nil {
					return err
				}
				isOptional = true
			} else {
				if err := renderParentInit(ctx, param, fields, true); err != nil {
					return err
				}
				setParentInitialized(param, fields)
			}
		}
		io.WriteString(ctx.out, goVarName+" = ")

		if underlying.IsValid() {
			io.WriteString(ctx.out, typ.ToLiteral())
			io.WriteString(ctx.out, "(")
		}

		io.WriteString(ctx.out, valueReadText)

		if underlying.IsValid() {
			io.WriteString(ctx.out, ")")
		}

		if isOptional {
			io.WriteString(ctx.out, "}")
		}
		return nil
	}

	fn = selectFunction(ctx.plugin, required, isArray, "string")
	if fn == nil {
		return errors.New("param '" + goVarName + "' of '" +
			method.FullName() +
			"' cannot determine a function")
	}

	webParamName := GetWebParamName(param, fields)
	var valueReadText string
	if fn.WithDefault {
		valueReadText = fmt.Sprintf(fn.Format, webParamName,
			defaultValue(fn.ResultType, param, fields))
	} else {
		valueReadText = fmt.Sprintf(fn.Format, webParamName)
	}

	convertFmt, needCast, retError, err := selectConvert(ctx.convertNS, fn.IsArray, fn.ResultType, typ.ToLiteral())
	if err != nil {
		originErr := err
		if underlying.IsValid() {
			convertFmt, needCast, retError, err = selectConvert(ctx.convertNS, fn.IsArray, fn.ResultType, underlying.ToLiteral())
		}
		if err != nil {
			return errors.New("param '" + goVarName + "' of '" +
				method.FullName() +
				"' hasnot convert function: " + originErr.Error())
		}
		needCast = true
	}

	if fn.Required {
		// 情况3

		if retError {
			method.SetErrorDeclared()
			if needCast {
				io.WriteString(ctx.out, fieldName(param, fields))
				io.WriteString(ctx.out, "Value, err := ")
			} else {
				io.WriteString(ctx.out, goVarName)
				io.WriteString(ctx.out, ", err := ")
			}
		} else {
			if needCast {
				io.WriteString(ctx.out, fieldName(param, fields))
				io.WriteString(ctx.out, "Value := ")
			} else {
				io.WriteString(ctx.out, goVarName)
				if len(fields) == 0 {
					io.WriteString(ctx.out, " := ")
				} else {
					io.WriteString(ctx.out, " = ")
				}
			}
		}

		io.WriteString(ctx.out, fmt.Sprintf(convertFmt, valueReadText))

		if retError {
			io.WriteString(ctx.out, "\r\n\tif err != nil {\r\n")
			renderCastError(ctx, method, webParamName, "err", valueReadText)
			io.WriteString(ctx.out, "\r\n\t}")
		}

		if needCast {
			if len(fields) > 0 {
				if err := renderParentInit(ctx, param, fields, false); err != nil {
					return err
				}
				setParentInitialized(param, fields)

				io.WriteString(ctx.out, "\r\n\t")
			} else {
				io.WriteString(ctx.out, "\r\n\tvar ")
			}

			io.WriteString(ctx.out, goVarName)
			io.WriteString(ctx.out, " = "+typ.ToLiteral()+"(")
			io.WriteString(ctx.out, fieldName(param, fields))
			io.WriteString(ctx.out, "Value)")
		}
		return nil
	}

	// 情况5
	if len(fields) == 0 {
		io.WriteString(ctx.out, "var ")
		io.WriteString(ctx.out, goVarName)
		io.WriteString(ctx.out, " ")
		io.WriteString(ctx.out, typ.ToLiteral())
		io.WriteString(ctx.out, "\r\n\t")
	}
	var tmpVarName = "s"
	if fn.IsArray {
		tmpVarName = "ss"
	}

	io.WriteString(ctx.out, "if "+tmpVarName+" := "+valueReadText)
	if fn.IsArray {
		io.WriteString(ctx.out, "; len("+tmpVarName+") != 0 {")
	} else {
		io.WriteString(ctx.out, "; "+tmpVarName+" != \"\" {")
	}

	if retError {
		io.WriteString(ctx.out, "\r\n\t\t"+fieldName(param, fields)+"Value")
		io.WriteString(ctx.out, ", err :="+fmt.Sprintf(convertFmt, tmpVarName))

		io.WriteString(ctx.out, "\r\n\t\tif err != nil {\r\n")
		renderCastError(ctx, method, webParamName, "err", tmpVarName)
		io.WriteString(ctx.out, "\r\n\t\t}")

		if err := renderParentInit(ctx, param, fields, false); err != nil {
			return err
		}
		if needCast {
			io.WriteString(ctx.out, "\r\n\t\t"+goVarName+" = "+typ.ToLiteral()+"("+fieldName(param, fields)+"Value)")
		} else {
			io.WriteString(ctx.out, "\r\n\t\t"+goVarName+" = "+fieldName(param, fields)+"Value")
		}
	} else {
		if err := renderParentInit(ctx, param, fields, false); err != nil {
			return err
		}

		if needCast {
			io.WriteString(ctx.out, "\r\n\t\t"+goVarName+" = "+typ.ToLiteral()+"("+fmt.Sprintf(convertFmt, tmpVarName)+")")
		} else {
			io.WriteString(ctx.out, "\r\n\t\t"+goVarName+" = "+fmt.Sprintf(convertFmt, tmpVarName))
		}
	}

	io.WriteString(ctx.out, "\r\n\t}")
	return nil
}

func (method *Method) renderNullableParam(ctx *GenContext, param *Param, fields []*Field) error {
	io.WriteString(ctx.out, "\r\n")

	required := false
	isArray := false
	var typ astutil.Type
	if len(fields) == 0 {
		typ = param.Type()
		required = param.option.In == "path"
	} else {
		typ = fields[len(fields)-1].Type()
	}
	if !typ.IsSqlNullableType() {
		return errors.New("type '" + typ.ToLiteral() + "' is unsupported for renderNullableParam")
	}

	goVarName := GetGoVarName(param, fields)

	if len(fields) == 0 && param.IsVariadic {
		return errors.New("param '" + goVarName + "' of '" +
			method.FullName() +
			"' is unsupported type - '..."+typ.ToLiteral()+"'")
	}


	fn := selectFunction(ctx.plugin, required, isArray, ElemTypeForNullable(typ))
	if fn != nil {
		webParamName := GetWebParamName(param, fields)
		var valueReadText string
		if fn.WithDefault {
			valueReadText = fmt.Sprintf(fn.Format, webParamName,
				defaultValue(fn.ResultType, param, fields))
		} else {
			valueReadText = fmt.Sprintf(fn.Format, webParamName)
		}

		if len(fields) == 0 {
			io.WriteString(ctx.out, "var ")
			io.WriteString(ctx.out, goVarName)
			io.WriteString(ctx.out, " ")
			io.WriteString(ctx.out, param.Type().ToLiteral())
			io.WriteString(ctx.out, "\r\n\t")
		} else {
			if err := renderParentInit(ctx, param, fields, true); err != nil {
				return err
			}
			setParentInitialized(param, fields)
		}

		if fn.ResultError || fn.ResultBool {
			if fn.ResultBool {
				io.WriteString(ctx.out, "if"+fieldName(param, fields)+"Value, ok := ")
				io.WriteString(ctx.out, valueReadText)
				io.WriteString(ctx.out, "; !ok {\r\n")
				renderCastError(ctx, method, webParamName, "nil", "\"\"")
			} else {
				io.WriteString(ctx.out, "if"+fieldName(param, fields)+"Value, err := ")
				io.WriteString(ctx.out, valueReadText)
				io.WriteString(ctx.out, "; err != nil {\r\n")
				renderCastError(ctx, method, webParamName, "err", "\"\"")
			}
			io.WriteString(ctx.out, "\r\n\t} else {")
			io.WriteString(ctx.out, "\r\n"+goVarName+".Valid = true")
			io.WriteString(ctx.out, "\r\n"+goVarName+"."+FieldNameForNullable(typ)+" = "+fieldName(param, fields)+"Value")
			io.WriteString(ctx.out, "\r\n\t}")
			return nil
		}

		if fn.Required {
			io.WriteString(ctx.out, goVarName+".Valid = true")
			io.WriteString(ctx.out, "\r\n"+goVarName+"."+FieldNameForNullable(typ)+" = ")
			io.WriteString(ctx.out, valueReadText)
		} else {
			io.WriteString(ctx.out, "if s := "+valueReadText+"; s != \"\" {")
			io.WriteString(ctx.out, "\r\n\t\t"+goVarName+".Valid = true")
			io.WriteString(ctx.out, "\r\n\t\t"+goVarName+"."+FieldNameForNullable(typ)+" = s")
			io.WriteString(ctx.out, "\r\n\t}")
		}

		return nil
	}

	fn = selectFunction(ctx.plugin, required, isArray, "string")
	if fn == nil {
		return errors.New("param '" + goVarName + "' of '" +
			method.FullName() +
			"' cannot determine a function")
	}

	webParamName := GetWebParamName(param, fields)
	var valueReadText string
	if fn.WithDefault {
		valueReadText = fmt.Sprintf(fn.Format, webParamName,
			defaultValue(fn.ResultType, param, fields))
	} else {
		valueReadText = fmt.Sprintf(fn.Format, webParamName)
	}

	convertFmt, needCast, retError, err := selectConvert(ctx.convertNS, fn.IsArray, fn.ResultType, ElemTypeForNullable(typ))
	if err != nil {
		return errors.New("param '" + goVarName + "' of '" +
			method.FullName() +
			"' hasnot convert function: " + err.Error())
	}

	// 情况7，8
	if len(fields) == 0 {
		io.WriteString(ctx.out, "var ")
		io.WriteString(ctx.out, goVarName)
		io.WriteString(ctx.out, " ")
		io.WriteString(ctx.out, param.Type().ToLiteral())
		io.WriteString(ctx.out, "\r\n\t")
	}

	io.WriteString(ctx.out, "if s := ")
	io.WriteString(ctx.out, valueReadText)
	io.WriteString(ctx.out, "; s != \"\" {")

	if retError {
		io.WriteString(ctx.out, "\r\n\t\t"+fieldName(param, fields)+"Value")
		if retError {
			io.WriteString(ctx.out, ", err ")
		}
		io.WriteString(ctx.out, " :=")
		io.WriteString(ctx.out, fmt.Sprintf(convertFmt, "s"))

		if retError {
			io.WriteString(ctx.out, "\r\n\t\tif err != nil {\r\n")
			renderCastError(ctx, method, webParamName, "err", "s")
			io.WriteString(ctx.out, "\r\n\t\t}")
		}

		if err := renderParentInit(ctx, param, fields, false); err != nil {
			return err
		}
		io.WriteString(ctx.out, "\r\n\t\t"+goVarName+".Valid = true")
		if needCast {
			io.WriteString(ctx.out, "\r\n\t\t"+goVarName+"."+FieldNameForNullable(typ)+" = "+ElemTypeForNullable(typ)+"("+fieldName(param, fields)+"Value)")
		} else {
			io.WriteString(ctx.out, "\r\n\t\t"+goVarName+"."+FieldNameForNullable(typ)+" = "+fieldName(param, fields)+"Value")
		}
	} else {
		if err := renderParentInit(ctx, param, fields, false); err != nil {
			return err
		}
		io.WriteString(ctx.out, "\r\n\t\t"+goVarName+".Valid = true")
		if needCast {
			io.WriteString(ctx.out, "\r\n\t\t"+goVarName+"."+FieldNameForNullable(typ)+" = "+ElemTypeForNullable(typ)+"("+fmt.Sprintf(convertFmt, "s")+")")
		} else {
			io.WriteString(ctx.out, "\r\n\t\t"+goVarName+"."+FieldNameForNullable(typ)+" = "+fmt.Sprintf(convertFmt, "s"))
		}
	}

	io.WriteString(ctx.out, "\r\n\t}")
	return nil
}

func (method *Method) renderPtrTypeParam(ctx *GenContext, param *Param, fields []*Field) error {
	io.WriteString(ctx.out, "\r\n")

	required := false
	isArray := false
	var typ astutil.Type
	if len(fields) == 0 {
		typ = param.Type()
		required = param.option.In == "path"
	} else {
		typ = fields[len(fields)-1].Type()
	}
	if !typ.IsPtrType() {
		return errors.New("type '" + typ.ToLiteral() + "' is unsupported for renderPtrTypeParam")
	}

	goVarName := GetGoVarName(param, fields)

	if len(fields) == 0 && param.IsVariadic {
		return errors.New("param '" + goVarName + "' of '" +
			method.FullName() +
			"' is unsupported type - '..."+typ.ToLiteral()+"'")
	}

	typ = typ.PtrElemType()

	if typ.IsSliceType() {
		return errors.New("type '" + typ.ToLiteral() + "' is unsupported for renderPtrTypeParam")
		// isArray = true
		// typ = typ.SliceElemType()
	}

	underlying := typ.GetUnderlyingType()

	// elemTypeStr := typ.ToLiteral()
	var fn *Function
	if underlying.IsValid() {
		fn = selectFunction(ctx.plugin, required, isArray, underlying.ToLiteral())
	} else {
		fn = selectFunction(ctx.plugin, required, isArray, typ.ToLiteral())
	}
	if fn != nil {
		webParamName := GetWebParamName(param, fields)
		var valueReadText string
		if fn.WithDefault {
			valueReadText = fmt.Sprintf(fn.Format, webParamName,
				defaultValue(fn.ResultType, param, fields))
		} else {
			valueReadText = fmt.Sprintf(fn.Format, webParamName)
		}

		if len(fields) == 0 {
			io.WriteString(ctx.out, "var ")
			io.WriteString(ctx.out, goVarName)
			io.WriteString(ctx.out, " ")

			if fn.Required {
				// io.WriteString(ctx.out, typ.ToLiteral())
				io.WriteString(ctx.out, " = ")
				io.WriteString(ctx.out, valueReadText)

				method.goArgumentLiterals[param.index] = "&" + goVarName
				return nil
			}
			io.WriteString(ctx.out, param.Type().ToLiteral())
			io.WriteString(ctx.out, "\r\n\t")
		}

		io.WriteString(ctx.out, "if s := ")
		io.WriteString(ctx.out, valueReadText)
		io.WriteString(ctx.out, "; s != \"\" {")
		if err := renderParentInit(ctx, param, fields, false); err != nil {
			return err
		}

		// if isNullableType(typeStr) {
		// 	io.WriteString(ctx.out, "\r\n\t"+goVarName+".Valid = true")
		// 	io.WriteString(ctx.out, "\r\n\t"+goVarName+"."+FieldNameForNullable(param.Type())+" = s")
		// } else {
		if underlying.IsValid() {
			io.WriteString(ctx.out, "\r\n\t"+goVarName+" = new(")
			io.WriteString(ctx.out, typ.ToLiteral())
			io.WriteString(ctx.out, ")")
			io.WriteString(ctx.out, "\r\n\t*"+goVarName+" = "+typ.ToLiteral()+"(s)")
		} else {
			io.WriteString(ctx.out, "\r\n\t"+goVarName+" = &s")
		}
		//}
		io.WriteString(ctx.out, "\r\n\t}")
		return nil
	}

	fn = selectFunction(ctx.plugin, required, isArray, "string")
	if fn == nil {
		return errors.New("param '" + goVarName + "' of '" +
			method.FullName() +
			"' cannot determine a function")
	}

	webParamName := GetWebParamName(param, fields)
	var valueReadText string
	if fn.WithDefault {
		valueReadText = fmt.Sprintf(fn.Format, webParamName,
			defaultValue(fn.ResultType, param, fields))
	} else {
		valueReadText = fmt.Sprintf(fn.Format, webParamName)
	}

	convertFmt, needCast, retError, err := selectConvert(ctx.convertNS, fn.IsArray, fn.ResultType, typ.ToLiteral())
	if err != nil {
		originErr := err
		if underlying.IsValid() {
			convertFmt, needCast, retError, err = selectConvert(ctx.convertNS, fn.IsArray, fn.ResultType, underlying.ToLiteral())
		}
		if err != nil {
			return errors.New("param '" + goVarName + "' of '" +
				method.FullName() +
				"' hasnot convert function: " + originErr.Error())
		}
		needCast = true
	}

	if fn.Required {
		// 情况13

		if retError {

			if len(fields) == 0 && !needCast {
				method.SetErrorDeclared()

				io.WriteString(ctx.out, goVarName)
				io.WriteString(ctx.out, ", err := ")
				io.WriteString(ctx.out, fmt.Sprintf(convertFmt, valueReadText))

				io.WriteString(ctx.out, "\r\n\t if err != nil {\r\n")
				renderCastError(ctx, method, webParamName, "err", valueReadText)
				io.WriteString(ctx.out, "\r\n\t}")

				method.goArgumentLiterals[param.index] = "&" + goVarName

			} else {

				if len(fields) == 0 {
					io.WriteString(ctx.out, "var ")
					io.WriteString(ctx.out, goVarName)
					io.WriteString(ctx.out, " ")
					io.WriteString(ctx.out, param.Type().ToLiteral())
					io.WriteString(ctx.out, "\r\n\t")
				}

				io.WriteString(ctx.out, "if ")
				io.WriteString(ctx.out, fieldName(param, fields))
				io.WriteString(ctx.out, "Value, err := ")

				io.WriteString(ctx.out, fmt.Sprintf(convertFmt, valueReadText))
				io.WriteString(ctx.out, "; err != nil {\r\n")
				renderCastError(ctx, method, webParamName, "err", valueReadText)
				io.WriteString(ctx.out, "\r\n\t} else {")

				if err := renderParentInit(ctx, param, fields, false); err != nil {
					return err
				}
				io.WriteString(ctx.out, "\r\n\t\t"+goVarName)

				if needCast {
					io.WriteString(ctx.out, " = new("+typ.ToLiteral()+")")
					io.WriteString(ctx.out, "\r\n\t\t*")
					io.WriteString(ctx.out, goVarName)
					io.WriteString(ctx.out, " = ")
					io.WriteString(ctx.out, fieldName(param, fields))
					io.WriteString(ctx.out, "Value")
				} else {
					io.WriteString(ctx.out, " = &")
					io.WriteString(ctx.out, fieldName(param, fields))
					io.WriteString(ctx.out, "Value")
				}
				io.WriteString(ctx.out, "\r\n\t}")
			}
		} else {
			if err := renderParentInit(ctx, param, fields, false); err != nil {
				return err
			}
			io.WriteString(ctx.out, goVarName)
			io.WriteString(ctx.out, " := ")
			if needCast {
				io.WriteString(ctx.out, typ.ToLiteral()+"("+fmt.Sprintf(convertFmt, valueReadText)+")")
			} else {
				io.WriteString(ctx.out, fmt.Sprintf(convertFmt, valueReadText))
			}
		}
	} else {
		if len(fields) == 0 {
			io.WriteString(ctx.out, "var ")
			io.WriteString(ctx.out, goVarName)
			io.WriteString(ctx.out, " ")
			io.WriteString(ctx.out, param.Type().ToLiteral())
			io.WriteString(ctx.out, "\r\n\t")
		}

		// 情况14
		io.WriteString(ctx.out, "if s := ")
		io.WriteString(ctx.out, valueReadText)
		io.WriteString(ctx.out, "; s != \"\" {")

		if retError {
			io.WriteString(ctx.out, "\r\n\t\t"+fieldName(param, fields)+"Value")
			io.WriteString(ctx.out, ", err :="+fmt.Sprintf(convertFmt, "s"))
			io.WriteString(ctx.out, "\r\n\t\tif err != nil {\r\n")
			renderCastError(ctx, method, webParamName, "err", "s")
			io.WriteString(ctx.out, "\r\n\t\t}")

			if err := renderParentInit(ctx, param, fields, false); err != nil {
				return err
			}
			if needCast {
				io.WriteString(ctx.out, "\r\n\t\t"+goVarName+" = new("+typ.ToLiteral()+")")
				io.WriteString(ctx.out, "\r\n\t\t*"+goVarName+" = "+typ.ToLiteral()+"("+fieldName(param, fields)+"Value)")
			} else {
				io.WriteString(ctx.out, "\r\n\t\t"+goVarName+" = &"+fieldName(param, fields)+"Value")
			}
		} else {
			if err := renderParentInit(ctx, param, fields, false); err != nil {
				return err
			}
			if needCast {
				io.WriteString(ctx.out, "\r\n\t\t*"+goVarName+"Value := "+typ.ToLiteral()+"("+fmt.Sprintf(convertFmt, "s")+")")
			} else {
				io.WriteString(ctx.out, "\r\n\t\t"+goVarName+"Value := "+fmt.Sprintf(convertFmt, "s"))
			}
			io.WriteString(ctx.out, "\r\n\t\t"+goVarName+" = &"+goVarName+"Value")
		}
		io.WriteString(ctx.out, "\r\n\t}")
	}
	return nil
}

func fieldName(param *Param, parents []*Field) string {
	if len(parents) > 0 {
		return toLowerCamelCase(parents[len(parents)-1].Name)
	}
	return toLowerCamelCase(param.Name)
}

func GetGoVarName(param *Param, parents []*Field, hideAnonymous ...bool) string {
	name := param.Name
	for idx := range parents {
		if parents[idx].IsAnonymous {
			if len(hideAnonymous) > 0 && hideAnonymous[0] {
				continue
			}
		}

		name = name + "." + parents[idx].Name
	}
	return name
}

func GetWebParamName(param *Param, parents []*Field) string {
	var name string 
	if param.option != nil {
		if !isExtendInline(param.option) {
			name = param.option.Name
		}
	} else {
		name = toSnakeCase(param.Name)
	}
	for idx := range parents {
		jsonName, _ := getTagValue(parents[idx].Field, "json")
		jsonName = getJSONName(jsonName)
		if parents[idx].Field.IsAnonymous {
			if jsonName == "" {
				continue
			}
		} else if jsonName == "" {
			jsonName = toSnakeCase(parents[idx].Field.Name)
		}
		if name != "" {
			name = name + "."
		}
		name = name + jsonName
	}
	return name
}

func needParentInitialize(param *Param, parents []*Field) bool {
	if len(parents) == 0 {
		return false
	}

	if len(parents) == 1 {
		return param.Type().IsPtrType()
	}

	for idx := len(parents) - 1; idx >= 0; idx-- {
		if parents[idx].Type().IsPtrType() {
			return true
		}
	}
	return param.Type().IsPtrType()
}

func setParentInitialized(param *Param, parents []*Field) {
	for i := range parents {
		parents[i].isInitialized = true
	}
	param.isInitialized = true
}

func renderParentInit(ctx *GenContext, param *Param, parents []*Field, noCRCF bool, isFirst ...bool) error {
	if len(parents) == 0 {
		return nil
	}

	isFirstValue := true
	if len(isFirst) > 0 {
		isFirstValue = isFirst[0]
	}
	if isFirstValue {
		for idx := len(parents) - 1; idx >= 0; idx-- {
			if !parents[idx].isFirstField {
				isFirstValue = false
				break
			}
		}
	}

	if param.Type().IsPtrType() {
		if !param.isInitialized {
			if !noCRCF {
				io.WriteString(ctx.out, "\r\n\t")
			}
			goVarName := GetGoVarName(param, nil, false)
			if isFirstValue {
				io.WriteString(ctx.out, goVarName+" = new("+
					param.Type().PtrElemType().ToLiteral()+")")
			} else {
				io.WriteString(ctx.out, "if "+goVarName+" == nil {")
				io.WriteString(ctx.out, "\r\n\t\t"+goVarName+" = new("+
					param.Type().PtrElemType().ToLiteral()+")")
				io.WriteString(ctx.out, "}")
			}
			if noCRCF {
				io.WriteString(ctx.out, "\r\n")
			}
		}
	}

	for idx := 0; idx < len(parents)-1; idx++ {
		isFirstValue := true
		if len(isFirst) > 0 {
			isFirstValue = isFirst[0]
		}
		if isFirstValue {
			for i := len(parents) - 1; i > idx; i-- {
				if !parents[i].isFirstField {
					isFirstValue = false
					break
				}
			}
		}

		if !parents[idx].Type().IsPtrType() {
			continue
		}

		if parents[idx].isInitialized {
			continue
		}

		if !noCRCF {
			io.WriteString(ctx.out, "\r\n\t")
		}

		goVarName := GetGoVarName(param, parents[:idx+1], false)

		if isFirstValue {
			io.WriteString(ctx.out, goVarName+" = new("+
				parents[idx].Type().PtrElemType().ToLiteral()+")")
		} else {
			io.WriteString(ctx.out, "if "+goVarName+" == nil {")
			io.WriteString(ctx.out, "\r\n\t\t"+goVarName+" = new("+
				parents[idx].Type().PtrElemType().ToLiteral()+")")
			io.WriteString(ctx.out, "}")
		}

		if noCRCF {
			io.WriteString(ctx.out, "\r\n")
		}
	}
	return nil
}

type SiblingName struct {
	Name     string
	IsPrefix bool
}

func (method *Method) renderMapParamWithAnonymous(ctx *GenContext, param *Param, parents []*Field, siblingNames []SiblingName, typeStr, valueIndex string) error {
	var goVarName = GetGoVarName(param, parents, true)

	if len(parents) > 0 {
		// 当前字段为匿名字段，GetGoVarName() 函数会不显示它，我们
		// 要访问它，所以这个要加上
		goVarName = goVarName + "." + parents[len(parents)-1].Name
	}

	if len(parents) == 0 {
		io.WriteString(ctx.out, "var "+goVarName+" = "+typeStr+"{}")
	}

	var parentPrefix = GetWebParamName(param, parents)
	if parentPrefix != "" {
		parentPrefix = parentPrefix + "."
	}

	var sb strings.Builder
	for idx, sname := range siblingNames {
		if idx > 0 {
			sb.WriteString(" ||\r\n\t\t\t")
		}
		if sname.IsPrefix {
			sb.WriteString("strings.HasPrefix(key, \"" + parentPrefix + sname.Name + "\")")
		} else {
			sb.WriteString("key == \"" + parentPrefix + sname.Name + "\"")
		}
	}

	io.WriteString(ctx.out, "\r\n\tfor key, values := range ")
	values, _ := ctx.plugin.GetSpecificTypeArgument("url.Values")
	io.WriteString(ctx.out, values+"{")
	if parentPrefix != "" {
		io.WriteString(ctx.out, "\r\n\t\tif !strings.HasPrefix(key, \""+parentPrefix+"\") {")
		io.WriteString(ctx.out, "\r\n\t\t\tcontinue")
		io.WriteString(ctx.out, "\r\n\t\t}")
	}
	if sb.Len() > 0 {
		io.WriteString(ctx.out, "\r\n\t\tif "+sb.String()+"{")
		io.WriteString(ctx.out, "\r\n\t\t\tcontinue")
		io.WriteString(ctx.out, "\r\n\t\t}")
	}

	if err := renderParentInit(ctx, param, parents, false); err != nil {
		return err
	}

	if len(parents) > 0 {
		io.WriteString(ctx.out, "\r\n\t\tif "+goVarName+" == nil {")
		io.WriteString(ctx.out, "\r\n\t\t\t"+goVarName+" = "+typeStr+"{}")
		io.WriteString(ctx.out, "\r\n\t\t}")
	}

	if parentPrefix == "" {
		io.WriteString(ctx.out, "\r\n\t\t"+goVarName+"[key] = values"+valueIndex)
	} else {
		io.WriteString(ctx.out, "\r\n\t\t"+goVarName+"[strings.TrimPrefix(key, \""+parentPrefix+"\")] = values"+valueIndex)
	}
	io.WriteString(ctx.out, "\r\n\t}")
	return nil
}

func (method *Method) renderMapParam(ctx *GenContext, param *Param, parents []*Field, typeStr, valueIndex string) error {
	var goVarName = GetGoVarName(param, parents, true)

	var tagName string
	if len(parents) == 0 {
		tagName = toLowerCamelCase(param.Name)
		if param.option != nil {
			if isExtendInline(param.option) {
				siblingNames, err := method.getSiblingParamNames([]string{"query"})
				if err != nil {
					return err
				}
				return method.renderMapParamWithAnonymous(ctx, param, parents, siblingNames, typeStr, valueIndex)
			}

			tagName = param.option.Name
		}
		io.WriteString(ctx.out, "var "+goVarName+" = "+typeStr+"{}")
	} else {
		if parents[len(parents)-1].IsAnonymous {
			offset := len(parents) - 1
			for offset >= 0 {
				if !parents[offset].IsAnonymous {
					break
				}
				offset--
			}

			var fieldType astutil.Type
			if offset >= 0 {
				fieldType = parents[offset].Type()
			} else {
				fieldType = param.Type()
			}
			siblingNames, err := getFieldSiblingNames(fieldType)
			if err != nil {
				return err
			}

			if offset < 0 && isExtendInline(param.option) {
				names, err := method.getSiblingParamNames([]string{"query"})
				if err != nil {
					return err
				}
				siblingNames = append(siblingNames, names...)
			}
			return method.renderMapParamWithAnonymous(ctx, param, parents, siblingNames, typeStr, valueIndex)
		}

		tagName = GetWebParamName(param, parents)
	}

	io.WriteString(ctx.out, "\r\n\tfor key, values := range ")
	values, _ := ctx.plugin.GetSpecificTypeArgument("url.Values")
	io.WriteString(ctx.out, values+"{")
	io.WriteString(ctx.out, "\r\n\t\tif !strings.HasPrefix(key, \""+tagName+".\") {")
	io.WriteString(ctx.out, "\r\n\t\t\tcontinue")
	io.WriteString(ctx.out, "\r\n\t\t}")

	if err := renderParentInit(ctx, param, parents, false); err != nil {
		return err
	}
	if len(parents) > 0 {
		io.WriteString(ctx.out, "\r\n\t\tif "+goVarName+" == nil {")
		io.WriteString(ctx.out, "\r\n\t\t\t"+goVarName+" = "+typeStr+"{}")
		io.WriteString(ctx.out, "\r\n\t\t}")
	}

	if tagName == "" {
		io.WriteString(ctx.out, "\r\n\t\t"+goVarName+"[key] = values"+valueIndex)
	} else {
		io.WriteString(ctx.out, "\r\n\t\t"+goVarName+"[strings.TrimPrefix(key, \""+tagName+".\")] = values"+valueIndex)
	}

	io.WriteString(ctx.out, "\r\n\t}")
	return nil
}

func (method *Method) renderBodyParams(ctx *GenContext, params []BodyParam) error {
	varName := params[0].Param.Name
	if len(params) == 1 && isExtendEntire(params[0].Option) {
		isString := params[0].Param.Type().IsStringType(false)
		isStringPtr := params[0].Param.Type().PtrElemType().IsValid() &&
			params[0].Param.Type().PtrElemType().IsStringType(false)

		if isString || isStringPtr {
			io.WriteString(ctx.out, "\r\n\tvar ")
			io.WriteString(ctx.out, varName)
			if isStringPtr {
				io.WriteString(ctx.out, "Builder")
				varName = varName + "Builder"
			}
			io.WriteString(ctx.out, " strings.Builder")

			s, _ := ctx.plugin.GetSpecificTypeArgument("io.Reader")

			io.WriteString(ctx.out, "\r\n\tif _, err := io.Copy(&"+varName+", "+s+"); err != nil {\r\n\t\t")
			txt := ctx.plugin.GetBodyErrorText(method, varName, "err")
			ctx.plugin.RenderReturnError(ctx.out, method, "http.StatusBadRequest", txt)
			io.WriteString(ctx.out, "\r\n}")

			if isStringPtr {
				io.WriteString(ctx.out, "\r\n\tvar ")
				io.WriteString(ctx.out, params[0].Param.Name)
				io.WriteString(ctx.out, " = ")
				io.WriteString(ctx.out, varName)
				io.WriteString(ctx.out, ".String()")

				method.goArgumentLiterals[params[0].Index] = "&" + params[0].Param.Name
			} else {
				method.goArgumentLiterals[params[0].Index] = varName + ".String()"
			}

			return nil
		} else {
			if params[0].Param.Type().PtrElemType().IsValid() {
				io.WriteString(ctx.out, "\r\n\tvar "+varName+" "+params[0].Param.Type().PtrElemType().ToLiteral())
				method.goArgumentLiterals[params[0].Index] = "&" + varName
			} else {
				io.WriteString(ctx.out, "\r\n\tvar "+varName+" "+params[0].Param.Type().ToLiteral())
				method.goArgumentLiterals[params[0].Index] = varName
			}
		}
	} else {
		varName = "bindArgs"
		io.WriteString(ctx.out, "\r\n\tvar bindArgs struct{")
		for idx := range params {
			fieldName := toUpperFirst(params[idx].Param.Name)
			io.WriteString(ctx.out, "\r\n\t\t")
			io.WriteString(ctx.out, fieldName)
			io.WriteString(ctx.out, "\t")
			io.WriteString(ctx.out, params[idx].Param.Type().ToLiteral())
			io.WriteString(ctx.out, "\t`json:\"")
			if params[idx].Option.Name != "" {
				io.WriteString(ctx.out, params[idx].Option.Name)
			} else {
				io.WriteString(ctx.out, toSnakeCase(params[idx].Param.Name))
			}
			io.WriteString(ctx.out, ",omitempty\"`")

			method.goArgumentLiterals[params[idx].Index] = "bindArgs." + fieldName
		}
		io.WriteString(ctx.out, "\r\n\t}")
	}

	io.WriteString(ctx.out, "\r\n\tif err := ")
	io.WriteString(ctx.out, ctx.plugin.ReadBodyFunc("&"+varName))
	io.WriteString(ctx.out, "; err != nil {\r\n")
	// txt := `NewBadArgument(err, "bindArgs", "body")`
	txt := ctx.plugin.GetBodyErrorText(method, varName, "err")

	ctx.plugin.RenderReturnError(ctx.out, method, "http.StatusBadRequest", txt)
	io.WriteString(ctx.out, "\r\n\t}")

	return nil
}

func (method *Method) renderInvokeAndReturn(ctx *GenContext) error {
	io.WriteString(ctx.out, "\r\n")
	/// 输出返回参数
	if len(method.Method.Results.List) > 2 {
		for idx, result := range method.Method.Results.List {
			if idx > 0 {
				io.WriteString(ctx.out, ", ")
			}
			if result.Type().IsErrorType() {
				io.WriteString(ctx.out, "err")
			} else {
				io.WriteString(ctx.out, result.Name)
			}
		}
		io.WriteString(ctx.out, " :=")
	} else if len(method.Method.Results.List) == 1 {
		if method.Method.Results.List[0].Type().IsErrorType() {
			if method.IsErrorDeclared() {
				io.WriteString(ctx.out, "err =")
			} else {
				io.WriteString(ctx.out, "err :=")
			}
		} else {
			io.WriteString(ctx.out, "result :=")
		}
	} else {
		io.WriteString(ctx.out, "result, err :=")
	}

	/// 输出调用
	io.WriteString(ctx.out, "svc.")
	io.WriteString(ctx.out, method.Method.Name)
	io.WriteString(ctx.out, "(")
	for idx, param := range method.goArgumentLiterals {
		if idx > 0 {
			io.WriteString(ctx.out, ", ")
		}
		io.WriteString(ctx.out, param)

		if method.Method.Params.List[idx].IsVariadic {
			io.WriteString(ctx.out, "...")
		}
	}
	io.WriteString(ctx.out, ")")

	noreturn := false
	if o := method.Operation.Extensions["x-gogen-noreturn"]; o != nil {
		noreturn = strings.ToLower(fmt.Sprint(o)) == "true"
	}

	/// 输出返回
	if len(method.Method.Results.List) > 2 {
		io.WriteString(ctx.out, "\r\n\tif err != nil {\r\n")
		ctx.plugin.RenderReturnError(ctx.out, method, "", "err")
		io.WriteString(ctx.out, "\r\n\t}")

		io.WriteString(ctx.out, "\r\n\tresult := map[string]interface{}{")

		for _, result := range method.Method.Results.List {
			if result.Type().IsErrorType() {
				continue
			}

			io.WriteString(ctx.out, "\r\n\t\"")
			io.WriteString(ctx.out, Underscore(result.Name))
			io.WriteString(ctx.out, "\":")
			io.WriteString(ctx.out, result.Name)
			io.WriteString(ctx.out, ",")
		}
		io.WriteString(ctx.out, "\r\n\t}\r\n")
		ctx.plugin.RenderReturnOK(ctx.out, method, "", "result")
	} else if len(method.Method.Results.List) == 1 {

		arg := method.Method.Results.List[0]
		if arg.Type().IsErrorType() {
			io.WriteString(ctx.out, "\r\nif err != nil {\r\n")
			ctx.plugin.RenderReturnError(ctx.out, method, "", "err")
			io.WriteString(ctx.out, "\r\n}")
			io.WriteString(ctx.out, "\r\n")
			if !noreturn {
				ctx.plugin.RenderReturnOK(ctx.out, method, "", "\"OK\"")
			} else {
				ctx.plugin.RenderReturnEmpty(ctx.out, method)
			}
		} else {
			// if methodParams.IsPlainText {
			//  	{{$.mux.PlainTextFunc $method "result"}}
			// } else {
			io.WriteString(ctx.out, "\r\n")
			ctx.plugin.RenderReturnOK(ctx.out, method, "", "result")
			// }
		}

	} else {
		io.WriteString(ctx.out, "\r\n\tif err != nil {\r\n")
		ctx.plugin.RenderReturnError(ctx.out, method, "", "err")
		io.WriteString(ctx.out, "\r\n\t}\r\n")

		// {{- if $methodParams.IsPlainText }}
		//   {{$.mux.PlainTextFunc $method "result"}}
		// {{- else}}
		ctx.plugin.RenderReturnOK(ctx.out, method, "", "result")
		// {{- end}}
	}

	return nil
}
