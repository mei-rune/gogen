package gengen

type context struct {
	errDefined   bool
	parentInited bool
	needQuery    bool
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

func (c *context) SetNeedQuery() string {
	c.needQuery = true
	return ""
}
func (c *context) IsNeedQuery() bool {
	return c.needQuery
}
