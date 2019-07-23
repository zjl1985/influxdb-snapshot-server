package config

import "time"

type Tag struct {
    Quality    int       `json:"quality" xorm:"-"`
    Id         int64     `json:"id" xorm:"autoincr pk INTEGER"`
    Value      float64   `json:"value" xorm:"-"`
    Time       int64     `json:"time" xorm:"-"`
    CreateTime time.Time `json:"createTime" xorm:"created"`
    Code       string    `json:"code"`
    Name       string    `json:"name"`
    Desc       string    `json:"desc"`
    Table      string    `json:"table"`
    Database   string    `json:"database"`
}

type TagSlice []Tag

func (s TagSlice) Len() int           { return len(s) }
func (s TagSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s TagSlice) Less(i, j int) bool { return s[i].Code < s[j].Code }
