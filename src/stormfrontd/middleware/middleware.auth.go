// middleware.auth.go

package middleware

import (
	"log"
	"net/http"
	"strings"

	"stormfrontd/client/auth"
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

func CheckTokenAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		splitToken := strings.Split(token, "Bearer ")
		if len(splitToken) != 2 {
			token := c.Request.Header.Get("X-Stormfront-API")

			status := auth.VerifyAPIToken(token)
			if status != http.StatusOK {
				c.Status(status)
				return
			}
		} else {
			token = splitToken[1]

			status := auth.VerifyAccessToken(token)
			if status != http.StatusOK {
				c.Status(status)
				return
			}
		}

		c.Next()
	}
}
