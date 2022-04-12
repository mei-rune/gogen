package gengen

import (
	"errors"
	"fmt"
	"io"

	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

func defaultValue(typ string, value interface{}) string {
	if value != nil {
		return fmt.Sprint(value)
	}
	if typ == "string" {
		return ""
	}
	return "0"
}

func RenderInvocation(out io.Writer, plugin Plugin, method *Method, param *Param, invocation *Invocation) error {
	typeStr := astutil.ToString(param.Type())
	resultExceptedType := invocation.ResultType == typeStr ||
		(invocation.IsArray && invocation.ResultType == astutil.ToString(param.SliceType())) ||
		(invocation.IsArray && param.IsVariadic() && invocation.ResultType == astutil.ToString(param.Type())) ||
		(isNullableType(typeStr) && !invocation.IsArray && !param.IsVariadic() && invocation.ResultType == nullableType(typeStr))

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

		if isPtrType && !invocation.IsArray && invocation.ResultType == astutil.ToString(param.PtrElemType()) {
			// 情况11, 12
			io.WriteString(out, "var ")
			io.WriteString(out, param.GoVarName())

			io.WriteString(out, " ")
			io.WriteString(out, astutil.ToString(param.Type()))

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

	var isStringPtrType = (isPtrType && !invocation.IsArray && invocation.ResultType == astutil.ToString(param.PtrElemType()))

	if resultExceptedType || isStringPtrType {
		// resultExceptedType = true 时，情况1, 2
		// isStringPtrType = true 时，情况9, 10

		if invocation.Required || !isStringPtrType {
			// resultExceptedType = true 时，情况1, 2

			io.WriteString(out, "var ")
			io.WriteString(out, param.GoVarName())
			io.WriteString(out, " = ")

			if invocation.WithDefault {
				fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
			} else {
				fmt.Fprintf(out, invocation.Format, param.Option.Name)
			}
		} else {

			// isStringPtrType = true 时，情况9, 10

			io.WriteString(out, "var ")
			io.WriteString(out, param.GoVarName())
			io.WriteString(out, " ")
			io.WriteString(out, astutil.ToString(param.Type()))

			io.WriteString(out, "\r\n\tif s := ")
			if invocation.WithDefault {
				fmt.Fprintf(out, invocation.Format, param.Option.Name, defaultValue(invocation.ResultType, param.Option.Default))
			} else {
				fmt.Fprintf(out, invocation.Format, param.Option.Name)
			}
			io.WriteString(out, "; s != \"\" {")
			io.WriteString(out, "\r\n\t"+param.GoVarName()+" = &s")
			io.WriteString(out, "\r\n\t}")
		}
		return nil
	}

	if !isPtrType {

		isNullable := isNullableType(typeStr)
		nullValueTypeStr := nullableType(typeStr)

		convertFmt, needCast, err := selectConvert(invocation.IsArray, invocation.ResultType, nullValueTypeStr)
		if err != nil {
			return errors.New("param '" + param.Param.Name + "' of '" + method.Method.Clazz.Name + "." + method.Method.Name + "' hasnot convert function: " + err.Error())
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
				io.WriteString(out, " = "+astutil.ToString(param.Type())+"(")
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
			io.WriteString(out, astutil.ToString(param.Type()))

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

	elemTypeStr := astutil.ToString(param.PtrElemType())

	io.WriteString(out, "var ")
	io.WriteString(out, param.GoVarName())
	io.WriteString(out, " ")
	io.WriteString(out, astutil.ToString(param.Type()))

	// resultExceptedType := !invocation.IsArray && invocation.ResultType == astutil.ToString(param.PtrElemType())
	// if resultExceptedType {
	// 	// 情况10, 11
	// 	// 前面已经处理了
	// 	return
	// }

	convertFmt, needCast, err := selectConvert(invocation.IsArray, invocation.ResultType,
		astutil.ToString(param.PtrElemType()))
	if err != nil {
		return errors.New("param '" + param.Param.Name + "' of '" + method.Method.Clazz.Name + "." + method.Method.Name + "' hasnot convert function: " + err.Error())
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
