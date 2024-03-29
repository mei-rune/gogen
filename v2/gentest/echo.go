//go:build echo
// +build echo

package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/runner-mei/gogen/v2/gentest/docs" // docs is generated by Swag CLI
	echoSwagger "github.com/swaggo/echo-swagger"
)

func httpCodeWith(err error) int {
	return http.StatusInternalServerError
}

func NewBadArgument(err error, method, param string) error {
	return err
}

func ToInt64Array(ss []string) ([]int64, error) {
	var results = make([]int64, len(ss))
	for _, s := range ss {
		i64, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		results = append(results, i64)
	}
	return results, nil
}

func ToDatetimes(ss []string) ([]time.Time, error) {
	var results = make([]time.Time, len(ss))
	for _, s := range ss {
		i64, err := time.Parse(s, time.RFC3339)
		if err != nil {
			return nil, err
		}
		results = append(results, i64)
	}
	return results, nil
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	var svc StringSvc

	// Routes
	InitStringSvc(e.Group("/test"), svc)
	// InitStringSvcImpl(e.Group("/test2"), &StringSvcImpl{})

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
