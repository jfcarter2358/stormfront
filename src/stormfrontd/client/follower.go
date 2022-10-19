package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jfcarter2358/ceresdb-go/connection"
)

func RegisterFollower(c *gin.Context) {
	var follower StormfrontNode
	c.BindJSON(&follower)

	currentTime := time.Now()
	Client.Updated = currentTime.Format(time.RFC3339)

	_, err := connection.Query(`get record stormfront.node | jq '.[0].succession[.[0].succession|length] |= . + "baz"' | put record stormfront.node`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func DeregisterFollower(c *gin.Context) {
	var follower StormfrontNode
	c.BindJSON(&follower)

	currentTime := time.Now()
	Client.Updated = currentTime.Format(time.RFC3339)

	_, err := connection.Query(fmt.Sprintf(`get record stormfront.node | jq 'del(.[0].succession[] | select(.host == "%s"))' | put record stormfront.node`, follower.Host))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
