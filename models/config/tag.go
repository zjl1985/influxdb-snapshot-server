package config

import "time"

type Tag struct {
	Id         int       `json:"id" xorm:"autoincr"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Desc       string    `json:"desc"`
	Table      string    `json:"table"`
	CreateTime time.Time `json:"createTime" xorm:"created"`
	Database   string    `json:"database"`
}
