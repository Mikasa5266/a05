package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"your-project/config"
	"your-project/internal/initializer"
	"your-project/internal/model"
	"your-project/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	repository.SetDB(db)

	if err := rebuildDemoData(db); err != nil {
		log.Fatalf("Failed to rebuild demo data: %v", err)
	}

	fmt.Println("Demo seed completed.")
	fmt.Println("Students: student01@test.com ... student05@test.com / 123456")
	fmt.Println("Enterprise: enterprise01@test.com ... enterprise03@test.com / 123456")
	fmt.Println("University: university01@test.com, university02@test.com / 123456")
}

func initDatabase() (*gorm.DB, error) {
	cfg := config.GetConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

type accountSeed struct {
	Username string
	Email    string
	Role     string
	Password string
	Avatar   string
}

type enterpriseSeed struct {
	Email         string
	CompanyName   string
	ContactName   string
	ContactPhone  string
	BusinessScope string
}

type universitySeed struct {
	Email          string
	UniversityName string
	ContactName    string
	ContactPhone   string
	Department     string
}

func loadConfig() error {
	candidates := []string{
		"config.yaml",
		filepath.Join("..", "..", "config.yaml"),
	}

	for _, path := range candidates {
		if err := config.LoadConfig(path); err == nil {
			return nil
		}
	}

	return fmt.Errorf("config.yaml not found in expected locations")
}

func rebuildDemoData(db *gorm.DB) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := cleanupSeedData(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := repository.NewPositionRepositoryWithDB(tx).EnsureDefaults(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to ensure positions: %w", err)
	}

	usersByEmail, err := seedUsers(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	enterprises, err := seedEnterprises(tx, usersByEmail)
	if err != nil {
		tx.Rollback()
		return err
	}

	universities, err := seedUniversities(tx, usersByEmail)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := seedJobs(tx, enterprises); err != nil {
		tx.Rollback()
		return err
	}

	if err := seedUniversityData(tx, universities, usersByEmail); err != nil {
		tx.Rollback()
		return err
	}

	if err := seedCommunityPosts(tx, usersByEmail); err != nil {
		tx.Rollback()
		return err
	}

	if err := initializer.InitSampleQuestions(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to seed questions: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func cleanupSeedData(tx *gorm.DB) error {
	if err := tx.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	models := []struct {
		name string
		item interface{}
	}{
		{name: "post_likes", item: &model.PostLike{}},
		{name: "post_comments", item: &model.PostComment{}},
		{name: "community_posts", item: &model.CommunityPost{}},
		{name: "mentor_bookings", item: &model.MentorBooking{}},
		{name: "talent_pushes", item: &model.TalentPush{}},
		{name: "courses", item: &model.Course{}},
		{name: "student_records", item: &model.StudentRecord{}},
		{name: "universities", item: &model.University{}},
		{name: "referrals", item: &model.Referral{}},
		{name: "capability_standards", item: &model.CapabilityStandard{}},
		{name: "interview_sessions", item: &model.InterviewSession{}},
		{name: "talent_records", item: &model.TalentRecord{}},
		{name: "jobs", item: &model.Job{}},
		{name: "enterprises", item: &model.Enterprise{}},
		{name: "question_favorites", item: &model.PracticeQuestionFavorite{}},
		{name: "question_wrong_books", item: &model.PracticeWrongBookEntry{}},
		{name: "question_list_items", item: &model.PracticeQuestionListItem{}},
		{name: "question_lists", item: &model.PracticeQuestionList{}},
		{name: "question_practice_records", item: &model.QuestionPracticeRecord{}},
		{name: "question_assessment_answers", item: &model.PracticeAssessmentAnswer{}},
		{name: "question_assessment_items", item: &model.QuestionAssessmentItem{}},
		{name: "question_assessments", item: &model.QuestionAssessment{}},
		{name: "answer_results", item: &model.AnswerResult{}},
		{name: "interview_questions", item: &model.InterviewQuestion{}},
		{name: "reports", item: &model.Report{}},
		{name: "human_interview_invitations", item: &model.HumanInterviewInvitation{}},
		{name: "interview_bookings", item: &model.InterviewBooking{}},
		{name: "interviews", item: &model.Interview{}},
		{name: "resume_parse_results", item: &model.ResumeParseResult{}},
		{name: "questions", item: &model.Question{}},
		{name: "users", item: &model.User{}},
		{name: "job_positions", item: &model.JobPosition{}},
	}

	for _, m := range models {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(m.item).Error; err != nil {
			_ = tx.Exec("SET FOREIGN_KEY_CHECKS = 1").Error
			return fmt.Errorf("failed to cleanup %s: %w", m.name, err)
		}
	}

	if err := tx.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		return fmt.Errorf("failed to re-enable foreign key checks: %w", err)
	}

	return nil
}

func seedUsers(tx *gorm.DB) (map[string]model.User, error) {
	seeds := []accountSeed{
		{Username: "student01", Email: "student01@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student01"},
		{Username: "student02", Email: "student02@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student02"},
		{Username: "student03", Email: "student03@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student03"},
		{Username: "student04", Email: "student04@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student04"},
		{Username: "student05", Email: "student05@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student05"},
		{Username: "enterprise01", Email: "enterprise01@test.com", Role: "enterprise", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=enterprise01"},
		{Username: "enterprise02", Email: "enterprise02@test.com", Role: "enterprise", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=enterprise02"},
		{Username: "enterprise03", Email: "enterprise03@test.com", Role: "enterprise", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=enterprise03"},
		{Username: "university01", Email: "university01@test.com", Role: "university", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=university01"},
		{Username: "university02", Email: "university02@test.com", Role: "university", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=university02"},
	}

	usersByEmail := make(map[string]model.User, len(seeds))
	for _, item := range seeds {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password for %s: %w", item.Email, err)
		}

		user := model.User{
			Username: item.Username,
			Email:    item.Email,
			Password: string(hashedPassword),
			Role:     item.Role,
			Avatar:   item.Avatar,
		}
		if err := tx.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to create user %s: %w", item.Email, err)
		}
		usersByEmail[item.Email] = user
	}

	return usersByEmail, nil
}

func seedEnterprises(tx *gorm.DB, usersByEmail map[string]model.User) (map[string]model.Enterprise, error) {
	seeds := []enterpriseSeed{
		{Email: "enterprise01@test.com", CompanyName: "星瀚科技", ContactName: "王晨", ContactPhone: "13800010001", BusinessScope: "企业级 SaaS 与中台研发"},
		{Email: "enterprise02@test.com", CompanyName: "远川数据", ContactName: "李航", ContactPhone: "13800010002", BusinessScope: "数据中台与 AI 应用平台"},
		{Email: "enterprise03@test.com", CompanyName: "海岳智能", ContactName: "周宁", ContactPhone: "13800010003", BusinessScope: "智能制造软件与工业互联网"},
	}

	result := make(map[string]model.Enterprise, len(seeds))
	for _, item := range seeds {
		user, ok := usersByEmail[item.Email]
		if !ok {
			return nil, fmt.Errorf("enterprise user not found: %s", item.Email)
		}

		entity := model.Enterprise{
			UserID:        user.ID,
			CompanyName:   item.CompanyName,
			ContactName:   item.ContactName,
			ContactPhone:  item.ContactPhone,
			BusinessScope: item.BusinessScope,
			AuditStatus:   "approved",
			AuditRemark:   "演示环境自动审核通过",
		}

		if err := tx.Create(&entity).Error; err != nil {
			return nil, fmt.Errorf("failed to create enterprise profile %s: %w", item.CompanyName, err)
		}
		result[item.Email] = entity
	}

	return result, nil
}

func seedUniversities(tx *gorm.DB, usersByEmail map[string]model.User) (map[string]model.University, error) {
	seeds := []universitySeed{
		{Email: "university01@test.com", UniversityName: "华东数字科技大学", ContactName: "陈老师", ContactPhone: "13900020001", Department: "计算机学院"},
		{Email: "university02@test.com", UniversityName: "南方应用理工大学", ContactName: "赵老师", ContactPhone: "13900020002", Department: "软件工程学院"},
	}

	result := make(map[string]model.University, len(seeds))
	for _, item := range seeds {
		user, ok := usersByEmail[item.Email]
		if !ok {
			return nil, fmt.Errorf("university user not found: %s", item.Email)
		}

		entity := model.University{
			UserID:         user.ID,
			UniversityName: item.UniversityName,
			ContactName:    item.ContactName,
			ContactPhone:   item.ContactPhone,
			Department:     item.Department,
			AuditStatus:    "approved",
			AuditRemark:    "演示环境自动审核通过",
		}

		if err := tx.Create(&entity).Error; err != nil {
			return nil, fmt.Errorf("failed to create university profile %s: %w", item.UniversityName, err)
		}
		result[item.Email] = entity
	}

	return result, nil
}

func seedJobs(tx *gorm.DB, enterprises map[string]model.Enterprise) error {
	type jobSeed struct {
		EnterpriseEmail string
		Title           string
		Department      string
		Location        string
		SalaryRange     string
		Description     string
		Requirements    string
	}

	seeds := []jobSeed{
		{EnterpriseEmail: "enterprise01@test.com", Title: "Java后端开发工程师", Department: "基础平台", Location: "上海", SalaryRange: "18k-30k", Description: "负责交易与订单核心服务研发", Requirements: "熟悉 Spring Boot、MySQL、Redis、消息队列"},
		{EnterpriseEmail: "enterprise01@test.com", Title: "前端开发工程师", Department: "用户体验", Location: "上海", SalaryRange: "16k-28k", Description: "负责 Web 前端架构与性能优化", Requirements: "熟悉 Vue3/React、TypeScript、工程化体系"},
		{EnterpriseEmail: "enterprise02@test.com", Title: "数据工程师", Department: "数据平台", Location: "杭州", SalaryRange: "20k-35k", Description: "建设离线/实时数据链路与指标体系", Requirements: "熟悉 Flink、Kafka、ClickHouse"},
		{EnterpriseEmail: "enterprise02@test.com", Title: "AI应用工程师", Department: "智能产品", Location: "杭州", SalaryRange: "22k-38k", Description: "负责 Agent 与 RAG 应用落地", Requirements: "熟悉 LLM API、向量检索、Prompt 工程"},
		{EnterpriseEmail: "enterprise03@test.com", Title: "测试开发工程师", Department: "质量平台", Location: "深圳", SalaryRange: "15k-24k", Description: "建设自动化测试与质量门禁平台", Requirements: "熟悉 Python/Go，CI/CD，测试框架"},
		{EnterpriseEmail: "enterprise03@test.com", Title: "DevOps工程师", Department: "运维平台", Location: "深圳", SalaryRange: "20k-32k", Description: "负责云原生发布与可观测体系", Requirements: "熟悉 Kubernetes、GitOps、监控告警"},
	}

	for _, item := range seeds {
		enterprise, ok := enterprises[item.EnterpriseEmail]
		if !ok {
			return fmt.Errorf("enterprise profile not found: %s", item.EnterpriseEmail)
		}

		job := model.Job{
			EnterpriseID: enterprise.ID,
			Title:        item.Title,
			Department:   item.Department,
			Location:     item.Location,
			SalaryRange:  item.SalaryRange,
			Description:  item.Description,
			Requirements: item.Requirements,
			Status:       "active",
		}

		if err := tx.Create(&job).Error; err != nil {
			return fmt.Errorf("failed to create job %s: %w", item.Title, err)
		}
	}

	return nil
}

func seedUniversityData(tx *gorm.DB, universities map[string]model.University, usersByEmail map[string]model.User) error {
	type studentSeed struct {
		Email          string
		UniversityMail string
		Major          string
		Grade          string
	}

	studentSeeds := []studentSeed{
		{Email: "student01@test.com", UniversityMail: "university01@test.com", Major: "计算机科学与技术", Grade: "2026届"},
		{Email: "student02@test.com", UniversityMail: "university01@test.com", Major: "软件工程", Grade: "2026届"},
		{Email: "student03@test.com", UniversityMail: "university01@test.com", Major: "人工智能", Grade: "2025届"},
		{Email: "student04@test.com", UniversityMail: "university02@test.com", Major: "数据科学与大数据技术", Grade: "2026届"},
		{Email: "student05@test.com", UniversityMail: "university02@test.com", Major: "网络工程", Grade: "2025届"},
	}

	for idx, item := range studentSeeds {
		student, ok := usersByEmail[item.Email]
		if !ok {
			return fmt.Errorf("student user not found: %s", item.Email)
		}
		university, ok := universities[item.UniversityMail]
		if !ok {
			return fmt.Errorf("university profile not found: %s", item.UniversityMail)
		}

		record := model.StudentRecord{
			UniversityID:     university.ID,
			UserID:           student.ID,
			Name:             student.Username,
			StudentNo:        fmt.Sprintf("S2026%03d", idx+1),
			Major:            item.Major,
			Grade:            item.Grade,
			RiskLevel:        "low",
			InterviewCount:   2 + idx,
			AverageScore:     72 + idx*4,
			EmploymentStatus: "interviewing",
		}

		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("failed to create student record for %s: %w", item.Email, err)
		}
	}

	courses := []model.Course{
		{UniversityID: universities["university01@test.com"].ID, Title: "后端高频面试题精讲", Category: "interview", StudentCount: 128, Duration: "8周", Instructor: "陈老师", Description: "覆盖 Java 后端核心八股与系统设计", CoverColor: "from-cyan-500 to-blue-600"},
		{UniversityID: universities["university01@test.com"].ID, Title: "简历优化与项目表达", Category: "resume", StudentCount: 96, Duration: "4周", Instructor: "王老师", Description: "聚焦 STAR 表达与项目量化呈现", CoverColor: "from-emerald-500 to-teal-600"},
		{UniversityID: universities["university02@test.com"].ID, Title: "前端工程化冲刺", Category: "interview", StudentCount: 84, Duration: "6周", Instructor: "赵老师", Description: "从构建工具到性能优化的系统训练", CoverColor: "from-orange-500 to-amber-600"},
		{UniversityID: universities["university02@test.com"].ID, Title: "AI应用开发实战", Category: "career", StudentCount: 66, Duration: "6周", Instructor: "刘老师", Description: "RAG、Agent、评测与上线闭环实践", CoverColor: "from-indigo-500 to-sky-600"},
	}

	for _, course := range courses {
		if err := tx.Create(&course).Error; err != nil {
			return fmt.Errorf("failed to create course %s: %w", course.Title, err)
		}
	}

	return nil
}

func seedCommunityPosts(tx *gorm.DB, usersByEmail map[string]model.User) error {
	type postSeed struct {
		Email       string
		Company     string
		Position    string
		Title       string
		Content     string
		Process     string
		Questions   string
		Review      string
		Difficulty  int
		OfferStatus string
		Rounds      int
		Tags        string
		Likes       int
		Comments    int
		Views       int
		DaysAgo     int
	}

	seeds := []postSeed{
		{Email: "student01@test.com", Company: "星瀚科技", Position: "Java后端工程师", Title: "秋招后端三面复盘：从项目深挖到场景设计", Content: "一面偏基础，二面重点问缓存一致性和分库分表，三面追问故障应急。", Process: "笔试 -> 一面技术 -> 二面架构 -> HR", Questions: "Redis 一致性、限流降级、慢 SQL 治理", Review: "项目指标要量化，回答先给结论再讲取舍", Difficulty: 4, OfferStatus: "Received", Rounds: 4, Tags: "后端,秋招,系统设计", Likes: 26, Comments: 8, Views: 320, DaysAgo: 18},
		{Email: "student02@test.com", Company: "远川数据", Position: "数据工程师", Title: "数据岗面试里被问到的实时数仓问题", Content: "重点是 Flink checkpoint、状态后端和数据延迟监控。", Process: "网申 -> 技术一面 -> 主管面", Questions: "Exactly-once、维表关联、数据回溯", Review: "建议提前准备一条完整的数据链路图", Difficulty: 4, OfferStatus: "Pending", Rounds: 3, Tags: "数据工程,Flink,实时数仓", Likes: 18, Comments: 5, Views: 210, DaysAgo: 15},
		{Email: "student03@test.com", Company: "海岳智能", Position: "AI应用工程师", Title: "RAG 场景项目表达模板分享", Content: "面试官对评测体系很感兴趣，追问了召回率和幻觉控制。", Process: "内推 -> 技术面 -> 总监面", Questions: "Chunk 策略、重排、Prompt 防御", Review: "不要只讲模型，数据质量和评测闭环更关键", Difficulty: 5, OfferStatus: "Received", Rounds: 3, Tags: "AI,RAG,项目表达", Likes: 32, Comments: 11, Views: 410, DaysAgo: 12},
		{Email: "student04@test.com", Company: "星瀚科技", Position: "前端工程师", Title: "前端性能优化面：LCP 和资源加载策略", Content: "现场给了一个首屏慢案例，要求快速定位瓶颈并给优化方案。", Process: "测评 -> 一面 -> 二面", Questions: "代码分割、预加载、长任务拆分", Review: "准备真实性能数据比背概念更有说服力", Difficulty: 3, OfferStatus: "Pending", Rounds: 3, Tags: "前端,性能优化,Vue", Likes: 21, Comments: 6, Views: 260, DaysAgo: 10},
		{Email: "student05@test.com", Company: "远川数据", Position: "测试开发工程师", Title: "测试开发岗：自动化体系搭建经验", Content: "主要考察如何把测试能力平台化，以及失败用例治理。", Process: "一面 -> 二面 -> HR", Questions: "UI 自动化稳定性、Mock 策略、回归优先级", Review: "要讲清楚投入产出和效率提升指标", Difficulty: 3, OfferStatus: "Received", Rounds: 3, Tags: "测试开发,自动化,质量平台", Likes: 14, Comments: 4, Views: 180, DaysAgo: 9},
		{Email: "enterprise01@test.com", Company: "星瀚科技", Position: "招聘官", Title: "企业视角：我们如何评估候选人的系统思维", Content: "我们更看重候选人如何拆解问题和做权衡，而不是背模板。", Process: "技术评估 + 场景追问", Questions: "容量估算、关键指标、降级方案", Review: "回答要贴近业务场景，少空泛术语", Difficulty: 2, OfferStatus: "Pending", Rounds: 1, Tags: "企业视角,面试官建议", Likes: 30, Comments: 12, Views: 500, DaysAgo: 8},
		{Email: "enterprise02@test.com", Company: "远川数据", Position: "数据平台负责人", Title: "为什么我们偏好有数据治理意识的候选人", Content: "能从血缘、口径、质量监控讲清楚就是加分项。", Process: "部门技术面", Questions: "口径管理、血缘追踪、异常告警", Review: "建议准备一个真实治理案例", Difficulty: 2, OfferStatus: "Pending", Rounds: 1, Tags: "数据治理,企业需求", Likes: 17, Comments: 3, Views: 205, DaysAgo: 7},
		{Email: "enterprise03@test.com", Company: "海岳智能", Position: "技术经理", Title: "从面试反馈看应届生常见短板", Content: "常见问题是方案不落地，忽略可观测性和故障处理。", Process: "复盘分享", Questions: "监控指标、压测、应急预案", Review: "建议在项目介绍中补充稳定性实践", Difficulty: 2, OfferStatus: "Pending", Rounds: 1, Tags: "技术经理,面试复盘", Likes: 19, Comments: 5, Views: 230, DaysAgo: 6},
		{Email: "university01@test.com", Company: "华东数字科技大学", Position: "就业指导老师", Title: "如何组织学生进行模拟面试与复盘", Content: "按岗位分组训练，采用结构化评分表做连续追踪。", Process: "训练营", Questions: "表达结构、技术深度、反问质量", Review: "每周至少一次 1v1 复盘", Difficulty: 1, OfferStatus: "Pending", Rounds: 1, Tags: "高校,就业指导,模拟面试", Likes: 22, Comments: 9, Views: 300, DaysAgo: 5},
		{Email: "university02@test.com", Company: "南方应用理工大学", Position: "辅导员", Title: "校企联动提升学生求职转化率的做法", Content: "通过企业真实 JD 反推训练计划，效果明显。", Process: "校企协同", Questions: "岗位画像、能力模型、训练节奏", Review: "建议建立可量化的阶段目标", Difficulty: 1, OfferStatus: "Pending", Rounds: 1, Tags: "高校,校企合作,就业率", Likes: 15, Comments: 2, Views: 140, DaysAgo: 4},
		{Email: "student01@test.com", Company: "模拟面试社区", Position: "Java后端工程师", Title: "团队冲突题怎么答更自然？我这次的复盘", Content: "我用事实-影响-行动-反思四步回答，效果比纯讲结论好。", Process: "行为面", Questions: "冲突背景、协作方式、结果复盘", Review: "别回避矛盾，重点体现推进与复盘能力", Difficulty: 3, OfferStatus: "Received", Rounds: 2, Tags: "行为面,团队协作,沟通", Likes: 28, Comments: 10, Views: 360, DaysAgo: 3},
		{Email: "student04@test.com", Company: "模拟面试社区", Position: "前端工程师", Title: "Vue 响应式原理一问到底：一次高压面经历", Content: "从 proxy 追问到 scheduler，再到 effect 执行时机。", Process: "技术二面", Questions: "依赖收集、触发更新、批量异步刷新", Review: "建议画图说明依赖关系和更新链路", Difficulty: 4, OfferStatus: "Pending", Rounds: 2, Tags: "Vue,响应式,前端面试", Likes: 24, Comments: 7, Views: 285, DaysAgo: 2},
	}

	now := time.Now()
	for _, item := range seeds {
		user, ok := usersByEmail[item.Email]
		if !ok {
			return fmt.Errorf("post author not found: %s", item.Email)
		}

		interviewDate := now.AddDate(0, 0, -item.DaysAgo)
		post := model.CommunityPost{
			UserID:        user.ID,
			Author:        user.Username,
			Avatar:        user.Avatar,
			Company:       item.Company,
			Position:      item.Position,
			Title:         item.Title,
			Content:       item.Content,
			Process:       item.Process,
			Questions:     item.Questions,
			Review:        item.Review,
			Difficulty:    item.Difficulty,
			OfferStatus:   item.OfferStatus,
			Rounds:        item.Rounds,
			InterviewDate: &interviewDate,
			Tags:          item.Tags,
			Likes:         item.Likes,
			Comments:      item.Comments,
			Views:         item.Views,
			IsIndexed:     false,
		}

		if err := tx.Create(&post).Error; err != nil {
			return fmt.Errorf("failed to create community post %s: %w", item.Title, err)
		}
	}

	return nil
}
