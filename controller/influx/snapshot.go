package influx

import (
    "fastdb-server/models"
    "fastdb-server/service"
    "fmt"
    "github.com/gin-gonic/gin"
    client "github.com/influxdata/influxdb1-client/v2"
    "github.com/sirupsen/logrus"
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
    returnData := GetSnapshotData(codes)
    c.JSON(200, returnData)
}

func GetSnapshotData(codes []string) []models.TagValue {
    returnData := make([]models.TagValue, len(codes))
    influxdClient, err := client.NewHTTPClient(client.HTTPConfig{
        Addr:     service.MyConfig.FastDBAddress,
        Username: service.MyConfig.FastUser,
        Password: service.MyConfig.FastPwd,
    })
    if err != nil {
        logrus.Error("Error creating InfluxDB Client: ", err.Error())
    }
    defer influxdClient.Close()
    for _, code := range codes {
        //LiveDataMap.Load(code)
        if value, ok := LiveDataMap.Load(code); ok {
            returnData = append(returnData, value.(models.TagValue))
        } else {
            whenNoCodeInLiveDataMap(code, influxdClient)
        }
    }
    return returnData
}

func whenNoCodeInLiveDataMap(code string, cl client.Client) {
    lastSql := `SELECT LAST(value) AS "value", quality FROM "tag_value" WHERE time>now()-2h and time<=now( )+5m AND code='%s' GROUP BY code`
    sql := fmt.Sprintf(lastSql, code)
    q := client.NewQuery(sql, "telegraf", "ms")
    response, err := cl.Query(q)
    var m []map[string]interface{}
    if err == nil && response.Error() == nil {
        m = GroupBy(response.Results[0])
    } else {
        if response != nil {
            logrus.Error(response.Error())
            return
        }
    }
    logrus.Info(convertToTagValue(m))
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
