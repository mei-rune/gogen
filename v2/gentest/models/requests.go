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
	OperatorID               int64 `json:"operator_id"`
	CreatorID                int64 `json:"creator_id"`
	RequesterID              int64 `json:"requester_id"`
	RequestTypeIDs           []int64 `json:"request_type_ids"`
	RequestTypeNames         []string `json:"request_type_names"`
	NameLike                 string  `json:"name_like"`
	CurrentStatus            int64 `json:"current_status"`
	IsUnclosed               sql.NullBool `json:"is_unclosed"`
	IsOverdued               sql.NullBool `json:"is_overdued"`
	IsSuspend                sql.NullBool `json:"is_suspend"`
	StartAt           time.Time `json:"start_at"`
	EndAt           time.Time `json:"end_at"`
	OverdueStart time.Time  `json:"overdue_start"`
	OverdueEnd time.Time `json:"overdue_end"`
	Settings                 map[string]string `json:"settings"`
	Args                     map[string]string `json:"ttargstt"`
	url.Values
}
