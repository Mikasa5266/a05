package handler

import "your-project/service"

var aiService *service.AIService

func SetAIService(svc *service.AIService) {
	aiService = svc
}

func mustAIService() *service.AIService {
	if aiService != nil {
		return aiService
	}
	return service.MustGetAIService()
}
