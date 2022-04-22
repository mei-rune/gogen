package gengen

import (
	"strings"

	"github.com/grsmv/inflect"
)

var TimeFormat = "TimeFormat"

type PathSegement struct {
	IsArgument bool
	Value      string
}

type ReplaceFunc func(PathSegement) string

var (
	colonReplace ReplaceFunc = func(segement PathSegement) string {
		return ":" + segement.Value
	}

	braceReplace ReplaceFunc = func(segement PathSegement) string {
		return "{" + segement.Value + "}"
	}
)

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
func parseURL(rawurl string) ([]PathSegement, []string, map[string]string) {
	i := strings.IndexByte(rawurl, '?')
	var pa, query string
	if i < 0 {
		pa = rawurl
	} else {
		pa = rawurl[:i]
		query = rawurl[i+1:]
	}

	pathList := strings.Split(strings.Trim(pa, "/"), "/")
	var pathNames []string
	var segements []PathSegement
	if pa != "" {
		for idx := range pathList {
			if strings.HasPrefix(pathList[idx], ":") {
				name := strings.TrimPrefix(pathList[idx], ":")
				pathNames = append(pathNames, name)
				segements = append(segements, PathSegement{IsArgument: true, Value: name})
			} else {
				segements = append(segements, PathSegement{IsArgument: false, Value: pathList[idx]})
			}
		}
	}
	return segements, pathNames, parseQuery(query)
}

func parseQuery(query string) map[string]string {
	values := map[string]string{}
	for query != "" {
		key := query
		if i := strings.IndexByte(key, '&'); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := key
		if i := strings.IndexByte(key, '='); i >= 0 {
			key, value = key[:i], key[i+1:]
			if strings.HasPrefix(value, ":") {
				value = strings.TrimPrefix(value, ":")
			}
		}

		if value == "<none>" {
			values[key] = value
		} else {
			values[value] = key
		}
	}
	return values
}

func convertToStringLiteral(param Param, isArray ...bool) string {
	return convertToStringLiteral2("", param, isArray...)
}

var convertNS string

func convertToStringLiteral2(suffix string, param Param, isArray ...bool) string {
	name := param.Name.Name + suffix

	var typ string
	if len(isArray) > 0 && isArray[0] {
		typ = typePrint(ElemType(param.Typ))
	} else {
		typ = typePrint(param.Typ)
		if strings.HasPrefix(typ, "[]") {
			return convertNS + strings.TrimPrefix(typ, "[]") + "ArrayToString(" + name + ")"
		}
	}
	isFirst := true
	needWrap := false

retry:
	switch typ {
	case "string":
		return name
	case "*string":
		return "*" + param.Name.Name + suffix
	case "int", "int8", "int16", "int32":
		return "strconv.FormatInt(int64(" + name + "), 10)"
	case "*int", "*int8", "*int16", "*int32":
		return "strconv.FormatInt(int64(*" + name + "), 10)"
	case "int64":
		if needWrap {
			return "strconv.FormatInt(int64(" + name + "), 10)"
		}
		return "strconv.FormatInt(" + name + ", 10)"
	case "*int64":
		if needWrap {
			return "strconv.FormatInt(int64(*" + name + "), 10)"
		}
		return "strconv.FormatInt(*" + name + ", 10)"
	case "uint", "uint8", "uint16", "uint32":
		return "strconv.FormatUint(uint64(" + name + "), 10)"
	case "*uint", "*uint8", "*uint16", "*uint32":
		return "strconv.FormatUint(uint64(*" + name + "), 10)"
	case "uint64":
		if needWrap {
			return "strconv.FormatUint(uint64(" + name + "), 10)"
		}
		return "strconv.FormatUint(" + param.Name.Name + ", 10)"
	case "*uint64":
		if needWrap {
			return "strconv.FormatUint(uint64(*" + name + "), 10)"
		}
		return "strconv.FormatUint(*" + name + ", 10)"
	case "bool":
		return convertNS + "BoolToString(" + name + ")"
	case "*bool":
		return convertNS + "BoolToString(*" + name + ")"
	case "time.Time", "*time.Time":
		return name + ".Format(" + TimeFormat + ")"
	case "time.Duration", "*time.Duration":
		return name + ".String()"
	case "net.IP", "*net.IP":
		return name + ".String()"
	case "sql.NullTime":
		return name + ".Time.Format(" + TimeFormat + ")"
	case "sql.NullBool":
		return convertNS + "BoolToString(" + name + ".Bool)"
	case "sql.NullInt64":
		return "strconv.FormatInt(" + name + ".Int64, 10)"
	case "sql.NullUint64":
		return "strconv.FormatUint(" + name + ".Uint64, 10)"
	case "sql.NullString":
		return name + ".String"
	default:

		underlying := param.Method.Ctx.GetType(typ)
		if underlying != nil {
			if isFirst {
				isFirst = false
				needWrap = true
				typ = typePrint(underlying.Type)
				goto retry
			}
		}

		return "fmt.Sprint(" + name + ")"

		// err := errors.New(param.Method.Ctx.PostionFor(param.Method.Node.Pos()).String() + ": path param '" + param.Name.Name + "' of '" + param.Method.Name.Name + "' is unsupport type - " + typ)

		// log.Fatalln(err)
		// panic(err)
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
