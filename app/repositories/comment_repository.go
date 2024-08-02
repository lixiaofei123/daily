package repositories

import (
	"github.com/lixiaofei123/daily/app/models"
	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(comment *models.Comment) error

	Get(cid uint) (*models.Comment, error)

	GetByPostIdAndIsApproved(pid uint, isApproved bool, count uint) ([]*models.Comment, error)

	GetByPostId(pid uint, count uint) ([]*models.Comment, error)

	ListByUserId(uid, page, pageSize uint) ([]*models.Comment, error)

	List(page, pageSize uint) ([]*models.Comment, error)

	DeleteByIdAndUserId(cid uint, uid uint) error

	DeleteById(cid uint) error

	Update(comment *models.Comment) error
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{
		db: db,
	}
}

type commentRepository struct {
	db *gorm.DB
}

func (c *commentRepository) Create(comment *models.Comment) error {
	return c.db.Create(comment).Error
}

func (c *commentRepository) Get(cid uint) (*models.Comment, error) {
	var comment *models.Comment = new(models.Comment)
	if err := c.db.First(comment, cid).Error; err != nil {
		return nil, err
	}
	return comment, nil
}

func (c *commentRepository) GetByPostIdAndIsApproved(pid uint, isApproved bool, count uint) ([]*models.Comment, error) {
	var comments []*models.Comment
	err := c.db.Model(&models.Comment{}).Where("postId = ?", pid).Where("isApproved = ?", isApproved).Limit(int(count)).Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (c *commentRepository) GetByPostId(pid uint, count uint) ([]*models.Comment, error) {
	var comments []*models.Comment
	err := c.db.Model(&models.Comment{}).Where("postId = ?", pid).Limit(int(count)).Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (c *commentRepository) ListByUserId(uid, page, pageSize uint) ([]*models.Comment, error) {
	var comments []*models.Comment
	err := c.db.Model(&models.Comment{}).Where("userId = ?", uid).
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (c *commentRepository) List(page, pageSize uint) ([]*models.Comment, error) {
	var comments []*models.Comment
	err := c.db.Model(&models.Comment{}).
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (c *commentRepository) DeleteByIdAndUserId(cid uint, uid uint) error {
	return c.db.Where("id = ?", cid).Where("userId = ?", uid).Delete(&models.Comment{}).Error
}

func (c *commentRepository) DeleteById(cid uint) error {
	return c.db.Where("id = ?", cid).Delete(&models.Comment{}).Error
}

func (c *commentRepository) Update(comment *models.Comment) error {
	return c.db.Save(comment).Error
}
