package config

type Database struct {
    Id     int     `json:"id" xorm:"autoincr pk INTEGER"`
    TagNum int     `json:"tagNum" xorm:"-"`
    Disk   float64 `json:"disk" xorm:"-"`
    Code   string  `json:"code"`
    Name   string  `json:"name"`
}
