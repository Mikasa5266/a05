package handler

import (
	"your-project/service"
	aidomain "your-project/service/ai"
)

var aiService aidomain.AIFacade
var resumeService service.ResumeServiceInterface

func SetAIService(svc aidomain.AIFacade) {
	aiService = svc
}

func mustAIService() aidomain.AIFacade {
	if aiService != nil {
		return aiService
	}
	return service.MustGetAIService()
}

func SetResumeService(svc service.ResumeServiceInterface) {
	resumeService = svc
}

func mustResumeService() service.ResumeServiceInterface {
	if resumeService != nil {
		return resumeService
	}
	return service.NewResumeService()
}
