package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lixiaofei123/daily/app/cache"
	apperr "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/models"
	"github.com/lixiaofei123/daily/app/repositories"
	"github.com/lixiaofei123/daily/app/utils"
	"gorm.io/gorm"
)

type UserService interface {
	AddNewUser(email, password string) (*models.UserDetail, error)

	Login(email, password string) (*models.UserDetail, error)

	GetById(id uint) (*models.UserDetail, error)

	ResetUserPassword(id uint, password string) error

	UpdateUserEnable(id uint, enable bool) (*models.UserDetail, error)

	UpdateUser(id uint, user *models.UserDetail) (*models.UserDetail, error)

	ListUsers() ([]*models.SimpleUser, error)

	CountUser() (uint, error)

	GetByEmail(email string) (*models.User, error)

	CreateOrGetByEmail(email string, username string) (*models.User, error)

	GetAdmin() (*models.UserDetail, error)

	GetUserOrAdmin(cuid uint) (*models.UserDetail, error)
}

func NewUserService(userRepository repositories.UserRepository, userProfileRepository repositories.UserProfileRepository, postRepository repositories.PostRepository, cache *cache.IDValueCache) UserService {
	return &userService{
		userRepository:        userRepository,
		userProfileRepository: userProfileRepository,
		postRepository:        postRepository,
		cache:                 cache,
	}
}

type userService struct {
	userRepository        repositories.UserRepository
	userProfileRepository repositories.UserProfileRepository
	postRepository        repositories.PostRepository
	cache                 *cache.IDValueCache
}

func (s *userService) AddNewUser(email, password string) (*models.UserDetail, error) {

	if len(password) < 8 {
		return nil, apperr.ErrPasswordIsTooShort
	}

	encryptedPassword := utils.Md5WithSalt(password, email)

	user := models.User{
		Email:    email,
		Username: fmt.Sprintf("用户%d", time.Now().Unix()),
		Password: encryptedPassword,
		Enable:   true,
		Role:     models.UserRole,
	}

	userCount, err := s.userRepository.Count()
	if err != nil {
		return nil, err
	}

	if userCount == 0 {
		user.Role = models.AdminRole
	}

	if err := s.userRepository.Create(&user); err != nil {

		var lasterror = err
		// 可能是邮箱已经被游客身份使用，将游客角色升级成用户
		existUser, err := s.userRepository.FindByEmail(email)
		if err != nil {
			return nil, err
		}

		if existUser.Role == models.VisitorRole {

			existUser.Role = models.UserRole
			existUser.Password = encryptedPassword
			err := s.userRepository.Update(existUser)
			if err != nil {
				return nil, err
			}

			user = *existUser

		} else {
			return nil, lasterror
		}
	}

	userProfile := models.UserProfile{
		UserID:    user.ID,
		Cover:     "/static/images/default_cover.jpg",
		Avatar:    "/static/images/default_avatar.jpg",
		Signature: "",
	}

	if err := s.userProfileRepository.Create(&userProfile); err != nil {
		s.userRepository.ForceDelete(user.ID)
		return nil, err
	}

	return &models.UserDetail{
		User:    &user,
		Profile: &userProfile,
	}, nil

}

func (s *userService) GetAdmin() (*models.UserDetail, error) {
	user, err := s.userRepository.FindAdmin()
	if err != nil {
		return nil, apperr.ErrUserNotFound
	}
	userProfile, err := s.userProfileRepository.FindByUserId(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.UserDetail{
		User:    user,
		Profile: userProfile,
	}, nil
}

func (s *userService) GetById(id uint) (*models.UserDetail, error) {

	cacheValue, _ := s.cache.Get(&models.UserDetail{
		ID: id,
	})
	if cacheValue != nil {
		return cacheValue.(*models.UserDetail), nil
	}

	user, err := s.userRepository.FindById(id)
	if err != nil {
		return nil, apperr.ErrUserNotFound
	}

	var userProfile *models.UserProfile

	if user.Role != models.VisitorRole {
		userProfile, err = s.userProfileRepository.FindByUserId(user.ID)
		if err != nil {
			return nil, err
		}
	}

	userDetail := &models.UserDetail{
		ID:      user.ID,
		User:    user,
		Profile: userProfile,
	}

	s.cache.Put(userDetail, time.Hour*24)

	return userDetail, nil
}

func (s *userService) Login(email, password string) (*models.UserDetail, error) {

	encryptedPassword := utils.Md5WithSalt(password, email)
	user, err := s.userRepository.FindByEmailAndPassword(email, encryptedPassword)
	if err != nil {
		return nil, apperr.ErrWrongUserOrPassword
	}

	if !user.Enable {
		return nil, apperr.ErrUserIsDisabled
	}

	userProfile, err := s.userProfileRepository.FindByUserId(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.UserDetail{
		User:    user,
		Profile: userProfile,
	}, nil

}

func (s *userService) ResetUserPassword(id uint, password string) error {

	if len(password) < 8 {
		return apperr.ErrPasswordIsTooShort
	}

	user, err := s.userRepository.FindById(id)
	if err != nil {
		return apperr.ErrUserNotFound
	}

	encryptedPassword := utils.Md5WithSalt(password, user.Email)
	return s.userRepository.UpdatePassword(id, encryptedPassword)
}

func (s *userService) UpdateUserEnable(id uint, enable bool) (*models.UserDetail, error) {

	user, err := s.userRepository.FindById(id)
	if err != nil {
		return nil, apperr.ErrUserNotFound
	}
	user.Enable = enable
	err = s.userRepository.Update(user)
	if err != nil {
		return nil, err
	}

	if user.Enable {
		s.postRepository.BatchUpdateApprovedStatusByUid(id, true)
	} else {
		s.postRepository.BatchUpdateApprovedStatusByUid(id, false)
	}

	s.cache.Delete(&models.UserDetail{
		ID: user.ID,
	})

	return s.GetById(id)
}

func (s *userService) UpdateUser(id uint, updateUser *models.UserDetail) (*models.UserDetail, error) {

	user, err := s.userRepository.FindById(id)
	if err != nil {
		return nil, apperr.ErrUserNotFound
	}

	userProfile, err := s.userProfileRepository.FindByUserId(user.ID)
	if err != nil {
		return nil, err
	}

	if updateUser.User.Username != user.Username && updateUser.User.Username != "" {
		user.Username = updateUser.User.Username
	}

	if updateUser.Profile.Avatar != userProfile.Avatar && updateUser.Profile.Avatar != "" {
		userProfile.Avatar = updateUser.Profile.Avatar
	}

	if updateUser.Profile.Cover != userProfile.Cover && updateUser.Profile.Cover != "" {
		userProfile.Cover = updateUser.Profile.Cover
	}

	if updateUser.Profile.Signature != userProfile.Signature && updateUser.Profile.Signature != "" {
		userProfile.Signature = updateUser.Profile.Signature
	}

	err = s.userRepository.Update(user)
	if err != nil {
		return nil, err
	}

	err = s.userProfileRepository.Update(userProfile)
	if err != nil {
		return nil, err
	}

	s.cache.Delete(&models.UserDetail{
		ID: user.ID,
	})

	return s.GetById(id)

}

func (s *userService) ListUsers() ([]*models.SimpleUser, error) {
	return s.userRepository.ListAll()
}

func (s *userService) CountUser() (uint, error) {
	return s.userRepository.Count()
}

func (s *userService) GetByEmail(email string) (*models.User, error) {

	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) CreateOrGetByEmail(email string, username string) (*models.User, error) {
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在这个用户，为这个用户创建账号
			user = &models.User{
				Email:    email,
				Role:     models.VisitorRole,
				Username: username,
				Password: uuid.NewString(),
				Enable:   true,
			}

			err = s.userRepository.Create(user)
			if err != nil {
				return nil, err
			}

		} else {
			return nil, err
		}
	}

	if user.Role != models.VisitorRole {
		return nil, apperr.ErrEmailIsUsed
	}

	if user.Username != username {
		err = s.userRepository.Update(user)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *userService) GetUserOrAdmin(cuid uint) (*models.UserDetail, error) {

	if cuid != 0 {
		userDetail, err := s.GetById(cuid)
		if err == nil {
			return userDetail, nil
		}
	}

	return s.GetAdmin()

}
