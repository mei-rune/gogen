package main

import (
	"context"
	"database/sql"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/runner-mei/gogen/v2/gentest/models"
)

type ErrStringSvc interface {

	// @Summary get files
	// @Description get files
	// @ID Get1
	// @Accept  json
	// @Produce  json
	// @Router /files1 [get]
	Get1() (list []string, total int64, err error)


	// @Summary get files
	// @Description get files
	// @ID Get2
	// @Accept  json
	// @Produce  json
	// @Router /files2 [get]
	Get2() (list []string, total int64, err error)

	// @Summary get files
	// @Description get files
	// @ID Get3
	// @Accept  json
	// @Produce  json
	// @Router /files3 [get]
	Get3() (err error)


	// @Summary get files
	// @Description get files
	// @ID Get4
	// @Param   id      query   int     true  "Some ID" Format(int32)
	// @Accept  json
	// @Produce  json
	// @Router /files4 [get]
	Get4(id int) (err error)
}
