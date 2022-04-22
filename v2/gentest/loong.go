//go:build loong
// +build loong

package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/runner-mei/loong"
	"github.com/runner-mei/log"
)

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
	e := loong.New()

	// Routes
	InitStringSvc(e.Group("/test"), &StringSvcImpl{})
	InitStringSvcImpl(e.Group("/test2"), &StringSvcImpl{})

	// Start server
	err := http.ListenAndServe("", e)
	if err != nil {
		if err != http.ErrServerClosed {
			e.Logger.Fatal("server failure:", log.Error(err))
		}
	}
}
