package gengen

var beeConfig = map[string]interface{}{
	// "features.buildTag":     "bee-gen",
	"features.httpCodeWith": true,
	// "features.boolConvert":     "toBool({{.name}})",
	// "features.datetimeConvert": "toDatetime({{.name}})",
	"imports": map[string]string{
		"github.com/astaxie/beego":         "beego",
		"github.com/astaxie/beego/context": "beecontext",
	},

	"func_signature":      "func(ctx *beecontext.Context) ",
	"ctx_name":            "ctx",
	"ctx_type":            "*beecontext.Context",
	"route_party_name":    "*beego.Namespace",
	"path_param_format":   "Input.Param",
	"query_param_format":  "Input.Query",
	"read_body_format":    "json.Unmarshal({{.ctx}}.Input.CopyBody(4 * 1024), &{{.name}})",
	"bad_argument_format": "fmt.Errorf(\"argument %%q is invalid - %%q\", %s, %s, %s)",

	"ok_func_format":  "ctx.Output.SetStatus({{.statusCode}})\r\n    ctx.Output.JSON({{.data}}, false, false)\r\n    return",
	"err_func_format": "ctx.Output.SetStatus({{.errCode}})\r\n    ctx.WriteString({{.err}}.Error())\r\n    return",

	"reserved": map[string]string{
		"*http.Request":       "ctx.Request",
		"http.ResponseWriter": "ctx.Response",
		"context.Context":     "ctx.Request.Context()",
		"*beecontext.Context": "ctx",
		"*context.Context":    "ctx",
	},
	"method_mapping": map[string]string{
		"GET":     "Get",
		"POST":    "Post",
		"DELETE":  "Delete",
		"PUT":     "Put",
		"HEAD":    "Head",
		"OPTIONS": "Options",
		"PATCH":   "Patch",
		"ANY":     "Any",
	},
	"types": map[string]interface{}{
		"required": map[string]interface{}{},
		"optional": map[string]interface{}{},
	},
}
