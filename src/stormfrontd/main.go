// main.go

package main

import (
	"fmt"
	"stormfrontd/api"
	"stormfrontd/config"
	"stormfrontd/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func main() {
	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	config.LoadConfig()

	routerPort := ":" + strconv.Itoa(config.Config.DaemonPort)

	fmt.Printf("Running with port: %v\n", config.Config.DaemonPort)

	api.Healthy = true

	// Set the router as the default one provided by Gin
	router = gin.Default()

	// Initialize the routes
	initializeRoutes()

	err := utils.EnsureDataDirectory()

	if err != nil {
		panic(err)
	}

	// Start serving the application
	router.Run(routerPort)
}
