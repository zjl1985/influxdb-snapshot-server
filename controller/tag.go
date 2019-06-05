package controller

import (
	"fastdb-server/models"
	"fastdb-server/models/config"
	"fastdb-server/service"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strconv"
)

func SelectPage(c *gin.Context) {
	database := c.Param("database")
	code := c.Param("code")
	limit, _ := strconv.Atoi(c.Query("ps"))
	pi, _ := strconv.Atoi(c.Query("pi"))
	offset := (pi - 1) * limit
	tags := make([]config.Tag, 0)
	sqlSession := service.Engine.Where("database=?", database)
	if code != "" {
		sqlSession.Where("code like '%'||?||'%'", code)
	}
	err := sqlSession.Limit(limit, offset).Find(&tags)
	if err != nil {
		log.Error(err)
		c.JSON(200, err)
	} else {
		c.JSON(200, tags)
	}
}

func SelectById(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 32)
	tag := new(config.Tag)
	has, err := service.Engine.Id(id).Get(tag)
	if err != nil {
		log.Error(err)
		c.JSON(200, err)
	} else {
		if has {
			c.JSON(200, tag)
		} else {
			c.JSON(200, nil)
		}
	}
}

func Create(c *gin.Context) {
	tag := new(config.Tag)
	_ = c.Bind(tag)
	_, err := service.Engine.InsertOne(tag)
	if err != nil {
		log.Error(err)
		c.JSON(200, models.Result{
			Success: false,
			Result:  "插入失败",
		})
	} else {
		c.JSON(200, models.Result{
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
		c.JSON(200, models.Result{
			Success: false,
			Result:  "数据校验失败",
		})
		return
	}

	if tags == nil || len(tags) == 0 {
		c.JSON(200, models.Result{
			Success: false,
			Result:  "没有上传数据",
		})
		return
	}
	sql := `replace into tag(code,name,desc,"table",database,create_time) values (?,?,?,'tag_value',?,datetime('now', 'localtime'))`
	for _, tag := range tags {
		_, _ = service.Engine.Exec(sql, tag.Code, tag.Name, tag.Desc, tag.Database)
	}

	c.JSON(200, models.Result{
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
		c.JSON(200, models.Result{
			Success: false,
			Result:  "更新失败",
		})
	} else {
		c.JSON(200, models.Result{
			Success: true,
			Result:  "success",
		})
	}
}

func Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 32)
	//tag := new(config.Tag)
	//todo 加入删除数据库测点
	_, err := service.Engine.Id(id).Delete(&config.Tag{})
	if err != nil {
		log.Error(err)
		c.JSON(200, models.Result{
			Success: false,
			Result:  "删除失败",
		})
	} else {
		c.JSON(200, models.Result{
			Success: true,
			Result:  "success",
		})
	}
}

func DeleteList(c *gin.Context) {
	ids := make([]int, 0)
	_ = c.Bind(&ids)
	tags := make([]config.Tag, 0)
	//tag := new(config.Tag)
	session := service.Engine.In("id", ids)
	err := session.Find(&tags)
	//todo 加入删除数据库测点
	if err != nil {
		log.Error(err)
	}
	_, err = service.Engine.Where(&config.Tag{}).In("id", ids).Delete(&config.Tag{})
	if err != nil {
		log.Error(err)
	}
	c.JSON(200, models.Result{
		Success: true,
		Result:  &tags,
	})
}
