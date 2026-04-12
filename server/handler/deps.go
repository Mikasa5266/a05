package handler

import (
	"your-project/service"
	aidomain "your-project/service/ai"
)

var aiService aidomain.AIFacade
var resumeService service.ResumeService
var practiceService service.PracticeService

// 平台瘦身后仅保留学生核心模块依赖。
// Enterprise/University 相关 Handler 不再通过此处注入初始化。

func SetAIService(svc aidomain.AIFacade) {
	aiService = svc
}

func mustAIService() aidomain.AIFacade {
	if aiService != nil {
		return aiService
	}
	return service.MustGetAIService()
}

func SetResumeService(svc service.ResumeService) {
	resumeService = svc
}

func mustResumeService() service.ResumeService {
	if resumeService != nil {
		return resumeService
	}
	return service.NewResumeService()
}

func SetPracticeService(svc service.PracticeService) {
	practiceService = svc
}

func mustPracticeService() service.PracticeService {
	if practiceService != nil {
		return practiceService
	}
	return service.NewPracticeService()
}
