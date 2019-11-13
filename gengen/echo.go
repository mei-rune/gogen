package gengen

var echoConfig = map[string]interface{}{
	"features.buildTag":     "echo-gen",
	"features.httpCodeWith": true,
	// "features.boolConvert":     "toBool({{.name}})",
	// "features.datetimeConvert": "toDatetime({{.name}})",
	"imports": map[string]string{
		"github.com/labstack/echo": "",
	},

	"func_signature":        "func(ctx echo.Context) error ",
	"ctx_name":              "ctx",
	"ctx_type":              "echo.Context",
	"route_party_name":      "*echo.Group",
	"required_param_format": "{{.ctx}}.Param(\"{{.name}}\")",
	"optional_param_format": "{{.ctx}}.QueryParam(\"{{.name}}\")",
	"read_body_format":      "{{.ctx}}.Bind(&{{.name}})",
	"bad_argument_format":   "fmt.Errorf(\"argument %%q is invalid - %%q\", \"%s\", %s, %s)",
	"read_format":           "{{.ctx}}.{{.readMethodName}}(\"{{.name}}\")",
	"ok_func_format":        "return ctx.JSON({{.statusCode}}, {{.data}})",
	"err_func_format":       "ctx.Error({{.err}})\r\n     return nil",

	"reserved": map[string]string{
		"url.Values":          "ctx.QueryParams()",
		"*http.Request":       "ctx.Request()",
		"http.ResponseWriter": "ctx.Response().Writer",
		"io.Writer":           "ctx.Response().Writer",
		"context.Context":     "ctx.Request().Context()",
		"echo.Context":        "ctx",
	},
	"types": map[string]interface{}{
		"optional": map[string]interface{}{
			"[]string": map[string]interface{}{
				"format": "{{.ctx}}.QueryParamArray(\"{{.name}}\")",
			},
		},
		"required": map[string]interface{}{
			"[]string": map[string]interface{}{
				"format": "{{.ctx}}.QueryParamArray(\"{{.name}}\")",
			},
		},
	},
}
