package controller

import (
    "fastdb-server/models"
    "fastdb-server/models/config"
    "fastdb-server/service"
    "github.com/gin-gonic/gin"
    client "github.com/influxdata/influxdb1-client/v2"
    log "github.com/sirupsen/logrus"
    "github.com/thoas/go-funk"
    "net/http"
    "strconv"
)

func SelectPage(c *gin.Context) {
    tags, total, err := TagPage(c)
    if err != nil {
        log.Error(err)
        c.JSON(http.StatusOK, models.Page{
            Total: 0,
            List:  tags,
        })
    } else {
        c.JSON(http.StatusOK, models.Page{
            Total: total,
            List:  tags,
        })
    }
}

func TagPage(c *gin.Context) ([]config.Tag, int64, error) {
    database := c.Param("database")
    code := c.Param("code")
    limit, _ := strconv.Atoi(c.Query("ps"))
    pi, _ := strconv.Atoi(c.Query("pi"))
    offset := (pi - 1) * limit
    tags := make([]config.Tag, 0)
    sqlSession := service.Engine.Where("database=?", database)
    if code != "" {
        sqlSession = service.Engine.Where("database=?",
            database).And("code like '%'||?||'%'", code)
    }
    err := sqlSession.Limit(limit, offset).Find(&tags)
    if err != nil {
        return tags, 0, err
    }

    sqlSession = service.Engine.Where("database=?", database)
    if code != "" {
        sqlSession = service.Engine.Where("database=?",
            database).And("code like '%'||?||'%'", code)
    }
    total, _ := sqlSession.Count(config.Tag{})
    return tags, total, nil
}

func SelectById(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 32)
    tag := new(config.Tag)
    has, err := service.Engine.Id(id).Get(tag)
    if err != nil {
        log.Error(err)
        c.JSON(http.StatusOK, err)
    } else {
        if has {
            c.JSON(http.StatusOK, tag)
        } else {
            c.JSON(http.StatusOK, nil)
        }
    }
}

func Create(c *gin.Context) {
    tag := new(config.Tag)
    _ = c.Bind(tag)
    _, err := service.Engine.InsertOne(tag)
    if err != nil {
        log.Error(err)
        c.JSON(http.StatusOK, models.Result{
            Success: false,
            Result:  "插入失败",
        })
    } else {
        c.JSON(http.StatusOK, models.Result{
            Success: true,
            Result:  "success",
        })
    }
}

func CreateList(c *gin.Context) {
    tags := make([]config.Tag, 0)
    err := c.Bind(&tags)
    if err != nil {
        log.Error(err)
        c.JSON(http.StatusOK, models.Result{
            Success: false,
            Result:  "数据校验失败",
        })
        return
    }

    if tags == nil || len(tags) == 0 {
        c.JSON(http.StatusOK, models.Result{
            Success: false,
            Result:  "没有上传数据",
        })
        return
    }

    needInsert := make([]config.Tag, 0)
    for _, tag := range tags {
        existTag := new(config.Tag)
        exist, _ := service.Engine.Where("code=?", tag.Code).And("database=?",
            tag.Database).And(`"table"=?`, "tag_value").Cols("id").Get(existTag)
        if exist {
            _, _ = service.Engine.ID(existTag.Id).Cols("name,desc").Update(tag)
        } else {
            needInsert = append(needInsert, tag)
        }
    }

    insert := funk.Chunk(needInsert, 150)
    for _, tags := range insert.([][]config.Tag) {
        _, _ = service.Engine.Insert(&tags)
    }

    c.JSON(http.StatusOK, models.Result{
        Success: true,
        Result:  "success",
    })
}

func Update(c *gin.Context) {
    tag := new(config.Tag)
    _ = c.Bind(tag)
    _, err := service.Engine.Id(tag.Id).Cols("name", "desc").Update(tag)
    if err != nil {
        log.Error(err)
        c.JSON(http.StatusOK, models.Result{
            Success: false,
            Result:  "更新失败",
        })
    } else {
        c.JSON(http.StatusOK, models.Result{
            Success: true,
            Result:  "success",
        })
    }
}

func Delete(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 32)
    tag := new(config.Tag)
    session := service.Engine.Id(id)
    _, _ = session.Get(tag)
    service.DropSeries(tag.Database, []string{tag.Code})
    _, err := service.Engine.Id(id).Delete(&config.Tag{})
    if err != nil {
        log.Error(err)
        c.JSON(http.StatusOK, models.Result{
            Success: false,
            Result:  "删除失败",
        })
    } else {
        c.JSON(http.StatusOK, models.Result{
            Success: true,
            Result:  "success",
        })
    }
}

func DeleteList(c *gin.Context) {
    database := c.Param("database")
    ids := make([]int, 0)
    _ = c.Bind(&ids)
    tags := make([]config.Tag, 0)
    err := service.Engine.In("id", ids).Find(&tags)
    codes := funk.Get(tags, "Code").([]string)
    service.DropSeries(database, codes)
    if err != nil {
        log.Error(err)
    }
    _, err = service.Engine.Where(&config.Tag{}).In("id", ids).Delete(&config.Tag{})
    if err != nil {
        log.Error(err)
        c.JSON(http.StatusOK, models.Result{
            Success: false,
            Result:  "删除失败",
        })
    } else {
        c.JSON(http.StatusOK, models.Result{
            Success: true,
            Result:  "success",
        })
    }
}

func Synchronize(c *gin.Context) {
    database := c.Param("database")
    tags := make([]config.Tag, 0)
    err := service.Engine.Where("database=?", database).And(`"table"=?`,
        "tag_value").Cols("id,code").Find(&tags)
    if err != nil {
        log.Error(err)
    }
    var exists = struct{}{}

    tagMap := make(map[string]struct{})
    oriTags := getReadTag(database)
    insertTags := make([]config.Tag, 0)
    for i := range tags {
        tagMap[tags[i].Code] = exists
    }
    for _, code := range oriTags {
        _, ok := tagMap[code]
        if !ok {
            insertTags = append(insertTags, config.Tag{
                Code:     code,
                Name:     code,
                Database: database,
                Table:    "tag_value",
            })
        }
    }
    insert := funk.Chunk(insertTags, 150)
    for _, tags := range insert.([][]config.Tag) {
        _, _ = service.Engine.Insert(&tags)
    }
    c.JSON(http.StatusOK, models.Result{
        Success: true,
        Result:  len(insertTags),
    })
}

func getReadTag(database string) []string {
    influxSql :=
        `show tag values on ` + database + ` with key="code"`;
    c, err := client.NewHTTPClient(client.HTTPConfig{
        Addr: service.MyConfig.FastDBAddress,
    })
    if err != nil {
        log.Error("Error creating InfluxDB Client: ", err.Error())
    }
    defer c.Close()
    q := client.NewQuery(influxSql, database, "")
    response, err := c.Query(q)
    tagCodes := make([]string, 0)

    if err == nil && response.Error() == nil {
        if len(response.Results[0].Series) > 0 {
            for i := range response.Results[0].Series {
                if response.Results[0].Series[i].Name == "tag_value" {
                    for _, value := range response.Results[0].Series[i].
                        Values {
                        tagCodes = append(tagCodes, value[1].(string))
                    }
                }
            }
        }
    } else {
        if err != nil {
            log.Error(err)
        }
        if response.Error() != nil {
            log.Error(response.Error())
        }
    }
    return tagCodes
}
