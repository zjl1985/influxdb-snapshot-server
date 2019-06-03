package fastdb_server

type TagValue struct {
	Quality  int8    `json:"quality"`
	DataTime int64   `json:"dataTime"`
	Value    float32 `json:"value"`
	TagCode  string  `json:"tagCode"`
}

type Config struct {
	Delay        int
	Port         string
	Mode         string
	RedisAddress string
	RedisPwd     string
}