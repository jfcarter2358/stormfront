package client

import (
	"net/http"
	"stormfrontd/client/lightning"

	"github.com/gin-gonic/gin"
)

func GetBolt(c *gin.Context) {
	boltId := c.Param("id")

	bolt, err := lightning.GetBolt(boltId)

	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, bolt)
}

func PostBolt(c *gin.Context) {
	var boltConstructor lightning.BoltConstructor
	c.BindJSON(&boltConstructor)

	bolt, idx := lightning.CreateBolt(boltConstructor.Command)

	go lightning.RunBolt(&lightning.Bolts[idx])

	c.JSON(http.StatusOK, bolt)
}
