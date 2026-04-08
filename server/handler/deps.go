package handler

import (
	"your-project/service"
	aidomain "your-project/service/ai"
)

var aiService aidomain.AIFacade
var resumeService service.ResumeService
var questionBankService service.QuestionBankService

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

func SetQuestionBankService(svc service.QuestionBankService) {
	questionBankService = svc
}

func mustQuestionBankService() service.QuestionBankService {
	if questionBankService != nil {
		return questionBankService
	}
	return service.NewQuestionBankService()
}
