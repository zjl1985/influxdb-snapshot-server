package config

type Tag struct {
	Id         string `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	CreateTime string `json:"createTime"`
	Database   string `json:"database"`
}
