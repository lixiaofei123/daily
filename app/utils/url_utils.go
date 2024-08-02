package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	apperror "github.com/lixiaofei123/daily/app/errors"
)

func IsValidHttpURL(urlstr string) bool {
	uurl, err := url.Parse(urlstr)
	return urlstr != "" && err == nil && (uurl.Scheme == "http" || uurl.Scheme == "https")

}

func HttpGet(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	return body, nil
}

func GenerateLocalPath(inputUrl string) string {
	parsedUrl, _ := url.Parse(inputUrl)

	pathPart := parsedUrl.Path
	extension := path.Ext(pathPart)
	if extension != "" {
		extension = strings.ToLower(extension)
	}

	now := time.Now()
	datePath := now.Format("storage/2006/01/02")

	randomString := generateRandomString()

	var localPath string
	if extension != "" {
		localPath = fmt.Sprintf("%s/%s%s", datePath, randomString, extension)
	} else {
		localPath = fmt.Sprintf("%s/%s", datePath, randomString)
	}

	return localPath
}

func generateRandomString() string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	randNum, _ := rand.Int(rand.Reader, big.NewInt(1000))
	randomNumber := randNum.Int64() + 1
	randomString := fmt.Sprintf("%d%d", randomNumber, timestamp)
	return base36Encode(randomString)
}

func base36Encode(input string) string {
	num := new(big.Int)
	num.SetString(input, 10)
	return num.Text(36)
}

func isPathWithinRoot(root, testpath string) bool {

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	absolutePath, err := filepath.Abs(testpath)
	if err != nil {
		return false
	}

	absoluteRoot = strings.ToLower(absoluteRoot)
	absolutePath = strings.ToLower(absolutePath)

	return strings.HasPrefix(absolutePath, absoluteRoot)
}

func SafeJoin(root, rpath string) (string, error) {

	cleanpath := filepath.Clean(rpath)
	abspath := path.Join(root, cleanpath)
	if !isPathWithinRoot(root, abspath) {
		return "", apperror.ErrInvalidPath
	}
	return abspath, nil
}
