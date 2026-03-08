package handler

import (
	"fmt"
	"net/http"
	"your-project/pkg/tts"

	"github.com/gin-gonic/gin"
)

// GenerateTTS generates speech from text using Edge TTS (Microsoft Azure Neural)
func GenerateTTS(c *gin.Context) {
	var req struct {
		Text  string `json:"text"`
		Voice string `json:"voice"` // Optional: "xiaoxiao", "yunxi", "yunjian", "xiaoyi", "panpan", "yunyang"
		Rate  string `json:"rate"`  // Optional: "+10%", "-5%"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
		return
	}

	etts := tts.NewEdgeTTS()
	
	// Voice Selection Logic
	// Default: Xiaoxiao (Warm, Gentle Female) - Similar to Doubao's default
	etts.Voice = "zh-CN-XiaoxiaoNeural" 
	etts.Rate = "+10%" // Slightly faster for natural flow

	switch req.Voice {
	case "yunxi": // Energetic Male
		etts.Voice = "zh-CN-YunxiNeural"
	case "yunjian": // Sports/News Male
		etts.Voice = "zh-CN-YunjianNeural"
	case "xiaoyi": // Digital Assistant Female
		etts.Voice = "zh-CN-XiaoyiNeural"
	case "panpan": // Warm Female (Storytelling)
		etts.Voice = "zh-CN-PanpanNeural"
	case "yunyang": // News Male
		etts.Voice = "zh-CN-YunyangNeural"
	case "liaoning": // Northeast Dialect (Fun)
		etts.Voice = "zh-CN-liaoning-XiaobeiNeural"
	case "shaanxi": // Shaanxi Dialect (Fun)
		etts.Voice = "zh-CN-shaanxi-XiaoniNeural"
	}

	if req.Rate != "" {
		etts.Rate = req.Rate
	}

	audioData, err := etts.Synthesize(req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tts failed: " + err.Error()})
		return
	}

	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Length", fmt.Sprintf("%d", len(audioData)))
	c.Header("Cache-Control", "no-cache")
	c.Writer.Write(audioData)
}
