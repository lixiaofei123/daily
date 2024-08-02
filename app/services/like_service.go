package services

import (
	apperr "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/models"
	"github.com/lixiaofei123/daily/app/repositories"
)

type LikeService interface {
	CreateByUid(uid uint, pid uint) (*models.Like, error)

	CreateByIp(ip string, pid uint) (*models.Like, error)

	GetByPostId(pid uint) (*models.LikeDetails, error)

	Delete(lid uint) error

	CheckIsLiked(ip string, uid uint, pid uint) (*models.Like, error)
}

func NewLikeService(likeRepository repositories.LikeRepository,
	userService UserService, postService PostService) LikeService {
	return &likeService{
		likeRepository: likeRepository,
		userService:    userService,
		postService:    postService,
	}
}

type likeService struct {
	likeRepository repositories.LikeRepository
	userService    UserService
	postService    PostService
}

func (ls *likeService) CreateByUid(uid uint, pid uint) (*models.Like, error) {

	exitlike, err := ls.likeRepository.GetByUserIdAndPostId(uid, pid)
	if err != nil {
		return nil, err
	}

	if exitlike != nil {
		return nil, apperr.ErrHasAlreadyLiked
	}

	like := models.Like{
		UserID: uid,
		PostId: pid,
		IP:     "0",
	}

	err = ls.likeRepository.Create(&like)
	if err != nil {
		return nil, err
	}

	return &like, nil

}

func (ls *likeService) CreateByIp(ip string, pid uint) (*models.Like, error) {

	exitlike, err := ls.likeRepository.GetByIPAndPostId(ip, pid)
	if err != nil {
		return nil, err
	}

	if exitlike != nil {
		return nil, apperr.ErrHasAlreadyLiked
	}

	like := models.Like{
		UserID: 0,
		PostId: pid,
		IP:     ip,
	}

	err = ls.likeRepository.Create(&like)
	if err != nil {
		return nil, err
	}

	return &like, nil

}

func (ls *likeService) GetByPostId(pid uint) (*models.LikeDetails, error) {

	likes, err := ls.likeRepository.GetByPostId(pid)

	if err != nil {
		return nil, err
	}

	var likeDetails []*models.LikeDetail
	for _, like := range likes {

		likeDetail := &models.LikeDetail{
			Like: like,
		}

		userDetail, _ := ls.userService.GetById(like.UserID)
		likeDetail.UserDetail = userDetail

		likeDetails = append(likeDetails, likeDetail)
	}

	visitorLikeCount, _ := ls.likeRepository.VisitorLikeCount(pid)

	return &models.LikeDetails{
		PostId:           pid,
		LikeDetails:      likeDetails,
		VisitorLikeCount: visitorLikeCount,
	}, nil

}

func (ls *likeService) Delete(lid uint) error {
	return ls.likeRepository.DeleteById(lid)
}

func (ls *likeService) CheckIsLiked(ip string, uid uint, pid uint) (*models.Like, error) {
	var like *models.Like
	var err error
	if uid != 0 {
		like, err = ls.likeRepository.GetByUserIdAndPostId(uid, pid)
	} else {
		like, err = ls.likeRepository.GetByIPAndPostId(ip, pid)
	}

	return like, err
}
