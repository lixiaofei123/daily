package repositories

import (
	"github.com/lixiaofei123/daily/app/models"
	"gorm.io/gorm"
)

type UserProfileRepository interface {
	Create(userProfile *models.UserProfile) error

	FindByUserId(userId uint) (*models.UserProfile, error)

	Update(userProfile *models.UserProfile) error
}

func NewUserProfileRepository(db *gorm.DB) UserProfileRepository {
	return &userProfileRepository{
		db: db,
	}
}

type userProfileRepository struct {
	db *gorm.DB
}

func (r *userProfileRepository) Create(userProfile *models.UserProfile) error {
	return r.db.Create(userProfile).Error
}

func (r *userProfileRepository) FindByUserId(userId uint) (*models.UserProfile, error) {
	var userProfile models.UserProfile
	if err := r.db.Where("userId = ?", userId).First(&userProfile).Error; err != nil {
		return nil, err
	}
	return &userProfile, nil
}

func (r *userProfileRepository) Update(userProfile *models.UserProfile) error {
	return r.db.Save(userProfile).Error
}
