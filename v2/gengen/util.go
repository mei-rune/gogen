package gengen

import (
	"errors"
	"reflect"
	"strings"
	"unicode"

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
	return "net.IP" == name ||
		"net.HardwareAddr" == name ||
		"time.Time" == name ||
		"time.Duration" == name
}

// func isUnderlyingBasicType(file *astutil.File, elmType ast.Expr) bool {
// 	return file.Package.Context.IsBasicType(file, elmType)
// }

// func underlyingType(file *astutil.File, elmType ast.Expr) (*astutil.File, ast.Expr) {
// 	return file.Package.Context.GetUnderlyingType(file, elmType)
// }

func isNullableType(name string) bool {
	return strings.HasPrefix(name, "sql.Null") || strings.HasPrefix(name, "null.")
}

func nullableType(name string) string {
	if strings.HasPrefix(name, "sql.Null") {
		name = strings.TrimPrefix(name, "sql.Null")
		return strings.ToLower(name)
	}
	if strings.HasPrefix(name, "null.") {
		name = strings.TrimPrefix(name, "null.")
		return strings.ToLower(name)
	}
	return name
}

func FieldNameForNullable(typ astutil.Type) string {
	// sql.NullBool, sql.NullInt64, sql.NullString, sql.NullTime ......
	name := typ.ToString()
	name = strings.TrimPrefix(name, "sql.Null")
	name = strings.TrimPrefix(name, "null.")
	return name
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

var ConvertHook func(isArray bool, paramType string) *ConvertFunc

func selectConvert(isArray bool, resultType, paramType string) (string, bool, bool, error) {
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

	switch paramType {
	case "int":
		if isArray {
			return "ToIntArray(%s)", false, true, nil
		}
		return "strconv.Atoi(%s)", false, true, nil
	case "int64":
		if isArray {
			return "ToInt64Array(%s)", false, true, nil
		}
		return "strconv.ParseInt(%s, 10, 64)", false, true, nil
	case "int32":
		if isArray {
			return "ToInt32Array(%s)", false, true, nil
		}
		return "strconv.ParseInt(%s, 10, 32)", true, true, nil
	case "int16":
		if isArray {
			return "ToInt16Array(%s)", false, true, nil
		}
		return "strconv.ParseInt(%s, 10, 16)", true, true, nil
	case "int8":
		if isArray {
			return "ToInt8Array(%s)", false, true, nil
		}
		return "strconv.ParseInt(%s, 10, 8)", true, true, nil
	case "uint":
		if isArray {
			return "ToUintArray(%s)", false, true, nil
		}
		return "strconv.ParseUint(%s, 10, 64)", true, true, nil
	case "uint64":
		if isArray {
			return "ToUint64Array(%s)", false, true, nil
		}
		return "strconv.ParseUint(%s, 10, 64)", false, true, nil
	case "uint32":
		if isArray {
			return "ToUint32Array(%s)", false, true, nil
		}
		return "strconv.ParseUint(%s, 10, 32)", true, true, nil
	case "uint16":
		if isArray {
			return "ToUint16Array(%s)", false, true, nil
		}
		return "strconv.ParseUint(%s, 10, 16)", true, true, nil
	case "uint8":
		if isArray {
			return "ToUint8Array(%s)", false, true, nil
		}
		return "strconv.ParseUint(%s, 10, 8)", true, true, nil
	case "bool":
		if isArray {
			return "ToBoolArray(%s)", false, true, nil
		}
		return "strconv.ParseBool(%s)", false, true, nil
	case "float64":
		if isArray {
			return "ToFloat64Array(%s)", false, true, nil
		}
		return "strconv.ParseFloat(%s, 64)", false, true, nil
	case "float32":
		if isArray {
			return "ToFloat32Array(%s)", false, true, nil
		}
		return "strconv.ParseFloat(%s, 32)", true, true, nil
	case "time.Time":
		if isArray {
			return "ToDatetimes(%s)", false, true, nil
		}
		return "ToDatetime(%s)", false, true, nil
	case "time.Duration":
		if isArray {
			return "ToDurations(%s)", false, true, nil
		}
		return "time.ParseDuration(%s)", false, true, nil
	case "net.IP":
		if isArray {
			return "ToIPList(%s)", false, true, nil
		}
		return "ToIPAddr(%s)", false, true, nil
	case "net.HardwareAddr":
		if isArray {
			return "ToMacList(%s)", false, true, nil
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

func Underscore(name string) string {
	ss := strings.Split(name, ".")
	for idx := range ss {
		ss[idx] = UnderscoreSimple(ss[idx])
	}
	return strings.Join(ss, ".")
}

func toSnakeCase(in string) string {
	return Underscore(in)
}

func toLowerCamelCase(in string) string {
	runes := []rune(in)

	var out []rune
	flag := false
	for i, curr := range runes {
		if (i == 0 && unicode.IsUpper(curr)) || (flag && unicode.IsUpper(curr)) {
			out = append(out, unicode.ToLower(curr))
			flag = true
		} else {
			out = append(out, curr)
			flag = false
		}
	}

	return string(out)
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
