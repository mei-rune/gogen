// +build echo

package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	InitStringSvc(e.Group("/test"), &StringSvcImpl{})
	InitStringSvcImpl(e.Group("/test2"), &StringSvcImpl{})

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
