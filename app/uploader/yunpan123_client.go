package uploader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	apperror "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/utils"
)

type yunpan123Client struct {
	client       http.Client
	clientID     string
	clientSecret string
	accessToken  *AccessToken
	fileIdCache  map[string]int
}

func newYunpan123Client(clientID string, clientSecret string) *yunpan123Client {
	return &yunpan123Client{
		client:       *http.DefaultClient,
		clientID:     clientID,
		clientSecret: clientSecret,
		fileIdCache:  map[string]int{},
	}
}

type AccessToken struct {
	AccessToken string    `json:"accessToken"`
	ExpiredAt   time.Time `json:"expiredAt"`
}

type OpenApiResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func parseResult(result interface{}, data []byte) (int, error) {
	resp := new(OpenApiResp)
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return 500, err
	}

	if resp.Code != 0 {
		return resp.Code, errors.New(resp.Message)
	}

	databyte, err := json.Marshal(resp.Data)
	if err != nil {
		return 500, err
	}

	err = json.Unmarshal(databyte, result)
	if err != nil {
		return 500, err
	}
	return 0, nil
}

func (c *yunpan123Client) requestAccessToken() error {
	var jsonStr = []byte(fmt.Sprintf(`{"clientID":"%s", "clientSecret": "%s"}`, c.clientID, c.clientSecret))
	req, err := http.NewRequest("POST", "https://open-api.123pan.com/api/v1/access_token", bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	req.Header.Set("Platform", "open_platform")
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	accesstoken := new(AccessToken)
	_, err = parseResult(accesstoken, data)
	if err != nil {
		return err
	}

	c.accessToken = accesstoken
	return nil
}

func (c *yunpan123Client) checkAndRefreshToken() error {
	if c.accessToken == nil {
		return c.requestAccessToken()
	}

	if time.Until(c.accessToken.ExpiredAt) < time.Hour*24 {
		return c.requestAccessToken()
	}

	return nil
}

type FileItem struct {
	FileID       int    `json:"fileID"`
	FileName     string `json:"filename"`
	Type         int    `json:"type"`
	ParentFileId int    `json:"parentFileId"`
}

type FileListData struct {
	Total    int         `json:"total"`
	FileList []*FileItem `json:"fileList"`
}

func (c *yunpan123Client) FindFileByName(fileid int, name string, tryTime int) (*FileItem, error) {
	err := c.checkAndRefreshToken()
	if err != nil {
		return nil, err
	}

	serchUrl := fmt.Sprintf(`https://open-api.123pan.com/api/v1/file/list?parentFileId=%d&page=1&limit=10&orderBy=file_name&orderDirection=desc&searchData=%s`, fileid, name)
	req, err := http.NewRequest("GET", serchUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Platform", "open_platform")
	req.Header.Set("Authorization", c.accessToken.AccessToken)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fileListData := new(FileListData)
	code, err := parseResult(fileListData, data)
	if err != nil {
		if code == 429 {
			log.Println("触发了限流规则")
			if tryTime == 0 {
				return nil, err
			} else {
				time.Sleep(time.Millisecond * 500)
				return c.FindFileByName(fileid, name, tryTime-1)
			}
		}
		return nil, err
	}

	if fileListData.Total == 0 {
		return nil, apperror.ErrFileIsNotExistsIn123Pan
	}

	for i := 0; i < len(fileListData.FileList); i++ {
		if fileListData.FileList[i].FileName == name && fileListData.FileList[i].ParentFileId == fileid {
			return fileListData.FileList[i], nil
		}
	}

	return nil, apperror.ErrFileIsNotExistsIn123Pan
}

type DirInfo struct {
	DirID int `json:"dirID"`
}

func (c *yunpan123Client) CreateDir(fileid int, name string, tryTime int) (*DirInfo, error) {
	err := c.checkAndRefreshToken()
	if err != nil {
		return nil, err
	}

	var jsonStr = []byte(fmt.Sprintf(`{"name":"%s", "parentID": %d}`, name, fileid))
	req, err := http.NewRequest("POST", `https://open-api.123pan.com/upload/v1/file/mkdir`, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Platform", "open_platform")
	req.Header.Set("Authorization", c.accessToken.AccessToken)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	dirInfo := new(DirInfo)

	code, err := parseResult(dirInfo, data)
	if err != nil {
		if code == 429 {
			log.Println("触发了限流规则")
			if tryTime == 0 {
				return nil, err
			} else {
				time.Sleep(time.Millisecond * 500)
				return c.CreateDir(fileid, name, tryTime-1)
			}
		}
		return nil, err
	}

	return dirInfo, nil
}

func (c *yunpan123Client) createDirs(path string) (int, error) {
	path = strings.Trim(path, "/")
	arr := strings.Split(path, "/")

	key := ""
	lastfileid := 0
	for i := 0; i < len(arr); i++ {
		key = key + "/" + arr[i]
		cacheFileId, ok := c.fileIdCache[key]
		if !ok {
			// 没有缓存，需要查找
			file, err := c.FindFileByName(lastfileid, arr[i], 10)
			if err != nil {
				if errors.Is(err, apperror.ErrFileIsNotExistsIn123Pan) {
					// 文件不存在，需要创建这个文件夹
					dirinfo, err := c.CreateDir(lastfileid, arr[i], 10)
					if err != nil {
						return 0, err
					}
					c.fileIdCache[key] = dirinfo.DirID
					lastfileid = dirinfo.DirID
				} else {
					return 0, err
				}
			} else {
				// 已经有文件夹了，
				c.fileIdCache[key] = file.FileID
				lastfileid = file.FileID
			}
		} else {
			lastfileid = cacheFileId
		}
	}

	return lastfileid, nil
}

type CreateFileResult struct {
	FileID      int    `json:"fileID"`
	PreuploadID string `json:"preuploadID"`
	Reuse       bool   `json:"reuse"`
	SliceSize   int    `json:"sliceSize"`
}

func (c *yunpan123Client) createFile(parentFileID int, filename, md5 string, size int, tryTime int) (*CreateFileResult, error) {
	err := c.checkAndRefreshToken()
	if err != nil {
		return nil, err
	}

	var jsonStr = []byte(fmt.Sprintf(`{"parentFileID":%d, "filename": "%s", "etag":"%s", "size": %d}`, parentFileID, filename, md5, size))
	req, err := http.NewRequest("POST", `https://open-api.123pan.com/upload/v1/file/create`, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Platform", "open_platform")
	req.Header.Set("Authorization", c.accessToken.AccessToken)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {

		return nil, err
	}

	createResp := new(CreateFileResult)
	code, err := parseResult(createResp, data)
	if err != nil {
		if code == 429 {
			if tryTime == 0 {
				return nil, err
			} else {
				time.Sleep(time.Millisecond * 500)
				return c.createFile(parentFileID, filename, md5, size, tryTime-1)
			}
		}
		return nil, err
	}

	return createResp, nil

}

type FileParts struct {
	Parts []*FilePart `json:"parts"`
}

type FilePart struct {
	PartNumber string `json:"partNumber"`
	Size       int    `json:"size"`
	Etag       string `json:"etag"`
}

func (c *yunpan123Client) listUploadParts(preuploadID string, tryTime int) (*FileParts, error) {
	err := c.checkAndRefreshToken()
	if err != nil {
		return nil, err
	}

	var jsonStr = []byte(fmt.Sprintf(`{"preuploadID":"%s"}`, preuploadID))
	req, err := http.NewRequest("POST", `https://open-api.123pan.com/upload/v1/file/list_upload_parts`, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Platform", "open_platform")
	req.Header.Set("Authorization", c.accessToken.AccessToken)

	resp, err := c.client.Do(req)

	if err != nil {

		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fileParts := new(FileParts)
	code, err := parseResult(fileParts, data)
	if err != nil {
		if code == 429 {
			log.Println("触发了限流规则")
			if tryTime == 0 {
				return nil, err
			} else {
				time.Sleep(time.Millisecond * 500)
				return c.listUploadParts(preuploadID, tryTime-1)
			}
		}
		return nil, err
	}

	return fileParts, nil

}

type PreUploadUrlResp struct {
	PresignedURL string `json:"presignedURL"`
}

func (c *yunpan123Client) getPreUploadUrl(preuploadID string, sliceNo int, tryTime int) (*PreUploadUrlResp, error) {
	err := c.checkAndRefreshToken()
	if err != nil {
		return nil, err
	}

	var jsonStr = []byte(fmt.Sprintf(`{"preuploadID":"%s", "sliceNo": %d}`, preuploadID, sliceNo))
	req, err := http.NewRequest("POST", `https://open-api.123pan.com/upload/v1/file/get_upload_url`, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Platform", "open_platform")
	req.Header.Set("Authorization", c.accessToken.AccessToken)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	preUploadUrlResp := new(PreUploadUrlResp)
	code, err := parseResult(preUploadUrlResp, data)
	if err != nil {
		if code == 429 {
			log.Println("触发了限流规则")
			if tryTime == 0 {
				return nil, apperror.ErrTooManyRequestIn123Pan
			} else {
				time.Sleep(time.Millisecond * 500)
				return c.getPreUploadUrl(preuploadID, sliceNo, tryTime-1)
			}
		}
		return nil, err
	}

	return preUploadUrlResp, nil

}

type UploadCompleteResp struct {
	FileID    int  `json:"fileID"`
	Async     bool `json:"async"`
	Completed bool `json:"completed"`
}

func (c *yunpan123Client) uploadComplete(preuploadID string, tryTime int) (*UploadCompleteResp, error) {
	err := c.checkAndRefreshToken()
	if err != nil {
		return nil, err
	}

	var jsonStr = []byte(fmt.Sprintf(`{"preuploadID":"%s"}`, preuploadID))
	req, err := http.NewRequest("POST", `https://open-api.123pan.com/upload/v1/file/upload_complete`, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Platform", "open_platform")
	req.Header.Set("Authorization", c.accessToken.AccessToken)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	uploadCompleteResp := new(UploadCompleteResp)
	code, err := parseResult(uploadCompleteResp, data)
	if err != nil {
		if code == 429 {
			log.Println("触发了限流规则")
			if tryTime == 0 {
				return nil, apperror.ErrTooManyRequestIn123Pan
			} else {
				time.Sleep(time.Millisecond * 500)
				return c.uploadComplete(preuploadID, tryTime-1)
			}
		}
		return nil, err
	}

	return uploadCompleteResp, nil

}

type UploadAsyncResultResp struct {
	FileID    int  `json:"fileID"`
	Completed bool `json:"completed"`
}

func (c *yunpan123Client) uploadAsyncResult(preuploadID string, tryTime int) (*UploadAsyncResultResp, error) {
	err := c.checkAndRefreshToken()
	if err != nil {
		return nil, err
	}

	var jsonStr = []byte(fmt.Sprintf(`{"preuploadID":"%s"}`, preuploadID))
	req, err := http.NewRequest("POST", `https://open-api.123pan.com/upload/v1/file/upload_async_result`, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Platform", "open_platform")
	req.Header.Set("Authorization", c.accessToken.AccessToken)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	uploadAsyncResultResp := new(UploadAsyncResultResp)
	code, err := parseResult(uploadAsyncResultResp, data)
	if err != nil {
		if code == 429 {
			log.Println("触发了限流规则")
			if tryTime == 0 {
				return nil, apperror.ErrTooManyRequestIn123Pan
			} else {
				time.Sleep(time.Millisecond * 500)
				return c.uploadAsyncResult(preuploadID, tryTime-1)
			}
		}
		return nil, err
	}

	return uploadAsyncResultResp, nil

}

func (c *yunpan123Client) writeData(savepath string, data []byte) error {

	dir := path.Dir(savepath)
	filename := path.Base(savepath)

	// 第一步先创建目录
	fileid, err := c.createDirs(dir)
	if err != nil {
		return err
	}

	// 第二步创建文件
	md5 := utils.Md5Data(data)
	size := len(data)

	createResult, err := c.createFile(fileid, filename, md5, size, 10)
	if err != nil {
		return err
	}

	if createResult.Reuse {
		// 秒传了
		return nil
	}

	chunkSize := createResult.SliceSize
	numChunks := (len(data) + chunkSize - 1) / chunkSize
	preuploadID := createResult.PreuploadID

	uploadMd5s := []string{}

	// 逐个处理分片
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize

		if end > len(data) {
			end = len(data)
		}

		chunk := data[start:end]
		getPreUploadResult, err := c.getPreUploadUrl(preuploadID, i+1, 10)
		if err != nil {
			return err
		}

		preUploadUrl := getPreUploadResult.PresignedURL

		req, err := http.NewRequest("PUT", preUploadUrl, bytes.NewBuffer(chunk))
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return apperror.ErrUploadError
		}

		sliceMd5 := utils.Md5Data(chunk)
		uploadMd5s = append(uploadMd5s, sliceMd5)

	}

	// 检查分片
	parts, err := c.listUploadParts(preuploadID, 10)
	if err != nil {
		return err
	}

	if len(parts.Parts) > 0 {
		if len(parts.Parts) == len(uploadMd5s) {
			for i := 0; i < len(parts.Parts); i++ {
				if parts.Parts[i].Etag != uploadMd5s[i] {
					return apperror.ErrUnknowPan123Error
				}
			}
		} else {
			return apperror.ErrUnknowPan123Error
		}
	}

	// 分片检查无误，发送确认信息
	uploadResult, err := c.uploadComplete(preuploadID, 10)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	if uploadResult.Async {

		// 异步获取结果
		for {
			resp, err := c.uploadAsyncResult(preuploadID, 10)
			if err != nil {
				log.Println(err.Error())
				return err
			}
			if resp.Completed {
				return nil
			} else {
				time.Sleep(time.Second)
			}
		}

	} else {
		return nil
	}

}
