package service

import (
	"encoding/json"
	"strings"

	"your-project/internal/model"
	"your-project/internal/repository"

	"github.com/gin-gonic/gin"
)

func RecordAuditLog(
	c *gin.Context,
	actorType string,
	actorName string,
	action string,
	outcome string,
	targetType string,
	targetID uint,
	detail map[string]interface{},
) {
	detailJSON := "{}"
	if len(detail) > 0 {
		if payload, err := json.Marshal(detail); err == nil {
			detailJSON = string(payload)
		}
	}

	method := ""
	path := ""
	sourceIP := ""
	if c != nil && c.Request != nil {
		method = truncateString(strings.TrimSpace(c.Request.Method), 10)
		path = strings.TrimSpace(c.FullPath())
		if path == "" && c.Request.URL != nil {
			path = strings.TrimSpace(c.Request.URL.Path)
		}
		ip, _ := splitHostPort(c.Request.RemoteAddr)
		sourceIP = ProtectSecurityLogField(ip)
	}

	entry := model.AuditLog{
		ActorType:  truncateString(strings.TrimSpace(actorType), 30),
		ActorName:  truncateString(strings.TrimSpace(actorName), 100),
		Action:     truncateString(strings.TrimSpace(action), 100),
		Outcome:    truncateString(strings.TrimSpace(outcome), 20),
		Method:     method,
		Path:       truncateString(path, 255),
		TargetType: truncateString(strings.TrimSpace(targetType), 30),
		TargetID:   targetID,
		SourceIP:   truncateString(sourceIP, 255),
		DetailJSON: detailJSON,
	}

	_ = repository.GetDB().Create(&entry).Error
}
