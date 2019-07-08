package gengen

type context struct {
	errDefined   bool
	parentInited bool
}

func (c *context) SetErrorDefined() string {
	c.errDefined = true
	return ""
}
func (c *context) IsErrorDefined() bool {
	return c.errDefined
}

func (c *context) SetParentInited() string {
	c.parentInited = true
	return ""
}
func (c *context) IsParentInited() bool {
	return c.parentInited
}
