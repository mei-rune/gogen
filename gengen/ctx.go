package gengen

import "fmt"

type context struct {
	errDefined bool
}

func (c *context) SetErrorDefined() string {
	fmt.Println("====")
	c.errDefined = true
	return ""
}
func (c *context) IsErrorDefined() bool {
	return c.errDefined
}
