package main

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/common-nighthawk/go-figure"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/robfig/cron"
	"log"
	"strconv"
	"strings"
	"time"
)

type TagValue struct {
	Code    string  `json:"code"`
	Value   float64 `json:"value"`
	Date    int64   `json:"date"`
	Quality int     `json:"quality"`
}

type Config struct {
	Delay        int
	Port         string
	Mode         string
	RedisAddress string
}

var liveDataMap map[string]*TagValue
var DelayTime int64
var config *Config
var client *redis.Client
var enableRedis bool

func main() {

	if _, err := toml.DecodeFile("config.conf", &config); err != nil {
		log.Fatal(err)
	}
	enableRedis = false
	if config.RedisAddress != "" {
		client = redis.NewClient(&redis.Options{
			Addr:     config.RedisAddress,
			Password: "", // no password set
			DB:       1,  // use default DB
		})
		c := cron.New()
		_ = c.AddFunc("*/10 * * * * ?", func() {
			fmt.Println("do job")
			checkRedis()
		})
		c.Start()
	}

	gin.SetMode(config.Mode)

	liveDataMap = make(map[string]*TagValue)
	DelayTime = int64(config.Delay * 1e9)
	r := gin.Default()
	r.POST("/snapshot", snapshot)
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
	arr := make([]string, 0)
	_ = c.Bind(&arr)
	returnData := make([]TagValue, 0)
	for _, code := range arr {
		if _, ok := liveDataMap[code]; ok {
			returnData = append(returnData, *liveDataMap[code])
		}
	}
	c.JSON(200, returnData)
}

func influxSub(c *gin.Context) {
	body, _ := c.GetRawData()
	processString(string(body))
	c.String(200, "ok")
}

func processString(body string) {
	lines := strings.Split(body, "\n")
	delaySub := (time.Now().UnixNano() - DelayTime) / 1e6
	for _, line := range lines {
		if !strings.HasPrefix(line, "tag_value,code=") {
			continue
		}
		tv := buildTagValue(line)
		if tv == nil {
			continue
		}
		if _, ok := liveDataMap[tv.Code]; ok {
			if tv.Date > liveDataMap[tv.Code].Date && tv.Date < delaySub {
				go setRedis(tv)
				liveDataMap[tv.Code].Value = tv.Value
				liveDataMap[tv.Code].Date = tv.Date
			}
		} else {
			go setRedis(tv)
			liveDataMap[tv.Code] = tv
		}
	}
}

func setRedis(tv *TagValue) {
	if enableRedis {
		jsonBytes, _ := json.Marshal(tv)
		client.Set(tv.Code, jsonBytes, 0)
	}
}

func buildTagValue(line string) *TagValue {
	line = strings.Replace(line, "tag_value,code=", "", -1)
	items := strings.Split(line, " ")
	code := items[0]
	sec, _ := strconv.ParseInt(items[len(items)-1], 10, 64)
	if sec < time.Now().UnixNano()-DelayTime {
		return nil
	}
	values := strings.Split(items[1], ",")
	var value float64
	if strings.HasPrefix(values[0], "value=") {
		val := strings.Replace(values[0], "value=", "", 1)
		value, _ = strconv.ParseFloat(val, 32)
	} else {
		val := strings.Replace(values[1], "value=", "", 1)
		value, _ = strconv.ParseFloat(val, 32)
	}
	return &TagValue{
		Code:  code,
		Value: value,
		Date:  sec / 1e6,
	}
}
