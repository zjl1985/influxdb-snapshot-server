package config

import "time"

type Tag struct {
    Id         int64     `json:"id" xorm:"autoincr pk INTEGER"`
    CreateTime time.Time `json:"createTime" xorm:"created"`
    Code       string    `json:"code"`
    Name       string    `json:"name"`
    Desc       string    `json:"desc"`
    Table      string    `json:"table"`
    Database   string    `json:"database"`
}
