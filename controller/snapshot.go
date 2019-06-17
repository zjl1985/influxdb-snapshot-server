package controller

import (
    "fastdb-server/models"
    "github.com/gin-gonic/gin"
    "strconv"
    "strings"
    "time"
)

var LiveDataMap map[string]*models.TagValue
var delayTime int64

func InitSnapshot(delay int) {
    delayTime = int64(delay * 1e9)
    LiveDataMap = make(map[string]*models.TagValue)
}

func Snapshot(c *gin.Context) {
    codes := make([]string, 0)
    _ = c.Bind(&codes)
    returnData := make([]models.TagValue, 0)
    for _, code := range codes {
        if _, ok := LiveDataMap[code]; ok {
            returnData = append(returnData, *LiveDataMap[code])
        }
    }
    c.JSON(200, returnData)
}

func InfluxSub(c *gin.Context) {
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
        if _, ok := LiveDataMap[tv.TagCode]; ok {
            if tv.DataTime > LiveDataMap[tv.TagCode].DataTime && tv.DataTime < delaySub {
                LiveDataMap[tv.TagCode].Value = tv.Value
                LiveDataMap[tv.TagCode].DataTime = tv.DataTime
                //setRedis(tv)
            }
        } else {
            LiveDataMap[tv.TagCode] = tv
            //setRedis(tv)
        }
        tv = nil
    }
    lines = nil
}

func buildTagValue(line string) *models.TagValue {
    line = strings.Replace(line, "tag_value,code=", "", -1)
    items := strings.Split(line, " ")
    code := items[0]
    sec, _ := strconv.ParseInt(items[len(items)-1], 10, 64)
    if sec < time.Now().UnixNano()-delayTime {
        return nil
    }
    values := strings.Split(items[1], ",")
    var value float64
    var quality int8 = 0
    if strings.HasPrefix(values[0], "value=") {
        val := strings.Replace(values[0], "value=", "", 1)
        value, _ = strconv.ParseFloat(val, 64)
        if len(values) > 1 {
            q := strings.Replace(values[1], "quality=", "", 1)
            qv, _ := strconv.ParseInt(q, 10, 8)
            quality = int8(qv)
        }
    } else {
        val := strings.Replace(strings.Replace(values[1], "value=", "", 1), "i", "", 1)
        value, _ = strconv.ParseFloat(val, 64)

        q := strings.Replace(strings.Replace(values[0], "quality=", "", 1), "i", "", 1)
        qv, _ := strconv.ParseInt(q, 10, 8)
        quality = int8(qv)

    }
    return &models.TagValue{
        TagCode:  code,
        Value:    value,
        DataTime: sec / 1e6,
        Quality:  quality,
    }
}
