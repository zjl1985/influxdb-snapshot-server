package models

import "fastdb-server/models/config"

type TagValue struct {
    Quality  int     `json:"quality"`
    DataTime int64   `json:"dataTime"`
    Value    float64 `json:"value"`
    TagCode  string  `json:"tagCode"`
}

type Page struct {
    List  []config.Tag `json:"list"`
    Total int64        `json:"total"`
}

type TagValueHistory struct {
    Quality interface{} `json:"quality"`
    Time    interface{} `json:"time"`
    Value   interface{} `json:"value"`
    Code    string      `json:"code"`
}

type TagValueHistorySlice []TagValueHistory

func (s TagValueHistorySlice) Len() int           { return len(s) }
func (s TagValueHistorySlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s TagValueHistorySlice) Less(i, j int) bool { return s[i].Code < s[j].Code }
