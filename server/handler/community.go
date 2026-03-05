package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"your-project/model"
	"your-project/repository"
	"your-project/service"

	"github.com/gin-gonic/gin"
)

// ===== Posts =====

func GetPosts(c *gin.Context) {
	db := repository.GetDB()
	var posts []model.CommunityPost
	query := db.Model(&model.CommunityPost{})

	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if tag := c.Query("tag"); tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}

	query.Order("created_at DESC").Limit(50).Find(&posts)
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

func GetPost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	db := repository.GetDB()
	var post model.CommunityPost
	if err := db.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		return
	}
	// Increment view count
	db.Model(&post).Update("views", post.Views+1)
	post.Views++
	c.JSON(http.StatusOK, gin.H{"post": post})
}

func CreatePost(c *gin.Context) {
	userID := c.GetUint("user_id")
	var post model.CommunityPost
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	post.UserID = userID

	// Get author from user profile
	var user model.User
	db := repository.GetDB()
	if err := db.First(&user, userID).Error; err == nil {
		post.Author = user.Username
		post.Avatar = user.Avatar
	}

	if err := db.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"post": post})
}

func LikePost(c *gin.Context) {
	userID := c.GetUint("user_id")
	postID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	db := repository.GetDB()

	// Check if already liked
	var existingLike model.PostLike
	if err := db.Where("post_id = ? AND user_id = ?", postID, userID).First(&existingLike).Error; err == nil {
		// Already liked, unlike
		db.Delete(&existingLike)
		db.Model(&model.CommunityPost{}).Where("id = ?", postID).UpdateColumn("likes", db.Raw("likes - 1"))
		c.JSON(http.StatusOK, gin.H{"message": "已取消点赞", "liked": false})
		return
	}

	like := model.PostLike{PostID: uint(postID), UserID: userID}
	db.Create(&like)
	db.Model(&model.CommunityPost{}).Where("id = ?", postID).UpdateColumn("likes", db.Raw("likes + 1"))
	c.JSON(http.StatusOK, gin.H{"message": "已点赞", "liked": true})
}

func CommentOnPost(c *gin.Context) {
	userID := c.GetUint("user_id")
	postID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var comment model.PostComment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comment.PostID = uint(postID)
	comment.UserID = userID

	// Get author from user
	db := repository.GetDB()
	var user model.User
	if err := db.First(&user, userID).Error; err == nil {
		comment.Author = user.Username
	}

	if err := db.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Increment comment count
	db.Model(&model.CommunityPost{}).Where("id = ?", postID).UpdateColumn("comments", db.Raw("comments + 1"))

	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}

func GetPostComments(c *gin.Context) {
	postID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	db := repository.GetDB()
	var comments []model.PostComment
	db.Where("post_id = ?", postID).Order("created_at DESC").Find(&comments)
	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

// ===== Mentors & Bookings =====

func BookMentor(c *gin.Context) {
	userID := c.GetUint("user_id")
	mentorID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var booking model.MentorBooking
	if err := c.ShouldBindJSON(&booking); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	booking.UserID = userID
	booking.MentorID = uint(mentorID)

	db := repository.GetDB()
	if err := db.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"booking": booking})
}

func GetMentors(c *gin.Context) {
	// Return top users who have completed many interviews as potential mentors
	c.JSON(http.StatusOK, gin.H{
		"mentors": []gin.H{
			{"id": 1, "name": "张学长", "company": "字节跳动", "position": "高级工程师", "avatar": ""},
			{"id": 2, "name": "李学姐", "company": "阿里巴巴", "position": "产品经理", "avatar": ""},
			{"id": 3, "name": "王学长", "company": "腾讯", "position": "算法工程师", "avatar": ""},
		},
	})
}

func GetBookings(c *gin.Context) {
	userID := c.GetUint("user_id")
	db := repository.GetDB()
	var bookings []model.MentorBooking
	db.Where("user_id = ?", userID).Order("created_at DESC").Find(&bookings)
	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

// ===== Knowledge Base =====

func QueryKnowledgeBase(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ragSvc := service.GetRAGService()
	chunks, err := ragSvc.SearchKnowledgeChunks(req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search knowledge base"})
		return
	}

	var contextBuilder strings.Builder
	var sources []gin.H

	for i, chunk := range chunks {
		contextBuilder.WriteString(fmt.Sprintf("[%d] %s\n%s\n\n", i+1, chunk.Source, chunk.Content))
		sources = append(sources, gin.H{
			"title":     chunk.ID, // Or chunk.Source
			"relevance": 0.9 - float64(i)*0.1, // Mock relevance score
			"content":   chunk.Content,
			"category":  chunk.Category,
		})
	}

	aiSvc := service.NewAIService()
	prompt := fmt.Sprintf("请根据以下参考资料回答问题。如果参考资料不足以回答，请根据你的通用知识补充，但请优先使用参考资料。\n\n参考资料：\n%s\n\n问题：%s", contextBuilder.String(), req.Query)
	
	answer, err := aiSvc.Chat(context.Background(), prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate answer: " + err.Error()})
		return
	}

	answer = aiSvc.EnsureChineseOutput(answer, "抱歉，我暂时无法回答这个问题。")

	c.JSON(http.StatusOK, gin.H{
		"answer":  answer,
		"sources": sources,
	})
}

// ===== Hot Content =====

func GetTopAlumni(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"alumni": []gin.H{
			{"name": "张学长", "company": "字节跳动", "likes": 256},
			{"name": "李学姐", "company": "阿里巴巴", "likes": 198},
			{"name": "王学长", "company": "腾讯", "likes": 176},
			{"name": "赵学姐", "company": "美团", "likes": 145},
		},
	})
}

func GetHotCompanies(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"companies": []gin.H{
			{"name": "字节跳动", "posts": 42},
			{"name": "阿里巴巴", "posts": 38},
			{"name": "腾讯", "posts": 35},
			{"name": "美团", "posts": 28},
			{"name": "华为", "posts": 25},
		},
	})
}
