package gengen

var chiConfig = map[string]interface{}{
	// "features.buildTag":     "loong-gen",
	"features.httpCodeWith": false,
	// "features.boolConvert":     "toBool({{.name}})",
	// "features.datetimeConvert": "chi.ToDatetime({{.name}})",
	"imports": map[string]string{
		"net/http":                     "",
		"github.com/go-chi/chi":        "",
		"github.com/go-chi/render":     "",
		"github.com/runner-mei/errors": "",
	},

	"func_head_str":         "queryParams := r.URL.Query()",
	"func_signature":        "func(w http.ResponseWriter, r *http.Request) ",
	"ctx_name":              "r",
	"ctx_type":              "*http.Request",
	"route_party_name":      "chi.Router",
	"required_param_format": "chi.URLParam({{.ctx}}, \"{{.name}}\")",
	"optional_param_format": "queryParams.Get(\"{{.name}}\")",
	"read_body_format":      "render.Decode({{.ctx}}, &{{.name}})",
	"bad_argument_format":   "errors.BadArgument(\"%s\", %s, %s)",
	"ok_func_format": `{{if .noreturn}}
  return
  {{- else if eq .method "POST" -}} 
	render.JSON(w, r, {{.data}})
  return
	{{- else if eq .method "PUT" -}}
	render.JSON(w, r, {{.data}})
  return
	{{- else if eq .method "DELETE" -}}
	render.JSON(w, r, {{.data}})
  return
	{{- else if eq .method "GET" -}}
	render.JSON(w, r, {{.data}})
  return
	{{- else -}}
  {{- if .statusCode }}
		render.Status(r, {{.statusCode}})
  {{- end}}
	render.JSON(w, r, {{.data}})
  return
	{{end}}`,

	"plain_text_func_format": `{{if .noreturn}}
  return
  {{- else if eq .method "POST" -}} 
  render.PlainText(w, r, {{.data}})
  return
  {{- else if eq .method "PUT" -}}
  render.PlainText(w, r, {{.data}})
  return
  {{- else if eq .method "DELETE" -}}
  render.PlainText(w, r, {{.data}})
  return
  {{- else if eq .method "GET" -}}
  render.PlainText(w, r, {{.data}})
  return
  {{- else -}}
  {{- if .statusCode }}
    render.Status(r, {{.statusCode}})
  {{- end}}
  render.PlainText(w, r, {{.data}})
  return
  {{end}}`,
	"err_func_format": `if he, ok := {{.err}}.(errors.HTTPError); ok {
    render.Status(r, he.HTTPCode())
  } else {
  {{- if and .errCode .hasRealErrorCode}}
    render.Status(r, {{.errCode}})
  {{- else}}
    render.Status(r, http.StatusInternalServerError)
  {{- end}}
  }
  render.JSON(w, r, {{.err}})
  return`,

	"reserved": map[string]string{
		"url.Values":          "r.URL.Query()",
		"*http.Request":       "r",
		"http.ResponseWriter": "w",
		"context.Context":     "r.Context()",
	},
	"types": map[string]interface{}{
		"optional": map[string]interface{}{
			"[]string": map[string]interface{}{
				"format": "queryParams[{{.name}}]",
			},
		},
		"required": map[string]interface{}{
			"[]string": map[string]interface{}{
				"format": "queryParams[{{.name}}]",
			},
		},
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
}
