package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"your-project/service"

	"github.com/gin-gonic/gin"
)

type algorithmProblem struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Prompt          string   `json:"prompt"`
	Level           string   `json:"level"`
	Tags            []string `json:"tags"`
	MethodHint      string   `json:"method_hint"`
	RequirementHint string   `json:"requirement_hint"`
	TestCaseCount   int      `json:"test_case_count"`
	JudgeToken      string   `json:"-"`
	FailCaseHint    string   `json:"-"`
}

type algorithmSessionState struct {
	SessionID string
	Problems  []algorithmProblem
	CreatedAt time.Time
}

var algorithmSessionStore = struct {
	sync.RWMutex
	items map[uint]algorithmSessionState
}{
	items: map[uint]algorithmSessionState{},
}

func buildAlgorithmProblemsByDifficulty(difficulty string) []algorithmProblem {
	if difficulty == "social_junior" {
		return []algorithmProblem{
			{
				ID:              "alg-hard-1",
				Title:           "最短路径并统计最少边数",
				Prompt:          "给定 n 个点和带权无向边，返回 1 到 n 的最短路径长度；若有多条最短路径，返回其中边数最少的一条。",
				Level:           "hard",
				Tags:            []string{"图论", "Dijkstra", "最短路"},
				MethodHint:      "Dijkstra + 状态维护",
				RequirementHint: "时间复杂度 O((n+m)logn)",
				TestCaseCount:   6,
				JudgeToken:      "priority_queue",
				FailCaseHint:    "case#3: 多条最短路时边数选择错误",
			},
			{
				ID:              "alg-hard-2",
				Title:           "区间最大重叠数",
				Prompt:          "给定若干闭区间 [l, r]，返回任意时刻最大重叠区间数。",
				Level:           "hard",
				Tags:            []string{"扫描线", "排序"},
				MethodHint:      "扫描线 + 事件排序",
				RequirementHint: "时间复杂度 O(nlogn)",
				TestCaseCount:   7,
				JudgeToken:      "sort",
				FailCaseHint:    "case#5: 边界点重叠处理错误",
			},
		}
	}

	return []algorithmProblem{
		{
			ID:              "alg-mid-1",
			Title:           "两数之和 II",
			Prompt:          "给定有序数组 nums 和目标值 target，返回两个下标使 nums[i] + nums[j] = target。",
			Level:           "easy",
			Tags:            []string{"双指针", "数组"},
			MethodHint:      "双指针",
			RequirementHint: "时间复杂度 O(n)",
			TestCaseCount:   5,
			JudgeToken:      "while",
			FailCaseHint:    "case#2: 没有正确移动左右指针",
		},
		{
			ID:              "alg-mid-2",
			Title:           "最长不重复子串",
			Prompt:          "给定字符串 s，返回不含重复字符的最长子串长度。",
			Level:           "medium",
			Tags:            []string{"滑动窗口", "哈希"},
			MethodHint:      "滑动窗口",
			RequirementHint: "时间复杂度 O(n)",
			TestCaseCount:   6,
			JudgeToken:      "map",
			FailCaseHint:    "case#4: 重复字符窗口收缩异常",
		},
		{
			ID:              "alg-mid-3",
			Title:           "合并区间",
			Prompt:          "给定若干区间，合并所有重叠区间并返回结果。",
			Level:           "medium",
			Tags:            []string{"排序", "区间"},
			MethodHint:      "排序后线性合并",
			RequirementHint: "时间复杂度 O(nlogn)",
			TestCaseCount:   5,
			JudgeToken:      "sort",
			FailCaseHint:    "case#1: 起始区间未正确初始化",
		},
	}
}

func GetAlgorithmSession(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}
	interviewID := uint(interviewID64)

	svc := service.NewInterviewService()
	type interviewLookupResult struct {
		difficulty string
		err        error
	}
	resultCh := make(chan interviewLookupResult, 1)
	go func() {
		interview, err := svc.GetInterviewByID(userID, interviewID)
		if err != nil {
			resultCh <- interviewLookupResult{
				err: err,
			}
			return
		}
		resultCh <- interviewLookupResult{
			difficulty: interview.Difficulty,
		}
	}()
	var interviewDifficulty string
	select {
	case <-time.After(coreRequestTimeout):
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "请求超时，请稍后重试"})
		return
	case <-c.Request.Context().Done():
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "请求已取消"})
		return
	case result := <-resultCh:
		if result.err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
			return
		}
		interviewDifficulty = result.difficulty
	}

	algorithmSessionStore.Lock()
	state, exists := algorithmSessionStore.items[interviewID]
	if !exists {
		problems := buildAlgorithmProblemsByDifficulty(interviewDifficulty)
		state = algorithmSessionState{
			SessionID: fmt.Sprintf("alg-%d-%d", interviewID, time.Now().Unix()),
			Problems:  problems,
			CreatedAt: time.Now(),
		}
		algorithmSessionStore.items[interviewID] = state
	}
	algorithmSessionStore.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"session_id": state.SessionID,
		"problems":   state.Problems,
	})
}

func RunAlgorithmCode(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}
	interviewID := uint(interviewID64)

	svc := service.NewInterviewService()
	checkCh := make(chan error, 1)
	go func() {
		_, err := svc.GetInterviewByID(userID, interviewID)
		checkCh <- err
	}()
	select {
	case <-time.After(coreRequestTimeout):
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "请求超时，请稍后重试"})
		return
	case <-c.Request.Context().Done():
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "请求已取消"})
		return
	case err := <-checkCh:
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
			return
		}
	}

	var req struct {
		SessionID string `json:"session_id"`
		ProblemID string `json:"problem_id" binding:"required"`
		Language  string `json:"language" binding:"required"`
		Code      string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	algorithmSessionStore.RLock()
	state, ok := algorithmSessionStore.items[interviewID]
	algorithmSessionStore.RUnlock()
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先初始化算法会话"})
		return
	}

	if req.SessionID != "" && req.SessionID != state.SessionID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "算法会话已失效，请刷新重试"})
		return
	}

	var target *algorithmProblem
	for i := range state.Problems {
		if state.Problems[i].ID == req.ProblemID {
			target = &state.Problems[i]
			break
		}
	}
	if target == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "题目不存在"})
		return
	}

	codeLower := strings.ToLower(req.Code)
	langLower := strings.ToLower(req.Language)

	baseChecks := []string{"solve", "return"}
	for _, check := range baseChecks {
		if !strings.Contains(codeLower, check) {
			c.JSON(http.StatusOK, gin.H{
				"passed":      false,
				"failed_case": "case#0: 缺少基础函数结构",
				"message":     "未通过，请补全函数结构后重试",
			})
			return
		}
	}

	if langLower == "java" && !strings.Contains(codeLower, "class") {
		c.JSON(http.StatusOK, gin.H{
			"passed":      false,
			"failed_case": "case#0: Java 代码缺少 class 结构",
			"message":     "Java 代码结构不完整",
		})
		return
	}

	passed := strings.Contains(codeLower, strings.ToLower(target.JudgeToken))
	if !passed {
		c.JSON(http.StatusOK, gin.H{
			"passed":      false,
			"failed_case": target.FailCaseHint,
			"message":     "未通过，至少有一个测试案例失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"passed":      true,
		"failed_case": "",
		"message":     "全部测试案例通过",
	})
}

func SkipAlgorithmProblem(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}
	interviewID := uint(interviewID64)

	svc := service.NewInterviewService()
	checkCh := make(chan error, 1)
	go func() {
		_, err := svc.GetInterviewByID(userID, interviewID)
		checkCh <- err
	}()
	select {
	case <-time.After(coreRequestTimeout):
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "请求超时，请稍后重试"})
		return
	case <-c.Request.Context().Done():
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "请求已取消"})
		return
	case err := <-checkCh:
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
			return
		}
	}

	var req struct {
		SessionID string `json:"session_id"`
		ProblemID string `json:"problem_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "已跳过，按 0 分处理",
		"problem_id": req.ProblemID,
	})
}
