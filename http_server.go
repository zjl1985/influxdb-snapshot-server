package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/common-nighthawk/go-figure"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	jsoniter "github.com/json-iterator/go"
	"github.com/robfig/cron"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type TagValue struct {
	TagCode  string  `json:"tagCode"`
	Value    float32 `json:"value"`
	DataTime int64   `json:"dataTime"`
	Quality  int8    `json:"quality"`
}

type Config struct {
	Delay        int
	Port         string
	Mode         string
	RedisAddress string
	RedisPwd     string
}

var liveDataMap map[string]*TagValue
var delayTime int64
var config *Config
var client *redis.Client
var enableRedis bool
var redisKey = "fastDBSnapshot"

func main() {
	if _, err := toml.DecodeFile("config.conf", &config); err != nil {
		log.Fatal(err)
	}
	enableRedis = false
	if config.RedisAddress != "" {
		client = redis.NewClient(&redis.Options{
			Addr:     config.RedisAddress,
			Password: config.RedisPwd, // no password set
			DB:       0,
		})
		c := cron.New()
		_ = c.AddFunc("*/30 * * * * ?", func() {
			checkRedis()
		})
		c.Start()
	}
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)
	gin.SetMode(config.Mode)

	liveDataMap = make(map[string]*TagValue)
	//data, err := client.HGetAll(redisKey).Result()
	//if err == nil {
	//	fmt.Println(data)
	//}
	delayTime = int64(config.Delay * 1e9)
	r := gin.Default()
	r.POST("/snapshot", snapshot)
	r.POST("/snapshot/redis", snapshotRedis)
	r.POST("/snapshot/write", influxSub)
	myFigure := figure.NewFigure("FastDB", "", true)
	myFigure.Print()
	_ = r.Run(config.Port)
}

func checkRedis() {
	_, err := client.Ping().Result()
	if err == nil {
		enableRedis = true
	} else {
		enableRedis = false
		fmt.Println(err)
	}
}

func snapshot(c *gin.Context) {
	codes := make([]string, 0)
	_ = c.Bind(&codes)
	returnData := make([]TagValue, 0)
	for _, code := range codes {
		if _, ok := liveDataMap[code]; ok {
			returnData = append(returnData, *liveDataMap[code])
		}
	}
	c.JSON(200, returnData)
}

func snapshotRedis(c *gin.Context) {
	codes := make([]string, 0)
	_ = c.Bind(&codes)
	returnData := make([]TagValue, 0)
	result, err := client.HMGet(redisKey, codes...).Result()
	if err != nil {
		c.JSON(500, err)
		return
	}
	var tag TagValue
	for _, item := range result {
		var jsonBlob = []byte(item.(string))
		_ = jsoniter.Unmarshal(jsonBlob, &tag)
		returnData = append(returnData, tag)
	}

	c.JSON(200, returnData)
}

func influxSub(c *gin.Context) {
	body, _ := c.GetRawData()
	processString(string(body))
	body = nil
	c.String(200, "ok")
}

func processString(body string) {
	lines := strings.Split(body, "\n")
	delaySub := (time.Now().UnixNano() + delayTime) / 1e6
	for _, line := range lines {
		if !strings.HasPrefix(line, "tag_value,code=") {
			continue
		}
		tv := buildTagValue(line)
		if tv == nil {
			continue
		}
		if _, ok := liveDataMap[tv.TagCode]; ok {
			if tv.DataTime > liveDataMap[tv.TagCode].DataTime && tv.DataTime < delaySub {
				go setRedis(tv)
				liveDataMap[tv.TagCode].Value = tv.Value
				liveDataMap[tv.TagCode].DataTime = tv.DataTime
			}
		} else {
			go setRedis(tv)
			liveDataMap[tv.TagCode] = tv
		}
	}
	lines = nil
}

func setRedis(tv *TagValue) {
	if enableRedis {
		jsonBytes, _ := jsoniter.Marshal(tv)
		client.HSet(redisKey, tv.TagCode, jsonBytes)
		//client.Set(tv.Code, jsonBytes, 0)
	}
}

func buildTagValue(line string) *TagValue {
	line = strings.Replace(line, "tag_value,code=", "", -1)
	items := strings.Split(line, " ")
	code := items[0]
	sec, _ := strconv.ParseInt(items[len(items)-1], 10, 64)
	if sec < time.Now().UnixNano()-delayTime {
		return nil
	}
	values := strings.Split(items[1], ",")
	var value float32
	var quality int8 = 0
	if strings.HasPrefix(values[0], "value=") {
		val := strings.Replace(values[0], "value=", "", 1)
		value64, _ := strconv.ParseFloat(val, 32)
		value = float32(value64)
		if len(values) > 1 {
			q := strings.Replace(values[1], "quality=", "", 1)
			qv, _ := strconv.ParseInt(q, 10, 8)
			quality = int8(qv)
		}
	} else {
		val := strings.Replace(values[1], "value=", "", 1)
		value64, _ := strconv.ParseFloat(val, 32)
		value = float32(value64)

		q := strings.Replace(values[0], "quality=", "", 1)
		qv, _ := strconv.ParseInt(q, 10, 8)
		quality = int8(qv)

	}
	return &TagValue{
		TagCode:  code,
		Value:    value,
		DataTime: sec / 1e6,
		Quality:  quality,
	}
}
