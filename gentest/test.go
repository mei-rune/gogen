package gentest

type StringSvc interface {
	// @http.GET(path="/concat")
	Concat(a, b string) (string, error)

	// @http.GET(path="/concat1")
	Concat1(a, b *string) (string, error)

	// @http.GET(path="/concat2/:a/:b")
	Concat2(a, b string) (string, error)

	// @http.GET(path="/concat3/:a/:b")
	Concat3(a, b *string) (string, error)

	// @http.GET(path="/sub")
	Sub(a string, start int64) (string, error)

	// @http.POST(path="/save/:a", data="b")
	Save(a, b string) (string, error)

	// @http.POST(path="/save2/:a", data="b")
	Save2(a, b *string) (string, error)

	// @http.GET(path="/add/:a/:b")
	Add(a, b int) (int, error)

	// @http.GET(path="/add2/:a/:b")
	Add2(a, b *int) (int, error)

	// @http.GET(path="/add3")
	Add3(a, b *int) (int, error)

	Misc() string
}
