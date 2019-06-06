package database

import (
	"fastdb-server/models"
	"fastdb-server/models/config"
	"fastdb-server/service"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strconv"
)

func SelectAll(c *gin.Context) {
	databases := make([]config.Database, 0)
	err := service.Engine.Where(&config.Database{}).Find(&databases)
	if err != nil {
		log.Error(err)
	}
	c.JSON(200, &databases)
}

func Select(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 32)
	database := new(config.Database)
	has, err := service.Engine.ID(id).Get(database)
	if err != nil {
		log.Error(err)
	}
	if has {
		c.JSON(200, database)
	} else {
		c.JSON(200, nil)
	}
}

func Create(c *gin.Context) {
	database := new(config.Database)
	_ = c.Bind(database)
	has, err := service.Engine.Where("code=?", database.Code).Get(&config.Database{})
	if err != nil {
		log.Error(err)
	}

	if has {
		c.JSON(200, models.Result{
			Success: false,
			Result:  "编码重复",
		})
		return
	}
	result := service.CreateDataBase(database.Code)
	if result {
		_, err = service.Engine.InsertOne(database)
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
	} else {
		c.JSON(200, models.Result{
			Success: false,
			Result:  "插入失败",
		})
	}
}

func Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 32)
	session := service.Engine.ID(id)
	database := new(config.Database)
	has, _ := session.Get(database)
	if has {
		result := service.DropDataBase(database.Code)
		if result {
			_, err := service.Engine.ID(id).Delete(&config.Database{})
			if err != nil {
				log.Error(err)
				c.JSON(200, models.Result{
					Success: false,
					Result:  "删除数据库失败",
				})
			} else {
				c.JSON(200, models.Result{
					Success: true,
					Result:  "success",
				})
			}
		} else {
			c.JSON(200, models.Result{
				Success: false,
				Result:  "删除数据库失败",
			})
		}
	} else {
		c.JSON(200, models.Result{
			Success: false,
			Result:  "没有找到对应数据库",
		})
	}
}

func Update(c *gin.Context) {
	database := new(config.Database)
	_ = c.Bind(database)
	_, err := service.Engine.ID(database.Id).Cols("name").Update(database)
	if err != nil {
		log.Error(err)
		c.JSON(200, models.Result{
			Success: false,
			Result:  "更新数据库失败",
		})
	} else {
		c.JSON(200, models.Result{
			Success: true,
			Result:  "success",
		})
	}
}
