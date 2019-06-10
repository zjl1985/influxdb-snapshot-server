package influx

import (
    "fastdb-server/controller"
    "fastdb-server/models"
    "fastdb-server/models/config"
    "fastdb-server/service"
    "github.com/gin-gonic/gin"
    client "github.com/influxdata/influxdb1-client/v2"
    log "github.com/sirupsen/logrus"
    "time"
)

func WriteHistoryData(context *gin.Context) {
    database := context.Param("database")
    tagValues := make([]config.Tag, 0)
    _ = context.Bind(&tagValues)

    c, err := client.NewHTTPClient(client.HTTPConfig{
        Addr: service.MyConfig.FastDBAddress,
    })
    if err != nil {
        log.Error("Error creating InfluxDB Client: ", err.Error())
        context.JSON(200, models.Result{
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

    for _, vtq := range tagValues {
        pt, err := client.NewPoint(
            "tag_value",
            map[string]string{
                "code": vtq.Code,
            },
            map[string]interface{}{
                "value":   vtq.Value,
                "quality": vtq.Quality,
            },
            time.Unix(vtq.Time/1e3, 0),
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
        context.JSON(200, models.Result{
            Success: false,
            Result:  "插入历史数据失败",
        })
    } else {
        context.JSON(200, models.Result{
            Success: true,
            Result:  "success",
        })
    }
}

func GetLiveData(c *gin.Context) {
    tags, total, err := controller.TagPage(c)
    if err != nil {
        log.Error(err)
        c.JSON(200, models.Page{
            List:  tags,
            Total: total,
        })
        return
    }

    for i := range tags {
        value, ok := controller.LiveDataMap[tags[i].Code]
        if ok {
            tags[i].Time = value.DataTime
            tags[i].Quality = value.Quality
            tags[i].Value = value.Value
        }
    }

    c.JSON(200, models.Page{
        List:  tags,
        Total: total,
    })
}
