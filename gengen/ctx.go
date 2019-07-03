package gengen

type context struct {
	errDefined bool
}

func (c *context) SetErrorDefined() {
	c.errDefined = true
}
func (c *context) IsErrorDefined() bool {
	return c.errDefined
}
