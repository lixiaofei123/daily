package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/lixiaofei123/daily/app/cache"
	"github.com/lixiaofei123/daily/app/controller"
	"github.com/lixiaofei123/daily/app/funcmap"
	"github.com/lixiaofei123/daily/app/middleware"
	"github.com/lixiaofei123/daily/app/models"
	"github.com/lixiaofei123/daily/app/mvc"
	"github.com/lixiaofei123/daily/app/repositories"
	"github.com/lixiaofei123/daily/app/services"
	"github.com/lixiaofei123/daily/configs"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	echo_middleware "github.com/labstack/echo/v4/middleware"

	echo "github.com/labstack/echo/v4"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func startServer(mode string) {

	var db *gorm.DB
	var err error

	if mode == "dev" {
		log.Println("当前是开发模式")
		db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{
			SkipDefaultTransaction: true,
			Logger: logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags),
				logger.Config{
					LogLevel:                  logger.Error,
					IgnoreRecordNotFoundError: true,
					ParameterizedQueries:      true,
					Colorful:                  false,
				},
			),
		})
	} else {
		log.Println("当前是生产模式")
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			configs.GlobalConfig.Database.User,
			configs.GlobalConfig.Database.Password,
			configs.GlobalConfig.Database.Host,
			configs.GlobalConfig.Database.Port,
			configs.GlobalConfig.Database.Name,
		)

		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			SkipDefaultTransaction: true,
			Logger: logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
				logger.Config{
					LogLevel:                  logger.Error,
					IgnoreRecordNotFoundError: true,
					ParameterizedQueries:      true,
					Colorful:                  false,
				},
			),
		})
	}

	if err != nil {
		log.Fatal("failed to connect database")
	}

	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.UserProfile{})
	db.AutoMigrate(&models.Post{})
	db.AutoMigrate(&models.Comment{})
	db.AutoMigrate(&models.Like{})

	userRepository := repositories.NewUserRepository(db)
	userProfileRepository := repositories.NewUserProfileRepository(db)
	postRepository := repositories.NewPostRepository(db)
	commentRepository := repositories.NewCommentRepository(db)
	likeRepository := repositories.NewLikeRepository(db)

	userService := services.NewUserService(userRepository, userProfileRepository, postRepository, cache.NewValueCache())
	postService := services.NewPostService(postRepository, userService)
	commentService := services.NewCommentService(commentRepository, postService, userService)
	likeService := services.NewLikeService(likeRepository, userService, postService)

	e := echo.New()
	e.IPExtractor = func(r *http.Request) string {
		IPAddress := r.Header.Get("X-Real-Ip")
		if IPAddress == "" {
			IPAddress = r.Header.Get("X-Forwarded-For")
		}
		if IPAddress == "" {
			IPAddress = r.RemoteAddr
		}
		return IPAddress
	}

	t := &Template{
		templates: template.Must(template.New("").Funcs(funcmap.CustonFuncMap).ParseGlob("public/views/*.html")),
	}
	e.Renderer = t

	// e.Use(echo_middleware.Logger())
	e.Use(echo_middleware.Recover())
	e.Use(echo_middleware.CORSWithConfig(echo_middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowCredentials: true,
	}))

	index := e.Group("")
	index.Use(middleware.IndexAuthHandler)
	mvc.New(index).Prefix("").Handle(controller.NewIndexController(postService, userService))

	apiv1 := e.Group("/api/v1")

	initapi := apiv1.Group("/init")
	mvc.New(initapi).Prefix("/api/v1/init").Handle(controller.NewInitController(userService))

	login := apiv1.Group("")
	mvc.New(login).Prefix("/api/v1").Handle(controller.NewLoginController(userService))

	userapi := apiv1.Group("/user")
	userapi.Use(middleware.UserAuthHandler)
	mvc.New(userapi).Prefix("/api/v1/user").Handle(controller.NewUserController(userService))

	adminapi := apiv1.Group("/admin/user")
	adminapi.Use(middleware.AdminAuthHandler)
	mvc.New(adminapi).Prefix("/api/v1/admin/user").Handle(controller.NewAdminUserController(userService))

	commentapi := apiv1.Group("/comment")
	commentapi.Use(middleware.CommentAuthHandler)
	mvc.New(commentapi).Prefix("/api/v1/comment").Handle(controller.NewCommentController(commentService, likeService, userService, postService))

	likeapi := apiv1.Group("/like")
	likeapi.Use(middleware.IndexAuthHandler)
	mvc.New(likeapi).Prefix("/api/v1/like").Handle(controller.NewLikeController(likeService))

	fileapi := apiv1.Group("/file")
	fileapi.Use(middleware.UserAuthHandler)

	uploaderapi := e.Group("/uploader")
	uploaderService := services.NewUploadService(uploaderapi)
	mvc.New(fileapi).Prefix("/api/v1/file").Handle(controller.NewFileController(uploaderService))

	postapi := apiv1.Group("/post")
	postapi.Use(middleware.UserAuthHandler)
	mvc.New(postapi).Prefix("/api/v1/post").Handle(controller.NewPostController(postService, uploaderService, userService))

	lbsapi := apiv1.Group("/lbs")
	mvc.New(lbsapi).Prefix("/api/v1/lbs").Handle(controller.NewLBSController(services.NewLBSService()))

	staticapi := e.Group("/static")
	staticapi.Static("", "./static")

	e.Logger.Fatal(e.Start(":8081"))
}

func main() {
	var mode string
	flag.StringVar(&mode, "mode", "product", "启动模式")
	flag.Parse()

	if configs.CheckConfigIsExist() {
		err := configs.InitConfig()
		if err != nil {
			log.Panic(err)
		}
		startServer(mode)
	} else {
		// 没有配置文件，站点未进行初始化
		log.Println("未检测配置文件，启动配置服务")
		configServer := http.Server{Addr: ":8081", Handler: nil}
		configServer.Handler = &ConfigHandler{
			ConfigServer: &configServer,
			StartServerFunc: func() {
				err := configs.InitConfig()
				if err != nil {
					log.Panic(err)
				}
				startServer(mode)
			},
		}
		configServer.ListenAndServe()
		shutdown := make(chan struct{})
		<-shutdown
	}

}
