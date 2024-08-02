package configs

import (
	"errors"
	"os"
	"strconv"

	"github.com/lixiaofei123/daily/app/card"
	"github.com/spf13/viper"
)

type Config struct {
	Database  DatabaseConfig
	Site      SiteConfig
	LBSConfig LBSConfig
	Auth      Auth
	Uploader  Uploader
}

type SiteConfig struct {
	Background    string
	Title         string
	ICP           string
	BaiDuTongJi   string
	Favicon       string
	ImageCompress bool
	CustomCardCss []string
	CustomCardJs  string
}

type LBSConfig struct {
	Name   string
	Config map[string]string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

type Auth struct {
	Secret string
}

type Uploader struct {
	Name      string            `yaml:"name"`
	Config    map[string]string `yaml:"config"`
	RateLimit *RateLimit
	Logger    *Logger
}

type Logger struct {
	Path string `yaml:"path"`
}

type RateLimit struct {
	Roles string //10:1;100:3600
}

var GlobalConfig *Config

func CheckConfigIsExist() bool {
	if _, err := os.Stat("config.yaml"); errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}

func InitConfig() error {
	var err error
	GlobalConfig, err = loadConfig()
	return err
}

func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DATABASE_HOST", viper.GetString("database.host")),
			Port:     getEnvAsInt("DATABASE_PORT", viper.GetInt("database.port")),
			User:     getEnv("DATABASE_USER", viper.GetString("database.user")),
			Password: getEnv("DATABASE_PASSWORD", viper.GetString("database.password")),
			Name:     getEnv("DATABASE_NAME", viper.GetString("database.name")),
		},
		Site: SiteConfig{
			Title:         getEnv("SITE_TITLE", viper.GetString("site.title")),
			ICP:           getEnv("SITE_ICP", viper.GetString("site.icp")),
			Background:    getEnv("SITE_BACKGROUND", viper.GetString("site.background")),
			BaiDuTongJi:   viper.GetString("site.BaiDuTongJi"),
			Favicon:       viper.GetString("site.favicon"),
			ImageCompress: viper.GetBool("site.imageCompress"),
			CustomCardCss: card.GetCardCssPathes(),
			CustomCardJs:  card.CardGlobalJavascript(),
		},
		Uploader: Uploader{
			Name:   viper.GetString("uploader.name"),
			Config: viper.GetStringMapString("uploader.config"),
			RateLimit: func() *RateLimit {
				ratelimit := viper.GetString("uploader.ratelimit")
				if ratelimit == "" {
					return nil
				} else {
					return &RateLimit{
						Roles: ratelimit,
					}
				}
			}(),
			Logger: func() *Logger {
				path := viper.GetString("uploader.logger.path")
				if path == "" {
					return nil
				} else {
					return &Logger{
						Path: path,
					}
				}
			}(),
		},
		Auth: Auth{
			Secret: getEnv("AUTH_SECRET", viper.GetString("auth.secret")),
		},
		LBSConfig: LBSConfig{
			Name:   viper.GetString("lbs.name"),
			Config: viper.GetStringMapString("lbs.config"),
		},
	}

	return config, nil
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultVal
}
