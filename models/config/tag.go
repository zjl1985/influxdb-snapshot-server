package config

import "time"

type Tag struct {
	Id         int64     `json:"id" xorm:"autoincr pk INTEGER"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Desc       string    `json:"desc"`
	Table      string    `json:"table"`
	CreateTime time.Time `json:"createTime" xorm:"created"`
	Database   string    `json:"database"`
}
