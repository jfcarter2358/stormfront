package client

func InitializeRoutes() {
	apiRoutes := Client.Router.Group("/api")
	{
		apiRoutes.GET("/health", GetHealth)
		apiRoutes.GET("/state", GetState)
		apiRoutes.POST("/register", RegisterFollower)
		apiRoutes.POST("/deregister", DeregisterFollower)
		apiRoutes.POST("/update/succession", UpdateFollowerSuccession)

	}
	authRoutes := Client.Router.Group("/auth")
	{
		authRoutes.GET("/join", GetJoinCommand)
		authRoutes.GET("/token", GetAccessToken)
		authRoutes.GET("/check", CheckAccessToken)
		authRoutes.GET("/refresh", RefreshAccessToken)
	}
}
