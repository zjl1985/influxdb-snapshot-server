package service

import (
	client "github.com/influxdata/influxdb1-client/v2"
	log "github.com/sirupsen/logrus"
)

func CreateDataBase(name string) bool {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: MyConfig.FastDBAddress,
	})
	if err != nil {
		log.Error("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()
	q := client.NewQuery("CREATE DATABASE "+name, "", "")
	response, err := c.Query(q);
	if err == nil && response.Error() == nil {
		log.Info(response.Results)
		return true
	} else {
		log.Error(err)
		log.Error(response.Results)
		return false
	}
}

func DropDataBase(name string) bool {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: MyConfig.FastDBAddress,
	})
	if err != nil {
		log.Error("Error creating InfluxDB Client: ", err.Error())
	}
	defer c.Close()
	q := client.NewQuery("DROP DATABASE "+name, "", "")

	response, err := c.Query(q);
	if err == nil && response.Error() == nil {
		log.Info(response.Results)
		return true
	} else {
		log.Error(err)
		log.Error(response.Results)
		return false
	}
}


