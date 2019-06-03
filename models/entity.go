package models

type TagValue struct {
	Quality  int8    `json:"quality"`
	DataTime int64   `json:"dataTime"`
	Value    float32 `json:"value"`
	TagCode  string  `json:"tagCode"`
}
