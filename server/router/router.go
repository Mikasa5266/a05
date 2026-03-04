package router

import (
	"your-project/handler"
	"your-project/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	api := router.Group("/api/v1")
	{
		public := api.Group("/")
		{
			public.POST("/register", handler.Register)
			public.POST("/login", handler.Login)
		}

		protected := api.Group("/")
		protected.Use(middleware.Auth())
		{
			protected.GET("/user/profile", handler.GetUserProfile)
			protected.PUT("/user/profile", handler.UpdateUserProfile)

			protected.POST("/interview/start", handler.StartInterview)
			protected.GET("/interview", handler.GetInterviews)
			protected.GET("/interview/:id", handler.GetInterview)
			protected.PUT("/interview/:id/answer", handler.SubmitAnswer)
			protected.POST("/interview/:id/end", handler.EndInterview)

			protected.GET("/questions", handler.GetQuestions)
			protected.GET("/questions/:id", handler.GetQuestion)
			protected.POST("/questions", handler.CreateQuestion)

			protected.GET("/reports", handler.GetReports)
			protected.GET("/reports/:id", handler.GetReport)
			protected.POST("/reports/generate", handler.GenerateReport)

			// Resume
			protected.POST("/resume/parse", handler.ParseResume)
			// protected.POST("/resume/match", handler.MatchJobs) // Merged into ParseResume for now

			// Growth
			protected.GET("/growth/stats", handler.GetGrowthStats)

			// AI对话接口
			protected.POST("/ai/chat", handler.AIChat)
			protected.POST("/interview/:id/ai-chat", handler.AIChatWithInterviewContext)
		}
	}

	return router
}
