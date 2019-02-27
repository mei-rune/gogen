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




type StringSvcImpl struct {
}

// @http.GET(path="/concat")
func (svc *StringSvcImpl)	Concat(a, b string) (string, error) {
	return "", nil
}

// @http.GET(path="/concat1")
func (svc *StringSvcImpl)	Concat1(a, b *string) (string, error){
	return "", nil
}

// @http.GET(path="/concat2/:a/:b")
func (svc *StringSvcImpl)	Concat2(a, b string) (string, error){
	return "", nil
}

// @http.GET(path="/concat3/:a/:b")
func (svc *StringSvcImpl)	Concat3(a, b *string) (string, error){
	return "", nil
}

// @http.GET(path="/sub")
func (svc *StringSvcImpl)	Sub(a string, start int64) (string, error){
	return "", nil
}

// @http.POST(path="/save/:a", data="b")
func (svc *StringSvcImpl)	Save(a, b string) (string, error){
	return "", nil
}

// @http.POST(path="/save2/:a", data="b")
func (svc *StringSvcImpl) Save2(a, b *string) (string, error){
	return "", nil
}

// @http.GET(path="/add/:a/:b")
func (svc *StringSvcImpl)	Add(a, b int) (int, error){
	return 0, nil
}

// @http.GET(path="/add2/:a/:b")
func (svc *StringSvcImpl)	Add2(a, b *int) (int, error){
	return 0, nil
}

// @http.GET(path="/add3")
func (svc *StringSvcImpl)	Add3(a, b *int) (int, error){
	return 0, nil
}

func (svc *StringSvcImpl)	Misc() string{
	return "", nil
}
