package influx

import (
    "fastdb-server/controller"
    "fastdb-server/models"
    "fastdb-server/models/config"
    "fastdb-server/service"
    "fmt"
    "github.com/360EntSecGroup-Skylar/excelize/v2"
    "github.com/gin-gonic/gin"
    client "github.com/influxdata/influxdb1-client/v2"
    log "github.com/sirupsen/logrus"
    "net/http"
    "sort"
    "strings"
    "time"
)

func WriteHistoryData(context *gin.Context) {
    writeInfluxData(context, false)
}

func WriteLiveData(context *gin.Context) {
    writeInfluxData(context, true)
}

func writeInfluxData(context *gin.Context, live bool) {
    database := context.Param("database")
    tagValues := make(config.TagSlice, 0)
    _ = context.Bind(&tagValues)
    sort.Stable(tagValues)

    c, err := client.NewHTTPClient(client.HTTPConfig{
        Addr: service.MyConfig.FastDBAddress,
        Username: service.MyConfig.FastUser,
        Password: service.MyConfig.FastPwd,
    })
    if err != nil {
        log.Error("Error creating InfluxDB Client: ", err.Error())
        context.JSON(http.StatusOK, models.Result{
            Success: false,
            Result:  err,
        })
        return
    }
    defer c.Close()

    bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
        Database:  database,
        Precision: "ms",
    })
    now := time.Now()
    for _, vtq := range tagValues {
        var insertTime time.Time
        if live {
            insertTime = now
        } else {
            insertTime = time.Unix(vtq.Time/1e3, 0)
        }
        pt, err := client.NewPoint(
            "tag_value",
            map[string]string{
                "code": vtq.Code,
            },
            map[string]interface{}{
                "value":   vtq.Value,
                "quality": vtq.Quality,
            },
            insertTime,
        )
        if err != nil {
            println("Error:", err.Error())
            continue
        }
        bp.AddPoint(pt)
    }
    err = c.Write(bp)
    if err != nil {
        log.Error(err)
        context.JSON(http.StatusOK, models.Result{
            Success: false,
            Result:  "插入历史数据失败",
        })
    } else {
        context.JSON(http.StatusOK, models.Result{
            Success: true,
            Result:  "success",
        })
    }
}

func GetLiveData(c *gin.Context) {
    tags, total, err := controller.TagPage(c)
    if err != nil {
        log.Error(err)
        c.JSON(http.StatusOK, models.Page{
            List:  tags,
            Total: total,
        })
        return
    }

    for i := range tags {
        value, ok := controller.LiveDataMap.Load(tags[i].Code)
        if ok {
            tags[i].Time = value.(models.TagValue).DataTime
            tags[i].Quality = value.(models.TagValue).Quality
            tags[i].Value = value.(models.TagValue).Value
        }
    }

    c.JSON(http.StatusOK, models.Page{
        List:  tags,
        Total: total,
    })
    tags = nil
}

type historyQuery struct {
    Type       string   `form:"type"`
    BeginTime  int64    `form:"beginTime"`
    EndTime    int64    `form:"endTime"`
    Interval   string   `form:"interval"`
    Tags       []string `form:"tags"`
    Symbol     string   `form:"symbol"`
    JudgeValue float64  `form:"judgeValue"`
}

func ImportData(context *gin.Context) {
    fileHeader, _ := context.FormFile("file")
    log.Info(fileHeader.Filename)
    file, _ := fileHeader.Open()
    defer file.Close()
    f, err := excelize.OpenReader(file)
    if err != nil {
        context.JSON(http.StatusBadRequest, gin.H{
            "status": "error",
        })
        log.Error(err)
        return
    }
    log.Info(f.GetSheetName(0))
    context.JSON(http.StatusOK, gin.H{
        "status": "success",
    })
}

func GeHistoryData(context *gin.Context) {
    database := context.Param("database")
    var query historyQuery
    _ = context.ShouldBindQuery(&query)
    rangeStr := ""
    if context.Query("judgeValue") != "" {
        rangeStr = fmt.Sprintf("and value %s%f", query.Symbol, query.JudgeValue)
    }
    log.Debug(query)
    tagStr := "and code=~ /^" + strings.Join(query.Tags, "$|^") + "$/"
    fullSql := `SELECT * FROM "tag_value" WHERE time>=%dms AND time<=%dms %s %s GROUP BY code`
    groupSql := `SELECT %s(value) as "value" FROM "tag_value" WHERE time>=%dms
AND time<=%dms %s %s GROUP BY time(%s),code fill(previous)`
    var sql string
    switch query.Type {
    case "full":
        sql = fmt.Sprintf(fullSql, query.BeginTime, query.EndTime, tagStr, rangeStr)
        break
    case "groupby":
        sql = fmt.Sprintf(groupSql, "FIRST", query.BeginTime,
            query.EndTime, tagStr, rangeStr,
            query.Interval)
        break
    case "max":
        sql = fmt.Sprintf(groupSql, "MAX", query.BeginTime,
            query.EndTime, tagStr, rangeStr, query.Interval)
        break
    case "min":
        sql = fmt.Sprintf(groupSql, "MIN", query.BeginTime,
            query.EndTime, tagStr, rangeStr, query.Interval)
        break
    default:
        break
    }
    log.Debug(sql)
    q := client.NewQuery(sql, database, "ms")
    c, err := client.NewHTTPClient(client.HTTPConfig{
        Addr: service.MyConfig.FastDBAddress,
        Username: service.MyConfig.FastUser,
        Password: service.MyConfig.FastPwd,
    })
    if err != nil {
        log.Error("Error creating InfluxDB Client: ", err.Error())
    }
    defer c.Close()
    response, err := c.Query(q)
    var m []map[string]interface{}
    if err == nil && response.Error() == nil {
        m = GroupBy(response.Results[0])
        context.JSON(http.StatusOK, models.Result{
            Success: true,
            Result:  m,
        })
    } else {
        if response != nil{
            log.Error(response.Error())
            context.JSON(http.StatusOK, models.Result{
                Success: false,
                Result:  response.Results,
            })
            return
        }
    }

    m = nil
}

func UserDefineQuery(context *gin.Context) {
    database := context.Param("database")
    query := context.Query("queryString")
    log.Debug(query)
    c, err := client.NewHTTPClient(client.HTTPConfig{
        Addr: service.MyConfig.FastDBAddress,
        Username: service.MyConfig.FastUser,
        Password: service.MyConfig.FastPwd,
    })
    if err != nil {
        log.Error("Error creating InfluxDB Client: ", err.Error())
    }
    defer c.Close()
    q := client.NewQuery(query, database, "ms")
    response, err := c.Query(q)
    var m []map[string]interface{}
    if err == nil && response.Error() == nil {
        m = GroupBy(response.Results[0])
    }
    context.JSON(http.StatusOK, m)
}

func DeleteData(context *gin.Context) {
    database := context.Param("database")
    tagValues := make([]config.Tag, 0)
    _ = context.Bind(&tagValues)

    c, err := client.NewHTTPClient(client.HTTPConfig{
        Addr: service.MyConfig.FastDBAddress,
        Username: service.MyConfig.FastUser,
        Password: service.MyConfig.FastPwd,
    })
    if err != nil {
        log.Error("Error creating InfluxDB Client: ", err.Error())
        context.JSON(http.StatusOK, models.Result{
            Success: false,
            Result:  "删除数据失败",
        })
        return
    }
    defer c.Close()
    var q client.Query

    for _, tagValue := range tagValues {
        sql := fmt.Sprintf("DELETE FROM tag_value WHERE time=%dms and code='%s'",
            tagValue.Time, tagValue.Code)
        q = client.NewQuery(sql, database, "")
        response, err := c.Query(q)
        if err != nil {
            log.Error(err)
            context.JSON(http.StatusOK, models.Result{
                Success: false,
                Result: fmt.Sprintf("删除编码%s,时间%dms的数据失败",
                    tagValue.Code, tagValue.Time),
            })
            return
        }
        if response.Error() != nil {
            log.Error(response.Error())
            context.JSON(http.StatusOK, models.Result{
                Success: false,
                Result: fmt.Sprintf("删除编码%s,时间%dms的数据失败",
                    tagValue.Code, tagValue.Time),
            })
            return
        }
    }

    context.JSON(http.StatusOK, gin.H{
        "success": true,
        "result":  "success",
    })
}
