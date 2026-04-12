package repository

import (
	"strings"

	"your-project/internal/model"

	"gorm.io/gorm"
)

type CommunityPostFilter struct {
	Search   string
	Tag      string
	Company  string
	Position string
}

type CommunityRepository interface {
	CreatePost(post *model.CommunityPost) error
	GetPostByID(id uint) (*model.CommunityPost, error)
	ListPosts(filter CommunityPostFilter, page, pageSize int) ([]model.CommunityPost, int64, error)
	UpdatePost(post *model.CommunityPost) error
	DeletePost(id uint) error

	CreateComment(comment *model.PostComment) error
	GetCommentByID(id uint) (*model.PostComment, error)
	ListCommentsByPostID(postID uint, page, pageSize int) ([]model.PostComment, error)
	UpdateComment(comment *model.PostComment) error
	DeleteComment(id uint) error

	CreateMentorBooking(booking *model.MentorBooking) error
	GetMentorBookingByID(id uint) (*model.MentorBooking, error)
	ListMentorBookingsByUserID(userID uint) ([]model.MentorBooking, error)
	UpdateMentorBooking(booking *model.MentorBooking) error
	DeleteMentorBooking(id uint) error

	CreatePostLike(like *model.PostLike) error
	GetPostLike(postID, userID uint) (*model.PostLike, error)
	DeletePostLike(id uint) error
}

type GormCommunityRepository struct {
	db *gorm.DB
}

var _ CommunityRepository = (*GormCommunityRepository)(nil)

func NewCommunityRepository() CommunityRepository {
	return &GormCommunityRepository{db: GetDB()}
}

func NewCommunityRepositoryWithDB(db *gorm.DB) CommunityRepository {
	if db == nil {
		db = GetDB()
	}
	return &GormCommunityRepository{db: db}
}

func (r *GormCommunityRepository) CreatePost(post *model.CommunityPost) error {
	return r.db.Create(post).Error
}

func (r *GormCommunityRepository) GetPostByID(id uint) (*model.CommunityPost, error) {
	var post model.CommunityPost
	if err := r.db.First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *GormCommunityRepository) ListPosts(filter CommunityPostFilter, page, pageSize int) ([]model.CommunityPost, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	query := r.db.Model(&model.CommunityPost{})
	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + search + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", like, like)
	}
	if tag := strings.TrimSpace(filter.Tag); tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}
	if company := strings.TrimSpace(filter.Company); company != "" {
		query = query.Where("company = ?", company)
	}
	if position := strings.TrimSpace(filter.Position); position != "" {
		query = query.Where("position = ?", position)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	posts := make([]model.CommunityPost, 0, pageSize)
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *GormCommunityRepository) UpdatePost(post *model.CommunityPost) error {
	return r.db.Save(post).Error
}

func (r *GormCommunityRepository) DeletePost(id uint) error {
	return r.db.Delete(&model.CommunityPost{}, id).Error
}

func (r *GormCommunityRepository) CreateComment(comment *model.PostComment) error {
	return r.db.Create(comment).Error
}

func (r *GormCommunityRepository) GetCommentByID(id uint) (*model.PostComment, error) {
	var comment model.PostComment
	if err := r.db.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *GormCommunityRepository) ListCommentsByPostID(postID uint, page, pageSize int) ([]model.PostComment, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	comments := make([]model.PostComment, 0, pageSize)
	err := r.db.Where("post_id = ?", postID).Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&comments).Error
	return comments, err
}

func (r *GormCommunityRepository) UpdateComment(comment *model.PostComment) error {
	return r.db.Save(comment).Error
}

func (r *GormCommunityRepository) DeleteComment(id uint) error {
	return r.db.Delete(&model.PostComment{}, id).Error
}

func (r *GormCommunityRepository) CreateMentorBooking(booking *model.MentorBooking) error {
	return r.db.Create(booking).Error
}

func (r *GormCommunityRepository) GetMentorBookingByID(id uint) (*model.MentorBooking, error) {
	var booking model.MentorBooking
	if err := r.db.First(&booking, id).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *GormCommunityRepository) ListMentorBookingsByUserID(userID uint) ([]model.MentorBooking, error) {
	bookings := make([]model.MentorBooking, 0)
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&bookings).Error
	return bookings, err
}

func (r *GormCommunityRepository) UpdateMentorBooking(booking *model.MentorBooking) error {
	return r.db.Save(booking).Error
}

func (r *GormCommunityRepository) DeleteMentorBooking(id uint) error {
	return r.db.Delete(&model.MentorBooking{}, id).Error
}

func (r *GormCommunityRepository) CreatePostLike(like *model.PostLike) error {
	return r.db.Create(like).Error
}

func (r *GormCommunityRepository) GetPostLike(postID, userID uint) (*model.PostLike, error) {
	var like model.PostLike
	if err := r.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&like).Error; err != nil {
		return nil, err
	}
	return &like, nil
}

func (r *GormCommunityRepository) DeletePostLike(id uint) error {
	return r.db.Delete(&model.PostLike{}, id).Error
}
