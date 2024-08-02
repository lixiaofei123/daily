package errors

import "errors"

var (
	ErrUnknownError         = errors.New("未知异常")
	ErrInvalidURL           = errors.New("invalid URL")
	ErrInvalidEmail         = errors.New("invalid email")
	ErrWrongUserOrPassword  = errors.New("用户名或者密码错误")
	ErrUserNotFound         = errors.New("用户不存在")
	ErrUserIsDisabled       = errors.New("此用户已经被禁用")
	ErrRequiredParamMissing = errors.New("缺少必要参数")
	ErrInvalidParam         = errors.New("参数无效")
	ErrPasswordIsTooShort   = errors.New("密码太短了")
	ErrNeedLogin            = errors.New("需要登陆")
	ErrRoleIsNotEnough      = errors.New("权限不足")
	ErrInvalidPath          = errors.New("非法访问")

	ErrPostNotFound    = errors.New("动态不存在")
	ErrCommentNotFound = errors.New("评论不存在")

	ErrPostIsEmpty      = errors.New("动态是空的")
	ErrCommentIsEmpty   = errors.New("评论是空的")
	ErrCommentIsTooLong = errors.New("评论太长了")

	ErrAdminUserIsAlreayAdd = errors.New("管理员用户已经添加，请勿再次添加")
	ErrEmailIsUsed          = errors.New("邮箱已经被使用了")

	ErrUploadError   = errors.New("上传异常")
	ErrDownloadError = errors.New("下载异常")

	ErrSignIsError = errors.New("签名不正确")

	ErrTooManyRequests = errors.New("触发限流规则")

	ErrHasAlreadyLiked = errors.New("已经点过赞了")

	ErrUnknowPan123Error = errors.New("连接123云盘出现未知异常")

	ErrTooManyRequestIn123Pan = errors.New("超过了123云盘的QPS")

	ErrFileIsNotExistsIn123Pan = errors.New("文件不存在")

	ErrUnsupportExternalMedia = errors.New("不支持的外部资源媒体")

	ErrNotYetImplementedMethod = errors.New("暂未实现")

	ErrRegisterCardError = errors.New("注册卡片类型失败，请检查代码")

	ErrCardRenderHtmlError = errors.New("渲染卡片失败")

	ErrNoSuchCard = errors.New("不支持的卡片类型")

	ErrLocationError = errors.New("定位失败")
)
