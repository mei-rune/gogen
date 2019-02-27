package gentest

type StringSvc interface {
	// @http.GET(path="/concat")
	Concat(a, b string) string

	// @http.GET(path="/sub")
	Sub(a string, start int64) string
}
