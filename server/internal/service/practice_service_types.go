package service

import (
	"context"
	"errors"
	"time"

	"your-project/internal/model"
)

type PracticeService interface {
	GetMeta(ctx context.Context) (*PracticeMetaResponse, error)
	GetFilterOptions(ctx context.Context, positionCode string) (*PracticeFilterOptions, error)
	ListQuestions(ctx context.Context, userID uint, req PracticeListQuestionsRequest) (*PracticeQuestionListResponse, error)
	GetQuestionLists(ctx context.Context, userID uint, positionCode string) ([]PracticeQuestionListDefinition, error)
	GetSpecialties(ctx context.Context, positionCode string) ([]string, error)
	DrawQuestion(ctx context.Context, userID uint, req PracticeDrawRequest) (*PracticeQuestionEnvelope, error)
	SubmitAnswer(ctx context.Context, userID uint, req PracticeAnswerRequest) (*PracticeAnswerFeedback, error)
	GetPointSummary(ctx context.Context, userID uint, positionCode, point string) (*PracticePointSummary, error)
	GetWrongRemedial(ctx context.Context, userID, wrongID uint) (*PracticeWrongRemedialResponse, error)
	GetSolution(ctx context.Context, questionID uint) (*PracticeSolutionResponse, error)
	StartAssessment(ctx context.Context, userID uint, req PracticeAssessmentStartRequest) (*PracticeAssessmentSession, error)
	SubmitAssessmentAnswer(ctx context.Context, userID uint, req PracticeAssessmentAnswerRequest) (*PracticeAnswerFeedback, error)
	CompleteAssessment(ctx context.Context, userID uint, assessmentID uint) (*PracticeAssessmentSummary, error)
	GetIntegrationSnapshot(ctx context.Context, userID uint, positionCode string) (*PracticeIntegrationSnapshotResponse, error)
	SubmitIntegrationFeedback(ctx context.Context, userID uint, req PracticeIntegrationFeedbackRequest) (*PracticeIntegrationFeedbackResponse, error)
	ListWrongs(ctx context.Context, userID uint, req PracticeWrongsRequest) ([]PracticeWrongBookItem, error)
	DeleteWrong(ctx context.Context, userID, wrongID uint) error
	ToggleWrongFavorite(ctx context.Context, userID, wrongID uint) (bool, error)
	ToggleQuestionFavorite(ctx context.Context, userID, questionID uint) (bool, error)
	GetDashboard(ctx context.Context, userID uint, positionCode string) (*PracticeDashboardResponse, error)
	ImportQuestionBank(ctx context.Context, req PracticeImportRequest) (int, error)
	ExportRecords(ctx context.Context, userID uint) (*PracticeRecordExport, error)
	GetQuestionByID(ctx context.Context, userID, questionID uint) (*PracticeQuestionEnvelope, error)
}

type PracticeListQuestionsRequest struct {
	PositionCode string
	Level        string
	QuestionType string
	Specialty    string
	Point        string
	CompanyType  string
	Status       string
	Keyword      string
	FavoriteOnly bool
	ListID       uint
	Page         int
	PageSize     int
}

type PracticeDrawRequest struct {
	PositionCode string
	Level        string
	QuestionType string
	Specialty    string
	Point        string
	CompanyType  string
	ListID       uint
}

type PracticeWrongsRequest struct {
	PositionCode string
	Point        string
	QuestionType string
	FavoriteOnly bool
}

type PracticeIntegrationFeedbackRequest struct {
	PositionCode string
	WeakPoints   []string
}

type PracticeImportRequest struct {
	Items []PracticeImportQuestionInput `json:"items"`
}

type PracticeImportQuestionInput struct {
	PositionCode    string                 `json:"position_code,omitempty"`
	Role            string                 `json:"role,omitempty"`
	Level           string                 `json:"level"`
	QuestionType    string                 `json:"question_type"`
	Specialty       string                 `json:"specialty"`
	Stem            string                 `json:"stem"`
	Answer          string                 `json:"answer"`
	Analysis        string                 `json:"analysis"`
	Tips            string                 `json:"tips"`
	Exemplar        string                 `json:"exemplar"`
	Points          string                 `json:"points"`
	CompanyType     string                 `json:"company_type"`
	DifficultyScore int                    `json:"difficulty_score"`
	Options         []model.QuestionOption `json:"options"`
}

type PracticeRoleMeta struct {
	Name        string   `json:"name"`
	Focus       []string `json:"focus"`
	Specialties []string `json:"specialties"`
}

type PracticeMetaResponse struct {
	Roles         map[string]PracticeRoleMeta `json:"roles"`
	Levels        map[string]string           `json:"levels"`
	QuestionTypes []string                    `json:"question_types"`
	Counts        map[string]int64            `json:"counts"`
}

type PracticeFilterOptions struct {
	Points        []string          `json:"points"`
	Specialties   []string          `json:"specialties"`
	CompanyTypes  []string          `json:"company_types"`
	Levels        map[string]string `json:"levels"`
	QuestionTypes []string          `json:"question_types"`
	StatusOptions []string          `json:"status_options"`
}

type PracticePagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type PracticeStatusStats struct {
	Todo   int `json:"todo"`
	Solved int `json:"solved"`
	Wrong  int `json:"wrong"`
}

type PracticeQuestionSummary struct {
	ID              uint   `json:"id"`
	Role            string `json:"role"`
	PositionCode    string `json:"position_code"`
	Level           string `json:"level"`
	QuestionType    string `json:"question_type"`
	Specialty       string `json:"specialty"`
	Points          string `json:"points"`
	CompanyType     string `json:"company_type"`
	DifficultyScore int    `json:"difficulty_score"`
	Stem            string `json:"stem"`
	HasOptions      bool   `json:"has_options"`
	Status          string `json:"status,omitempty"`
	IsFavorite      bool   `json:"is_favorite"`
}

type PracticeQuestionDetail struct {
	ID              uint                   `json:"id"`
	Role            string                 `json:"role"`
	PositionCode    string                 `json:"position_code"`
	Position        string                 `json:"position,omitempty"`
	Difficulty      string                 `json:"difficulty,omitempty"`
	Level           string                 `json:"level"`
	QuestionType    string                 `json:"question_type"`
	Specialty       string                 `json:"specialty"`
	Stem            string                 `json:"stem"`
	Options         []model.QuestionOption `json:"options,omitempty"`
	Points          string                 `json:"points"`
	CompanyType     string                 `json:"company_type"`
	DifficultyScore int                    `json:"difficulty_score"`
	HasOptions      bool                   `json:"has_options"`
	Status          string                 `json:"status,omitempty"`
	IsFavorite      bool                   `json:"is_favorite"`
}

type PracticeQuestionEnvelope struct {
	Question PracticeQuestionDetail `json:"question"`
}

type PracticeQuestionListResponse struct {
	Items       []PracticeQuestionSummary `json:"items"`
	Pagination  PracticePagination        `json:"pagination"`
	StatusStats PracticeStatusStats       `json:"status_stats"`
}

type PracticeQuestionListDefinition struct {
	ID           uint     `json:"id"`
	Role         string   `json:"role"`
	PositionCode string   `json:"position_code"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	TotalCount   int      `json:"total_count"`
	SolvedCount  int      `json:"solved_count"`
	Progress     float64  `json:"progress"`
}

type PracticeWrongBookItem struct {
	WrongID         uint      `json:"wrong_id"`
	ID              uint      `json:"id"`
	Role            string    `json:"role"`
	PositionCode    string    `json:"position_code"`
	Level           string    `json:"level"`
	QuestionType    string    `json:"question_type"`
	Specialty       string    `json:"specialty"`
	Stem            string    `json:"stem"`
	Points          string    `json:"points"`
	CompanyType     string    `json:"company_type"`
	DifficultyScore int       `json:"difficulty_score"`
	HasOptions      bool      `json:"has_options"`
	LastUserAnswer  string    `json:"last_user_answer,omitempty"`
	ErrorReason     string    `json:"error_reason,omitempty"`
	IsFavorite      bool      `json:"is_favorite"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PracticeRemedialQuestion struct {
	ID         uint   `json:"id"`
	Stem       string `json:"stem"`
	Level      string `json:"level"`
	Difficulty int    `json:"difficulty"`
}

type PracticeWrongRemedialResponse struct {
	Role              string                     `json:"role"`
	PositionCode      string                     `json:"position_code"`
	Point             string                     `json:"point"`
	Packet            PracticePointPacket        `json:"packet"`
	BaseQuestions     []PracticeRemedialQuestion `json:"base_questions"`
	AdvancedQuestions []PracticeRemedialQuestion `json:"advanced_questions"`
}

type PracticeRoleSnapshotItem struct {
	Role          string  `json:"role"`
	PositionCode  string  `json:"position_code"`
	TotalAttempts int64   `json:"total_attempts"`
	Accuracy      float64 `json:"accuracy"`
}

type PracticePointMasteryItem struct {
	Role         string  `json:"role"`
	PositionCode string  `json:"position_code"`
	Point        string  `json:"point"`
	Mastery      float64 `json:"mastery"`
}

type PracticeIntegrationSnapshotResponse struct {
	GeneratedAt  string                     `json:"generated_at"`
	RoleStats    []PracticeRoleSnapshotItem `json:"role_stats"`
	PointMastery []PracticePointMasteryItem `json:"point_mastery"`
}

type PracticeRemedialSet struct {
	Point     string                     `json:"point"`
	Questions []PracticeRemedialQuestion `json:"questions"`
}

type PracticeIntegrationFeedbackResponse struct {
	Role         string                `json:"role"`
	PositionCode string                `json:"position_code"`
	RemedialSets []PracticeRemedialSet `json:"remedial_sets"`
}

type PracticeDashboardRoleProgress struct {
	AttemptCount int64   `json:"attempt_count"`
	Accuracy     float64 `json:"accuracy"`
	Progress     float64 `json:"progress"`
}

type PracticeDashboardTrendItem struct {
	Day      string  `json:"day"`
	Count    int64   `json:"count"`
	Accuracy float64 `json:"accuracy"`
}

type PracticeDashboardRadarItem struct {
	Dimension string  `json:"dimension"`
	Mastery   float64 `json:"mastery"`
}

type PracticeDashboardResponse struct {
	TotalAttempts int64                                    `json:"total_attempts"`
	Accuracy      float64                                  `json:"accuracy"`
	RoleProgress  map[string]PracticeDashboardRoleProgress `json:"role_progress"`
	Trend         []PracticeDashboardTrendItem             `json:"trend"`
	Radar         []PracticeDashboardRadarItem             `json:"radar"`
}

type PracticeRecordExport struct {
	Filename string
	Content  string
}

type PracticeErrorCode string

const (
	PracticeErrorInvalidArgument PracticeErrorCode = "invalid_argument"
	PracticeErrorUnauthorized    PracticeErrorCode = "unauthorized"
	PracticeErrorNotFound        PracticeErrorCode = "not_found"
	PracticeErrorInternal        PracticeErrorCode = "internal"
)

type PracticeError struct {
	Code    PracticeErrorCode
	Message string
	Cause   error
}

func (e *PracticeError) Error() string {
	return e.Message
}

func (e *PracticeError) Unwrap() error {
	return e.Cause
}

func AsPracticeError(err error) (*PracticeError, bool) {
	var target *PracticeError
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

func newPracticeError(code PracticeErrorCode, message string, cause error) error {
	return &PracticeError{Code: code, Message: message, Cause: cause}
}

func invalidPracticeArgument(message string) error {
	return newPracticeError(PracticeErrorInvalidArgument, message, nil)
}

func unauthorizedPracticeError(message string) error {
	return newPracticeError(PracticeErrorUnauthorized, message, nil)
}

func practiceNotFound(message string, cause error) error {
	return newPracticeError(PracticeErrorNotFound, message, cause)
}

func practiceInternalError(message string, cause error) error {
	return newPracticeError(PracticeErrorInternal, message, cause)
}

type PracticeAnswerRequest struct {
	QuestionID     uint
	UserAnswer     string
	ErrorReason    string
	ElapsedSeconds *int
	TimedMode      bool
	IsTimeout      bool
	AssessmentID   *uint
	SourceKind     string
}

type PracticePointPacket struct {
	PositionCode        string   `json:"position_code"`
	Point               string   `json:"point"`
	Memo                string   `json:"memo"`
	InterviewExtensions []string `json:"interview_extensions"`
}

type PracticePointProgress struct {
	Total      int     `json:"total"`
	Solved     int     `json:"solved"`
	Completion float64 `json:"completion"`
}

type PracticeAnswerFeedback struct {
	IsCorrect       bool                  `json:"is_correct"`
	AnswerMode      string                `json:"answer_mode"`
	StandardAnswer  string                `json:"standard_answer"`
	Analysis        string                `json:"analysis"`
	Tips            string                `json:"tips,omitempty"`
	Exemplar        string                `json:"exemplar,omitempty"`
	MatchedKeywords []string              `json:"matched_keywords,omitempty"`
	MissingKeywords []string              `json:"missing_keywords,omitempty"`
	PointPacket     PracticePointPacket   `json:"point_packet"`
	PointState      PracticePointProgress `json:"point_state"`
}

type PracticeSolutionResponse struct {
	StandardAnswer string `json:"standard_answer"`
	Analysis       string `json:"analysis"`
	Tips           string `json:"tips,omitempty"`
	Exemplar       string `json:"exemplar,omitempty"`
}

type PracticePointSummary struct {
	PracticePointPacket
	Progress         PracticePointProgress `json:"progress"`
	IsPointCompleted bool                  `json:"is_point_completed"`
}

type PracticeAssessmentStartRequest struct {
	PositionCode string
	Difficulty   string
	TotalCount   int
}

type PracticeAssessmentAnswerRequest struct {
	AssessmentID   uint
	QuestionID     uint
	UserAnswer     string
	ElapsedSeconds *int
	IsTimeout      bool
}

type PracticeAssessmentSession struct {
	AssessmentID uint                     `json:"assessment_id"`
	Questions    []PracticeQuestionDetail `json:"questions"`
}

type PracticePointMastery struct {
	Point   string  `json:"point"`
	Mastery float64 `json:"mastery"`
}

type PracticeAssessmentSummary struct {
	AssessmentID      uint                   `json:"assessment_id"`
	PositionCode      string                 `json:"position_code"`
	Score             float64                `json:"score"`
	CorrectCount      int                    `json:"correct_count"`
	TotalCount        int                    `json:"total_count"`
	TargetCompanyType string                 `json:"target_company_type"`
	PointReport       []PracticePointMastery `json:"point_report"`
	NeedImprovePoints []string               `json:"need_improve_points"`
}
