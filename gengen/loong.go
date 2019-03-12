package gengen

var loongConfig = map[string]interface{}{
	// "features.buildTag":     "loong-gen",
	"features.httpCodeWith": false,
	// "features.boolConvert":     "toBool({{.name}})",
	// "features.datetimeConvert": "toDatetime({{.name}})",
	"imports": map[string]string{
		"github.com/runner-mei/loong": "",
	},

	"func_signature":      "func(ctx *loong.Context) error ",
	"ctx_name":            "ctx",
	"ctx_type":            "*loong.Context",
	"route_party_name":    "loong.Party",
	"path_param_format":   "Param",
	"query_param_format":  "QueryParam",
	"read_body_format":    "{{.ctx}}.Bind(&{{.name}})",
	"bad_argument_format": "loong.ErrBadArgument(%q, %s, %s)",
	"read_format":         "{{.ctx}}.{{.readMethodName}}(\"{{.name}}\")",
	"ok_func_format":      "return ctx.Return({{.statusCode}}, {{.data}})",
	"err_func_format":     "ctx.Error({{.err}})\r\n     return nil",

	"reserved": map[string]string{
		"*http.Request":       "ctx.Request()",
		"http.ResponseWriter": "ctx.Response().Writer",
		"context.Context":     "ctx.StdContext",
		"*loong.Context":      "ctx",
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
