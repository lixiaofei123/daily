package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lixiaofei123/daily/app/uploader"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ConfigHandler struct {
	ConfigServer    *http.Server
	StartServerFunc func()
}

type SimpleUploaderConfig struct {
	Name   string            `json:"name"`
	Config map[string]string `json:"config"`
}

type SimpleDBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (ch *ConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		path := r.URL.Path
		if path == "" || path == "/" {
			path = "/static/install.html"
		}

		fileData, err := os.ReadFile(path[1:])
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", getContentType(path))
		w.Write(fileData)
	} else if r.Method == http.MethodPost {
		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {

			var config map[string]interface{} = map[string]interface{}{}
			err := json.Unmarshal(data, &config)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			// 检查存储配置
			uploaderConfig := new(SimpleUploaderConfig)
			uploaderData, _ := json.Marshal(config["uploader"])
			err = json.Unmarshal(uploaderData, uploaderConfig)
			if err == nil {
				uploaderSrv := uploader.NewUploader(uploaderConfig.Name, uploaderConfig.Config)
				err = uploaderSrv.Put("测试文件_可以随时删除.txt", []byte("这是一个测试存储是否能够正常使用过的文件，可以删除"))
			}
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("存储配置测试失败，原因是 %s", err.Error())))
				return
			}

			// 检查数据库
			dbConfig := new(SimpleDBConfig)
			dbConfigData, _ := json.Marshal(config["database"])
			err = json.Unmarshal(dbConfigData, dbConfig)
			if err == nil {
				dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
					dbConfig.User,
					dbConfig.Password,
					dbConfig.Host,
					dbConfig.Port,
					dbConfig.Name,
				)

				_, err = gorm.Open(mysql.Open(dsn))

			}
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("连接数据库失败，原因是 %s", err.Error())))
				return
			}

			// All Pass

			w.WriteHeader(http.StatusOK)
			go func(config map[string]interface{}) {
				yamlConfig, _ := yaml.Marshal(config)
				err = os.WriteFile("config.yaml", yamlConfig, 0755)
				if err != nil {
					log.Panicln("创建配置失败，原因是 --->", err.Error())
				}

				err = ch.ConfigServer.Shutdown(context.Background())
				if err != nil {
					log.Panicln("停止配置服务器失败，原因是 --->", err.Error())
				}

				ch.StartServerFunc()
			}(config)

		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func getContentType(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".json":
		return "application/json; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}
