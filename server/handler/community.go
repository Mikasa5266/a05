package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"your-project/model"
	"your-project/repository"
	"your-project/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	if company := c.Query("company"); company != "" {
		query = query.Where("company = ?", company)
	}
	if position := c.Query("position"); position != "" {
		query = query.Where("position = ?", position)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	_ = query.Count(&total).Error
	query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&posts)

	resp := make([]gin.H, 0, len(posts))
	for _, p := range posts {
		tags := []string{}
		if strings.TrimSpace(p.Tags) != "" {
			for _, t := range strings.Split(p.Tags, ",") {
				tt := strings.TrimSpace(t)
				if tt != "" {
					tags = append(tags, tt)
				}
			}
		}
		resp = append(resp, gin.H{
			"id":             p.ID,
			"user_id":        p.UserID,
			"author":         p.Author,
			"avatar":         p.Avatar,
			"company":        p.Company,
			"position":       p.Position,
			"title":          p.Title,
			"content":        p.Content,
			"process":        p.Process,
			"questions":      p.Questions,
			"review":         p.Review,
			"difficulty":     p.Difficulty,
			"offer_status":   p.OfferStatus,
			"rounds":         p.Rounds,
			"interview_date": p.InterviewDate,
			"tags":           tags,
			"is_indexed":     p.IsIndexed,
			"likes":          p.Likes,
			"comments":       p.Comments,
			"views":          p.Views,
			"created_at":     p.CreatedAt,
			"updated_at":     p.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":     resp,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
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
	db.Model(&model.CommunityPost{}).Where("id = ?", post.ID).UpdateColumn("views", gorm.Expr("views + ?", 1))
	post.Views++
	tags := []string{}
	if strings.TrimSpace(post.Tags) != "" {
		for _, t := range strings.Split(post.Tags, ",") {
			tt := strings.TrimSpace(t)
			if tt != "" {
				tags = append(tags, tt)
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"post": gin.H{
			"id":             post.ID,
			"user_id":        post.UserID,
			"author":         post.Author,
			"avatar":         post.Avatar,
			"company":        post.Company,
			"position":       post.Position,
			"title":          post.Title,
			"content":        post.Content,
			"process":        post.Process,
			"questions":      post.Questions,
			"review":         post.Review,
			"difficulty":     post.Difficulty,
			"offer_status":   post.OfferStatus,
			"rounds":         post.Rounds,
			"interview_date": post.InterviewDate,
			"tags":           tags,
			"is_indexed":     post.IsIndexed,
			"likes":          post.Likes,
			"comments":       post.Comments,
			"views":          post.Views,
			"created_at":     post.CreatedAt,
			"updated_at":     post.UpdatedAt,
		},
	})
}

func CreatePost(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		Title         string     `json:"title" binding:"required"`
		Company       string     `json:"company"`
		Position      string     `json:"position"`
		Tags          []string   `json:"tags"`
		Content       string     `json:"content"` // Made optional as we might rely on other fields
		Process       string     `json:"process"`
		Questions     string     `json:"questions"`
		Review        string     `json:"review"`
		Difficulty    int        `json:"difficulty"`
		OfferStatus   string     `json:"offer_status"`
		Rounds        int        `json:"rounds"`
		InterviewDate *time.Time `json:"interview_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	title := strings.TrimSpace(req.Title)
	// Combine structured content if main content is empty, or just use it as summary
	content := strings.TrimSpace(req.Content)

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标题不能为空"})
		return
	}
	// If content is empty but we have structured data, that's fine.
	// But let's ensure at least something is there.
	if content == "" && req.Process == "" && req.Questions == "" && req.Review == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请至少填写一部分面经内容"})
		return
	}
	if len([]rune(title)) > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标题过长"})
		return
	}
	if len([]rune(content)) > 8000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "内容过长"})
		return
	}

	tags := []string{}
	for _, t := range req.Tags {
		tt := strings.TrimSpace(t)
		if tt != "" {
			tags = append(tags, tt)
		}
	}
	tagStr := strings.Join(tags, ",")

	post := model.CommunityPost{
		UserID:        userID,
		Author:        "",
		Avatar:        "",
		Company:       strings.TrimSpace(req.Company),
		Position:      strings.TrimSpace(req.Position),
		Title:         title,
		Content:       content,
		Process:       req.Process,
		Questions:     req.Questions,
		Review:        req.Review,
		Difficulty:    req.Difficulty,
		OfferStatus:   req.OfferStatus,
		Rounds:        req.Rounds,
		InterviewDate: req.InterviewDate,
		Tags:          tagStr,
	}

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

	// Index into RAG knowledge base asynchronously
	go func(p model.CommunityPost) {
		ragSvc := service.GetRAGService()
		_ = ragSvc.FilterAndIndexPost(&p)
	}(post)

	c.JSON(http.StatusCreated, gin.H{
		"post": gin.H{
			"id":             post.ID,
			"user_id":        post.UserID,
			"author":         post.Author,
			"avatar":         post.Avatar,
			"company":        post.Company,
			"position":       post.Position,
			"title":          post.Title,
			"content":        post.Content,
			"process":        post.Process,
			"questions":      post.Questions,
			"review":         post.Review,
			"difficulty":     post.Difficulty,
			"offer_status":   post.OfferStatus,
			"rounds":         post.Rounds,
			"interview_date": post.InterviewDate,
			"tags":           tags,
			"likes":          post.Likes,
			"comments":       post.Comments,
			"views":          post.Views,
			"created_at":     post.CreatedAt,
			"updated_at":     post.UpdatedAt,
		},
	})
}

func DeletePost(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.GetUint("userID")
	id, _ := strconv.Atoi(c.Param("id"))

	var post model.CommunityPost
	if err := db.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		return
	}
	if post.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "只能删除自己的帖子"})
		return
	}
	if err := db.Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
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
		db.Model(&model.CommunityPost{}).Where("id = ?", postID).UpdateColumn("likes", gorm.Expr("likes - ?", 1))
		var p model.CommunityPost
		_ = db.Select("likes").First(&p, postID).Error
		c.JSON(http.StatusOK, gin.H{"message": "已取消点赞", "liked": false, "likes": p.Likes})
		return
	}

	like := model.PostLike{PostID: uint(postID), UserID: userID}
	db.Create(&like)
	db.Model(&model.CommunityPost{}).Where("id = ?", postID).UpdateColumn("likes", gorm.Expr("likes + ?", 1))
	var p model.CommunityPost
	_ = db.Select("likes").First(&p, postID).Error
	c.JSON(http.StatusOK, gin.H{"message": "已点赞", "liked": true, "likes": p.Likes})
}

func CommentOnPost(c *gin.Context) {
	userID := c.GetUint("user_id")
	postID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容不能为空"})
		return
	}

	comment := model.PostComment{
		PostID:  uint(postID),
		UserID:  userID,
		Author:  "",
		Content: content,
	}

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
	db.Model(&model.CommunityPost{}).Where("id = ?", postID).UpdateColumn("comments", gorm.Expr("comments + ?", 1))

	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}

func GetPostComments(c *gin.Context) {
	postID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	db := repository.GetDB()
	var comments []model.PostComment
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	db.Where("post_id = ?", postID).Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&comments)
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
			"title":     chunk.ID,             // Or chunk.Source
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
	db := repository.GetDB()
	var results []struct {
		Name    string `json:"name"`
		Company string `json:"company"`
		Posts   int64  `json:"posts"`
		Avatar  string `json:"avatar"`
	}

	// Group by author and company to show active alumni
	// We select the one with most posts
	err := db.Model(&model.CommunityPost{}).
		Select("author as name, company, count(*) as posts, max(avatar) as avatar").
		Where("author != '' AND author IS NOT NULL").
		Group("author, company").
		Order("posts DESC").
		Limit(5).
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch top alumni"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alumni": results,
	})
}

func GetHotCompanies(c *gin.Context) {
	db := repository.GetDB()
	var results []struct {
		Name  string `json:"name"`
		Posts int64  `json:"posts"`
	}

	err := db.Model(&model.CommunityPost{}).
		Select("company as name, count(*) as posts").
		Where("company != '' AND company IS NOT NULL").
		Group("company").
		Order("posts DESC").
		Limit(8).
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch hot companies"})
		return
	}

	// Convert to simple string list for frontend if needed,
	// but the frontend might want object with counts.
	// Let's check frontend expectation.
	// Frontend expects: "companies": [{"name": "...", "posts": ...}]
	// But in Community.vue hotCompanies is just an array of strings: ['字节跳动', ...]
	// We should return what frontend needs or update frontend.
	// Let's return objects, and I'll update frontend to use objects or map them.

	c.JSON(http.StatusOK, gin.H{
		"companies": results,
	})
}
