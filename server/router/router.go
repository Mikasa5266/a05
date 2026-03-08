package router

import (
	"your-project/handler"
	"your-project/middleware"
	"your-project/pkg/websocket"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.New()
	router.MaxMultipartMemory = 512 << 20

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Static("/uploads", "./uploads")

	// WebSocket route
	router.GET("/ws", func(c *gin.Context) {
		websocket.GetHub().HandleWebSocket(c.Writer, c.Request)
	})

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
			protected.POST("/user/avatar", handler.UpdateAvatar)
			protected.PUT("/user/password", handler.UpdatePassword)

			protected.POST("/interview/start", handler.StartInterview)
			protected.GET("/interview", handler.GetInterviews)
			protected.GET("/interview/config", handler.GetInterviewConfig)
			protected.GET("/interview/:id", handler.GetInterview)
			protected.GET("/interview/:id/session", handler.GetInterviewSession)
			protected.PUT("/interview/:id/answer", handler.SubmitAnswer)
			protected.POST("/interview/:id/end", handler.EndInterview)
			protected.POST("/interview/:id/recording", handler.UploadInterviewRecording)
			protected.POST("/interview/:id/speech-analyze", handler.AnalyzeSpeechChunk)
			protected.POST("/interview/:id/human-feedback", handler.SubmitHumanFeedback)
			protected.GET("/interview/:id/reveal-style", handler.RevealRandomStyle)
			protected.POST("/interview/blindbox/draw", handler.DrawBlindBoxScenario)
			protected.GET("/interview/blindbox/scenarios", handler.GetBlindBoxScenarios)

			// Human Interviewers & Bookings
			protected.GET("/interview/human-interviewers", handler.GetHumanInterviewers)
			protected.GET("/interview/human-interviewers/:id", handler.GetHumanInterviewer)
			protected.POST("/interview/booking", handler.BookHumanInterview)
			protected.GET("/interview/bookings", handler.GetUserBookings)

			protected.GET("/questions", handler.GetQuestions)
			protected.GET("/questions/:id", handler.GetQuestion)
			protected.POST("/questions", handler.CreateQuestion)

			protected.GET("/reports", handler.GetReports)
			protected.GET("/reports/:id", handler.GetReport)
			protected.GET("/reports/:id/download", handler.DownloadReport)
			protected.POST("/reports/generate", handler.GenerateReport)

			// Resume
			protected.POST("/resume/parse", handler.ParseResume)
			protected.POST("/resume/generate-questions", handler.GenerateQuestions)
			// protected.POST("/resume/match", handler.MatchJobs) // Merged into ParseResume for now

			// Growth
			protected.GET("/growth/stats", handler.GetGrowthStats)

			// AI对话接口
			protected.POST("/ai/chat", handler.AIChat)
			protected.POST("/interview/:id/ai-chat", handler.AIChatWithInterviewContext)
			protected.POST("/tts", handler.GenerateTTS)

			// 系统诊断
			protected.GET("/system/ocr/status", handler.OCRStatus)

			// ===== Enterprise 企业端 =====
			enterprise := protected.Group("/enterprise")
			{
				enterprise.GET("/dashboard", handler.GetEnterpriseDashboard)

				enterprise.GET("/talent-pool", handler.GetTalentPool)
				enterprise.POST("/talent-pool/:id/invite", handler.InviteTalent)
				enterprise.POST("/talent-pool/:id/save", handler.SaveTalent)

				enterprise.GET("/jobs", handler.GetJobs)
				enterprise.POST("/jobs", handler.CreateJob)
				enterprise.PUT("/jobs/:id", handler.UpdateJob)
				enterprise.DELETE("/jobs/:id", handler.DeleteJob)
				enterprise.GET("/jobs/:id/ability-atlas", handler.GetAbilityAtlas)

				enterprise.GET("/interview-sessions", handler.GetInterviewSessions)
				enterprise.POST("/scenarios", handler.CreateCustomScenario)
				enterprise.GET("/scenarios", handler.GetCustomScenarios)

				enterprise.GET("/analytics", handler.GetRecruitmentAnalytics)
				enterprise.GET("/analytics/funnel", handler.GetRecruitmentFunnel)
				enterprise.GET("/analytics/quality", handler.GetCandidateQualityDistribution)

				enterprise.GET("/standards", handler.GetCapabilityStandards)
				enterprise.POST("/standards", handler.CreateCapabilityStandard)
				enterprise.PUT("/standards/:id", handler.UpdateCapabilityStandard)

				enterprise.GET("/certified", handler.GetCertifiedCandidates)

				enterprise.GET("/referrals", handler.GetReferralChannels)
				enterprise.POST("/referrals", handler.CreateReferral)
			}

			// ===== University 高校端 =====
			university := protected.Group("/university")
			{
				university.GET("/dashboard", handler.GetUniversityDashboard)

				university.GET("/students", handler.GetStudentTracking)
				university.GET("/students/:id", handler.GetStudentDetail)
				university.PUT("/students/:id/risk", handler.UpdateStudentRisk)

				university.GET("/risk-groups", handler.GetRiskGroups)
				university.POST("/mentor/assign", handler.AssignMentor)
				university.POST("/support/batch", handler.BatchSupport)
				university.POST("/support/recommend-course", handler.RecommendCourse)

				university.GET("/courses", handler.GetCourses)
				university.POST("/courses", handler.CreateCourse)
				university.GET("/resources", handler.GetResources)

				university.GET("/employment/stats", handler.GetEmploymentStats)
				university.GET("/employment/by-major", handler.GetMajorEmployment)
				university.GET("/employment/salary", handler.GetSalaryDistribution)
				university.GET("/employment/city", handler.GetCityDistribution)
				university.GET("/employment/industry", handler.GetIndustryDistribution)

				university.GET("/talent-push/recommended", handler.GetRecommendedStudents)
				university.POST("/talent-push", handler.PushStudentsToEnterprise)
				university.GET("/talent-push/history", handler.GetPushHistory)
			}

			// ===== Community 社区 =====
			community := protected.Group("/community")
			{
				community.GET("/posts", handler.GetPosts)
				community.GET("/posts/:id", handler.GetPost)
				community.POST("/posts", handler.CreatePost)
				community.DELETE("/posts/:id", handler.DeletePost)
				community.POST("/posts/:id/like", handler.LikePost)
				community.POST("/posts/:id/comments", handler.CommentOnPost)
				community.GET("/posts/:id/comments", handler.GetPostComments)

				community.POST("/mentors/:id/book", handler.BookMentor)
				community.GET("/mentors", handler.GetMentors)
				community.GET("/bookings", handler.GetBookings)

				community.POST("/knowledge/query", handler.QueryKnowledgeBase)

				community.GET("/top-alumni", handler.GetTopAlumni)
				community.GET("/hot-companies", handler.GetHotCompanies)
			}
		}
	}

	return router
}
