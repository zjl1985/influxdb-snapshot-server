package influx

import (
	"fastdb-server/service"
	"fmt"
	"github.com/gin-gonic/gin"
	client "github.com/influxdata/influxdb1-client/v2"
	log "github.com/sirupsen/logrus"
)

type online struct {
	Online bool `json:"online"`
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

func Status(context *gin.Context) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: service.MyConfig.FastDBAddress,
	})
	if err != nil {
		log.Error("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()
	q := client.NewQuery("SHOW DIAGNOSTICS", "_internal", "");
	response, err := c.Query(q);
	var m map[string]interface{}
	if err == nil && response.Error() == nil {
		fmt.Println(response.Results)
		m = flat(response.Results)
	}
	context.JSON(200, m)
}

// 扁平化数据结构
func flat(result []client.Result) map[string]interface{} {
	m := make(map[string]interface{})
	for _, ser := range result[0].Series {
		for index := range ser.Columns {
			m[ser.Columns[index]] = ser.Values[0][index]
		}
	}
	return m
}

func diagnostics() {

}
