package client

func InitializeRoutes(clientType string) {
	apiRoutes := Client.Router.Group("/api")
	{
		apiRoutes.GET("/health", GetHealth)
		apiRoutes.GET("/state", GetState)
		apiRoutes.POST("/update/succession", UpdateFollowerSuccession)
	}
	authRoutes := Client.Router.Group("/auth")
	{
		authRoutes.GET("/check", CheckAccessToken)
	}
	lightningRoutes := Client.Router.Group("/lightning")
	{
		lightningRoutes.GET("/:id", GetBolt)
		lightningRoutes.POST("/", PostBolt)
	}
	if clientType == "Leader" {
		authRoutes.GET("/join", GetJoinCommand)
		authRoutes.GET("/token", GetAccessToken)
		authRoutes.GET("/refresh", RefreshAccessToken)

		apiRoutes.POST("/register", RegisterFollower)
		apiRoutes.POST("/deregister", DeregisterFollower)
	}
}
