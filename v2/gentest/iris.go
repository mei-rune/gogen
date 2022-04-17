//go:build iris
// +build iris

package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/kataras/iris/v12"
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
	app := iris.New()

	var svc CaseSvc

	// Routes
	InitCaseSvc(app.WWW().Party("/test"), svc)

	// Start server
	app.Listen(":8080")
}
