package services

import (
	"strings"

	"github.com/lixiaofei123/daily/app/card"
	apperr "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/models"
	"github.com/lixiaofei123/daily/app/repositories"
)

type PostService interface {
	Create(post *models.Post) error

	Get(cuid, pid uint) (*models.PostDetail, error)

	List(cuid, page, pageSize uint) (*models.PageResult, error)

	ListByUid(cuid, uid, page, pageSize uint) (*models.PageResult, error)

	DeleteByIdAndUserId(pid uint, uid uint) error

	UpdatePriority(pid uint, uid uint, priority uint) (*models.PostDetail, error)
}

func NewPostService(postRepository repositories.PostRepository, userService UserService) PostService {
	return &postService{
		postRepository: postRepository,
		userService:    userService,
	}
}

type postService struct {
	postRepository repositories.PostRepository
	userService    UserService
}

func (p *postService) Create(post *models.Post) error {
	post.Content = strings.Trim(post.Content, " ")
	if len(post.Medias) == 0 {
		if post.Url != nil {
			post.PostType = models.UrlPostType
		} else if post.ExternalMedia != nil {
			post.PostType = models.ExternalPostType
			err := models.FastCheckExternalMedia(post.ExternalMedia)
			if err != nil {
				return err
			}
		} else if post.Card != nil {
			post.PostType = models.CardPostType
			user, _ := p.userService.GetById(post.UserID)
			_, err := card.Html(post.Card.Name, post.Card.Model, &models.PostDetail{
				Post: &models.Post{
					Content: post.Content,
				},
				UserDetail: user,
			})
			if err != nil {
				return err
			}
		} else {
			post.PostType = models.TextPostType
			if len(post.Content) == 0 {
				return apperr.ErrPostIsEmpty
			}
		}
	} else {
		firstMedia := post.Medias[0]

		if firstMedia.Type == models.FileMediaType {
			post.PostType = models.FilePostType
		}
		if firstMedia.Type == models.ImageMediaType {
			post.PostType = models.ImagePostType
		}
		if firstMedia.Type == models.MusicMediaType {
			post.PostType = models.MusicPostType
		}

		if firstMedia.Type == models.VideoMediaType {
			post.PostType = models.VideoPostType
		}

		if post.PostType != models.ImagePostType {
			post.Medias = post.Medias[0:1]
		}

	}

	post.IsApproved = true

	if post.Visibility == "" {
		post.Visibility = models.PublicVisibility
	}

	var err error
	if post.ID == 0 {
		err = p.postRepository.Create(post)
	} else {
		post.IsUpdated = true
		err = p.postRepository.Update(post)
	}

	return err
}

func (p *postService) Get(cuid, pid uint) (*models.PostDetail, error) {
	post, err := p.postRepository.Get(pid)
	if err != nil || (post.Visibility != models.PublicVisibility && post.UserID != cuid) {
		return nil, apperr.ErrPostNotFound
	}

	user, err := p.userService.GetById(post.UserID)
	if err != nil {
		return nil, err
	}

	return &models.PostDetail{
		Post:       post,
		UserDetail: user,
	}, nil
}

func (p *postService) List(cuid, page, pageSize uint) (*models.PageResult, error) {
	posts, err := p.postRepository.List(page, pageSize)
	if err != nil {
		return nil, err
	}

	newposts := []*models.Post{}
	for _, post := range posts {
		if post.Visibility == models.PublicVisibility || (post.Visibility == models.PrivateVisibility && post.UserID == cuid) {
			newposts = append(newposts, post)
		}
	}

	postDetails := []*models.PostDetail{}
	for _, post := range newposts {
		userid := post.UserID
		user, err := p.userService.GetById(userid)
		if err == nil {
			postDetails = append(postDetails, &models.PostDetail{
				UserDetail: user,
				Post:       post,
			})
		}
	}

	return &models.PageResult{
		Page:     page,
		PageSize: pageSize,
		Data:     postDetails,
	}, nil
}

func (p *postService) DeleteByIdAndUserId(pid uint, uid uint) error {
	return p.postRepository.DeleteByIdAndUserId(pid, uid)
}

func (p *postService) ListByUid(cuid, uid, page, pageSize uint) (*models.PageResult, error) {
	var posts []*models.Post
	var err error
	if cuid == uid {
		posts, err = p.postRepository.ListByUserId(uid, page, pageSize)
	} else {
		posts, err = p.postRepository.ListByUserIdAndVisibility(uid, models.PublicVisibility, page, pageSize)
	}
	if err != nil {
		return nil, err
	}

	postDetails := []*models.PostDetail{}
	for _, post := range posts {
		userid := post.UserID
		user, err := p.userService.GetById(userid)
		if err == nil {
			postDetails = append(postDetails, &models.PostDetail{
				UserDetail: user,
				Post:       post,
			})
		}
	}

	return &models.PageResult{
		Page:     page,
		PageSize: pageSize,
		Data:     postDetails,
	}, nil
}

func (p *postService) UpdatePriority(pid uint, uid uint, priority uint) (*models.PostDetail, error) {
	post, err := p.postRepository.Get(pid)
	if err != nil {
		return nil, err
	}

	if post.UserID != uid {
		return nil, apperr.ErrRoleIsNotEnough
	}

	if priority > 0 {
		err = p.postRepository.BatchUpdatePriorityByUid(post.UserID, 0)
		if err != nil {
			return nil, err
		}
	}

	post.Priority = priority
	err = p.postRepository.Update(post)
	if err != nil {
		return nil, err
	}

	return p.Get(uid, pid)
}
