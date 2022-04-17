//go:build chi
// +build chi

package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
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
	r := chi.NewRouter()

	var svc StringSvc

	// Routes
	r.Route("/test", func(r chi.Router) {
		InitStringSvc(r, svc)
	})
	// r.Route("/test2", func(r chi.Router) {
	// 	InitStringSvcImpl(r, svc)
	// })

	// Start server
	http.ListenAndServe(":3000", r)
}
