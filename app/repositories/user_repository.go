package repositories

import (
	"github.com/lixiaofei123/daily/app/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error

	FindById(id uint) (*models.User, error)

	FindByEmail(email string) (*models.User, error)

	FindByEmailAndPassword(email string, password string) (*models.User, error)

	Update(user *models.User) error

	UpdatePassword(userid uint, password string) error

	ForceDelete(userid uint) error

	ListAll() ([]*models.SimpleUser, error)

	Count() (uint, error)

	FindAdmin() (*models.User, error)
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

type userRepository struct {
	db *gorm.DB
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindById(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) UpdatePassword(userid uint, password string) error {

	return r.db.Model(&models.User{}).Where("id = ?", userid).Update("password", password).Error
}

func (r *userRepository) ForceDelete(userid uint) error {
	return r.db.Unscoped().Delete(&models.User{}, userid).Error
}

func (r *userRepository) FindByEmailAndPassword(email string, password string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).Where("password = ?", password).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ListAll() ([]*models.SimpleUser, error) {
	var users []*models.SimpleUser
	err := r.db.Model(&models.User{}).Select("users.id, users.uname as username, users.email, users.role, users.enable, user_profiles.avatar").
		Where("role = ?", "admin").Or("role = ? ", "user").
		Joins("left join user_profiles on user_profiles.userId = users.id").
		Find(&users).Error

	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Count() (uint, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return uint(count), nil
}

func (r *userRepository) FindAdmin() (*models.User, error) {
	var user models.User
	if err := r.db.Where("role = ?", "admin").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
