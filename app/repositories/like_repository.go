package repositories

import (
	"errors"

	"github.com/lixiaofei123/daily/app/models"
	"gorm.io/gorm"
)

type LikeRepository interface {
	Create(comment *models.Like) error

	GetByPostId(pid uint) ([]*models.Like, error)

	GetByUserIdAndPostId(uid uint, pid uint) (*models.Like, error)

	GetByIPAndPostId(ip string, pid uint) (*models.Like, error)

	DeleteById(lid uint) error

	Get(lid uint) (*models.Like, error)

	VisitorLikeCount(pid uint) (uint, error)
}

func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepository{
		db: db,
	}
}

type likeRepository struct {
	db *gorm.DB
}

func (l *likeRepository) Create(like *models.Like) error {
	return l.db.Create(like).Error
}

func (c *likeRepository) GetByPostId(pid uint) ([]*models.Like, error) {
	var likes []*models.Like
	err := c.db.Model(&models.Like{}).Where("postId = ?", pid).Where("userId != 0").Order("id ASC").Find(&likes).Error
	if err != nil {
		return nil, err
	}
	return likes, nil
}

func (c *likeRepository) GetByUserIdAndPostId(uid uint, pid uint) (*models.Like, error) {
	var likes *models.Like = new(models.Like)
	err := c.db.Model(&models.Like{}).Where("postId = ?", pid).
		Where("userId = ?", uid).First(&likes).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}
	return likes, nil
}

func (c *likeRepository) GetByIPAndPostId(ip string, pid uint) (*models.Like, error) {
	var likes *models.Like = new(models.Like)
	err := c.db.Model(&models.Like{}).Where("postId = ?", pid).
		Where("ip = ?", ip).Where("userId = 0").First(&likes).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}
	return likes, nil
}

func (c *likeRepository) DeleteById(lid uint) error {
	return c.db.Where("id = ?", lid).Unscoped().Delete(&models.Like{}).Error
}

func (c *likeRepository) Get(lid uint) (*models.Like, error) {
	like := new(models.Like)
	if err := c.db.First(like, lid).Error; err != nil {
		return nil, err
	}
	return like, nil
}

func (c *likeRepository) VisitorLikeCount(pid uint) (uint, error) {
	var count int64 = 0
	err := c.db.Model(&models.Like{}).Where("postId = ?", pid).
		Where("userId = ?", 0).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return uint(count), err
}
