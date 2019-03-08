package gengen

import (
	"errors"
	"log"
	"strings"
)

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

func JoinPathSegments(segements []PathSegement, replace ReplaceFunc) string {
	if len(segements) == 0 {
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
	for idx := range pathList {
		if strings.HasPrefix(pathList[idx], ":") {
			name := strings.TrimPrefix(pathList[idx], ":")
			pathNames = append(pathNames, name)
			segements = append(segements, PathSegement{IsArgument: true, Value: name})
		} else {
			segements = append(segements, PathSegement{IsArgument: false, Value: pathList[idx]})
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
		values[value] = key
	}
	return values
}

func convertToStringLiteral(param Param) string {
	typ := typePrint(param.Typ)
	switch typ {
	case "string":
		return param.Name.Name
	case "int", "int8", "int16", "int32":
		return "strconv.FormatInt(int64(" + param.Name.Name + "), 10)"
	case "int64":
		return "strconv.FormatInt(" + param.Name.Name + ", 10)"
	case "uint", "uint8", "uint16", "uint32":
		return "strconv.FormatUint(uint64(" + param.Name.Name + "), 10)"
	case "uint64":
		return "strconv.FormatUint(" + param.Name.Name + ", 10)"
	case "boolean":
		return "BoolToString(" + param.Name.Name + ", 10)"
	default:
		log.Fatalln(errors.New("path param '" + param.Name.Name + "' is unsupport type - " + typ))
		panic("")
	}
}
