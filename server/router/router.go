package router

import (
	"your-project/handler"
	"your-project/internal/service"
	aidomain "your-project/internal/service/ai"
	"your-project/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(aiService aidomain.AIFacade) *gin.Engine {
	handler.SetAIService(aiService)
	handler.SetResumeService(service.NewResumeService())
	handler.SetPracticeService(service.NewPracticeService())
	router := gin.New()
	router.MaxMultipartMemory = 512 << 20

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Static("/uploads", "./uploads")

	api := router.Group("/api/v1")
	{
		public := api.Group("/")
		{
			public.GET("/health", handler.Health)
			public.GET("/ready", handler.Ready)
			public.POST("/register", handler.Register)
			public.POST("/enterprise/apply", handler.ApplyEnterprise)
			public.POST("/university/apply", handler.ApplyUniversity)
			public.POST("/login", handler.Login)
			public.GET("/ws/interview/live", handler.InterviewLiveWS)
			public.GET("/ws/interview/group", handler.InterviewGroupWS)
			public.GET("/interview/live/ws", handler.InterviewSignalWS)
		}

		protected := api.Group("/")
		protected.Use(middleware.Auth())
		{
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("admin"))
			{
				admin.POST("/applications/:role/audit", handler.AuditApplication)
			}

			protected.GET("/user/profile", handler.GetUserProfile)
			protected.PUT("/user/profile", handler.UpdateUserProfile)
			protected.POST("/user/avatar", handler.UpdateAvatar)
			protected.PUT("/user/password", handler.UpdatePassword)

			protected.POST("/interview/start", handler.StartInterview)
			protected.POST("/interview/start/standard", handler.StartStandardInterview)
			protected.POST("/interview/start/algorithm", handler.StartAlgorithmInterview)
			protected.POST("/interview/live/join", handler.JoinInterview)
			protected.POST("/interview/live/start", handler.StartLiveInterview)
			protected.GET("/interview/live/workbench", handler.GetLiveInterviewWorkbench)
			protected.GET("/interview", handler.GetInterviews)
			protected.GET("/interview/config", handler.GetInterviewConfig)
			protected.GET("/interview/:id", handler.GetInterview)
			protected.PUT("/interview/:id/answer", handler.SubmitAnswer)
			protected.POST("/interview/:id/mock-code", handler.SubmitMockCode)
			protected.POST("/interview/:id/end", handler.EndInterview)
			protected.POST("/interview/:id/recording", handler.UploadInterviewRecording)
			protected.POST("/interview/:id/speech-analyze", handler.AnalyzeSpeechChunk)
			protected.POST("/interview/:id/shadow-hint", handler.GetShadowCoachHint)
			protected.POST("/interview/:id/tts", handler.SynthesizeInterviewSpeech)
			protected.GET("/interview/:id/algorithm/session", handler.GetAlgorithmSession)
			protected.POST("/interview/:id/algorithm/run", handler.RunAlgorithmCode)
			protected.POST("/interview/:id/algorithm/skip", handler.SkipAlgorithmProblem)
			protected.POST("/interview/:id/human-feedback", handler.SubmitHumanFeedback)
			protected.GET("/interview/:id/reveal-style", handler.RevealRandomStyle)
			protected.POST("/interview/blindbox/draw", handler.DrawBlindBoxScenario)
			protected.GET("/interview/blindbox/scenarios", handler.GetBlindBoxScenarios)

			// Human Interviewers & Bookings
			protected.GET("/interview/human-interviewers", handler.GetHumanInterviewers)
			protected.GET("/interview/human-interviewers/:id", handler.GetHumanInterviewer)
			protected.POST("/interview/booking", handler.BookHumanInterview)
			protected.GET("/interview/bookings", handler.GetUserBookings)
			protected.GET("/interview/invite-candidates", handler.ListInviteCandidates)
			protected.POST("/interview/invitations", handler.CreateHumanInvitation)
			protected.GET("/interview/invitations", handler.GetHumanInvitations)
			protected.GET("/interview/invitations/received", handler.GetReceivedHumanInvitations)
			protected.POST("/interview/invitations/:id/respond", handler.RespondHumanInvitation)
			protected.DELETE("/interview/invitations/:id", handler.DeleteHumanInvitation)

			protected.POST("/resume/parse", handler.ParseResume)
			protected.GET("/resume/latest", handler.GetLatestResumeAnalysis)

			practice := protected.Group("/practice")
			{
				practice.GET("/meta", handler.GetPracticeMeta)
				practice.GET("/bank/options", handler.GetPracticeOptions)
				practice.GET("/questions", handler.ListPracticeQuestions)
				practice.GET("/question-lists", handler.GetPracticeQuestionLists)
				practice.GET("/question-lists/:id/questions", handler.GetPracticeQuestionsOfList)
				practice.GET("/specialties", handler.GetPracticeSpecialties)
				practice.GET("/question/random", handler.DrawPracticeQuestion)
				practice.POST("/answer", handler.SubmitPracticeAnswer)
				practice.GET("/point/summary", handler.GetPracticePointSummary)
				practice.GET("/wrongs/:id/remedial", handler.GetPracticeWrongRemedial)
				practice.GET("/question/:id/solution", handler.GetPracticeSolution)
				practice.POST("/assessment/start", handler.StartPracticeAssessment)
				practice.POST("/assessment/answer", handler.SubmitPracticeAssessmentAnswer)
				practice.POST("/assessment/:id/complete", handler.CompletePracticeAssessment)
				practice.GET("/integration/snapshot", handler.GetPracticeIntegrationSnapshot)
				practice.POST("/integration/feedback", handler.SubmitPracticeIntegrationFeedback)
				practice.GET("/wrongs", handler.GetPracticeWrongs)
				practice.DELETE("/wrongs/:id", handler.DeletePracticeWrong)
				practice.POST("/wrongs/:id/favorite", handler.TogglePracticeWrongFavorite)
				practice.POST("/favorites/:question_id", handler.TogglePracticeQuestionFavorite)
				practice.GET("/dashboard", handler.GetPracticeDashboard)
				practice.POST("/questionbank/import", handler.ImportPracticeQuestionBank)
				practice.GET("/records/export", handler.ExportPracticeRecords)
				practice.GET("/question/:id", handler.GetPracticeQuestionByID)
			}

			protected.GET("/reports", handler.GetReports)
			protected.GET("/reports/:id", handler.GetReport)
			protected.GET("/reports/:id/download", handler.DownloadReport)
			protected.POST("/reports/generate", handler.GenerateReport)

			// Growth
			protected.GET("/growth/stats", handler.GetGrowthStats)

			// AI对话接口
			protected.POST("/ai/chat", handler.AIChat)
			protected.POST("/interview/:id/ai-chat", handler.AIChatWithInterviewContext)

			// 系统诊断
			protected.GET("/system/ocr/status", handler.OCRStatus)

			// ===== Enterprise 企业端 =====
			enterprise := protected.Group("/enterprise")
			enterprise.Use(middleware.RequireRole("enterprise"))
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
			university.Use(middleware.RequireRole("university"))
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
