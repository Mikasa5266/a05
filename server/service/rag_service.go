package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"your-project/config"
	"your-project/model"
	ragpkg "your-project/pkg/rag"
	"your-project/repository"
)

type KnowledgeChunk struct {
	ID       string
	Content  string
	Category string
	Source   string
}

type RAGService struct {
	questionRepo *repository.QuestionRepository
	vectorStore  ragpkg.VectorStore
	embedder     ragpkg.Embedder
	splitter     ragpkg.DocumentSplitter
}

var (
	globalRAGService *RAGService
	ragOnce          sync.Once
)

func GetRAGService() *RAGService {
	ragOnce.Do(func() {
		vectorStore, embedder, splitter := buildDefaultRAGBackends()
		globalRAGService = NewRAGServiceWithBackends(vectorStore, embedder, splitter)
		// Asynchronously load knowledge base on startup
		go func() {
			if err := globalRAGService.LoadKnowledgeBase("knowledge_base"); err != nil {
				fmt.Printf("Failed to load knowledge base: %v\n", err)
			}
			// Load existing community posts
			if err := globalRAGService.LoadCommunityPosts(); err != nil {
				fmt.Printf("Failed to load community posts: %v\n", err)
			}
		}()
	})
	return globalRAGService
}

func NewRAGService() *RAGService {
	return GetRAGService()
}

func buildDefaultRAGBackends() (ragpkg.VectorStore, ragpkg.Embedder, ragpkg.DocumentSplitter) {
	var vectorStore ragpkg.VectorStore
	var embedder ragpkg.Embedder

	splitter := ragpkg.NewDefaultSlidingWindowChunker()

	if store, err := ragpkg.NewQdrantStoreFromEnv(); err != nil {
		fmt.Printf("[RAG] Failed to init Qdrant store: %v\n", err)
	} else {
		pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := store.Ping(pingCtx); err != nil {
			fmt.Printf("[RAG] Qdrant unavailable, vector retrieval disabled: %v\n", err)
		} else {
			vectorStore = store
		}
	}

	if envEmbedder, err := ragpkg.NewOpenAIEmbedderFromEnv(); err == nil {
		embedder = envEmbedder
	} else {
		cfg := config.GetConfig()
		if cfg != nil {
			model := strings.TrimSpace(os.Getenv("EMBEDDING_MODEL"))
			if model == "" {
				model = strings.TrimSpace(cfg.LLM.Models["embedding"])
			}
			emb, buildErr := ragpkg.NewOpenAIEmbedder(ragpkg.OpenAIEmbedderConfig{
				APIKey:  strings.TrimSpace(cfg.LLM.APIKey),
				BaseURL: strings.TrimSpace(cfg.LLM.BaseURL),
				Model:   model,
			})
			if buildErr == nil {
				embedder = emb
			} else {
				fmt.Printf("[RAG] Failed to init embedder from config: %v\n", buildErr)
			}
		} else {
			fmt.Printf("[RAG] Config not loaded, embedder unavailable: %v\n", err)
		}
	}

	if vectorStore == nil || embedder == nil {
		fmt.Printf("[RAG] Vector pipeline not fully initialized (vectorStore=%t, embedder=%t)\n", vectorStore != nil, embedder != nil)
	}

	return vectorStore, embedder, splitter
}

func NewRAGServiceWithBackends(vectorStore ragpkg.VectorStore, embedder ragpkg.Embedder, splitter ragpkg.DocumentSplitter) *RAGService {
	return &RAGService{
		questionRepo: repository.NewQuestionRepository(),
		vectorStore:  vectorStore,
		embedder:     embedder,
		splitter:     splitter,
	}
}

func (s *RAGService) SetBackends(vectorStore ragpkg.VectorStore, embedder ragpkg.Embedder, splitter ragpkg.DocumentSplitter) {
	s.vectorStore = vectorStore
	s.embedder = embedder
	s.splitter = splitter
}

func (s *RAGService) LoadKnowledgeBase(rootPath string) error {
	var chunks []KnowledgeChunk

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			category := filepath.Base(filepath.Dir(path))
			doc := ragpkg.Document{
				ID:      info.Name(),
				Content: string(content),
				Metadata: map[string]string{
					"category": category,
					"source":   path,
				},
			}

			for i, chunk := range s.splitDocument(context.Background(), doc) {
				trimmed := strings.TrimSpace(chunk.Content)
				if len([]rune(trimmed)) < 20 {
					continue
				}
				chunkID := strings.TrimSpace(chunk.ID)
				if chunkID == "" {
					chunkID = fmt.Sprintf("%s_%d", info.Name(), i)
				}
				chunkCategory := category
				chunkSource := path
				if chunk.Metadata != nil {
					if v := strings.TrimSpace(chunk.Metadata["category"]); v != "" {
						chunkCategory = v
					}
					if v := strings.TrimSpace(chunk.Metadata["source"]); v != "" {
						chunkSource = v
					}
				}
				chunks = append(chunks, KnowledgeChunk{
					ID:       chunkID,
					Content:  trimmed,
					Category: chunkCategory,
					Source:   chunkSource,
				})
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return s.indexKnowledgeChunks(context.Background(), chunks)
}

func (s *RAGService) LoadCommunityPosts() error {
	db := repository.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	var posts []model.CommunityPost
	if err := db.Find(&posts).Error; err != nil {
		return err
	}
	for _, post := range posts {
		// Use a simple length check first to avoid too many AI calls on startup
		content := post.Content + post.Process + post.Questions + post.Review
		if len([]rune(content)) > 100 {
			_ = s.FilterAndIndexPost(&post)
		}
	}
	return nil
}

func (s *RAGService) SearchKnowledgeChunks(query string) ([]KnowledgeChunk, error) {
	return s.SearchKnowledgeChunksWithLimit(query, 3)
}

func (s *RAGService) SearchKnowledgeChunksWithLimit(query string, limit int) ([]KnowledgeChunk, error) {
	if limit <= 0 {
		limit = 1
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return []KnowledgeChunk{}, nil
	}
	if !s.hasVectorBackends() {
		return []KnowledgeChunk{}, nil
	}

	queryVector, err := s.embedder.Embed(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed search query: %w", err)
	}

	results, err := s.vectorStore.Search(context.Background(), queryVector, limit)
	if err != nil {
		return nil, err
	}

	chunks := make([]KnowledgeChunk, 0, len(results))
	for _, res := range results {
		content := strings.TrimSpace(res.Content)
		if content == "" {
			continue
		}
		chunk := KnowledgeChunk{
			ID:      strings.TrimSpace(res.ID),
			Content: content,
		}
		if res.Metadata != nil {
			chunk.Category = strings.TrimSpace(res.Metadata["category"])
			chunk.Source = strings.TrimSpace(res.Metadata["source"])
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

func (s *RAGService) FilterAndIndexPost(post *model.CommunityPost) error {
	fmt.Printf("[RAG] Processing post ID %d: %s\n", post.ID, post.Title)

	// 1. Basic length check
	content := post.Content + " " + post.Process + " " + post.Questions + " " + post.Review
	contentLen := len([]rune(content))
	fmt.Printf("[RAG] Post content length: %d\n", contentLen)

	if contentLen < 40 {
		fmt.Printf("[RAG] Post ID %d rejected: too short\n", post.ID)
		return nil
	}

	// 2. Force indexing for high-quality posts (long content + keywords)
	isHighQuality := contentLen > 200 ||
		strings.Contains(post.Title, "面经") ||
		strings.Contains(post.Title, "面试") ||
		strings.Contains(post.Title, "复盘")

	shouldIndex := false
	if isHighQuality {
		fmt.Printf("[RAG] Post ID %d: marked as high quality, force indexing\n", post.ID)
		shouldIndex = true
	} else {
		// Use AI to check if it's a valid interview experience
		aiSvc := MustGetAIService()
		prompt := fmt.Sprintf(`
		请分析以下面试经验内容，判断其是否包含有价值的技术面试信息（如面试流程、具体面试题、技术点复盘等）。
		如果内容是乱发的、无意义的、或者完全不包含技术面试相关信息，请返回 "INVALID"。
		如果是有效的技术面试经验，请返回 "VALID"。

		标题：%s
		公司：%s
		岗位：%s
		内容：%s
		面试流程：%s
		面试题：%s
		复盘建议：%s
	`, post.Title, post.Company, post.Position, post.Content, post.Process, post.Questions, post.Review)

		decision, err := aiSvc.Chat(context.Background(), prompt)
		if err != nil {
			fmt.Printf("[RAG] AI check failed for post %d: %v, falling back to manual check\n", post.ID, err)
			// Fallback to manual validation if AI fails
			if len([]rune(post.Questions)) > 10 || len([]rune(post.Process)) > 10 {
				shouldIndex = true
			}
		} else {
			fmt.Printf("[RAG] AI decision for post %d: %s\n", post.ID, decision)
			if strings.Contains(strings.ToUpper(decision), "VALID") {
				shouldIndex = true
			}
		}
	}

	if !shouldIndex {
		fmt.Printf("[RAG] Post ID %d rejected by filter\n", post.ID)
		return nil
	}

	// 3. Create content and save to disk
	author := post.Author
	if author == "" {
		author = "匿名"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", post.Title))
	sb.WriteString(fmt.Sprintf("> 来源: 校友社区 (作者: %s)\n", author))
	sb.WriteString(fmt.Sprintf("> 发布时间: %s\n\n", post.CreatedAt.Format("2006-01-02 15:04:05")))

	if post.Company != "" {
		sb.WriteString(fmt.Sprintf("## 面试信息\n- **公司**: %s\n", post.Company))
		if post.Position != "" {
			sb.WriteString(fmt.Sprintf("- **岗位**: %s\n", post.Position))
		}
		sb.WriteString("\n")
	}

	if post.Process != "" {
		sb.WriteString(fmt.Sprintf("## 面试流程\n%s\n\n", post.Process))
	}
	if post.Questions != "" {
		sb.WriteString(fmt.Sprintf("## 高频面试题\n%s\n\n", post.Questions))
	}
	if post.Review != "" {
		sb.WriteString(fmt.Sprintf("## 复盘建议\n%s\n\n", post.Review))
	}
	if post.Content != "" && len(post.Content) > 20 {
		sb.WriteString(fmt.Sprintf("## 详细记录\n%s\n\n", post.Content))
	}

	// Save to disk for persistence and visibility
	// We try to find the project root knowledge_base folder
	filename := fmt.Sprintf("post_%d.md", post.ID)

	// Try multiple possible paths to find the right knowledge_base
	possiblePaths := []string{
		filepath.Join("..", "knowledge_base", "community"), // If running from server/
		filepath.Join("knowledge_base", "community"),       // If running from root
	}

	var finalDirPath string
	for _, p := range possiblePaths {
		// Check if the parent 'knowledge_base' exists in this path
		parent := filepath.Dir(p)
		if _, err := os.Stat(parent); err == nil {
			finalDirPath = p
			break
		}
	}

	if finalDirPath == "" {
		finalDirPath = filepath.Join("knowledge_base", "community") // Default
	}

	fmt.Printf("[RAG] Saving post %d to: %s\n", post.ID, filepath.Join(finalDirPath, filename))

	_ = os.MkdirAll(finalDirPath, 0755)
	filePath := filepath.Join(finalDirPath, filename)

	err := os.WriteFile(filePath, []byte(sb.String()), 0644)
	if err != nil {
		fmt.Printf("[RAG] Failed to save post to knowledge base file: %v\n", err)
		return err
	}

	chunk := KnowledgeChunk{
		ID:       fmt.Sprintf("post_%d", post.ID),
		Content:  sb.String(),
		Category: "community",
		Source:   filePath,
	}

	// Update DB status
	db := repository.GetDB()
	if db != nil {
		db.Model(&model.CommunityPost{}).Where("id = ?", post.ID).Update("is_indexed", true)
	}

	return s.indexKnowledgeChunks(context.Background(), []KnowledgeChunk{chunk})
}

func (s *RAGService) SearchKnowledgeBase(query string) (string, error) {
	chunks, err := s.SearchKnowledgeChunks(query)
	if err != nil {
		return "", err
	}

	if len(chunks) == 0 {
		return "未找到相关知识点", nil
	}

	var sb strings.Builder
	for _, chunk := range chunks {
		sb.WriteString(chunk.Content)
		sb.WriteString("\n---\n")
	}
	return sb.String(), nil
}

func (s *RAGService) SearchSimilarQuestions(query string, position, difficulty string, limit int) ([]*model.Question, error) {
	return s.SearchSimilarQuestionsWithExclude(query, position, difficulty, limit, nil)
}

func (s *RAGService) SearchSimilarQuestionsWithExclude(query string, position, difficulty string, limit int, excludeQuestionIDs []uint) ([]*model.Question, error) {
	if limit <= 0 {
		limit = 1
	}

	allQuestions, err := s.questionRepo.GetQuestionsByPositionAndDifficultyWithExclude(position, difficulty, excludeQuestionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get questions: %w", err)
	}
	if len(allQuestions) == 0 {
		return nil, nil
	}
	if !s.hasVectorBackends() {
		return limitQuestions(allQuestions, limit), nil
	}

	points := make([]ragpkg.VectorPoint, 0, len(allQuestions))
	questionByPointID := make(map[string]*model.Question, len(allQuestions))

	for _, question := range allQuestions {
		if question == nil {
			continue
		}
		text := strings.TrimSpace(question.Title + "\n" + question.Content)
		if text == "" {
			continue
		}
		vector, embedErr := s.embedder.Embed(context.Background(), text)
		if embedErr != nil {
			continue
		}
		pointID := fmt.Sprintf("question_%d", question.ID)
		questionByPointID[pointID] = question
		points = append(points, ragpkg.VectorPoint{
			ID:      pointID,
			Vector:  vector,
			Content: text,
			Metadata: map[string]string{
				"kind":        "question",
				"question_id": fmt.Sprintf("%d", question.ID),
				"position":    question.Position,
				"difficulty":  question.Difficulty,
			},
		})
	}

	if len(points) == 0 {
		return limitQuestions(allQuestions, limit), nil
	}

	if err := s.vectorStore.Upsert(context.Background(), points); err != nil {
		return nil, fmt.Errorf("failed to index questions: %w", err)
	}

	queryVector, err := s.embedder.Embed(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	similarResults, err := s.vectorStore.SearchWithOptions(context.Background(), queryVector, limit*2, ragpkg.SearchOptions{
		ExcludePointIDs: buildQuestionExcludePointIDs(excludeQuestionIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search similar questions: %w", err)
	}

	filteredQuestions := make([]*model.Question, 0, limit)
	seen := make(map[uint]struct{}, limit)
	for _, result := range similarResults {
		question := questionByPointID[result.ID]
		if question == nil {
			continue
		}
		if uintSliceContains(excludeQuestionIDs, question.ID) {
			continue
		}
		if _, ok := seen[question.ID]; ok {
			continue
		}
		seen[question.ID] = struct{}{}
		filteredQuestions = append(filteredQuestions, question)
		if len(filteredQuestions) >= limit {
			break
		}
	}
	if len(filteredQuestions) == 0 {
		return limitQuestions(allQuestions, limit), nil
	}

	return filteredQuestions, nil
}

func buildQuestionExcludePointIDs(excludeQuestionIDs []uint) []string {
	if len(excludeQuestionIDs) == 0 {
		return nil
	}
	pointIDs := make([]string, 0, len(excludeQuestionIDs))
	seen := make(map[uint]struct{}, len(excludeQuestionIDs))
	for _, id := range excludeQuestionIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		pointIDs = append(pointIDs, fmt.Sprintf("question_%d", id))
	}
	return pointIDs
}

func uintSliceContains(items []uint, target uint) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func (s *RAGService) GenerateQuestionBasedOnContext(context string, position, difficulty string) (*model.Question, error) {
	similarQuestions, err := s.SearchSimilarQuestions(context, position, difficulty, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar questions: %w", err)
	}

	if len(similarQuestions) == 0 {
		question, createErr := s.createDefaultQuestion(position, difficulty)
		if createErr != nil {
			return nil, createErr
		}
		s.ensureQuestionLocalized(question)
		return question, nil
	}

	bestQuestion := similarQuestions[0]
	adapted := s.adaptQuestion(bestQuestion, context)
	s.ensureQuestionLocalized(adapted)
	return adapted, nil
}

func (s *RAGService) createDefaultQuestion(position, difficulty string) (*model.Question, error) {
	question := &model.Question{
		Title:      fmt.Sprintf("%s岗位技术问题", position),
		Content:    fmt.Sprintf("请描述你在%s方面的经验，以及你如何处理相关的技术挑战。", position),
		Position:   position,
		Difficulty: difficulty,
		Category:   "general",
	}
	question.SetTags([]string{position, difficulty, "experience"})

	return question, nil
}

func (s *RAGService) adaptQuestion(original *model.Question, context string) *model.Question {
	adapted := *original
	tags := adapted.GetTags()
	if strings.Contains(context, "项目") || strings.Contains(context, "project") {
		adapted.Content = fmt.Sprintf("结合你之前的项目经验，%s", original.Content)
		tags = append(tags, "project-based")
	}

	if strings.Contains(context, "团队") || strings.Contains(context, "team") {
		adapted.Content = fmt.Sprintf("在团队协作的场景下，%s", original.Content)
		tags = append(tags, "team-collaboration")
	}
	adapted.SetTags(tags)

	return &adapted
}

func (s *RAGService) ensureQuestionLocalized(question *model.Question) {
	if question == nil {
		return
	}

	defer func() {
		_ = recover()
	}()

	aiSvc := MustGetAIService()
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	aiSvc.EnsureQuestionChinese(ctx, question)
}

func (s *RAGService) hasVectorBackends() bool {
	return s.vectorStore != nil && s.embedder != nil
}

func (s *RAGService) splitDocument(ctx context.Context, doc ragpkg.Document) []ragpkg.Chunk {
	if s.splitter != nil {
		chunks, err := s.splitter.Split(ctx, doc)
		if err == nil && len(chunks) > 0 {
			return chunks
		}
	}

	parts := strings.Split(doc.Content, "\n\n")
	chunks := make([]ragpkg.Chunk, 0, len(parts))
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		chunks = append(chunks, ragpkg.Chunk{
			ID:      fmt.Sprintf("%s_%d", doc.ID, i),
			Content: trimmed,
			Metadata: map[string]string{
				"category": strings.TrimSpace(doc.Metadata["category"]),
				"source":   strings.TrimSpace(doc.Metadata["source"]),
			},
		})
	}
	return chunks
}

func (s *RAGService) indexKnowledgeChunks(ctx context.Context, chunks []KnowledgeChunk) error {
	if len(chunks) == 0 || !s.hasVectorBackends() {
		return nil
	}

	points := make([]ragpkg.VectorPoint, 0, len(chunks))
	for _, chunk := range chunks {
		content := strings.TrimSpace(chunk.Content)
		if content == "" {
			continue
		}
		vector, err := s.embedder.Embed(ctx, content)
		if err != nil {
			continue
		}
		pointID := strings.TrimSpace(chunk.ID)
		if pointID == "" {
			pointID = fmt.Sprintf("knowledge_%d", len(points)+1)
		}
		metadata := map[string]string{"kind": "knowledge"}
		if v := strings.TrimSpace(chunk.Category); v != "" {
			metadata["category"] = v
		}
		if v := strings.TrimSpace(chunk.Source); v != "" {
			metadata["source"] = v
		}
		points = append(points, ragpkg.VectorPoint{
			ID:       pointID,
			Vector:   vector,
			Content:  content,
			Metadata: metadata,
		})
	}

	if len(points) == 0 {
		return nil
	}

	return s.vectorStore.Upsert(ctx, points)
}

func limitQuestions(questions []*model.Question, limit int) []*model.Question {
	if limit <= 0 {
		limit = 1
	}
	if len(questions) <= limit {
		return questions
	}
	return questions[:limit]
}
