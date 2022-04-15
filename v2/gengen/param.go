package gengen

import (
	"errors"
	"fmt"
	"io"

	"github.com/go-openapi/spec"
	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

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
	var isStringPtrType = param.IsPtrType() && param.Type().PtrElemType().IsStringType(true)
	if isStringPtrType && param.Option.In == "path" {
		return "&" + param.Param.Name
	}
	return param.Param.Name
}

func (param *Param) GoType() string {
	return param.Param.Type().ToLiteral()
}

func (param *Param) IsPtrType() bool {
	return param.Type().IsPtrType()
}

func (param *Param) IsArrayType() bool {
	return param.Param.Type().IsSliceType()
}

func (param *Param) Type() astutil.Type {
	return param.Param.Type()
}

// func (param *Param) SliceType() ast.Expr {
// 	return astutil.SliceElemType(param.Param.Typ)
// }

// func (param *Param) PtrElemType() ast.Expr {
// 	return astutil.PtrElemType(param.Param.Typ)
// }

// func (param *Param) PtrOrSliceElemType() (*astutil.File, ast.Expr) {
// 	return param.Param.Method.Clazz.File.Package.Context.GetElemType(param.Param.Method.Clazz.File, param.Param.Typ, true)
// }

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

func (param *Param) RenderDeclareAndInit(plugin Plugin, out io.Writer, ts *astutil.TypeSpec, method *Method) error {
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

	if isBasicType || isUnderlyingBasicType || isBultinType(typeStr) || isNullableType(typeStr) {

		invocations := plugin.Invocations()
		foundIndex := -1
		for idx := range invocations {
			if (param.Option.In == "path") != invocations[idx].Required {
				continue
			}
			if param.IsArrayType() != invocations[idx].IsArray {
				continue
			}
			if invocations[idx].ResultType == typeStr {
				foundIndex = idx
				break
			}
			if underlyingType.IsValid() &&
				invocations[idx].ResultType == underlyingType.ToString() {
				foundIndex = idx
				break
			}
		}

		if foundIndex < 0 {
			for idx := range invocations {
				if (param.Option.In == "path") != invocations[idx].Required {
					continue
				}
				if param.IsArrayType() != invocations[idx].IsArray {
					continue
				}

				if invocations[idx].ResultType == "string" {
					foundIndex = idx
					break
				}
			}

			if foundIndex < 0 {
				return errors.New("param '" + param.Param.Name + "' of '" + method.Method.Clazz.Name + "." + method.Method.Name + "' cannot determine a invocation")
			}
		}

		io.WriteString(out, "\r\n")
		err := param.RenderBasic(out, plugin, method, &invocations[foundIndex])
		if err != nil {
			return err
		}
		return nil
	}

	// if param.IsStructType() {

	// }

	return errors.New("param '" + param.Param.Name + "' of '" + method.Method.Clazz.Name + "." + method.Method.Name + "' unsupported")
}

func (param *Param) RenderBasic(out io.Writer, plugin Plugin, method *Method, invocation *Invocation) error {
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

	var isPtrType = param.IsPtrType()

	if invocation.ResultError || invocation.ResultBool {
		if resultExceptedType {
			// 情况4, 6
			io.WriteString(out, param.GoVarName())
			if invocation.ResultBool {
				io.WriteString(out, ", ok = ")
				fmt.Fprintf(out, invocation.Format, param.Option.Name)

				if invocation.WithDefault {
					fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
				} else {
					fmt.Fprintf(out, invocation.Format, param.Option.Name)
				}

				io.WriteString(out, "\r\n\tif !ok {\r\n")
				renderBadArgument(out, plugin, method, param, "nil")
				io.WriteString(out, "\r\n\t}")
			} else {
				io.WriteString(out, ", err = ")

				if invocation.WithDefault {
					fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
				} else {
					fmt.Fprintf(out, invocation.Format, param.Option.Name)
				}

				io.WriteString(out, "\r\n\tif err != nil {\r\n")
				renderBadArgument(out, plugin, method, param, "err")
				io.WriteString(out, "\r\n\t}")
			}
			return nil
		}

		if isPtrType && !invocation.IsArray && invocation.ResultType == param.Type().PtrElemType().ToString() {
			// 情况11, 12
			io.WriteString(out, "var ")
			io.WriteString(out, param.GoVarName())

			io.WriteString(out, " ")
			io.WriteString(out, param.Type().ToString())

			if invocation.Required {
				// 情况12
				if invocation.ResultBool {
					io.WriteString(out, "\r\n\tif "+param.GoVarName()+"Value, ok = ")

					if invocation.WithDefault {
						fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
					} else {
						fmt.Fprintf(out, invocation.Format, param.Option.Name)
					}

					io.WriteString(out, "; ok {\r\n")
					io.WriteString(out, param.GoVarName()+" = &"+param.GoVarName()+"Value")
					io.WriteString(out, "\r\n\t} else {\r\n")
					renderBadArgument(out, plugin, method, param, "nil")
					io.WriteString(out, "\r\n\t}")
				} else {
					io.WriteString(out, "\r\n\tif "+param.GoVarName()+"Value, err = ")

					if invocation.WithDefault {
						fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
					} else {
						fmt.Fprintf(out, invocation.Format, param.Option.Name)
					}

					io.WriteString(out, "; err == nil {\r\n")
					io.WriteString(out, param.GoVarName()+" = &"+param.GoVarName()+"Value")
					io.WriteString(out, "\r\n\t} else {\r\n")
					renderBadArgument(out, plugin, method, param, "err")
					io.WriteString(out, "\r\n\t}")
				}
			} else {
				// 情况11
				if invocation.ResultBool {
					io.WriteString(out, "\r\n\tif "+param.GoVarName()+"Value, ok = ")

					if invocation.WithDefault {
						fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
					} else {
						fmt.Fprintf(out, invocation.Format, param.Option.Name)
					}
					io.WriteString(out, "; ok {\r\n")
					io.WriteString(out, param.GoVarName()+" = &"+param.GoVarName()+"Value")
					io.WriteString(out, "\r\n\t}")
				} else {
					io.WriteString(out, "\r\n\tif "+param.GoVarName()+"Value, err = ")
					if invocation.WithDefault {
						fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
					} else {
						fmt.Fprintf(out, invocation.Format, param.Option.Name)
					}
					io.WriteString(out, "; err == nil {\r\n")
					io.WriteString(out, param.GoVarName()+" = &"+param.GoVarName()+"Value")
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

		return errors.New("param '" + param.Param.Name + "' of " + method.Method.Clazz.Name + "." + method.Method.Name + " cannot resolved for '" + invocation.Format + "'")
	}

	var isStringPtrType = (isPtrType && !invocation.IsArray && invocation.ResultType == param.Type().PtrElemType().ToString())

	if resultExceptedType || isStringPtrType {
		// resultExceptedType = true 时，情况1, 2
		// isStringPtrType = true 时，情况9, 10

		if invocation.Required || !isStringPtrType {
			// resultExceptedType = true 时，情况1, 2

			io.WriteString(out, "var ")
			io.WriteString(out, param.GoVarName())
			io.WriteString(out, " = ")

			if isUnderlyingBasicType {
				io.WriteString(out, typeStr)
				io.WriteString(out, "(")
			}

			if invocation.WithDefault {
				fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
			} else {
				fmt.Fprintf(out, invocation.Format, param.Option.Name)
			}

			if isUnderlyingBasicType {
				io.WriteString(out, ")")
			}
		} else {

			// isStringPtrType = true 时，情况9, 10

			io.WriteString(out, "var ")
			io.WriteString(out, param.GoVarName())
			io.WriteString(out, " ")
			io.WriteString(out, param.Type().ToString())

			io.WriteString(out, "\r\n\tif s := ")
			if invocation.WithDefault {
				fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
			} else {
				fmt.Fprintf(out, invocation.Format, param.Option.Name)
			}
			io.WriteString(out, "; s != \"\" {")
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
				return errors.New("param '" + param.Param.Name + "' of '" + method.Method.Clazz.Name + "." + method.Method.Name + "' hasnot convert function: " + err.Error())
			}

			needCast = true
		}

		if invocation.Required {
			// 情况3

			if needCast {
				io.WriteString(out, param.GoVarName())
				io.WriteString(out, "Value, err := ")
			} else {
				io.WriteString(out, param.GoVarName())
				io.WriteString(out, ", err := ")
			}

			var paramValue string
			if invocation.WithDefault {
				paramValue = fmt.Sprintf(invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
			} else {
				paramValue = fmt.Sprintf(invocation.Format, param.Option.Name)
			}

			io.WriteString(out, fmt.Sprintf(convertFmt, paramValue))
			io.WriteString(out, "\r\n\tif err != nil {\r\n")
			renderBadArgument(out, plugin, method, param, "err")
			io.WriteString(out, "\r\n\t}")

			if needCast {
				io.WriteString(out, "\r\n\t")
				io.WriteString(out, param.GoVarName())
				io.WriteString(out, " = "+typeStr+"(")
				io.WriteString(out, param.GoVarName())
				io.WriteString(out, "Value)")
			}
			return nil
		}

		if isNullable {
			if invocation.IsArray {
				return errors.New("param '" + param.Param.Name + "' of " + method.Method.Clazz.Name + "." + method.Method.Name + " cannot resolved")
			}

			// 情况7，8
			io.WriteString(out, "var ")
			io.WriteString(out, param.GoVarName())
			io.WriteString(out, " ")
			io.WriteString(out, param.Type().ToString())

			io.WriteString(out, "\r\n\tif s := ")
			if invocation.WithDefault {
				fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
			} else {
				fmt.Fprintf(out, invocation.Format, param.Option.Name)
			}
			io.WriteString(out, "; s != \"\" {")

			io.WriteString(out, "\r\n\t\t"+param.GoVarName()+"Value")
			io.WriteString(out, ", err :="+fmt.Sprintf(convertFmt, "s"))
			io.WriteString(out, "\r\n\t\tif err != nil {\r\n")
			renderBadArgument(out, plugin, method, param, "err")
			io.WriteString(out, "\r\n\t\t}")

			if needCast {
				io.WriteString(out, "\r\n\t\t"+param.GoVarName()+"."+FieldNameForNullable(param.Type())+" = "+nullValueTypeStr+"("+param.GoVarName()+"Value)")
			} else {
				io.WriteString(out, "\r\n\t\t"+param.GoVarName()+"."+FieldNameForNullable(param.Type())+" = "+param.GoVarName()+"Value")
			}
			io.WriteString(out, "\r\n\t\t"+param.GoVarName()+".Valid = true")
			io.WriteString(out, "\r\n\t}")
			return nil
		}

		// 情况5
		io.WriteString(out, "var ")
		io.WriteString(out, param.GoVarName())
		io.WriteString(out, " ")
		io.WriteString(out, typeStr)

		if invocation.IsArray {
			io.WriteString(out, "\r\n\tif ss := ")
		} else {
			io.WriteString(out, "\r\n\tif s := ")
		}
		if invocation.WithDefault {
			fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
		} else {
			fmt.Fprintf(out, invocation.Format, param.Option.Name)
		}
		if invocation.IsArray {
			io.WriteString(out, "; len(ss) != 0 {")
		} else {
			io.WriteString(out, "; s != \"\" {")
		}

		io.WriteString(out, "\r\n\t\t"+param.GoVarName()+"Value")
		if invocation.IsArray {
			io.WriteString(out, ", err :="+fmt.Sprintf(convertFmt, "ss"))
		} else {
			io.WriteString(out, ", err :="+fmt.Sprintf(convertFmt, "s"))
		}
		io.WriteString(out, "\r\n\t\tif err != nil {\r\n")
		renderBadArgument(out, plugin, method, param, "err")
		io.WriteString(out, "\r\n\t\t}")

		if needCast {
			io.WriteString(out, "\r\n\t\t"+param.GoVarName()+" = "+typeStr+"("+param.GoVarName()+"Value)")
		} else {
			io.WriteString(out, "\r\n\t\t"+param.GoVarName()+" = "+param.GoVarName()+"Value")
		}
		io.WriteString(out, "\r\n\t}")
		return nil
	}

	elemTypeStr := param.Type().PtrElemType().ToString()

	io.WriteString(out, "var ")
	io.WriteString(out, param.GoVarName())
	io.WriteString(out, " ")
	io.WriteString(out, param.Type().ToString())

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
			return errors.New("param '" + param.Param.Name + "' of '" + method.Method.Clazz.Name + "." + method.Method.Name + "' hasnot convert function: " + err.Error())
		}

		needCast = true
	}

	if invocation.Required {
		// 情况13

		io.WriteString(out, "\r\nif ")
		io.WriteString(out, param.GoVarName())
		io.WriteString(out, "Value, err = ")

		var paramValue string
		if invocation.WithDefault {
			paramValue = fmt.Sprintf(invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
		} else {
			paramValue = fmt.Sprintf(invocation.Format, param.Option.Name)
		}

		io.WriteString(out, fmt.Sprintf(convertFmt, paramValue))
		io.WriteString(out, "; err != nil {\r\n")
		renderBadArgument(out, plugin, method, param, "err")
		io.WriteString(out, "\r\n\t} else {")
		io.WriteString(out, "\r\n\t\t"+param.GoVarName())

		if needCast {
			io.WriteString(out, " = new("+elemTypeStr+")")
			io.WriteString(out, "\r\n\t\t*")
			io.WriteString(out, param.GoVarName())
			io.WriteString(out, " = ")
			io.WriteString(out, param.GoVarName())
			io.WriteString(out, "Value")
		} else {
			io.WriteString(out, " = &")
			io.WriteString(out, param.GoVarName())
			io.WriteString(out, "Value")
		}
		io.WriteString(out, "\r\n\t}")
	} else {
		// 情况14
		io.WriteString(out, "\r\n\tif s := ")
		if invocation.WithDefault {
			fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
		} else {
			fmt.Fprintf(out, invocation.Format, param.Option.Name)
		}
		io.WriteString(out, "; s != \"\" {")

		io.WriteString(out, "\r\n\t\t"+param.GoVarName()+"Value")
		io.WriteString(out, ", err :="+fmt.Sprintf(convertFmt, "s"))
		io.WriteString(out, "\r\n\t\tif err != nil {\r\n")
		renderBadArgument(out, plugin, method, param, "err")
		io.WriteString(out, "\r\n\t\t}")

		if needCast {
			io.WriteString(out, "\r\n\t\t"+param.GoVarName()+" = new("+elemTypeStr+")")
			io.WriteString(out, "\r\n\t\t*"+param.GoVarName()+" = "+elemTypeStr+"("+param.GoVarName()+"Value)")
		} else {
			io.WriteString(out, "\r\n\t\t"+param.GoVarName()+" = &"+param.GoVarName()+"Value")
		}
		io.WriteString(out, "\r\n\t}")
	}

	return nil
}
