package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
	"strings"
	"time"
)

type tagValue struct {
	Code    string  `json:"code"`
	Value   float64 `json:"value"`
	Date    int64   `json:"date"`
	Quality int     `json:"quality"`
}

type config struct {
	Spec      string
	Count     int
	TableName string
	Host      string
	Port      int
}

var liveDataMap map[string]*tagValue
var ExpirationTime int64

func main() {
	gin.SetMode(gin.ReleaseMode)
	liveDataMap = make(map[string]*tagValue)
	ExpirationTime = 5 * 60 * 1e9
	r := gin.Default()
	r.POST("/live", hello)
	r.POST("/bye/write", influxSub)
	log.Println("Starting v2")
	_ = r.Run(":1210")
}

func hello(c *gin.Context) {
	arr := make([]string, 0)
	_ = c.Bind(&arr)
	fmt.Println(len(arr))
	returnData := make([]tagValue, 0)
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
	fmt.Println(len(lines))
	for _, line := range lines {
		if !strings.HasPrefix(line, "tag_value,code=") {
			continue
		}
		tv := buildTagValue(line)
		if tv == nil {
			continue
		}
		if _, ok := liveDataMap[tv.Code]; ok {
			if tv.Date > liveDataMap[tv.Code].Date && tv.Date < time.Now().UnixNano()-ExpirationTime {
				liveDataMap[tv.Code].Value = tv.Value
				liveDataMap[tv.Code].Date = tv.Date
			}
		} else {
			liveDataMap[tv.Code] = tv
		}
	}
}

func buildTagValue(line string) *tagValue {
	line = strings.Replace(line, "tag_value,code=", "", -1)
	items := strings.Split(line, " ")
	code := items[0]
	sec, _ := strconv.ParseInt(items[len(items)-1], 10, 64)
	if sec < time.Now().UnixNano()-ExpirationTime {
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
	return &tagValue{
		Code:  code,
		Value: value,
		Date:  sec / 1e6,
	}
}
