package gengen

type context struct {
	errDefined bool
}

func (c *context) SetErrorDefined() string {
	c.errDefined = true
	return ""
}
func (c *context) IsErrorDefined() bool {
	return c.errDefined
}
