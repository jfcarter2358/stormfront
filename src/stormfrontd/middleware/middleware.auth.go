// middleware.auth.go

package middleware

import (
	"log"
	"net/http"
	"strings"

	"stormfrontd/config"
	"stormfrontd/utils"

	"github.com/gin-gonic/gin"
)

func EnsureLocalhost() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.Config.RestrictRequestHost {
			remoteAddr := c.Request.RemoteAddr
			host := strings.Split(remoteAddr, ":")[0]
			if !utils.Contains(config.Config.AllowedIPs, host) {
				log.Printf("Invalid control request from IP %v", host)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}
	}
}
