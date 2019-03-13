package gengen

var ginConfig = map[string]interface{}{
	// "features.buildTag":     "gin",
	"features.httpCodeWith": true,
	// "features.boolConvert":     "toBool({{.name}})",
	// "features.datetimeConvert": "toDatetime({{.name}})",
	"imports": map[string]string{
		"github.com/gin-gonic/gin": "",
	},

	"func_signature":      "func(ctx *gin.Context) ",
	"ctx_name":            "ctx",
	"ctx_type":            "*gin.Context",
	"route_party_name":    "gin.IRouter",
	"path_param_format":   "Param",
	"query_param_format":  "Query",
	"read_body_format":    "{{.ctx}}.Bind(&{{.name}})",
	"bad_argument_format": "fmt.Errorf(\"argument %%q is invalid - %%q\", %s, %s, %s)",

	"ok_func_format":  "ctx.JSON({{.statusCode}}, {{.data}})\r\n    return",
	"err_func_format": "ctx.String({{.errCode}}, {{.err}}.Error())\r\n    return",

	"reserved": map[string]string{
		"*http.Request":       "ctx.Request",
		"http.ResponseWriter": "ctx.Writer",
		"context.Context":     "ctx.Request.Context()",
		"*gin.Context":        "ctx",
	},
	"method_mapping": map[string]string{
		// "GET":     "Get",
		// "POST":    "Post",
		// "DELETE":  "Delete",
		// "PUT":     "Put",
		// "HEAD":    "Head",
		// "OPTIONS": "Options",
		// "PATCH":   "Patch",
		"ANY": "Any",
	},
	"types": map[string]interface{}{
		// "required": map[string]interface{}{
		// 	"int": map[string]interface{}{
		//     "name": "IntParam",
		//   },
		// },
		// "optional": map[string]interface{}{
		// 	"int": map[string]interface{}{
		//     "name": "QueryIntParam",
		//   },
		// },
	},
}
