package controller

import (
    "fastdb-server/models"
    "fastdb-server/service"
    "github.com/gin-gonic/gin"
    "strconv"
    "strings"
    "sync"
    "time"
    "unsafe"
)

var LiveDataMap sync.Map
var delayTime int64

func InitSnapshot(delay int) {
    delayTime = int64(delay * 1e9)
}

func Snapshot(c *gin.Context) {
    codes := make([]string, 0)
    _ = c.Bind(&codes)
    returnData := make([]models.TagValue, 0)
    for _, code := range codes {
        LiveDataMap.Load(code)
        if value, ok := LiveDataMap.Load(code); ok {
            returnData = append(returnData, value.(models.TagValue))
        }
    }
    c.JSON(200, returnData)
}

func whenNoCodeInLiveDataMap(code string){

}

func InfluxSub(c *gin.Context) {
    body, _ := c.GetRawData()
    go processString(string(body))
    body = nil
    c.String(200, "ok")
}

func processString(body string) {
    lines := service.LinesToMapList(body)
    delaySub := (time.Now().UnixNano() + delayTime) / 1e6
    for _, line := range lines {
        tv := buildTagValue(line)
        if unsafe.Sizeof(tv) == 0 {
            continue
        }
        if value, ok := LiveDataMap.Load(tv.TagCode); ok {
            if tv.DataTime > value.(models.TagValue).DataTime && tv.
                DataTime < delaySub {
                //setRedis(tv)
                LiveDataMap.Store(tv.TagCode, tv)
            }
        } else {
            LiveDataMap.Store(tv.TagCode, tv)
            //setRedis(tv)
        }
        tv = models.TagValue{}
    }
    lines = nil
}

func buildTagValue(line map[string]interface{}) models.TagValue {
    sec := line["timestamp"].(int64)
    if sec < time.Now().UnixNano()-delayTime {
        return models.TagValue{}
    }
    tags := line["tags"].(map[string]string)
    fields := line["fields"].(map[string]string)

    value, _ := strconv.ParseFloat(fields["value"], 64)
    quality, _ := strconv.Atoi(strings.Replace(fields["quality"], "i", "", 1))
    code := tags["code"]

    return models.TagValue{
        TagCode:  code,
        Value:    value,
        DataTime: sec / 1e6,
        Quality:  quality,
    }
}
