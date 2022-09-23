package api

import (
	"log"
	"net/http"

	"stormfrontd/daemon"

	"github.com/gin-gonic/gin"
)

var Healthy = false

func Test(c *gin.Context) {
	remoteAddr := c.Request.RemoteAddr
	log.Printf("Remote Address: %v", remoteAddr)
	c.Status(http.StatusOK)
}

func GetHealth(c *gin.Context) {
	if Healthy {
		c.Status(http.StatusOK)
		return
	} else {
		c.Status(http.StatusServiceUnavailable)
	}
}

func Deploy(c *gin.Context) {
	err := daemon.Deploy()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func Destroy(c *gin.Context) {
	err := daemon.Destroy()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func Join(c *gin.Context) {
	var data map[string]interface{}
	c.BindJSON(&data)

	if _, ok := data["leader_host"]; !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing required 'leader_host' key in request payload"})
	}
	if _, ok := data["leader_port"]; !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing required 'leader_port' key in request payload"})
	}
	if _, ok := data["join_token"]; !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing required 'join_token' key in request payload"})
	}

	leaderHost := data["leader_host"]
	leaderPort := data["leader_port"]
	joinToken := data["join_token"]

	err := daemon.Join(leaderHost.(string), int(leaderPort.(float64)), joinToken.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func Restart(c *gin.Context) {
	err := daemon.Restart()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
