package repositories

import (
	"github.com/lixiaofei123/daily/app/models"
	"gorm.io/gorm"
)

type PostRepository interface {
	Create(post *models.Post) error

	Get(pid uint) (*models.Post, error)

	ListByUserIdAndVisibility(uid uint, visibility models.Visibility, page, pageSize uint) ([]*models.Post, error)

	ListByUserId(uid uint, page, pageSize uint) ([]*models.Post, error)

	List(page, pageSize uint) ([]*models.Post, error)

	DeleteByIdAndUserId(pid uint, uid uint) error

	Update(post *models.Post) error

	BatchUpdateApprovedStatusByUid(uid uint, newStatus bool) error

	BatchUpdatePriorityByUid(uid uint, newPriority uint) error
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{
		db: db,
	}
}

type postRepository struct {
	db *gorm.DB
}

func (pr *postRepository) Create(post *models.Post) error {
	return pr.db.Create(post).Error
}

func (pr *postRepository) Get(pid uint) (*models.Post, error) {
	var post *models.Post = new(models.Post)
	if err := pr.db.First(post, pid).Error; err != nil {
		return nil, err
	}
	return post, nil
}

func (pr *postRepository) ListByUserIdAndVisibility(uid uint, visibility models.Visibility, page, pageSize uint) ([]*models.Post, error) {
	var posts []*models.Post
	err := pr.db.Where(&models.Post{
		UserID:     uid,
		Visibility: visibility,
		IsApproved: true,
	}).Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Order("priority desc").Order("id desc").Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (pr *postRepository) List(page, pageSize uint) ([]*models.Post, error) {
	var posts []*models.Post
	err := pr.db.Model(&models.Post{}).Where(&models.Post{
		IsApproved: true,
	}).Offset(int((page - 1) * pageSize)).Order("priority desc").Order("id desc").Limit(int(pageSize)).Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (pr *postRepository) DeleteByIdAndUserId(pid, uid uint) error {
	return pr.db.Where("id = ?", pid).Where("userId = ?", uid).Delete(&models.Post{}).Error
}

func (pr *postRepository) ListByUserId(uid uint, page, pageSize uint) ([]*models.Post, error) {
	var posts []*models.Post
	err := pr.db.Where(&models.Post{
		UserID: uid,
	}).Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Order("priority desc").Order("id desc").Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (pr *postRepository) Update(post *models.Post) error {
	return pr.db.Save(post).Error
}

func (pr *postRepository) BatchUpdateApprovedStatusByUid(uid uint, newStatus bool) error {
	return pr.db.Model(&models.Post{}).Where("userId = ?", uid).Update("isApproved", newStatus).Error
}

func (pr *postRepository) BatchUpdatePriorityByUid(uid uint, newPriority uint) error {
	return pr.db.Model(&models.Post{}).Where("userId = ?", uid).Update("priority", newPriority).Error
}
