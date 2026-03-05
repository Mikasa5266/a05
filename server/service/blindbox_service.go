package service

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"your-project/model"
)

// BlindBoxScenario represents a randomly generated interview scenario
type BlindBoxScenario struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Pressure    string   `json:"pressure"` // "low", "medium", "high", "extreme"
	Tags        []string `json:"tags"`
	Style       string   `json:"style"`      // overrides interview style
	TimeLimit   int      `json:"time_limit"` // seconds per question (0 = unlimited)
	Icon        string   `json:"icon"`       // emoji for frontend
}

// scenarioPool defines all possible blind box scenarios with gradient pressure levels
var scenarioPool = []BlindBoxScenario{
	// === Low Pressure ===
	{
		ID: "casual_chat", Name: "轻松闲聊", Pressure: "low",
		Description: "像朋友聊天一样，面试官会以轻松方式了解你的技术背景和项目经历。",
		Tags:        []string{"轻松", "自由表达"}, Style: "gentle", TimeLimit: 0, Icon: "☕",
	},
	{
		ID: "show_and_tell", Name: "技术分享", Pressure: "low",
		Description: "向面试官展示你最自豪的项目或技术方案，像做一次 TED 演讲。",
		Tags:        []string{"表达力", "项目经验"}, Style: "gentle", TimeLimit: 0, Icon: "🎤",
	},
	{
		ID: "pair_thinking", Name: "结对思考", Pressure: "low",
		Description: "面试官和你一起探讨一个技术问题，没有标准答案，重在思考过程。",
		Tags:        []string{"协作", "开放思维"}, Style: "gentle", TimeLimit: 0, Icon: "🤝",
	},
	// === Medium Pressure ===
	{
		ID: "rapid_fire", Name: "快问快答", Pressure: "medium",
		Description: "连续回答简短的技术概念题，考验你的知识广度和反应速度。每题限时 60 秒。",
		Tags:        []string{"速度", "广度"}, Style: "gentle", TimeLimit: 60, Icon: "⚡",
	},
	{
		ID: "case_study", Name: "案例分析", Pressure: "medium",
		Description: "给你一个真实的系统设计案例，需要分析问题并给出解决方案。",
		Tags:        []string{"系统设计", "分析力"}, Style: "deep", TimeLimit: 180, Icon: "📊",
	},
	{
		ID: "debug_challenge", Name: "Debug 挑战", Pressure: "medium",
		Description: "面试官描述一个线上 Bug 场景，你需要快速定位问题根因并给出修复思路。",
		Tags:        []string{"排查能力", "实战"}, Style: "deep", TimeLimit: 120, Icon: "🐛",
	},
	{
		ID: "reverse_interview", Name: "反向面试", Pressure: "medium",
		Description: "面试官扮演初级开发者向你请教问题，考验你解释复杂概念的能力。",
		Tags:        []string{"表达力", "教学能力"}, Style: "gentle", TimeLimit: 0, Icon: "🔄",
	},
	// === High Pressure ===
	{
		ID: "deep_dive", Name: "灵魂追问", Pressure: "high",
		Description: "面试官会对每个回答追问 3-5 层「为什么」，直到触及知识边界。准备好被挑战！",
		Tags:        []string{"深度", "原理"}, Style: "deep", TimeLimit: 90, Icon: "🔍",
	},
	{
		ID: "challenge_mode", Name: "质疑模式", Pressure: "high",
		Description: "面试官会主动质疑你的每个回答，即使你说得对也会反驳，考验你在压力下的自信和逻辑。",
		Tags:        []string{"抗压", "辩论"}, Style: "stress", TimeLimit: 90, Icon: "⚔️",
	},
	{
		ID: "system_crash", Name: "线上炸了", Pressure: "high",
		Description: "模拟场景：核心服务宕机、数据库崩溃、流量暴增……考验你处理突发状况的思路。",
		Tags:        []string{"应急", "架构"}, Style: "stress", TimeLimit: 120, Icon: "🔥",
	},
	{
		ID: "whiteboard", Name: "白板编码", Pressure: "high",
		Description: "限时完成算法题，面试官会追问时间复杂度和优化思路。",
		Tags:        []string{"算法", "编码"}, Style: "deep", TimeLimit: 300, Icon: "📝",
	},
	// === Extreme Pressure ===
	{
		ID: "final_boss", Name: "终极 Boss 面", Pressure: "extreme",
		Description: "CTO 级别的终面模拟：跨领域追问 + 系统设计 + 价值观考察，全方位无死角！",
		Tags:        []string{"全方位", "综合"}, Style: "stress", TimeLimit: 120, Icon: "👑",
	},
	{
		ID: "silence_pressure", Name: "沉默施压", Pressure: "extreme",
		Description: "面试官在你回答后长时间沉默，不给任何反馈，考验你在沉默压力下的心态和补充能力。",
		Tags:        []string{"心理素质", "抗压"}, Style: "stress", TimeLimit: 60, Icon: "🤫",
	},
	{
		ID: "impossible_question", Name: "不可能的问题", Pressure: "extreme",
		Description: "面试官会提出一些看似不可能回答的开放性问题，考验你的创造力和思维方式。",
		Tags:        []string{"创造力", "极限思维"}, Style: "stress", TimeLimit: 90, Icon: "🌀",
	},
}

// pressureWeights defines weighted selection by gradient level
var pressureWeights = map[string]int{
	"low":     20,
	"medium":  40,
	"high":    30,
	"extreme": 10,
}

// BlindBoxService handles blind box scenario generation
type BlindBoxService struct {
	aiService *AIService
	rng       *rand.Rand
}

func NewBlindBoxService() *BlindBoxService {
	return &BlindBoxService{
		aiService: NewAIService(),
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// DrawScenario randomly selects a scenario using weighted pressure distribution
func (s *BlindBoxService) DrawScenario() *BlindBoxScenario {
	// Build weighted pool
	var pool []BlindBoxScenario
	for _, sc := range scenarioPool {
		weight := pressureWeights[sc.Pressure]
		for i := 0; i < weight; i++ {
			pool = append(pool, sc)
		}
	}
	selected := pool[s.rng.Intn(len(pool))]
	return &selected
}

// DrawScenarioByPressure draws a scenario with a minimum pressure level
func (s *BlindBoxService) DrawScenarioByPressure(minPressure string) *BlindBoxScenario {
	pressureLevels := map[string]int{"low": 0, "medium": 1, "high": 2, "extreme": 3}
	minLevel := pressureLevels[minPressure]

	var candidates []BlindBoxScenario
	for _, sc := range scenarioPool {
		if pressureLevels[sc.Pressure] >= minLevel {
			candidates = append(candidates, sc)
		}
	}
	if len(candidates) == 0 {
		candidates = scenarioPool
	}
	selected := candidates[s.rng.Intn(len(candidates))]
	return &selected
}

// GenerateBlindBoxQuestions generates questions tailored to a specific scenario
func (s *BlindBoxService) GenerateBlindBoxQuestions(scenario *BlindBoxScenario, position, difficulty string, count int) ([]*model.Question, error) {
	pressureInstruction := s.buildPressurePrompt(scenario)

	prompt := fmt.Sprintf(`
你是一位资深面试官。当前面试采用"盲盒面试"模式，抽到的场景如下：

【场景名称】%s
【场景描述】%s
【压力等级】%s
【面试风格】%s

候选人信息：
- 应聘岗位：%s
- 难度级别：%s

%s

请根据场景特点生成 %d 个面试题目。

要求：
1. 题目必须贴合场景主题和压力等级。
2. 如果是高压/极限场景，题目应该具有挑战性、需要深入思考、甚至带有一定压迫感。
3. 如果是低压场景，题目应该开放、友好、鼓励自由表达。
4. 所有内容必须使用简体中文。
5. 返回格式必须为 JSON 数组：
[
  {"title": "题目标题", "content": "具体问题内容", "expected_answer": "参考答案要点"}
]
`, scenario.Name, scenario.Description, scenario.Pressure, scenario.Style,
		position, difficulty, pressureInstruction, count)

	response, err := s.aiService.callLLM(prompt, "chat")
	if err != nil {
		return nil, fmt.Errorf("failed to generate blindbox questions: %w", err)
	}

	var questionsData []struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}

	cleanResponse := extractJSONContent(response)
	if err := json.Unmarshal([]byte(cleanResponse), &questionsData); err != nil {
		return nil, fmt.Errorf("failed to parse blindbox questions: %w, raw: %s", err, response)
	}

	var questions []*model.Question
	for _, qd := range questionsData {
		q := &model.Question{
			Title:          qd.Title,
			Content:        qd.Content,
			ExpectedAnswer: qd.ExpectedAnswer,
			Position:       position,
			Difficulty:     difficulty,
			Category:       "blindbox_" + scenario.ID,
		}
		s.aiService.EnsureQuestionChinese(q)
		questions = append(questions, q)
	}

	return questions, nil
}

// buildPressurePrompt generates scenario-specific pressure instructions for the AI
func (s *BlindBoxService) buildPressurePrompt(scenario *BlindBoxScenario) string {
	switch scenario.Pressure {
	case "low":
		return `【低压模式指令】
- 题目语气亲切友好，像同事间的技术交流。
- 多使用"你觉得"、"可以聊聊"等引导词。
- 允许候选人自由发挥，不必追求标准答案。`

	case "medium":
		return `【中等压力指令】
- 题目需要有一定技术深度，但保持公平。
- 可以设置具体的约束条件（如时间限制、资源限制）。
- 鼓励候选人展示分析过程。`

	case "high":
		return fmt.Sprintf(`【高压模式指令】
- 题目必须具有挑战性，考察知识深度和临场应变。
- 在题目中加入追问方向（如"请进一步解释为什么"），模拟面试官连续追问。
- 场景主题：%s — 请围绕这个主题设计压力情景。
- 每题限时 %d 秒，请在题目中提示候选人注意时间。`, scenario.Name, scenario.TimeLimit)

	case "extreme":
		return fmt.Sprintf(`【极限压力指令】
- 这是最高难度的面试场景，题目需要极强的综合能力。
- 可以包含：跨领域知识、开放性问题、看似矛盾的条件。
- 不要给候选人太多提示，考验独立思考能力。
- 场景特色：%s
- 每题限时 %d 秒，营造紧迫感。
- 面试官风格：不轻易认可，会反复质疑。`, scenario.Description, scenario.TimeLimit)

	default:
		return ""
	}
}

// GetAllScenarios returns all available scenarios (for admin/preview)
func GetAllScenarios() []BlindBoxScenario {
	return scenarioPool
}

// ScenarioToJSON serializes a scenario for storage in the Interview model
func ScenarioToJSON(scenario *BlindBoxScenario) string {
	data, _ := json.Marshal(scenario)
	return string(data)
}

// ScenarioFromJSON deserializes a scenario from Interview.Scenario field
func ScenarioFromJSON(data string) *BlindBoxScenario {
	if data == "" {
		return nil
	}
	var s BlindBoxScenario
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return nil
	}
	return &s
}
