package client

import (
	"stormfrontd/middleware"
)

func InitializeRoutes(clientType string) {
	Client.Router.Use(middleware.CORSMiddleware())
	apiRoutes := Client.Router.Group("/api")
	{
		apiRoutes.GET("/health", middleware.CheckTokenAuthentication(), GetHealth)
		apiRoutes.GET("/state", middleware.CheckTokenAuthentication(), GetState)
		apiRoutes.POST("/register", middleware.CheckTokenAuthentication(), RegisterFollower)
		apiRoutes.DELETE("/register", middleware.CheckTokenAuthentication(), DeregisterFollower)
		apiRoutes.GET("/application", middleware.CheckTokenAuthentication(), GetAllApplications)
		apiRoutes.GET("/application/:id/logs", middleware.CheckTokenAuthentication(), GetApplicationLogs)
		apiRoutes.GET("/application/:id/restart", middleware.CheckTokenAuthentication(), RestartApplication)
		apiRoutes.GET("/application/:id", middleware.CheckTokenAuthentication(), GetApplication)
		apiRoutes.POST("/application", middleware.CheckTokenAuthentication(), CreateApplication)
		// apiRoutes.PATCH("/application/:id", middleware.CheckTokenAuthentication(), UpdateApplication)
		apiRoutes.DELETE("/application/:id", middleware.CheckTokenAuthentication(), DeleteApplication)
		apiRoutes.GET("/client", middleware.CheckTokenAuthentication(), GetAllClients)
		apiRoutes.GET("/client/:id", middleware.CheckTokenAuthentication(), GetClient)
		apiRoutes.GET("/node", middleware.CheckTokenAuthentication(), GetAllNodes)
		apiRoutes.GET("/node/:id", middleware.CheckTokenAuthentication(), GetNode)
	}
	authRoutes := Client.Router.Group("/auth")
	{
		// authRoutes.GET("/check/access", CheckAccessToken)
		// authRoutes.GET("/check/api", CheckAPIToken)
		authRoutes.GET("/join/command", middleware.CheckTokenAuthentication(), GetJoinCommand)
		authRoutes.GET("/join", middleware.CheckTokenAuthentication(), GetJoinToken)
		authRoutes.DELETE("/join/:token", middleware.CheckTokenAuthentication(), RevokeJoinToken)
		authRoutes.GET("/token", GetAccessToken)
		authRoutes.GET("/refresh", RefreshAccessToken)
		authRoutes.GET("/api", middleware.CheckTokenAuthentication(), GetAPIToken)
		authRoutes.DELETE("/api", middleware.CheckTokenAuthentication(), RevokeAPIToken)
	}
	lightningRoutes := Client.Router.Group("/lightning")
	{
		lightningRoutes.GET("/:id", middleware.CheckTokenAuthentication(), GetBolt)
		lightningRoutes.POST("/", middleware.CheckTokenAuthentication(), PostBolt)
	}
}
