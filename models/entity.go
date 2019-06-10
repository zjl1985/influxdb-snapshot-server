package models

type TagValue struct {
    Quality  int8    `json:"quality"`
    DataTime int64   `json:"dataTime"`
    Value    float32 `json:"value"`
    TagCode  string  `json:"tagCode"`
}

type Vtq struct {
    Quality int8    `json:"quality"`
    Time    int64   `json:"time"`
    Value   float32 `json:"value"`
    Code    string  `json:"code"`
}
