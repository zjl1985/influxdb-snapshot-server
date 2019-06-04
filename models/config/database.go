package config

type Database struct {
	Id   int `json:"id" xorm:"autoincr pk INTEGER"`
	Code string `json:"code"`
	Name string `json:"name"`
}
