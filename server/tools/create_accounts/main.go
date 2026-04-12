package main

import (
	"crypto/rand"
	"encoding/hex"
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
		{Username: "林书远", Email: "student01@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student01"},
		{Username: "顾清禾", Email: "student02@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student02"},
		{Username: "唐知行", Email: "student03@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student03"},
		{Username: "魏景初", Email: "student04@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student04"},
		{Username: "许明川", Email: "student05@test.com", Role: "student", Password: "123456", Avatar: "https://api.dicebear.com/7.x/adventurer/svg?seed=student05"},
		{Username: "星澜智联科技有限公司", Email: "enterprise01@test.com", Role: "enterprise", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=enterprise01"},
		{Username: "远川数据智能有限公司", Email: "enterprise02@test.com", Role: "enterprise", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=enterprise02"},
		{Username: "海岳工业软件有限公司", Email: "enterprise03@test.com", Role: "enterprise", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=enterprise03"},
		{Username: "江南数字工程大学就业中心", Email: "university01@test.com", Role: "university", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=university01"},
		{Username: "华南应用科技大学就业中心", Email: "university02@test.com", Role: "university", Password: "123456", Avatar: "https://api.dicebear.com/7.x/shapes/svg?seed=university02"},
	}

	usersByEmail := make(map[string]model.User, len(seeds))
	for _, item := range seeds {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password for %s: %w", item.Email, err)
		}

		user := model.User{
			UUID:     generateSeedUUID(),
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
		{Email: "enterprise01@test.com", CompanyName: "星澜智联科技有限公司", ContactName: "沈知衡", ContactPhone: "13800010001", BusinessScope: "企业级协同平台、交易中台和智能客服系统研发"},
		{Email: "enterprise02@test.com", CompanyName: "远川数据智能有限公司", ContactName: "邓一凡", ContactPhone: "13800010002", BusinessScope: "实时数据平台、指标治理和 AI 决策支持产品"},
		{Email: "enterprise03@test.com", CompanyName: "海岳工业软件有限公司", ContactName: "宋承远", ContactPhone: "13800010003", BusinessScope: "工业互联网平台、设备运维系统和制造执行系统开发"},
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
		{Email: "university01@test.com", UniversityName: "江南数字工程大学", ContactName: "陈雅宁", ContactPhone: "13900020001", Department: "计算机与人工智能学院"},
		{Email: "university02@test.com", UniversityName: "华南应用科技大学", ContactName: "赵文博", ContactPhone: "13900020002", Department: "软件与网络空间安全学院"},
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
		{EnterpriseEmail: "enterprise01@test.com", Title: "Java后端开发工程师", Department: "交易平台研发部", Location: "上海", SalaryRange: "20k-32k", Description: "负责订单、支付与履约链路核心服务开发，优化高并发场景下的稳定性与可观测性。", Requirements: "熟悉 Spring Boot、MySQL、Redis、消息队列；具备分布式事务与性能调优经验。"},
		{EnterpriseEmail: "enterprise01@test.com", Title: "前端开发工程师", Department: "产品体验技术部", Location: "上海", SalaryRange: "18k-30k", Description: "负责 B 端运营系统和 C 端门户前端架构建设，持续提升首屏性能与交互体验。", Requirements: "熟悉 Vue3/React、TypeScript、工程化体系；有性能优化和监控接入经验。"},
		{EnterpriseEmail: "enterprise02@test.com", Title: "数据工程师", Department: "实时数据平台部", Location: "杭州", SalaryRange: "22k-36k", Description: "建设离线与实时一体化数据链路，负责数据模型治理、任务稳定性和成本优化。", Requirements: "熟悉 Flink、Kafka、ClickHouse；理解数仓分层、血缘治理和数据质量体系。"},
		{EnterpriseEmail: "enterprise02@test.com", Title: "AI应用工程师", Department: "智能应用研发部", Location: "杭州", SalaryRange: "24k-40k", Description: "负责企业知识问答与智能助手产品落地，推进 RAG、Agent 与评测体系工程化。", Requirements: "熟悉 LLM API、向量检索、Prompt 工程；有上线监控和效果迭代经验。"},
		{EnterpriseEmail: "enterprise03@test.com", Title: "测试开发工程师", Department: "质量工程平台部", Location: "深圳", SalaryRange: "18k-28k", Description: "建设自动化测试平台与质量门禁，推动研发流程左移并提升交付可靠性。", Requirements: "熟悉 Python/Go，CI/CD 与测试框架；具备接口自动化和稳定性测试经验。"},
		{EnterpriseEmail: "enterprise03@test.com", Title: "DevOps工程师", Department: "云原生运维平台部", Location: "深圳", SalaryRange: "22k-34k", Description: "负责云原生发布平台、可观测体系和故障应急机制建设。", Requirements: "熟悉 Kubernetes、GitOps、Prometheus/Grafana；具备生产环境排障经验。"},
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

func generateSeedUUID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("seed-%d", time.Now().UnixNano())
	}

	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80

	hexStr := hex.EncodeToString(buf)
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexStr[0:8],
		hexStr[8:12],
		hexStr[12:16],
		hexStr[16:20],
		hexStr[20:32],
	)
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
		{UniversityID: universities["university01@test.com"].ID, Title: "后端高频面试题精讲", Category: "interview", StudentCount: 128, Duration: "8周", Instructor: "陈伯安", Description: "覆盖 Java 后端核心八股与系统设计", CoverColor: "from-cyan-500 to-blue-600"},
		{UniversityID: universities["university01@test.com"].ID, Title: "简历优化与项目表达", Category: "resume", StudentCount: 96, Duration: "4周", Instructor: "王知远", Description: "聚焦 STAR 表达与项目量化呈现", CoverColor: "from-emerald-500 to-teal-600"},
		{UniversityID: universities["university02@test.com"].ID, Title: "前端工程化冲刺", Category: "interview", StudentCount: 84, Duration: "6周", Instructor: "赵宁川", Description: "从构建工具到性能优化的系统训练", CoverColor: "from-orange-500 to-amber-600"},
		{UniversityID: universities["university02@test.com"].ID, Title: "AI应用开发实战", Category: "career", StudentCount: 66, Duration: "6周", Instructor: "刘书宁", Description: "RAG、Agent、评测与上线闭环实践", CoverColor: "from-indigo-500 to-sky-600"},
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
		{Email: "student01@test.com", Company: "星澜智联科技有限公司", Position: "Java后端工程师", Title: "秋招后端四轮复盘：从接口性能到故障演练我踩过的坑", Content: "这次面试最难的是二面和三面。二面要求我把简历里的“订单服务优化”讲成可复用方案，我给了背景流量、瓶颈 SQL、缓存命中率提升前后数据；三面现场追问线上故障处理，我按“发现-止损-定位-复盘”讲了 45 分钟，面试官重点看我有没有把风险隔离做在前面。最终拿到 offer，但能明显感受到只背八股完全不够，必须把项目里的指标和取舍讲透。", Process: "网申 -> 在线笔试（2道算法+2道系统题） -> 一面基础与项目 -> 二面系统设计 -> 三面故障处置与压测 -> HR终面", Questions: "Redis 缓存一致性怎么保证；库存超卖如何兜底；慢 SQL 如何定位与治理；服务雪崩时限流、熔断和降级怎么组合；压测数据与线上差异怎么解释", Review: "建议准备 2 个能量化结果的项目故事，每个故事都要带“业务目标、技术方案、指标变化、失败教训”。回答系统设计题先给边界和核心指标，再给方案树，最后说明取舍。HR 面会看稳定性和协作，提前准备一次跨团队推进案例会很加分。", Difficulty: 4, OfferStatus: "Received", Rounds: 5, Tags: "后端,秋招,系统设计,故障复盘", Likes: 39, Comments: 14, Views: 522, DaysAgo: 18},
		{Email: "student02@test.com", Company: "远川数据智能有限公司", Position: "数据工程师", Title: "实时数仓面经：面试官如何追问你是不是真的做过 Flink", Content: "我投的是实时数据工程师岗，技术面几乎都围绕生产链路展开。面试官先让我画从埋点到报表的链路，再问 checkpoint 对吞吐的影响、状态膨胀如何处理、指标延迟怎么报警。我把之前做过的“广告看板分钟级延迟从 12 分钟降到 2 分钟”案例讲出来，包括 watermark、维表缓存和重试策略，最后才顺利通过。", Process: "网申 -> 线上测评 -> 技术一面（Flink+数仓） -> 技术二面（架构+治理） -> 主管面", Questions: "Exactly-once 与 at-least-once 的边界；维表 Join 热点怎么处理；任务反压出现时如何定位；数据回溯和重放方案；指标口径变更如何灰度发布", Review: "不要只讲“我会用 Flink”，要讲“我怎么在具体业务里控制延迟、保障准确率、降低成本”。建议提前画好一页链路图，标清数据源、清洗、计算、存储、消费方以及监控点位，现场讲解会非常顺。", Difficulty: 4, OfferStatus: "Pending", Rounds: 4, Tags: "数据工程,Flink,实时数仓,面经", Likes: 27, Comments: 9, Views: 388, DaysAgo: 15},
		{Email: "student03@test.com", Company: "海岳工业软件有限公司", Position: "AI应用工程师", Title: "RAG 项目面怎么答：我把“能跑”讲成了“可上线”", Content: "这次面试最关键的不是模型参数，而是完整闭环。面试官一直问：召回效果怎么评估、幻觉怎么监控、失败样本如何回流。我把项目拆成检索、重排、生成、评测四段，给了线下评测和线上 A/B 的结果，还补了失败案例：某些长文档切块过粗导致召回缺失，后来按语义边界切块并加 rerank 才稳定。", Process: "内推 -> 技术一面（LLM 基础） -> 技术二面（RAG 方案） -> 总监面（上线与成本）", Questions: "Chunk 大小怎么定；多路召回如何融合；Prompt 注入如何防御；评测集如何构建；token 成本和响应时延怎么平衡", Review: "如果你做过 RAG，一定要带数据说话：命中率、准确率、拒答率、平均响应时延。面试官最怕听到“效果不错但没有指标”，建议至少准备一个失败-修复-复验的完整故事。", Difficulty: 5, OfferStatus: "Received", Rounds: 4, Tags: "AI,RAG,项目表达,评测", Likes: 45, Comments: 17, Views: 640, DaysAgo: 12},
		{Email: "student04@test.com", Company: "星澜智联科技有限公司", Position: "前端工程师", Title: "前端性能专项面：LCP 从 4.8s 优化到 2.1s 的完整答题模板", Content: "我这轮面试是现场给业务页面做性能诊断。先通过 Performance 面板定位是首屏资源阻塞和图片体积过大，再补充打包分析发现公共包过重。我的回答顺序是“先测量、再定位、后改造、最后验证”，分别给出 preload、按路由分包、图片格式升级和长任务拆分策略，并说明改完后 LCP、TTI 的变化。", Process: "在线测评 -> 技术一面（基础） -> 技术二面（性能实战） -> 业务面", Questions: "LCP 和 CLS 如何实战优化；代码分割边界怎么定；SSR 和 CSR 在该场景如何选；埋点如何验证优化收益", Review: "性能题别背概念，直接说“我如何拿数据定位问题”。建议随身准备一套常用指标口径（LCP/FID/INP/CLS）和一页优化 checklist，面试官会觉得你可立即上手。", Difficulty: 4, OfferStatus: "Pending", Rounds: 4, Tags: "前端,性能优化,Vue,工程化", Likes: 31, Comments: 12, Views: 476, DaysAgo: 10},
		{Email: "student05@test.com", Company: "远川数据智能有限公司", Position: "测试开发工程师", Title: "测试开发岗三面总结：如何证明你不是只会点点点", Content: "我的准备重点是把“测试执行”升级成“质量体系建设”。一面问自动化框架我如何设计，二面追问 flaky case 如何治理，三面让讲一次线上事故复盘。我把质量门禁、冒烟流水线、回归分层和失败用例归因拆开讲，补了效率数据：核心回归耗时从 6 小时降到 1.5 小时。", Process: "简历筛选 -> 技术一面（自动化能力） -> 技术二面（质量平台） -> 经理面（协作与推动）", Questions: "UI 自动化稳定性如何保障；Mock 与真实依赖如何取舍；质量指标如何定义；上线前灰度验证怎么做", Review: "测试开发一定要准备“技术 + 业务价值”两套答案。技术方案讲可维护性，业务价值讲提效和降故障。只讲工具名字很难打动面试官。", Difficulty: 3, OfferStatus: "Received", Rounds: 4, Tags: "测试开发,自动化,质量平台,面试复盘", Likes: 24, Comments: 8, Views: 334, DaysAgo: 9},
		{Email: "enterprise01@test.com", Company: "星澜智联科技有限公司", Position: "招聘负责人", Title: "企业视角：我们如何判断候选人是否具备系统思维", Content: "面试中我们最看重三件事：问题拆解能力、指标意识、以及方案取舍。候选人如果只会给标准答案而不能结合业务约束，通常很难通过二面。我们会故意加入变化条件，例如流量翻倍、预算减半、上线时间提前，观察候选人如何调整方案。", Process: "岗位初筛 -> 结构化技术面 -> 场景追问面 -> 交叉复盘", Questions: "容量估算是否合理；关键指标是否可观测；极端场景下有没有降级与兜底；跨团队协作怎么推进", Review: "建议候选人在回答前先确认目标、约束和成功标准。回答中尽量给出“为什么这么做而不是另一种做法”，这比背术语更能体现成熟度。", Difficulty: 2, OfferStatus: "Pending", Rounds: 1, Tags: "企业视角,招聘标准,系统思维", Likes: 37, Comments: 15, Views: 610, DaysAgo: 8},
		{Email: "enterprise02@test.com", Company: "远川数据智能有限公司", Position: "数据平台负责人", Title: "为什么我们偏好有数据治理意识的应届生", Content: "很多同学能写任务，但忽略了数据可复用性和可追溯性。我们招人更关注是否理解口径一致、血缘可查、质量可控。即便是初级岗位，只要能把“数据从哪里来、哪里去、错了怎么发现”讲清楚，就会明显加分。", Process: "业务筛选 -> 技术面 -> 负责人面 -> 团队互评", Questions: "指标口径冲突如何处理；数据血缘怎样建设；异常数据如何告警并止损；历史数据修复策略", Review: "建议准备一个真实治理案例，至少包含“问题背景、治理动作、线上收益、后续机制”。只讲架构图不讲治理闭环，会显得项目深度不足。", Difficulty: 2, OfferStatus: "Pending", Rounds: 1, Tags: "数据治理,企业需求,数据平台", Likes: 23, Comments: 6, Views: 298, DaysAgo: 7},
		{Email: "enterprise03@test.com", Company: "海岳工业软件有限公司", Position: "技术经理", Title: "从 30 场校招面试看应届生的共性短板与改进建议", Content: "我们今年校招面试发现，很多同学技术点覆盖广，但问题是讲不深、讲不落地。比如提到微服务却说不清观测指标，提到高并发却没有压测结论。真正打动面试官的是“完整闭环”：背景、方案、风险、结果、复盘。", Process: "周度面试复盘会", Questions: "监控体系如何设计；压测结论是否可信；故障应急流程是否完整；上线后如何持续迭代", Review: "建议同学准备 2 个“从问题到结果”的项目模板，每个模板都带可量化指标。面试不是知识竞赛，而是验证你能否在真实场景解决问题。", Difficulty: 2, OfferStatus: "Pending", Rounds: 1, Tags: "技术经理,校招复盘,面试建议", Likes: 29, Comments: 9, Views: 372, DaysAgo: 6},
		{Email: "university01@test.com", Company: "江南数字工程大学", Position: "就业指导中心", Title: "高校实操：我们如何把模拟面试做成可持续训练体系", Content: "过去我们做模拟面试偏活动型，学生参与热情高但效果难追踪。今年改成“岗位分层 + 周期训练 + 数据复盘”后，效果明显提升。每位学生都有能力画像，老师每周根据薄弱维度安排针对性训练，企业导师参与末轮评审，形成闭环。", Process: "岗位分组建档 -> 每周模拟面试 -> 数据评分复盘 -> 企业导师联评", Questions: "表达结构是否清晰；技术深度是否达标；案例是否可量化；反问是否体现岗位理解", Review: "高校侧建议至少保存四类数据：训练次数、平均得分、薄弱维度、岗位投递结果。这样才能把“经验判断”升级为“数据驱动辅导”。", Difficulty: 1, OfferStatus: "Pending", Rounds: 1, Tags: "高校,就业指导,模拟面试,训练体系", Likes: 34, Comments: 13, Views: 468, DaysAgo: 5},
		{Email: "university02@test.com", Company: "华南应用科技大学", Position: "就业服务办公室", Title: "校企联动项目复盘：求职转化率提升 22% 的三步法", Content: "我们和三家企业共建了岗位能力标准，把企业 JD 拆成可训练的课程模块，再把模拟面试成绩与真实投递结果关联。经过一学期迭代，学生的简历通过率和面试通过率都有明显提升。关键在于高校、企业、学生三方都用同一套能力语言沟通。", Process: "企业需求采集 -> 能力模型拆解 -> 训练营执行 -> 招聘季数据回收", Questions: "岗位画像如何共建；能力标准如何量化；训练反馈如何落到个人；企业评价如何回流课程", Review: "校企合作不是一次宣讲，而是持续共创。建议高校把企业反馈纳入课程评价体系，形成“教学-训练-就业”一体化机制。", Difficulty: 1, OfferStatus: "Pending", Rounds: 1, Tags: "高校,校企合作,就业数据,转化提升", Likes: 28, Comments: 7, Views: 342, DaysAgo: 4},
		{Email: "student01@test.com", Company: "模拟面试社区", Position: "Java后端工程师", Title: "行为面高频题：团队冲突怎么答才不像背模板", Content: "我以前回答团队冲突题总是“我们沟通后解决了”，面试官基本不会追问。后来改用“事实-影响-行动-复盘”结构，先交代冲突背景，再讲我做了什么、结果如何、下次怎么预防。这次在行为面被追问了 4 轮，最后拿到了“沟通能力强”的评价。", Process: "行为面专项训练 -> 1v1 模拟 -> 复盘修订 -> 正式面试", Questions: "冲突起因是什么；你承担了什么责任；怎样推动达成一致；结果可量化吗", Review: "行为面最忌讳空话。建议准备 2 个真实冲突案例，重点突出“你如何推进问题解决”，并补充客观结果和反思，不要只强调“团队最终很和谐”。", Difficulty: 3, OfferStatus: "Received", Rounds: 2, Tags: "行为面,团队协作,沟通表达", Likes: 33, Comments: 12, Views: 490, DaysAgo: 3},
		{Email: "student04@test.com", Company: "模拟面试社区", Position: "前端工程师", Title: "Vue 响应式原理被追问 40 分钟后，我总结出这套作答顺序", Content: "这轮面试从 reactive 和 ref 的区别开始，直接追到 effect 调度与批量更新。我一开始回答得很散，后来改成“依赖收集 -> 触发更新 -> 调度执行 -> 边界场景”四段，面试官明显更容易理解。最后还被问到 watch 和 computed 的底层差异，我用缓存与副作用时机做了对比。", Process: "技术二面（源码向）", Questions: "依赖收集如何建立；scheduler 在哪里介入；微任务队列如何影响更新时机；watch 与 computed 的核心差异", Review: "源码题建议一定画流程图。先讲主链路，再补边界条件（嵌套 effect、递归更新、异步批处理），这样既有深度也不容易跑题。", Difficulty: 4, OfferStatus: "Pending", Rounds: 2, Tags: "Vue,响应式原理,前端面试,源码", Likes: 36, Comments: 11, Views: 538, DaysAgo: 2},
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
