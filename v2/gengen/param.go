package gengen

import (
	"errors"
	"fmt"
	"io"

	"github.com/go-openapi/spec"
	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

type Param struct {
	Parent      *Param
	Field       *astutil.Field

	Method      *Method
	Param       *astutil.Param
	Option      spec.Parameter

	isInitialized bool
}

// GoVarName 申明时的变量名
func (param *Param) GoVarName() string {
	if param.Parent != nil {
		if param.Field != nil {
			return param.Parent.GoVarName() + "."+param.Field.Name
		}
		return param.Parent.GoVarName() + "."+param.Param.Name
	}
	if param.Parent != nil {
		panic("field '"+param.Parent.GoVarName() + param.Field.Name+"' isnot null")
	}
	return param.Param.Name
}

func (param *Param) fieldName() string {
	if param.Field != nil {
		return toLowerCamelCase(param.Field.Name)
	}
	return param.Param.Name
}

func (param *Param) isField() bool {
	return param.Parent != nil
}

// GoMethodParamName 方法定义时的参数名
func (param *Param) GoMethodParamName() string {
	return param.Param.Name
}

// GoVarName 函数调用时变量，如变量名为 a, 调用可能是 &a
func (param *Param) GoArgumentLiteral() string {
	var isStringPtrType = param.Type().IsPtrType() && param.Type().PtrElemType().IsStringType(true)
	if isStringPtrType && param.Option.In == "path" {
		return "&" + param.Param.Name
	}
	return param.Param.Name
}

// WebParamName 请求中变量名
func (param *Param) WebParamName() string {
	return param.Option.Name
}

func (param *Param) GoMethodFullName() string {
	return param.Method.FullName()
}

// func (param *Param) GoType() string {
// 	return param.Param.Type().ToLiteral()
// }

func (param *Param) Type() astutil.Type {
	return param.Param.Type()
}

func (param *Param) IsVariadic() bool {
	return param.Param.IsVariadic
}

func defaultValue(typ string, value interface{}) string {
	if value != nil {
		return fmt.Sprint(value)
	}
	if typ == "string" {
		return ""
	}
	return "0"
}


func (param *Param) SetInitialized() {
	param.isInitialized = true
}


func (param *Param) renderParentInit(plugin Plugin, out io.Writer, noCRCF ...bool) error {
	if param.Parent == nil {
		return nil
	}

	if !param.Parent.Type().IsPtrType() {
		return nil
	}

	if param.Parent.isInitialized {
		return nil
	}

	if len(noCRCF) == 0 || !noCRCF[0] {
		io.WriteString(out, "\r\n\t")
	}

	io.WriteString(out, "if " + param.Parent.GoVarName() + " == nil {")
	io.WriteString(out, "\r\n\t\t" + param.Parent.GoVarName() + " = new("+
		param.Parent.Type().PtrElemType().ToLiteral()+")")
	io.WriteString(out, "}")

	if len(noCRCF) > 0 && noCRCF[0] {
		io.WriteString(out, "\r\n")
	}

	return nil
}

func (param *Param) RenderDeclareAndInit(plugin Plugin, out io.Writer) error {
	// isPtr := astutil.IsPtrType(param.Type())
	typ := param.Type().PtrElemType()
	if !typ.IsValid() {
		typ = param.Type()
	}

	// isSlice := astutil.IsSliceType(typ)
	elmType := typ.SliceElemType()
	if !elmType.IsValid() {
		elmType = typ
	}

	typeStr := elmType.ToLiteral()
	isBasicType := elmType.IsBasicType(false)
	underlyingType := elmType.GetUnderlyingType()
	isUnderlyingBasicType := !isBasicType && underlyingType.IsValid() && underlyingType.IsBasicType(true)

	// var underlyingTypeStr string
	// if underlyingType.IsValid() {
	// 	underlyingTypeStr = underlyingType.ToString()
	// }

	if isBasicType ||
		isUnderlyingBasicType ||
		isBultinType(typeStr) ||
		isNullableType(typeStr) {

		if param.Option.Name == "" {
			return errors.New("param '" + param.GoMethodParamName() + "' of '" +
				param.GoMethodFullName() +
				"' missing in the swagger annotations")
		}
		var err error
		if underlyingType.IsValid() {
			err = param.renderBySimpleType(out, plugin, underlyingType.ToString())
		} else {
			err = param.renderBySimpleType(out, plugin, typeStr)
		}
		if err != nil {
			return err
		}
		return nil
	}

	if param.Type().IsStructType() || (param.Type().IsPtrType() && param.Type().PtrElemType().IsStructType()) {

		typ := param.Type()
		if typ.IsPtrType() {
			typ = typ.PtrElemType()
		}

		ts, err := typ.ToTypeSpec()
		if err != nil {
			return errors.New("param '" + param.GoMethodParamName() + "' of '" +
					param.GoMethodFullName() +
					"' cannot convert to type spec: " + err.Error())
		}

		if !param.isField() {
			io.WriteString(out, "\r\nvar "+param.GoVarName()+" "+param.Type().ToLiteral())
		}

		var fields = ts.Fields()
		for idx := range fields {
			oidx := param.Method.SearchSwaggerParameter(param.Param.Name, fields[idx].Name)
			if oidx < 0 {
				return errors.New("param '" + param.GoMethodParamName() +
					"' of '" + param.GoMethodFullName() +
					"' not found in the swagger annotations")
			}

			subparam := Param{
				Parent:      param,
				Field:       &fields[idx],
				Method:      param.Method,
				Param: &astutil.Param{
					Method:     param.Method.Method,
					Name:       fields[idx].Name,
					IsVariadic: false,
					Expr:       fields[idx].Expr,
				},
				Option: param.Method.Operation.Parameters[oidx],
			}


			err = subparam.RenderDeclareAndInit(plugin, out)
			if err != nil {
				return errors.New("param '" + param.GoMethodParamName() + "' of '" +
					param.GoMethodFullName() +
					"' cannot convert to type spec: " + err.Error())
			}
		}
		return nil
	}

	return errors.New("param '" + param.GoMethodParamName() + "' of '" +
		param.GoMethodFullName() +
		"' unsupported")
}

func (param *Param) renderBySimpleType(out io.Writer, plugin Plugin, typeStr string) error {

	invocations := plugin.Invocations()
	foundIndex := -1
	for idx := range invocations {
		if (param.Option.In == "path") != invocations[idx].Required {
			continue
		}
		if param.Type().IsSliceType() != invocations[idx].IsArray {
			continue
		}
		if invocations[idx].ResultType == typeStr {
			foundIndex = idx
			break
		}
	}

	if foundIndex < 0 {
		for idx := range invocations {
			if (param.Option.In == "path") != invocations[idx].Required {
				continue
			}
			if param.Type().IsSliceType() != invocations[idx].IsArray {
				continue
			}

			if invocations[idx].ResultType == "string" {
				foundIndex = idx
				break
			}
		}

		if foundIndex < 0 {
			return errors.New("param '" + param.GoMethodParamName() + "' of '" +
				param.GoMethodFullName() +
				"' cannot determine a invocation")
		}
	}

	io.WriteString(out, "\r\n")
	err := param.renderBasic(out, plugin, &invocations[foundIndex])
	if err != nil {
		return err
	}
	return nil
}

func (param *Param) renderBasic(out io.Writer, plugin Plugin, invocation *Invocation) error {
	typeStr := param.Type().ToString()

	elmType := param.Type().GetElemType(true)
	isBasicType := elmType.IsBasicType(false)

	underlyingType := elmType.GetUnderlyingType()
	isUnderlyingBasicType := !isBasicType && underlyingType.IsValid() && underlyingType.IsBasicType(true)

	resultExceptedType := (!invocation.IsArray && invocation.ResultType == typeStr) ||
		(invocation.IsArray && invocation.ResultType == elmType.ToString()) ||
		(invocation.IsArray && param.IsVariadic() && invocation.ResultType == param.Type().ToString()) ||
		(isNullableType(typeStr) && !invocation.IsArray && !param.IsVariadic() && invocation.ResultType == nullableType(typeStr))

	if !resultExceptedType && isUnderlyingBasicType {
		underlyingTypeStr := underlyingType.ToString()

		resultExceptedType = (!invocation.IsArray && invocation.ResultType == underlyingTypeStr) ||
			(invocation.IsArray && invocation.ResultType == underlyingTypeStr) ||
			(invocation.IsArray && param.IsVariadic() && invocation.ResultType == underlyingTypeStr)
	}

	var isPtrType = param.Type().IsPtrType()

	if invocation.ResultError || invocation.ResultBool {
		if resultExceptedType {
			// 情况4, 6
			io.WriteString(out, param.GoVarName())
			if invocation.ResultBool {
				io.WriteString(out, ", ok := ")
				fmt.Fprintf(out, invocation.Format, param.WebParamName())

				if invocation.WithDefault {
					fmt.Fprintf(out, invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
				} else {
					fmt.Fprintf(out, invocation.Format, param.WebParamName())
				}

				io.WriteString(out, "\r\n\tif !ok {\r\n")
				renderBadArgument(out, plugin, param, "nil")
				io.WriteString(out, "\r\n\t}")
			} else {
				io.WriteString(out, ", err := ")

				if invocation.WithDefault {
					fmt.Fprintf(out, invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
				} else {
					fmt.Fprintf(out, invocation.Format, param.WebParamName())
				}

				io.WriteString(out, "\r\n\tif err != nil {\r\n")
				renderBadArgument(out, plugin, param, "err")
				io.WriteString(out, "\r\n\t}")

				param.Method.SetErrorDeclared()
			}
			return nil
		}

		if isPtrType && !invocation.IsArray && invocation.ResultType == param.Type().PtrElemType().ToString() {
			// 情况11, 12

			if !param.isField() {
				io.WriteString(out, "var ")
				io.WriteString(out, param.GoVarName())

				io.WriteString(out, " ")
				io.WriteString(out, param.Type().ToLiteral())
			}

			if invocation.Required {
				// 情况12
				if invocation.ResultBool {
					io.WriteString(out, "\r\n\tif "+param.fieldName()+"Value, ok := ")

					if invocation.WithDefault {
						fmt.Fprintf(out, invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
					} else {
						fmt.Fprintf(out, invocation.Format, param.WebParamName())
					}

					io.WriteString(out, "; ok {")
					if err := param.renderParentInit(plugin, out); err != nil {
						return err
					}
					io.WriteString(out, "\r\n\t\t" + param.GoVarName()+" = &"+param.fieldName()+"Value")
					io.WriteString(out, "\r\n\t} else {\r\n")
					renderBadArgument(out, plugin, param, "nil")
					io.WriteString(out, "\r\n\t}")
				} else {
					io.WriteString(out, "\r\n\tif "+param.fieldName()+"Value, err := ")

					if invocation.WithDefault {
						fmt.Fprintf(out, invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
					} else {
						fmt.Fprintf(out, invocation.Format, param.WebParamName())
					}

					io.WriteString(out, "; err == nil {")
					if err := param.renderParentInit(plugin, out); err != nil {
						return err
					}
					io.WriteString(out, "\r\n\t" + param.GoVarName()+" = &"+param.fieldName()+"Value")
					io.WriteString(out, "\r\n\t} else {\r\n")
					renderBadArgument(out, plugin, param, "err")
					io.WriteString(out, "\r\n\t}")
				}
			} else {
				// 情况11
				if invocation.ResultBool {
					io.WriteString(out, "\r\n\tif "+param.fieldName()+"Value, ok := ")

					if invocation.WithDefault {
						fmt.Fprintf(out, invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
					} else {
						fmt.Fprintf(out, invocation.Format, param.WebParamName())
					}
					io.WriteString(out, "; ok {")
					if err := param.renderParentInit(plugin, out); err != nil {
						return err
					}
					io.WriteString(out, "\r\n\t" +param.GoVarName()+" = &"+param.fieldName()+"Value")
					io.WriteString(out, "\r\n\t}")
				} else {
					io.WriteString(out, "\r\n\tif "+param.fieldName()+"Value, err := ")
					if invocation.WithDefault {
						fmt.Fprintf(out, invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
					} else {
						fmt.Fprintf(out, invocation.Format, param.WebParamName())
					}
					io.WriteString(out, "; err == nil {")
					if err := param.renderParentInit(plugin, out); err != nil {
						return err
					}
					io.WriteString(out, "\r\n\t" +param.GoVarName()+" = &"+param.fieldName()+"Value")
					io.WriteString(out, "\r\n\t}")
				}
			}

			return nil
		}

		// 这个情况目前没有遇到过
		// // s, err := ctx.GetXXXXParam("id")
		// // if err != nil {
		// // 	ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", id, ctx.Param("id"), err).Error())
		// // 	return
		// // }
		// // id, err := strconv.ParseInt(s, 10, 64)
		// // if err != nil {
		// // 	ctx.String(http.StatusBadRequest, fmt.Errorf("argument %q is invalid - %q", id, ctx.Param("id"), err).Error())
		// // 	return
		// // }

		return errors.New("param '" + param.GoMethodParamName() + "' of '" +
			param.GoMethodFullName() +
			"' cannot resolved for '" + invocation.Format + "'")
	}

	var isStringPtrType = (isPtrType && !invocation.IsArray && invocation.ResultType == param.Type().PtrElemType().ToString())

	if resultExceptedType || isStringPtrType {
		// resultExceptedType = true 时，情况1, 2
		// isStringPtrType = true 时，情况9, 10

		if invocation.Required || !isStringPtrType {
			// resultExceptedType = true 时，情况1, 2

			if !param.isField() {
				io.WriteString(out, "\tvar ")
			} else {
				if err := param.renderParentInit(plugin, out, true); err != nil {
					return err
				}
				param.Parent.SetInitialized()
			}
			io.WriteString(out, param.GoVarName() + " = ")

			if isUnderlyingBasicType {
				io.WriteString(out, typeStr)
				io.WriteString(out, "(")
			}

			if invocation.WithDefault {
				fmt.Fprintf(out, invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
			} else {
				fmt.Fprintf(out, invocation.Format, param.WebParamName())
			}

			if isUnderlyingBasicType {
				io.WriteString(out, ")")
			}
		} else {

			// isStringPtrType = true 时，情况9, 10

			if !param.isField() {
				io.WriteString(out, "var ")
				io.WriteString(out, param.GoVarName())
				io.WriteString(out, " ")
				io.WriteString(out, param.Type().ToLiteral())
				io.WriteString(out, "\r\n\t")
			}

			io.WriteString(out, "if s := ")
			if invocation.WithDefault {
				fmt.Fprintf(out, invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
			} else {
				fmt.Fprintf(out, invocation.Format, param.WebParamName())
			}
			io.WriteString(out, "; s != \"\" {")
			if err := param.renderParentInit(plugin, out); err != nil {
				return err
			}
			if isUnderlyingBasicType {
				io.WriteString(out, "\r\n\t"+param.GoVarName()+" = new(")
				io.WriteString(out, typeStr)
				io.WriteString(out, ")")
				io.WriteString(out, "\r\n\t*"+param.GoVarName()+" = "+typeStr+"(s)")
			} else {
				io.WriteString(out, "\r\n\t"+param.GoVarName()+" = &s")
			}
			io.WriteString(out, "\r\n\t}")
		}
		return nil
	}

	if !isPtrType {
		isNullable := isNullableType(typeStr)
		nullValueTypeStr := nullableType(typeStr)

		convertFmt, needCast, err := selectConvert(invocation.IsArray, invocation.ResultType, nullValueTypeStr)
		if err != nil {
			underlyingType := elmType.GetUnderlyingType()
			if underlyingType.IsValid() {
				convertFmt, needCast, err = selectConvert(invocation.IsArray, invocation.ResultType, underlyingType.ToString())
			}
			if err != nil {
				return errors.New("param '" + param.GoMethodParamName() + "' of '" +
					param.GoMethodFullName() +
					"' hasnot convert function: " + err.Error())
			}

			needCast = true
		}

		if invocation.Required {
			// 情况3


			param.Method.SetErrorDeclared()
			if needCast {
				io.WriteString(out, param.fieldName())
				io.WriteString(out, "Value, err := ")
			} else {
				io.WriteString(out, param.GoVarName())
				io.WriteString(out, ", err := ")
			}

			var paramValue string
			if invocation.WithDefault {
				paramValue = fmt.Sprintf(invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
			} else {
				paramValue = fmt.Sprintf(invocation.Format, param.WebParamName())
			}

			io.WriteString(out, fmt.Sprintf(convertFmt, paramValue))
			io.WriteString(out, "\r\n\tif err != nil {\r\n")
			renderBadArgument(out, plugin, param, "err")
			io.WriteString(out, "\r\n\t}")

			if needCast {
				if param.isField() {
					if err := param.renderParentInit(plugin, out); err != nil {
						return err
					}
					param.Parent.SetInitialized()

					io.WriteString(out, "\r\n\t")
				} else {
					io.WriteString(out, "\r\n\tvar ")
				}

				io.WriteString(out, param.GoVarName())
				io.WriteString(out, " = "+typeStr+"(")
				io.WriteString(out, param.fieldName())
				io.WriteString(out, "Value)")
			}
			return nil
		}

		if isNullable {
			if invocation.IsArray {
				return errors.New("param '" + param.GoMethodParamName() + "' of '" +
					param.GoMethodFullName() +
					"' cannot resolved")
			}

			// 情况7，8
			if !param.isField() {
				io.WriteString(out, "var ")
				io.WriteString(out, param.GoVarName())
				io.WriteString(out, " ")
				io.WriteString(out, param.Type().ToLiteral())
				io.WriteString(out, "\r\n\t")
			}

			io.WriteString(out, "if s := ")
			if invocation.WithDefault {
				fmt.Fprintf(out, invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
			} else {
				fmt.Fprintf(out, invocation.Format, param.WebParamName())
			}
			io.WriteString(out, "; s != \"\" {")

			io.WriteString(out, "\r\n\t\t"+param.fieldName()+"Value")
			io.WriteString(out, ", err :="+fmt.Sprintf(convertFmt, "s"))
			io.WriteString(out, "\r\n\t\tif err != nil {\r\n")
			renderBadArgument(out, plugin, param, "err")
			io.WriteString(out, "\r\n\t\t}")


			if err := param.renderParentInit(plugin, out); err != nil {
				return err
			}
			if needCast {
				io.WriteString(out, "\r\n\t\t"+param.GoVarName()+"."+FieldNameForNullable(param.Type())+" = "+nullValueTypeStr+"("+param.fieldName()+"Value)")
			} else {
				io.WriteString(out, "\r\n\t\t"+param.GoVarName()+"."+FieldNameForNullable(param.Type())+" = "+param.fieldName()+"Value")
			}
			io.WriteString(out, "\r\n\t\t"+param.GoVarName()+".Valid = true")
			io.WriteString(out, "\r\n\t}")
			return nil
		}

		// 情况5
		if !param.isField() {
			io.WriteString(out, "var ")
			io.WriteString(out, param.GoVarName())
			io.WriteString(out, " ")
			io.WriteString(out, typeStr)
			io.WriteString(out, "\r\n\t")
		}

		if invocation.IsArray {
			io.WriteString(out, "if ss := ")
		} else {
			io.WriteString(out, "if s := ")
		}
		if invocation.WithDefault {
			fmt.Fprintf(out, invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
		} else {
			fmt.Fprintf(out, invocation.Format, param.WebParamName())
		}
		if invocation.IsArray {
			io.WriteString(out, "; len(ss) != 0 {")
		} else {
			io.WriteString(out, "; s != \"\" {")
		}

		io.WriteString(out, "\r\n\t\t"+param.fieldName()+"Value")
		if invocation.IsArray {
			io.WriteString(out, ", err :="+fmt.Sprintf(convertFmt, "ss"))
		} else {
			io.WriteString(out, ", err :="+fmt.Sprintf(convertFmt, "s"))
		}
		io.WriteString(out, "\r\n\t\tif err != nil {\r\n")
		renderBadArgument(out, plugin, param, "err")
		io.WriteString(out, "\r\n\t\t}")


		if err := param.renderParentInit(plugin, out); err != nil {
			return err
		}
		if needCast {
			io.WriteString(out, "\r\n\t\t"+param.GoVarName()+" = "+typeStr+"("+param.fieldName()+"Value)")
		} else {
			io.WriteString(out, "\r\n\t\t"+param.GoVarName()+" = "+param.fieldName()+"Value")
		}
		io.WriteString(out, "\r\n\t}")
		return nil
	}

	elemTypeStr := param.Type().PtrElemType().ToString()


	if !param.isField() {
		io.WriteString(out, "var ")
		io.WriteString(out, param.GoVarName())
		io.WriteString(out, " ")
		io.WriteString(out, param.Type().ToLiteral())
		io.WriteString(out, "\r\n\t")
	}

	// resultExceptedType := !invocation.IsArray && invocation.ResultType == astutil.ToString(param.PtrElemType())
	// if resultExceptedType {
	// 	// 情况10, 11
	// 	// 前面已经处理了
	// 	return
	// }

	convertFmt, needCast, err := selectConvert(invocation.IsArray, invocation.ResultType,
		param.Type().PtrElemType().ToString())
	if err != nil {
		underlyingType := elmType.GetUnderlyingType()
		if underlyingType.IsValid() {
			convertFmt, needCast, err = selectConvert(invocation.IsArray, invocation.ResultType, underlyingType.ToString())
		}
		if err != nil {
			return errors.New("param '" + param.GoMethodParamName() + "' of '" +
				param.GoMethodFullName() +
				"' hasnot convert function: " + err.Error())
		}

		needCast = true
	}

	if invocation.Required {
		// 情况13

		io.WriteString(out, "if ")
		io.WriteString(out, param.fieldName())
		io.WriteString(out, "Value, err := ")

		var paramValue string
		if invocation.WithDefault {
			paramValue = fmt.Sprintf(invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
		} else {
			paramValue = fmt.Sprintf(invocation.Format, param.WebParamName())
		}

		io.WriteString(out, fmt.Sprintf(convertFmt, paramValue))
		io.WriteString(out, "; err != nil {\r\n")
		renderBadArgument(out, plugin, param, "err")
		io.WriteString(out, "\r\n\t} else {")


		if err := param.renderParentInit(plugin, out); err != nil {
			return err
		}
		io.WriteString(out, "\r\n\t\t"+param.GoVarName())

		if needCast {
			io.WriteString(out, " = new("+elemTypeStr+")")
			io.WriteString(out, "\r\n\t\t*")
			io.WriteString(out, param.GoVarName())
			io.WriteString(out, " = ")
			io.WriteString(out, param.fieldName())
			io.WriteString(out, "Value")
		} else {
			io.WriteString(out, " = &")
			io.WriteString(out, param.fieldName())
			io.WriteString(out, "Value")
		}
		io.WriteString(out, "\r\n\t}")
	} else {
		// 情况14
		io.WriteString(out, "if s := ")
		if invocation.WithDefault {
			fmt.Fprintf(out, invocation.Format, param.WebParamName(), defaultValue(invocation.ResultType, param.Option.Default))
		} else {
			fmt.Fprintf(out, invocation.Format, param.WebParamName())
		}
		io.WriteString(out, "; s != \"\" {")

		io.WriteString(out, "\r\n\t\t"+param.fieldName()+"Value")
		io.WriteString(out, ", err :="+fmt.Sprintf(convertFmt, "s"))
		io.WriteString(out, "\r\n\t\tif err != nil {\r\n")
		renderBadArgument(out, plugin, param, "err")
		io.WriteString(out, "\r\n\t\t}")


					if err := param.renderParentInit(plugin, out); err != nil {
						return err
					}
		if needCast {
			io.WriteString(out, "\r\n\t\t"+param.GoVarName()+" = new("+elemTypeStr+")")
			io.WriteString(out, "\r\n\t\t*"+param.GoVarName()+" = "+elemTypeStr+"("+param.fieldName()+"Value)")
		} else {
			io.WriteString(out, "\r\n\t\t"+param.GoVarName()+" = &"+param.fieldName()+"Value")
		}
		io.WriteString(out, "\r\n\t}")
	}

	return nil
}
