package config

type Database struct {
    Id     int     `json:"id" xorm:"autoincr pk INTEGER"`
    Code   string  `json:"code"`
    Name   string  `json:"name"`
    Disk   float64 `json:"disk" xorm:"-"`
    TagNum int     `json:"tagNum" xorm:"-"`
}
