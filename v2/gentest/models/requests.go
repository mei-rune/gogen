package models

import (
	"database/sql"
	"net/url"
	"time"
)

type Request struct {
	ID   int64  `json:"id" xorm:"id pk autoincr"`
	Uuid string `json:"uuid" xorm:"uuid null"`
}

type RequestQuery struct {
	OperatorID               int64
	CreatorID                int64
	RequesterID              int64
	RequestTypeIDs           []int64
	RequestTypeNames         []string
	NameLike                 string
	CurrentStatus            int64
	IsUnclosed               sql.NullBool
	IsOverdued               sql.NullBool
	IsSuspend                sql.NullBool
	StartAt, EndAt           time.Time
	OverdueStart, OverdueEnd time.Time
	Settings                 map[string]string `gogen:"true" swaggerignore:"true"`
	Args                     map[string]string `json:"ttargstt" gogen:"true" swaggerignore:"true"`
	url.Values               `gogen:"true" swaggerignore:"true"`
}
