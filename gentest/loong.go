// +build loong

package main

import (
	"github.com/runner-mei/loong"
)

func main() {
	// Echo instance
	e := loong.New()

	// Routes
	InitStringSvc(e.Group("/test"), &StringSvcImpl{})
	InitStringSvcImpl(e.Group("/test2"), &StringSvcImpl{})

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
