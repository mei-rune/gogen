package gengen

import (
	"errors"
	"go/ast"
	"go/token"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-openapi/spec"
	"github.com/grsmv/inflect"
	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

var TimeFormat = "TimeFormat"

type PathSegement struct {
	IsArgument bool
	Value      string
}

type ReplaceFunc func(PathSegement) string

var (
	Colon ReplaceFunc = func(segement PathSegement) string {
		return ":" + segement.Value
	}

	Brace ReplaceFunc = func(segement PathSegement) string {
		return "{" + segement.Value + "}"
	}
)

func ConvertURL(pa string, canEmpty bool, replace ReplaceFunc) (string, error) {
	segements, err := parseURL(pa)
	if err != nil {
		return "", err
	}
	return JoinPathSegments(segements, canEmpty, replace), nil
}

func JoinPathSegments(segements []PathSegement, canEmpty bool, replace ReplaceFunc) string {
	if len(segements) == 0 {
		if canEmpty {
			return ""
		}
		return "/"
	}

	var sb strings.Builder
	for idx := range segements {
		sb.WriteString("/")

		if segements[idx].IsArgument {
			sb.WriteString(replace(segements[idx]))
		} else {
			sb.WriteString(segements[idx].Value)
		}
	}
	return sb.String()
}

// parse parses a URL from a string in one of two contexts. If
// viaRequest is true, the URL is assumed to have arrived via an HTTP request,
// in which case only absolute URLs or path-absolute relative URLs are allowed.
// If viaRequest is false, all forms of relative URLs are allowed.
func parseURL(rawurl string) ([]PathSegement, error) {
	i := strings.IndexByte(rawurl, '?')
	var pa string
	if i < 0 {
		pa = rawurl
	} else {
		pa = rawurl[:i]
	}

	pathList := strings.Split(strings.Trim(pa, "/"), "/")
	var pathNames []string
	var segements []PathSegement
	if pa != "" {
		for idx := range pathList {
			if strings.HasPrefix(pathList[idx], "{") {
				if !strings.HasSuffix(pathList[idx], "}") {
					return nil, errors.New("url path '" + pa + "' is invalid")
				}
				name := strings.TrimPrefix(pathList[idx], "{")
				name = strings.TrimSuffix(name, "}")
				pathNames = append(pathNames, name)
				segements = append(segements, PathSegement{IsArgument: true, Value: name})
			} else {
				segements = append(segements, PathSegement{IsArgument: false, Value: pathList[idx]})
			}
		}
	}
	return segements, nil
}

func isBultinType(name string) bool {
	return isExceptedType(name, bultinTypes)
}

var bultinTypes = []string{
	"net.IP",
	"net.HardwareAddr",
	"time.Time",
	"time.Duration",
}

func isExceptedType(name string, anyTypes []string) bool {
	for _, anyType := range anyTypes {
		if anyType == name {
			return true
		}
	}

	return false
}

// func isUnderlyingBasicType(file *astutil.File, elmType ast.Expr) bool {
// 	return file.Package.Context.IsBasicType(file, elmType)
// }

// func underlyingType(file *astutil.File, elmType ast.Expr) (*astutil.File, ast.Expr) {
// 	return file.Package.Context.GetUnderlyingType(file, elmType)
// }

// func isNullableType(name string) bool {
// 	return strings.HasPrefix(name, "sql.Null") || strings.HasPrefix(name, "null.")
// }

// func nullableType(name string) string {
// 	if strings.HasPrefix(name, "sql.Null") {
// 		name = strings.TrimPrefix(name, "sql.Null")
// 		return strings.ToLower(name)
// 	}
// 	if strings.HasPrefix(name, "null.") {
// 		name = strings.TrimPrefix(name, "null.")
// 		return strings.ToLower(name)
// 	}
// 	return name
// }

func ElemTypeForNullable(typ astutil.Type) string {
	return astutil.ElemTypeForSqlNullable(typ)
}

func FieldNameForNullable(typ astutil.Type) string {
	return astutil.FieldNameForSqlNullable(typ)
}

var methodNames = map[string]string{
	"get":    "Get",
	"put":    "Put",
	"delete": "Delete",
	"post":   "Post",
	"patch":  "Patch",
	"head":   "Head",
}

func ConvertMethodNameToCamelCase(name string) string {
	newname := methodNames[strings.ToLower(name)]
	if newname == "" {
		panic(errors.New("'" + name + "' is unsupported"))
	}
	return newname
}

type ConvertFunc struct {
	Format      string
	NeedCast    bool
	HasRetError bool
}


var convertParamTypes []string

var ConvertHook func(isArray bool, paramType string) *ConvertFunc

func selectConvert(convertNS string, isArray bool, resultType, paramType string) (string, bool, bool, error) {
	if isArray {
		if !strings.HasPrefix(paramType, "[]") {
			return "", false, false, errors.New("cannot convert to '" + paramType + "', param type isnot match")
		}
		paramType = strings.TrimPrefix(paramType, "[]")
	}
	if ConvertHook != nil {
		r := ConvertHook(isArray, paramType)
		if r != nil {
			return r.Format, r.NeedCast, r.HasRetError, nil
		}
	}

	for  _, name := range convertParamTypes {
		if name == paramType {
			if isArray {
				return "Parse"+name+"(%s)", false, true, nil
			}
			return "Parse"+name+"(%s)", false, true, nil
		}
	}
	dot := strings.IndexByte(paramType, '.')
	if dot > 0 {
		ns := paramType[:dot]
		typeName := paramType[dot+1:]

		for  _, name := range convertParamTypes {
			if name == typeName {
				if isArray {
					return ns + ".Parse"+name+"(%s)", false, true, nil
				}
				return ns + ".Parse"+name+"(%s)", false, true, nil
			}
		}
	}

	switch paramType {
	case "int":
		if isArray {
			return convertNS + "ToIntArray(%s)", false, true, nil
		}
		return "strconv.Atoi(%s)", false, true, nil
	case "int64":
		if isArray {
			return convertNS + "ToInt64Array(%s)", false, true, nil
		}
		return "strconv.ParseInt(%s, 10, 64)", false, true, nil
	case "int32":
		if isArray {
			return convertNS + "ToInt32Array(%s)", false, true, nil
		}
		return "strconv.ParseInt(%s, 10, 32)", true, true, nil
	case "int16":
		if isArray {
			return convertNS + "ToInt16Array(%s)", false, true, nil
		}
		return "strconv.ParseInt(%s, 10, 16)", true, true, nil
	case "int8":
		if isArray {
			return convertNS + "ToInt8Array(%s)", false, true, nil
		}
		return "strconv.ParseInt(%s, 10, 8)", true, true, nil
	case "uint":
		if isArray {
			return convertNS + "ToUintArray(%s)", false, true, nil
		}
		return "strconv.ParseUint(%s, 10, 64)", true, true, nil
	case "uint64":
		if isArray {
			return convertNS + "ToUint64Array(%s)", false, true, nil
		}
		return "strconv.ParseUint(%s, 10, 64)", false, true, nil
	case "uint32":
		if isArray {
			return convertNS + "ToUint32Array(%s)", false, true, nil
		}
		return "strconv.ParseUint(%s, 10, 32)", true, true, nil
	case "uint16":
		if isArray {
			return convertNS + "ToUint16Array(%s)", false, true, nil
		}
		return "strconv.ParseUint(%s, 10, 16)", true, true, nil
	case "uint8":
		if isArray {
			return convertNS + "ToUint8Array(%s)", false, true, nil
		}
		return "strconv.ParseUint(%s, 10, 8)", true, true, nil
	case "bool":
		if isArray {
			return convertNS + "ToBoolArray(%s)", false, true, nil
		}

		// return r.Format, r.NeedCast, r.HasRetError, nil
		// return "ToBool(%s)", false, false, nil
		return "strconv.ParseBool(%s)", false, true, nil
	case "float64":
		if isArray {
			return convertNS + "ToFloat64Array(%s)", false, true, nil
		}
		return "strconv.ParseFloat(%s, 64)", false, true, nil
	case "float32":
		if isArray {
			return convertNS + "ToFloat32Array(%s)", false, true, nil
		}
		return "strconv.ParseFloat(%s, 32)", true, true, nil
	case "time.Time":
		if isArray {
			return convertNS + "ToDatetimes(%s)", false, true, nil
		}
		return convertNS + "ToDatetime(%s)", false, true, nil
	case "time.Duration":
		if isArray {
			return convertNS + "ToDurations(%s)", false, true, nil
		}
		return "time.ParseDuration(%s)", false, true, nil
	case "net.IP":
		if isArray {
			return convertNS + "ToIPList(%s)", false, true, nil
		}
		return convertNS + "ToIPAddr(%s)", false, true, nil
	case "net.HardwareAddr":
		if isArray {
			return convertNS + "ToMacList(%s)", false, true, nil
		}
		return "net.ParseMAC(%s)", false, true, nil
	default:
		if isArray {
			return "", false, false, errors.New("cannot convert to '[]" + paramType + "'")
		}
		return "", false, false, errors.New("cannot convert to '" + paramType + "'")
	}
}

func UnderscoreSimple(name string) string {
	return strings.Replace(inflect.Underscore(name), "_i_d", "_id", -1)
}

func CamelCase(name string) string {
	return strings.Replace(inflect.Camelize(name), "Id", "ID", -1)
}

func Underscore(name string) string {
	ss := strings.Split(name, ".")
	for idx := range ss {
		ss[idx] = UnderscoreSimple(ss[idx])
	}
	return strings.Join(ss, ".")
}

func Singularize(word string) string {
	return inflect.Singularize(word)
}

func toSnakeCase(in string) string {
	return Underscore(in)
}

func FieldNameEqual(name1,  name2 string) bool {
	return strings.EqualFold(name1, name2) ||
				strings.EqualFold(toSnakeCase(name1), name2) ||
				strings.EqualFold(Singularize(name1), name2) 
}

func toLowerCamelCase(in string) string {
	return toLowerFirst(CamelCase(in))
}

func toLowerFirst(in string) string {
	if in == "" {
		return in
	}

	runes := []rune(in)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func toUpperFirst(in string) string {
	if in == "" {
		return in
	}

	runes := []rune(in)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func getTagValue(field *astutil.Field, name string) (string, bool) {
	if field.Tag == nil {
		return "", false
	}
	s := strings.Trim(field.Tag.Value, "`")
	value, ok := reflect.StructTag(s).Lookup(name)
	return strings.Trim(value, "\""), ok
}

var zeroLits = map[string]string{
	"bool":      "false",
	"time.Time": "time.Time{}",
	"string":    "\"\"",
}

func zeroValueLiteral(typ astutil.Type) string {
	switch typ.Expr.(type) {
	case *ast.StarExpr:
		return "nil"
	case *ast.ArrayType:
		return "nil"
	case *ast.MapType:
		return "nil"
	}

	if typ.IsStringType(true) {
		return "\"\""
	}

	s := typ.ToLiteral()
	if lit, ok := zeroLits[s]; ok {
		return lit
	}
	return "0"
}

func isExtendEntire(param *spec.Parameter) bool {
	s, ok := param.Extensions.GetString("x-gogen-entire-body")
	if !ok {
		return true
	}
	return strings.ToLower(s) == "true"
}

func isExtendInline(param *spec.Parameter) bool {
	s, _ := param.Extensions.GetString("x-gogen-extend")
	return strings.ToLower(s) == "inline"
}

func getJSONName(s string) string {
	if s == "" {
		return ""
	}

	ss := strings.Split(s, ",")
	s = strings.TrimSpace(ss[0])
	if s == "-" {
		return ""
	}
	return s
}

func castToExceptedParamType(param *Param, pos token.Pos) ast.Expr {
	switch param.option.Type {
	case "int", "int64", "int32", "int16", "int8":
		return &ast.Ident{
			NamePos: pos,
			Name:    "int64",
		}
	case "uint", "uint64", "uint32", "uint16", "uint8":
		return &ast.Ident{
			NamePos: pos,
			Name:    "uint64",
		}
	case "bool":
		return &ast.Ident{
			NamePos: pos,
			Name:    "bool",
		}
	default:
		return &ast.Ident{
			NamePos: pos,
			Name:    "string",
		}
	}
}
