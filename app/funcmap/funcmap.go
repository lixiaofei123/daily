package funcmap

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/lixiaofei123/daily/app/card"
	"github.com/lixiaofei123/daily/app/models"
	"github.com/lixiaofei123/daily/configs"
)

var CustonFuncMap template.FuncMap
var buildtime []byte

func init() {
	CustonFuncMap = template.FuncMap{
		"DateBeautify":        DateBeautify,
		"ExternalMediaToHtml": ExternalMediaToHtml,
		"RenderCard":          RenderCard,
		"ConvertContent":      ConvertContent,
		"GetSiteConfig":       GetSiteConfig,
		"GetBuildTime":        GetBuildTime,
	}
	var err error
	buildtime, err = os.ReadFile("buildtime")
	if err != nil {
		log.Panicln(err.Error())
	}
}

func GetSiteConfig() configs.SiteConfig {
	return configs.GlobalConfig.Site
}

func DateBeautify(inputDate time.Time) string {
	now := time.Now()
	duration := now.Sub(inputDate)

	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%d秒前", int(duration.Seconds()))
	case duration < time.Hour:
		return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%d小时前", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		return fmt.Sprintf("%d天前", int(duration.Hours()/24))
	default:
		return inputDate.Format("2006/01/02")
	}
}

func ExternalMediaToHtml(isMobile bool, externalMedia *models.ExternalMedia) string {
	externalPlatform, err := models.GetExternalPlatform(externalMedia)
	if err != nil {
		return `<div>不支持的视频平台</div>`
	}
	return externalPlatform.Html(isMobile)
}

func RenderCard(card0 *models.Card, post *models.PostDetail) string {
	html, err := card.Html(card0.Name, card0.Model, post)
	if err != nil {
		return "卡片渲染失败，这是一个程序故障，无法通过重启恢复。错误原因:" + err.Error()
	}
	return html
}

func escapeHTML(input string) string {
	result := strings.Replace(input, "<", "&lt;", -1)
	result = strings.Replace(result, ">", "&gt;", -1)
	return result
}

func ConvertContent(oldstring string) string {
	newstring := escapeHTML(oldstring)
	newstring = strings.Replace(newstring, "\n", "<br>", -1)
	pattern := `\[url href="(.*?)"\](.*?)\[/url\]`
	re := regexp.MustCompile(pattern)
	newstring = re.ReplaceAllString(newstring, `<a target="_blank" href="$1">$2</a>`)
	return newstring
}

func GetBuildTime() string {

	return string(buildtime)
}
