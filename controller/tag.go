package controller

import (
	"fastdb-server/models/config"
	"fastdb-server/service"
	"github.com/gin-gonic/gin"
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
		c.JSON(200, err)
	} else {
		c.JSON(200, tags)
	}
}

func SelectById(c *gin.Context) {
	id := c.Param("id")
	tag := new(config.Tag)
	has, err := service.Engine.Id(id).Get(tag)
	if err != nil {
		c.JSON(200, err)
	} else {
		if has {
			c.JSON(200, tag)
		} else {
			c.JSON(200, nil)
		}
	}
}
