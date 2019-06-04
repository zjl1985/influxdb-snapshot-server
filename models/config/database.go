package config

type Database struct {
	Id   int `json:"id" xorm:"autoincr"`
	Code string `json:"code"`
	Name string `json:"name"`
}
