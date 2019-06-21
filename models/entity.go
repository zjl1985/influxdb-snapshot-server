package models

import "fastdb-server/models/config"

type TagValue struct {
    Quality  int    `json:"quality"`
    DataTime int64   `json:"dataTime"`
    Value    float64 `json:"value"`
    TagCode  string  `json:"tagCode"`
}

type Page struct {
    List  []config.Tag `json:"list"`
    Total int64        `json:"total"`
}
