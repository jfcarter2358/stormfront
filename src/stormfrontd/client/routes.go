package client

import (
	"stormfrontd/middleware"
)

func InitializeRoutes(clientType string) {
	Client.Router.Use(middleware.CORSMiddleware())
	apiRoutes := Client.Router.Group("/api")
	{
		apiRoutes.GET("/health", GetHealth)
		apiRoutes.GET("/state", GetState)
		apiRoutes.POST("/update/succession", UpdateFollowerSuccession)
		apiRoutes.POST("/register", RegisterFollower)
		apiRoutes.POST("/deregister", DeregisterFollower)
		apiRoutes.GET("/nodes", GetNodes)
	}
	authRoutes := Client.Router.Group("/auth")
	{
		authRoutes.GET("/check", CheckAccessToken)
		authRoutes.GET("/join", GetJoinCommand)
		authRoutes.GET("/token", GetAccessToken)
		authRoutes.GET("/refresh", RefreshAccessToken)
		authRoutes.GET("/api", GetAPIToken)
		authRoutes.DELETE("/api", RevokeAPIToken)
	}
	lightningRoutes := Client.Router.Group("/lightning")
	{
		lightningRoutes.GET("/:id", GetBolt)
		lightningRoutes.POST("/", PostBolt)
	}
}
