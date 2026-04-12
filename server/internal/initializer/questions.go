package initializer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"your-project/config"
	"your-project/internal/model"
	ragpkg "your-project/pkg/rag"

	"gorm.io/gorm"
)

const (
	seedSourceKnowledgeBase = "knowledge_base_seed"
	seedSourceCurated       = "curated_seed"
	knowledgeEmbedBatchSize = 48
)

var (
	headingLevel3Pattern = regexp.MustCompile(`^###\s+(.+)$`)
	optionLinePattern    = regexp.MustCompile(`^[A-D][\.、．]\s+.+`)
	answerPattern        = regexp.MustCompile(`(?s)\*\*参考答案：\*\*\s*(.+?)(\n\*\*(答案解析|解析)：\*\*|$)`)
	analysisPattern      = regexp.MustCompile(`(?s)\*\*(答案解析|解析)：\*\*\s*(.+)`)
	numberPrefixPattern  = regexp.MustCompile(`^\d+[\.、．]\s*`)
)

// InitSampleQuestions initializes question bank from knowledge_base markdown files.
func InitSampleQuestions(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	questions := make([]model.Question, 0, 320)

	kbRoot, err := resolveKnowledgeBaseRoot()
	if err != nil {
		log.Printf("Warning: knowledge base not found, continue with curated seeds only: %v", err)
	} else {
		fromKB, parseErr := parseQuestionsFromKnowledgeBase(kbRoot)
		if parseErr != nil {
			return parseErr
		}
		questions = append(questions, fromKB...)
	}

	questions = append(questions, buildCuratedSeedQuestions()...)
	questions = deduplicateSeedQuestions(questions)

	if len(questions) == 0 {
		log.Printf("No seed question available for initialization")
		return nil
	}

	persisted, err := upsertSeedQuestions(db, questions)
	if err != nil {
		return err
	}
	log.Printf("Question seed complete: parsed=%d, upserted=%d", len(questions), len(persisted))

	if err := upsertQuestionsToQdrant(context.Background(), persisted); err != nil {
		log.Printf("Warning: failed to upsert seeded questions to Qdrant: %v", err)
	}

	return nil
}

func deduplicateSeedQuestions(questions []model.Question) []model.Question {
	if len(questions) == 0 {
		return []model.Question{}
	}

	seen := make(map[string]struct{}, len(questions))
	out := make([]model.Question, 0, len(questions))
	for _, item := range questions {
		key := normalizeQuestionFingerprint(item.Position, item.Title, item.Content)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}

	return out
}

func buildCuratedSeedQuestions() []model.Question {
	type template struct {
		Title          string
		Content        string
		Position       string
		Difficulty     string
		Category       string
		ExpectedAnswer string
		Tags           []string
	}

	templates := []template{
		{Title: "如何设计一个可落地的分布式锁方案？", Content: "请基于 Redis 或 ZooKeeper 说明分布式锁的实现要点，并解释互斥性、可重入、续约与故障恢复策略。", Position: "Java后端工程师", Difficulty: "social_junior", Category: "distributed_system", ExpectedAnswer: "说明锁粒度、唯一标识、原子加锁释放、看门狗续约、主从切换风险与降级策略，并给出伪代码。", Tags: []string{"backend", "分布式锁", "redis", "zookeeper"}},
		{Title: "Redis 缓存穿透如何治理？", Content: "从攻击流量与正常请求混合场景出发，设计防止缓存穿透的组合方案。", Position: "Java后端工程师", Difficulty: "campus_graduate", Category: "cache", ExpectedAnswer: "应覆盖布隆过滤器、空值缓存、参数校验、热点保护、限流熔断与监控告警。", Tags: []string{"backend", "redis", "缓存穿透", "高并发"}},
		{Title: "Redis 缓存击穿和雪崩有什么区别，如何应对？", Content: "请分别说明缓存击穿与雪崩的触发条件，并给出工程化应对策略。", Position: "Java后端工程师", Difficulty: "campus_graduate", Category: "cache", ExpectedAnswer: "需要区分单 key 失效与大面积失效，说明互斥更新、随机过期、预热、降级和多级缓存。", Tags: []string{"backend", "redis", "缓存击穿", "缓存雪崩"}},
		{Title: "MySQL 慢查询排查流程", Content: "线上接口 RT 抖动，怀疑数据库瓶颈，请给出系统排查与优化流程。", Position: "Java后端工程师", Difficulty: "campus_graduate", Category: "database", ExpectedAnswer: "包含慢日志分析、执行计划、索引设计、回表问题、SQL 改写、连接池与容量评估。", Tags: []string{"backend", "mysql", "慢查询", "索引"}},
		{Title: "消息队列如何保证至少一次消费且避免重复副作用？", Content: "请设计生产端、消费端与业务侧的幂等策略。", Position: "Java后端工程师", Difficulty: "social_junior", Category: "middleware", ExpectedAnswer: "说明发送确认、重试、死信、去重表/幂等键、状态机和补偿任务。", Tags: []string{"backend", "mq", "幂等", "一致性"}},
		{Title: "如何设计秒杀系统的限流与削峰方案？", Content: "要求兼顾高并发、库存正确性和用户体验。", Position: "Java后端工程师", Difficulty: "social_junior", Category: "system_design", ExpectedAnswer: "应覆盖前置限流、令牌桶、异步排队、库存预扣、超卖防护与监控回滚机制。", Tags: []string{"backend", "系统设计", "限流", "秒杀"}},
		{Title: "Vue3 响应式系统是如何工作的？", Content: "请解释 Vue3 中基于 Proxy 的响应式实现机制，并说明依赖收集与触发更新流程。", Position: "前端工程师", Difficulty: "campus_graduate", Category: "frontend_framework", ExpectedAnswer: "需说明 reactive/ref、effect、track/trigger、scheduler 与批量更新策略。", Tags: []string{"frontend", "vue", "响应式", "proxy"}},
		{Title: "Vue 中 watch 和 computed 的使用边界", Content: "结合一个真实页面场景，说明 watch 与 computed 应如何选择。", Position: "前端工程师", Difficulty: "campus_intern", Category: "frontend_framework", ExpectedAnswer: "强调 computed 的缓存与声明式推导、watch 的副作用场景、常见误用与性能影响。", Tags: []string{"frontend", "vue", "watch", "computed"}},
		{Title: "前端性能优化你会从哪些指标入手？", Content: "请给出从监控到优化落地的完整链路。", Position: "前端工程师", Difficulty: "campus_graduate", Category: "performance", ExpectedAnswer: "应包含 Web Vitals、资源加载、代码分包、长任务治理、缓存策略与持续监控。", Tags: []string{"frontend", "性能优化", "web vitals", "工程化"}},
		{Title: "浏览器从输入 URL 到页面渲染经历了什么？", Content: "请按阶段说明网络、解析、渲染与脚本执行关键流程。", Position: "前端工程师", Difficulty: "campus_intern", Category: "browser", ExpectedAnswer: "覆盖 DNS、TCP/TLS、HTTP、HTML/CSS 解析、渲染树、布局绘制与事件循环。", Tags: []string{"frontend", "浏览器原理", "网络", "渲染"}},
		{Title: "React 与 Vue 在状态管理上的核心差异", Content: "请从响应式机制、更新模型和工程实践角度进行对比。", Position: "前端工程师", Difficulty: "campus_graduate", Category: "frontend_framework", ExpectedAnswer: "说明单向数据流、不可变数据、依赖追踪、组件更新粒度与团队协作成本。", Tags: []string{"frontend", "react", "vue", "状态管理"}},
		{Title: "二叉树层序遍历的实现与复杂度分析", Content: "请给出 BFS 实现，并分析时间和空间复杂度。", Position: "算法工程师", Difficulty: "campus_intern", Category: "algorithm", ExpectedAnswer: "应给出队列解法，复杂度 O(n) 和最坏空间 O(n)。", Tags: []string{"algorithm", "二叉树", "bfs", "复杂度"}},
		{Title: "TopK 问题有哪些常见解法？", Content: "请比较排序、堆、快速选择在不同数据规模下的适用性。", Position: "算法工程师", Difficulty: "campus_graduate", Category: "algorithm", ExpectedAnswer: "需要对比时间复杂度、空间复杂度与稳定性，说明工程实践选择。", Tags: []string{"algorithm", "topk", "heap", "quickselect"}},
		{Title: "如何设计一个高效的 LRU 缓存结构？", Content: "给定 get/put O(1) 要求，请描述数据结构设计。", Position: "算法工程师", Difficulty: "campus_graduate", Category: "data_structure", ExpectedAnswer: "双向链表 + 哈希表，说明淘汰流程、更新策略与并发扩展思路。", Tags: []string{"algorithm", "lru", "哈希", "链表"}},
		{Title: "动态规划题目如何构建状态转移？", Content: "请以背包或路径类问题为例说明状态定义与转移。", Position: "算法工程师", Difficulty: "campus_graduate", Category: "algorithm", ExpectedAnswer: "明确状态、边界、转移顺序和空间优化策略。", Tags: []string{"algorithm", "dp", "状态转移", "优化"}},
		{Title: "RAG 系统如何做召回与重排？", Content: "请说明向量召回、关键词召回、重排模型与最终融合策略。", Position: "AI工程师", Difficulty: "social_junior", Category: "llm_application", ExpectedAnswer: "应覆盖召回通道、rerank、阈值、上下文构建、离线评测与在线监控。", Tags: []string{"ai", "rag", "召回", "重排"}},
		{Title: "Prompt 注入攻击有哪些防御手段？", Content: "请从输入治理、工具调用和输出审计三层面给出方案。", Position: "AI工程师", Difficulty: "social_junior", Category: "llm_security", ExpectedAnswer: "需说明系统提示隔离、策略检查、工具白名单、敏感操作确认和日志审计。", Tags: []string{"ai", "prompt", "安全", "agent"}},
		{Title: "向量数据库索引参数如何调优？", Content: "以 HNSW 为例，说明构建参数和查询参数的权衡。", Position: "AI工程师", Difficulty: "campus_graduate", Category: "llm_infra", ExpectedAnswer: "解释 M、efConstruction、efSearch 对召回率、延迟和内存的影响。", Tags: []string{"ai", "向量数据库", "hnsw", "调优"}},
		{Title: "多模态模型落地时如何做评测？", Content: "请设计离线与在线结合的评测指标体系。", Position: "AI工程师", Difficulty: "campus_graduate", Category: "evaluation", ExpectedAnswer: "需包含准确性、鲁棒性、延迟、成本和业务目标达成率。", Tags: []string{"ai", "多模态", "评测", "指标"}},
		{Title: "当线上模型效果回退时你的排查路径是什么？", Content: "请说明数据、特征、模型、服务链路四层排查方式。", Position: "AI工程师", Difficulty: "social_junior", Category: "mlops", ExpectedAnswer: "应提到数据漂移监控、灰度对照、版本回溯与快速止损机制。", Tags: []string{"ai", "mlops", "效果回退", "排障"}},
		{Title: "请分享一次你在团队协作中处理冲突的经历", Content: "请描述冲突背景、你的具体行动和最终结果，并说明你的反思。", Position: "Java后端工程师", Difficulty: "campus_graduate", Category: "behavioral", ExpectedAnswer: "回答需包含事实背景、沟通策略、推进动作、结果和可复用方法。", Tags: []string{"behavioral", "团队冲突", "沟通", "复盘"}},
		{Title: "项目延期且团队压力较大时你如何推进？", Content: "请说明你如何拆解目标、协同资源并稳定团队节奏。", Position: "Java后端工程师", Difficulty: "campus_graduate", Category: "behavioral", ExpectedAnswer: "应体现优先级管理、风险透明、跨团队协同和阶段性里程碑。", Tags: []string{"behavioral", "项目管理", "协作", "执行力"}},
		{Title: "面对不清晰需求你通常如何澄清并落地？", Content: "请结合一次具体经历说明你的做法。", Position: "前端工程师", Difficulty: "campus_graduate", Category: "behavioral", ExpectedAnswer: "包括需求澄清、验收标准对齐、方案评审和交付闭环。", Tags: []string{"behavioral", "需求澄清", "沟通", "交付"}},
		{Title: "你如何在技术方案分歧中推动达成共识？", Content: "请描述你处理技术争议的过程与结果。", Position: "AI工程师", Difficulty: "campus_graduate", Category: "behavioral", ExpectedAnswer: "强调数据驱动比较、风险评估、试点验证和复盘机制。", Tags: []string{"behavioral", "技术决策", "协作", "共识"}},
	}

	questions := make([]model.Question, 0, len(templates))
	for _, item := range templates {
		q := model.Question{
			Title:          truncateText(item.Title, 120),
			Content:        truncateText(item.Content, 500),
			Position:       item.Position,
			Difficulty:     item.Difficulty,
			Category:       item.Category,
			Source:         seedSourceCurated,
			RAGEligible:    true,
			ExpectedAnswer: truncateText(item.ExpectedAnswer, 2600),
		}
		tags := append([]string{"curated", item.Category, item.Difficulty}, item.Tags...)
		q.SetTags(tags)
		questions = append(questions, q)
	}

	return questions
}

func resolveKnowledgeBaseRoot() (string, error) {
	candidates := []string{
		filepath.Join("..", "knowledge_base"),
		"knowledge_base",
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if info.IsDir() {
			abs, absErr := filepath.Abs(candidate)
			if absErr != nil {
				return candidate, nil
			}
			return abs, nil
		}
	}

	return "", fmt.Errorf("knowledge_base directory not found")
}

func parseQuestionsFromKnowledgeBase(root string) ([]model.Question, error) {
	allowedDomains := map[string]bool{
		"backend":     true,
		"frontend":    true,
		"algorithm":   true,
		"ai_engineer": true,
		"behavioral":  true,
	}

	result := make([]model.Question, 0, 256)
	seen := make(map[string]struct{}, 256)

	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".md" {
			return nil
		}
		if strings.EqualFold(info.Name(), "README.md") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")
		if len(parts) < 2 {
			return nil
		}
		domain := strings.TrimSpace(parts[0])
		if !allowedDomains[domain] {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		meta := questionSeedMeta{
			Domain:      domain,
			RelativeSrc: rel,
			Category:    strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
			Position:    mapDomainToPosition(domain),
		}

		parsed := extractQuestionsFromMarkdown(string(content), meta)
		for _, item := range parsed {
			key := normalizeQuestionFingerprint(item.Position, item.Title, item.Content)
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, item)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

type questionSeedMeta struct {
	Domain      string
	RelativeSrc string
	Category    string
	Position    string
}

func extractQuestionsFromMarkdown(doc string, meta questionSeedMeta) []model.Question {
	lines := strings.Split(doc, "\n")
	questions := make([]model.Question, 0, 32)

	currentHeading := ""
	var bodyBuilder strings.Builder

	flushSection := func() {
		heading := strings.TrimSpace(currentHeading)
		body := strings.TrimSpace(bodyBuilder.String())
		bodyBuilder.Reset()
		if heading == "" {
			return
		}

		q, ok := buildQuestionFromSection(heading, body, meta)
		if !ok {
			return
		}
		questions = append(questions, q)
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if matches := headingLevel3Pattern.FindStringSubmatch(trimmed); len(matches) > 1 {
			flushSection()
			currentHeading = strings.TrimSpace(matches[1])
			continue
		}
		if currentHeading == "" {
			continue
		}
		bodyBuilder.WriteString(line)
		bodyBuilder.WriteString("\n")
	}
	flushSection()

	return questions
}

func buildQuestionFromSection(rawHeading, body string, meta questionSeedMeta) (model.Question, bool) {
	heading := cleanQuestionHeading(rawHeading)
	if !looksLikeQuestionHeading(heading, body) {
		return model.Question{}, false
	}

	questionContent := extractQuestionContent(heading, body)
	if questionContent == "" {
		questionContent = fmt.Sprintf("请围绕“%s”进行系统说明。", heading)
	}

	expectedAnswer := extractExpectedAnswer(body)
	if expectedAnswer == "" {
		expectedAnswer = truncateText(body, 1200)
	}
	if expectedAnswer == "" {
		expectedAnswer = "请结合知识点给出结构化回答，覆盖定义、原理、实现和边界。"
	}

	difficulty := inferDifficultyFromSection(heading, body)
	tags := []string{meta.Domain, meta.Category, seedSourceKnowledgeBase, difficulty}

	q := model.Question{
		Title:          truncateText(heading, 120),
		Content:        truncateText(questionContent, 500),
		Position:       meta.Position,
		Difficulty:     difficulty,
		Category:       meta.Category,
		Source:         seedSourceKnowledgeBase,
		RAGEligible:    true,
		ExpectedAnswer: truncateText(expectedAnswer, 2600),
	}
	q.SetTags(tags)
	return q, true
}

func cleanQuestionHeading(raw string) string {
	heading := strings.TrimSpace(raw)
	heading = numberPrefixPattern.ReplaceAllString(heading, "")
	heading = strings.TrimPrefix(heading, "- ")
	heading = strings.TrimSpace(heading)
	return heading
}

func looksLikeQuestionHeading(heading, body string) bool {
	if heading == "" {
		return false
	}
	headingLower := strings.ToLower(heading)
	if strings.Contains(heading, "知识点部分") || strings.Contains(heading, "题库部分") || strings.Contains(heading, "扩展") || strings.Contains(heading, "加量") {
		return false
	}
	if strings.Contains(headingLower, "knowledge points") || strings.Contains(headingLower, "appendix") {
		return false
	}
	if strings.Contains(heading, "单选题") || strings.Contains(heading, "简答题") {
		return true
	}
	if strings.ContainsAny(heading, "？?") {
		return true
	}
	if strings.Contains(body, "**解析：**") || strings.Contains(body, "**参考答案：**") {
		return true
	}
	return false
}

func extractQuestionContent(heading, body string) string {
	content := heading
	if idx := strings.IndexAny(heading, "？?"); idx >= 0 && idx+1 < len(heading) {
		after := strings.TrimSpace(heading[idx+1:])
		if after != "" {
			content = after
		}
	}

	options := collectOptionLines(body)
	if len(options) > 0 {
		content = strings.TrimSpace(content + "\n" + strings.Join(options, "\n"))
	}
	return content
}

func collectOptionLines(body string) []string {
	lines := strings.Split(body, "\n")
	options := make([]string, 0, 4)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if optionLinePattern.MatchString(trimmed) {
			options = append(options, trimmed)
		}
	}
	return options
}

func extractExpectedAnswer(body string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}

	if matches := answerPattern.FindStringSubmatch(trimmed); len(matches) > 1 {
		answer := strings.TrimSpace(matches[1])
		if answer != "" {
			return answer
		}
	}

	if matches := analysisPattern.FindStringSubmatch(trimmed); len(matches) > 2 {
		analysis := strings.TrimSpace(matches[2])
		if analysis != "" {
			return analysis
		}
	}

	return ""
}

func inferDifficultyFromSection(heading, body string) string {
	text := strings.ToLower(strings.TrimSpace(heading + " " + body))
	switch {
	case strings.Contains(text, "单选题"):
		return "campus_intern"
	case strings.Contains(text, "高并发"), strings.Contains(text, "分布式"), strings.Contains(text, "多活"), strings.Contains(text, "系统设计"), strings.Contains(text, "架构"):
		return "social_junior"
	case strings.Contains(text, "简答题"):
		return "campus_graduate"
	default:
		return "campus_graduate"
	}
}

func mapDomainToPosition(domain string) string {
	switch strings.TrimSpace(domain) {
	case "backend":
		return "Java后端工程师"
	case "frontend":
		return "前端工程师"
	case "algorithm":
		return "算法工程师"
	case "ai_engineer":
		return "AI工程师"
	case "behavioral":
		return "Java后端工程师"
	default:
		return "Java后端工程师"
	}
}

func normalizeQuestionFingerprint(position, title, content string) string {
	joined := strings.ToLower(strings.TrimSpace(position + "|" + title + "|" + content))
	joined = strings.ReplaceAll(joined, " ", "")
	joined = strings.ReplaceAll(joined, "\n", "")
	joined = strings.ReplaceAll(joined, "\t", "")
	joined = strings.ReplaceAll(joined, "。", "")
	joined = strings.ReplaceAll(joined, "，", "")
	joined = strings.ReplaceAll(joined, "：", "")
	joined = strings.ReplaceAll(joined, ":", "")
	joined = strings.ReplaceAll(joined, "？", "")
	joined = strings.ReplaceAll(joined, "?", "")
	joined = strings.ReplaceAll(joined, "（", "")
	joined = strings.ReplaceAll(joined, "）", "")
	joined = strings.ReplaceAll(joined, "(", "")
	joined = strings.ReplaceAll(joined, ")", "")
	return joined
}

func truncateText(text string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= max {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[:max]))
}

func upsertSeedQuestions(db *gorm.DB, questions []model.Question) ([]*model.Question, error) {
	persisted := make([]*model.Question, 0, len(questions))
	for _, candidate := range questions {
		stored, err := upsertSingleQuestion(db, candidate)
		if err != nil {
			return nil, err
		}
		if stored != nil {
			persisted = append(persisted, stored)
		}
	}
	return persisted, nil
}

func upsertSingleQuestion(db *gorm.DB, candidate model.Question) (*model.Question, error) {
	var existing model.Question
	err := db.Where("title = ? AND content = ? AND position = ? AND difficulty = ?", candidate.Title, candidate.Content, candidate.Position, candidate.Difficulty).First(&existing).Error
	if err == nil {
		existing.Category = candidate.Category
		existing.Source = candidate.Source
		existing.RAGEligible = true
		existing.ExpectedAnswer = candidate.ExpectedAnswer
		existing.Tags = candidate.Tags
		if saveErr := db.Save(&existing).Error; saveErr != nil {
			return nil, fmt.Errorf("failed to update seed question: %w", saveErr)
		}
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query seed question: %w", err)
	}

	if createErr := db.Create(&candidate).Error; createErr != nil {
		return nil, fmt.Errorf("failed to create seed question: %w", createErr)
	}
	return &candidate, nil
}

func upsertQuestionsToQdrant(ctx context.Context, questions []*model.Question) error {
	if len(questions) == 0 {
		return nil
	}

	store, err := ragpkg.NewQdrantStoreFromEnv()
	if err != nil {
		return err
	}

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()
	if err := store.Ping(pingCtx); err != nil {
		return err
	}

	embedder, err := buildInitializerEmbedder()
	if err != nil {
		return err
	}

	points := make([]ragpkg.VectorPoint, 0, len(questions))
	for _, q := range questions {
		if q == nil || q.ID == 0 {
			continue
		}
		text := buildQuestionEmbeddingText(q)
		if strings.TrimSpace(text) == "" {
			continue
		}

		vector, embErr := embedder.Embed(ctx, text)
		if embErr != nil {
			continue
		}

		points = append(points, ragpkg.VectorPoint{
			ID:      fmt.Sprintf("question_%d", q.ID),
			Vector:  vector,
			Content: text,
			Metadata: map[string]string{
				"kind":        "question",
				"question_id": fmt.Sprintf("%d", q.ID),
				"position":    strings.TrimSpace(q.Position),
				"difficulty":  strings.TrimSpace(q.Difficulty),
				"category":    strings.TrimSpace(q.Category),
				"source":      strings.TrimSpace(q.Source),
			},
		})
	}

	if len(points) == 0 {
		return nil
	}

	for start := 0; start < len(points); start += knowledgeEmbedBatchSize {
		end := start + knowledgeEmbedBatchSize
		if end > len(points) {
			end = len(points)
		}
		if err := store.Upsert(ctx, points[start:end]); err != nil {
			return err
		}
	}
	return nil
}

func buildInitializerEmbedder() (ragpkg.Embedder, error) {
	embedder, err := ragpkg.NewOpenAIEmbedderFromEnv()
	if err == nil {
		return embedder, nil
	}

	cfg := config.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("embedding config not found")
	}

	modelName := strings.TrimSpace(os.Getenv("EMBEDDING_MODEL"))
	if modelName == "" {
		modelName = strings.TrimSpace(cfg.LLM.Models["embedding"])
	}

	return ragpkg.NewOpenAIEmbedder(ragpkg.OpenAIEmbedderConfig{
		APIKey:  strings.TrimSpace(cfg.LLM.APIKey),
		BaseURL: strings.TrimSpace(cfg.LLM.BaseURL),
		Model:   modelName,
	})
}

func buildQuestionEmbeddingText(q *model.Question) string {
	if q == nil {
		return ""
	}
	parts := []string{
		strings.TrimSpace(q.Title),
		strings.TrimSpace(q.Content),
	}
	if strings.TrimSpace(q.ExpectedAnswer) != "" {
		parts = append(parts, "参考答案："+strings.TrimSpace(q.ExpectedAnswer))
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
