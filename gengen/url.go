package gengen

import (
	"strings"
)

// parse parses a URL from a string in one of two contexts. If
// viaRequest is true, the URL is assumed to have arrived via an HTTP request,
// in which case only absolute URLs or path-absolute relative URLs are allowed.
// If viaRequest is false, all forms of relative URLs are allowed.
func parseURL(rawurl string) (string, []string, map[string]string) {
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
	for idx := range pathList {
		if strings.HasPrefix(pathList[idx], ":") {
			pathNames = append(pathNames, strings.TrimPrefix(pathList[idx], ":"))
		}
	}
	return pa, pathNames, parseQuery(query)
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
