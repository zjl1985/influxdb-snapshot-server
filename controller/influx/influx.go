package influx

import (
    "encoding/json"
    "fastdb-server/models/config"
    "fastdb-server/service"
    "fmt"
    "github.com/gin-gonic/gin"
    client "github.com/influxdata/influxdb1-client/v2"
    log "github.com/sirupsen/logrus"
    "strings"
)

type online struct {
    Online bool `json:"online"`
}

type Status struct {
    TotalDisk float64           `json:"totalDisk"`
    TotalNum  int               `json:"totalNum"`
    DataBases []config.Database `json:"dataBases"`
}

type infoStatus struct {
    Status Status                 `json:"status"`
    Info   map[string]interface{} `json:"info"`
}

func ConnectionState(context *gin.Context) {
    c, err := client.NewHTTPClient(client.HTTPConfig{
        Addr: service.MyConfig.FastDBAddress,
    })
    if err != nil {
        log.Error("Error creating InfluxDB Client: ", err.Error())
    }
    defer c.Close()
    _, _, err = c.Ping(500)
    if err != nil {
        context.JSON(200, online{
            Online: false,
        })
    } else {
        context.JSON(200, online{
            Online: true,
        })
    }
}

func StatusInfo(context *gin.Context) {
    databases := make([]config.Database, 0)
    status := new(Status)
    err := service.Engine.Where(&config.Database{}).Find(&databases)
    if err != nil {
        log.Error(err)
    }
    status.DataBases = databases
    whereQuery := make([]string, len(databases))
    for index, database := range databases {
        whereQuery[index] = database.Code
    }
    baseStr := `"database"='` + strings.Join(whereQuery, `' or "database"='`) + "'"
    c, err := client.NewHTTPClient(client.HTTPConfig{
        Addr: service.MyConfig.FastDBAddress,
    })
    if err != nil {
        log.Error("Error creating InfluxDB Client: ", err.Error())
    }
    defer c.Close()
    q := client.NewQuery("SHOW DIAGNOSTICS", "_internal", "")
    response, err := c.Query(q)
    var m map[string]interface{}
    if err == nil && response.Error() == nil {
        m = flat(response.Results[0])
    }
    statusSql := fmt.Sprintf(`select sum(*) from (select last(
diskBytes)  from "tsm1_filestore" where %s group by "database",
id) group by "database"`, baseStr)
    q = client.NewQuery(statusSql, "_internal", "")
    response, err = c.Query(q)
    if err == nil && response.Error() == nil {
        var totalDisk float64
        totalNum := 0
        dataMap := getDataBaseMap(response.Results[0])
        tagMap := getTagMap()
        log.Info(dataMap)
        for index := range databases {
            if _, ok := dataMap[databases[index].Code]; ok {
                databases[index].Disk = dataMap[databases[index].Code]
            }
            totalDisk += databases[index].Disk

            if _, ok := tagMap[databases[index].Code]; ok {
                databases[index].TagNum = tagMap[databases[index].Code]
            }
            totalNum += databases[index].TagNum
        }
        log.Info(databases)
        status.TotalDisk = totalDisk
        status.TotalNum = totalNum
    }

    context.JSON(200, infoStatus{
        Info:   m,
        Status: *status,
    })
}

type TagMap struct {
    Database string
    Count    int
}

func getTagMap() map[string]int {
    sql := `select COUNT(id) AS "count",database from tag where database in (
        select code from database) GROUP BY database`
    rows := make([]TagMap, 0)
    tagMap := make(map[string]int)
    err := service.Engine.SQL(sql).Find(&rows)
    if err != nil {
        log.Error(err)
    }
    for _, row := range rows {
        tagMap[row.Database] = row.Count
    }
    return tagMap
}

func getDataBaseMap(result client.Result) map[string]float64 {
    dataMap := make(map[string]float64)
    databases := groupBy(result)
    for _, database := range databases {
        disk, _ := database["sum_last"].(json.Number).Float64()
        dataMap[database["database"].(string)] = disk / (1024 * 1024)
    }
    return dataMap
}

// 扁平化数据结构
func flat(result client.Result) map[string]interface{} {
    m := make(map[string]interface{})
    for _, ser := range result.Series {
        if ser.Tags != nil {
            for k, v := range ser.Tags {
                m[k] = v
            }
        }
        for index := range ser.Columns {
            m[ser.Columns[index]] = ser.Values[0][index]
        }
    }
    return m
}

func groupBy(result client.Result) []map[string]interface{} {
    rows := make([]map[string]interface{}, 0)
    for _, ser := range result.Series {
        m := make(map[string]interface{})
        if ser.Tags != nil {
            for k, v := range ser.Tags {
                m[k] = v
            }
        }
        for index := range ser.Columns {
            m[ser.Columns[index]] = ser.Values[0][index]
        }
        rows = append(rows, m)
    }
    return rows
}

//func goodLook(result []client.Result) []map[string]interface{} {
//    r := make([]map[string]interface{}, 0)
//    for _, item := range result {
//        r = append(r, groupBy(item))
//    }
//    return r
//}

func diagnostics() {

}
