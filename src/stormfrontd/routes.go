// routes.go

package main

import (
	"stormfrontd/api"
	"stormfrontd/middleware"
)

func initializeRoutes() {
	daemonRoutes := router.Group("/api")
	{
		daemonRoutes.GET("/health", middleware.EnsureLocalhost(), api.GetHealth)
		daemonRoutes.POST("/deploy", middleware.EnsureLocalhost(), api.Deploy)
		daemonRoutes.DELETE("/destroy", middleware.EnsureLocalhost(), api.Destroy)
		daemonRoutes.POST("/join", middleware.EnsureLocalhost(), api.Join)
		daemonRoutes.POST("/restart", middleware.EnsureLocalhost(), api.Restart)
	}
}
