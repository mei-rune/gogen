package gengen

var loongConfig = map[string]interface{}{
	// "features.buildTag":     "loong-gen",
	"features.httpCodeWith": false,
	// "features.boolConvert":     "toBool({{.name}})",
	"features.datetimeConvert": "loong.ToDatetime({{.name}})",
	"imports": map[string]string{
		"github.com/runner-mei/loong": "",
	},

	"func_signature":        "func(ctx *loong.Context) error ",
	"ctx_name":              "ctx",
	"ctx_type":              "*loong.Context",
	"route_party_name":      "loong.Party",
	"required_param_format": "{{.ctx}}.Param(\"{{.name}}\")",
	"optional_param_format": "{{.ctx}}.QueryParam(\"{{.name}}\")",
	"read_body_format":      "{{.ctx}}.Bind(&{{.name}})",
	"bad_argument_format":   "loong.ErrBadArgument(\"%s\", %s, %s)",
	"read_format":           "{{.ctx}}.{{.readMethodName}}(\"{{.name}}\")",
	"ok_func_format": `{{- if eq .method "POST" -}} 
	return ctx.ReturnCreatedResult({{.data}})
	{{- else if eq .method "PUT" -}}
	return ctx.ReturnUpdatedResult({{.data}})
	{{- else if eq .method "DELETE" -}}
	return ctx.ReturnDeletedResult({{.data}})
	{{- else if eq .method "GET" -}}
	return ctx.ReturnQueryResult({{.data}})
	{{- else -}}
	return ctx.ReturnResult({{.statusCode}}, {{.data}})
	{{end}}`,
	"err_func_format": "return ctx.ReturnError({{.err}}{{if and .errCode .hasRealErrorCode}},{{.errCode}}{{end}})",

	"reserved": map[string]string{
		"url.Values":          "ctx.QueryParams()",
		"*http.Request":       "ctx.Request()",
		"http.ResponseWriter": "ctx.Response().Writer",
		"context.Context":     "ctx.StdContext",
		"*loong.Context":      "ctx",
	},
	"types": map[string]interface{}{
		"optional": map[string]interface{}{
			"[]string": map[string]interface{}{
				"name": "QueryParamArray",
			},
		},
	},
}

var loongClientConfig = map[string]interface{}{
	// "features.buildTag": "loong",
	//"features.httpCodeWith": true,
	// "features.boolConvert":     "toBool({{.name}})",
	// "features.datetimeConvert": "toDatetime({{.name}})",
	"imports": map[string]string{
		"github.com/runner-mei/loong/resty": "",
	},
}
