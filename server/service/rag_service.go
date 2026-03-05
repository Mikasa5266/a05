package service

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"your-project/model"
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
	vectorStore  VectorStore
}

type VectorStore interface {
	Search(query string, limit int) ([]SimilarityResult, error)
	IndexQuestions(questions []*model.Question) error
	IndexDocuments(docs []KnowledgeChunk) error
}

type SimilarityResult struct {
	Question *model.Question // Optional, if it's a question match
	Document *KnowledgeChunk // Optional, if it's a doc match
	Score    float64
}

var (
	globalRAGService *RAGService
	ragOnce          sync.Once
)

func GetRAGService() *RAGService {
	ragOnce.Do(func() {
		globalRAGService = &RAGService{
			questionRepo: repository.NewQuestionRepository(),
			vectorStore:  NewSimpleVectorStore(),
		}
		// Asynchronously load knowledge base on startup
		go func() {
			if err := globalRAGService.LoadKnowledgeBase("knowledge_base"); err != nil {
				fmt.Printf("Failed to load knowledge base: %v\n", err)
			}
		}()
	})
	return globalRAGService
}

func NewRAGService() *RAGService {
	return GetRAGService()
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
			
			// Simple chunking strategy: split by headers or paragraphs
			// Here we do a simplified split by double newlines for paragraphs
			// In a real scenario, use a better markdown parser/splitter
			text := string(content)
			parts := strings.Split(text, "\n\n")
			
			category := filepath.Base(filepath.Dir(path))
			
			for i, part := range parts {
				trimmed := strings.TrimSpace(part)
				if len(trimmed) < 20 { // Skip too short segments
					continue
				}
				chunks = append(chunks, KnowledgeChunk{
					ID:       fmt.Sprintf("%s_%d", info.Name(), i),
					Content:  trimmed,
					Category: category,
					Source:   path,
				})
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return s.vectorStore.IndexDocuments(chunks)
}

func (s *RAGService) SearchKnowledgeChunks(query string) ([]KnowledgeChunk, error) {
	results, err := s.vectorStore.Search(query, 3)
	if err != nil {
		return nil, err
	}

	var chunks []KnowledgeChunk
	for _, res := range results {
		if res.Document != nil {
			chunks = append(chunks, *res.Document)
		} else if res.Question != nil {
			// Convert question to chunk if needed, or skip
			chunks = append(chunks, KnowledgeChunk{
				ID:       fmt.Sprintf("q_%d", res.Question.ID),
				Content:  res.Question.Content,
				Category: "question_bank",
				Source:   "Interview Question DB",
			})
		}
	}
	return chunks, nil
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
	allQuestions, err := s.questionRepo.GetQuestionsByPositionAndDifficulty(position, difficulty)
	if err != nil {
		return nil, fmt.Errorf("failed to get questions: %w", err)
	}

	if err := s.vectorStore.IndexQuestions(allQuestions); err != nil {
		return nil, fmt.Errorf("failed to index questions: %w", err)
	}

	similarResults, err := s.vectorStore.Search(query, limit*2)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar questions: %w", err)
	}

	var filteredQuestions []*model.Question
	for _, result := range similarResults {
		if result.Question != nil && result.Score > 0.3 {
			filteredQuestions = append(filteredQuestions, result.Question)
		}
		if len(filteredQuestions) >= limit {
			break
		}
	}

	return filteredQuestions, nil
}

func (s *RAGService) GenerateQuestionBasedOnContext(context string, position, difficulty string) (*model.Question, error) {
	similarQuestions, err := s.SearchSimilarQuestions(context, position, difficulty, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar questions: %w", err)
	}

	if len(similarQuestions) == 0 {
		return s.createDefaultQuestion(position, difficulty)
	}

	bestQuestion := similarQuestions[0]
	return s.adaptQuestion(bestQuestion, context), nil
}

func (s *RAGService) createDefaultQuestion(position, difficulty string) (*model.Question, error) {
	question := &model.Question{
		Title:      fmt.Sprintf("%s - %s Level Question", position, difficulty),
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

type SimpleVectorStore struct {
	questions []*model.Question
	documents []KnowledgeChunk
}

func NewSimpleVectorStore() *SimpleVectorStore {
	return &SimpleVectorStore{
		questions: make([]*model.Question, 0),
		documents: make([]KnowledgeChunk, 0),
	}
}

func (s *SimpleVectorStore) IndexQuestions(questions []*model.Question) error {
	s.questions = questions
	return nil
}

func (s *SimpleVectorStore) IndexDocuments(docs []KnowledgeChunk) error {
	s.documents = docs
	return nil
}

func (s *SimpleVectorStore) Search(query string, limit int) ([]SimilarityResult, error) {
	var results []SimilarityResult

	// Search Questions
	for _, question := range s.questions {
		similarity := s.calculateSimilarity(query, question.Content+" "+question.Title)
		if similarity > 0 {
			results = append(results, SimilarityResult{
				Question: question,
				Score:    similarity,
			})
		}
	}

	// Search Documents
	for _, doc := range s.documents {
		similarity := s.calculateSimilarity(query, doc.Content)
		if similarity > 0 {
			// Copy loop var
			d := doc
			results = append(results, SimilarityResult{
				Document: &d,
				Score:    similarity,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (s *SimpleVectorStore) calculateSimilarity(query string, targetText string) float64 {
	queryWords := s.tokenize(query)
	targetWords := s.tokenize(targetText)

	if len(queryWords) == 0 || len(targetWords) == 0 {
		return 0
	}

	commonWords := 0
	querySet := make(map[string]bool)
	for _, word := range queryWords {
		querySet[word] = true
	}

	for _, word := range targetWords {
		if querySet[word] {
			commonWords++
		}
	}

	return float64(commonWords) / math.Sqrt(float64(len(queryWords)*len(targetWords)))
}

func (s *SimpleVectorStore) tokenize(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	var filtered []string

	for _, word := range words {
		if len(word) > 2 && !s.isStopWord(word) {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

func (s *SimpleVectorStore) isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "and": true, "or": true, "but": true, "in": true,
		"on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true,
	}
	return stopWords[word]
}
