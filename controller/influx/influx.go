package influx

import (
    "encoding/json"
    "fastdb-server/models"
    "fastdb-server/models/config"
    "fastdb-server/service"
    "fmt"
    "github.com/gin-gonic/gin"
    client "github.com/influxdata/influxdb1-client/v2"
    "github.com/mitchellh/mapstructure"
    log "github.com/sirupsen/logrus"
    "net/http"
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
        Addr:     service.MyConfig.FastDBAddress,
        Username: service.MyConfig.FastUser,
        Password: service.MyConfig.FastPwd,
    })
    if err != nil {
        log.Error("Error creating InfluxDB Client: ", err.Error())
    }
    defer c.Close()
    _, _, err = c.Ping(500)
    if err != nil {
        context.JSON(http.StatusOK, online{
            Online: false,
        })
    } else {
        context.JSON(http.StatusOK, online{
            Online: true,
        })
    }
}

func CreatAdmin() {
    log.Info("使用超级管理员登陆")
    c, err := client.NewHTTPClient(client.HTTPConfig{
        Addr:     service.MyConfig.FastDBAddress,
        Username: "super_admin",
        Password: "gxems#21",
    })
    if err != nil {
        log.Fatal("连接数据库引擎失败", err.Error())
    }
    defer c.Close()
    q := client.NewQuery("create user super_admin with password 'gxems#21' with all privileges", "_internal", "")
    _, _ = c.Query(q)

    q = client.NewQuery("SHOW USERS", "_internal", "")
    response, err := c.Query(q)
    if err == nil && response.Error() == nil {
        databases := GroupBy(response.Results[0])
        var user string
        for _, v := range databases {
            if strings.Compare(service.MyConfig.FastUser,
                v["user"].(string)) == 0 {
                user = service.MyConfig.FastUser
                break
            }
        }
        if user != "" {
            log.Info("用户存在,修改密码")
            q = client.NewQuery(fmt.Sprintf(
                `SET PASSWORD FOR "%s" = '%s'`,
                user, service.MyConfig.FastPwd), "_internal", "")
            response, err = c.Query(q)
            if err == nil && response.Error() == nil {
                log.Info("修改密码成功")
            }
        } else {
            log.Info("用户不存在,创建用户")
            q = client.NewQuery(fmt.Sprintf(
                "create user %s with password '%s' with all privileges",
                service.MyConfig.FastUser,
                service.MyConfig.FastPwd), "_internal", "")
            response, err = c.Query(q)
            if err == nil && response.Error() == nil {
                log.Info("创建用户成功")
            }
        }
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
        Addr:     service.MyConfig.FastDBAddress,
        Username: service.MyConfig.FastUser,
        Password: service.MyConfig.FastPwd,
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
        for index := range databases {
            if v, ok := dataMap[databases[index].Code]; ok {
                databases[index].Disk = v
            }
            totalDisk += databases[index].Disk

            if v, ok := tagMap[databases[index].Code]; ok {
                databases[index].TagNum = v
            }
            totalNum += databases[index].TagNum
        }
        status.TotalDisk = totalDisk
        status.TotalNum = totalNum
    } else {
        if err != nil {
            log.Error(err)
        }
        if response.Error() != nil {
            log.Error(response.Error())
        }
    }

    context.JSON(http.StatusOK, infoStatus{
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
    databases := GroupBy(result)
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

func GroupBy(result client.Result) []map[string]interface{} {
    rows := make([]map[string]interface{}, 0)
    for _, ser := range result.Series {
        for i := range ser.Values {
            m := make(map[string]interface{})
            if ser.Tags != nil {
                for k, v := range ser.Tags {
                    m[k] = v
                }
            }
            for index := range ser.Columns {
                m[ser.Columns[index]] = ser.Values[i][index]
            }
            rows = append(rows, m)
        }
    }
    return rows
}

func convertToTagValue(inputs []map[string]interface{}) []models.TagValue {
    result := make([]models.TagValue, 0)
    err := mapstructure.Decode(inputs, &result)
    if err != nil {
        log.Error(err)
    }
    return result
}

func convertToTagValueHistory(inputs []map[string]interface{}) []models.TagValueHistory {
    result := make(models.TagValueHistorySlice, 0)
    err := mapstructure.Decode(inputs, &result)
    if err != nil {
        log.Error(err)
    }
    return result
}
