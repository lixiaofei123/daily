package models

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/barkimedes/go-deepcopy"

	apperror "github.com/lixiaofei123/daily/app/errors"
)

func init() {
	RegesterExternalPlatform(&BiliBiliVideo{})
	RegesterExternalPlatform(&YoukuVideo{})
	RegesterExternalPlatform(&TencentVideo{})
	RegesterExternalPlatform(&YouTuBeVideo{})
	RegesterExternalPlatform(&Music163{})
	RegesterExternalPlatform(&QQMusic{})
}

type ExternalPlatform interface {
	GetName() string
	GetDisplayName() string
	GetMediaType() MediaType
	Init(config map[string]string)
	Html(isMobile bool) string
}

var platformsMap map[string][]ExternalPlatform = map[string][]ExternalPlatform{}
var name2PlatformMap map[string]ExternalPlatform = map[string]ExternalPlatform{}

type SupportPlatformProp struct {
	Name        string `json:"name"`
	Displayname string `json:"displayname"`
}

type SupportPlatformProps []*SupportPlatformProp

type SupportPlatform struct {
	Name        string                `json:"name"`
	Displayname string                `json:"displayname"`
	Props       *SupportPlatformProps `json:"props"`
}

type SupportPlatforms map[string][]*SupportPlatform

func GetSupportPlatforms() SupportPlatforms {
	var supportPlatforms = SupportPlatforms{}
	for mediatype, paltforms := range platformsMap {
		supportPlatforms[mediatype] = []*SupportPlatform{}
		for _, platform := range paltforms {
			name := platform.GetName()
			displayname := platform.GetDisplayName()
			supportPlatform := &SupportPlatform{
				Name:        name,
				Displayname: displayname,
				Props:       &SupportPlatformProps{},
			}
			objType := reflect.ValueOf(platform).Elem().Type()
			for i := 0; i < objType.NumField(); i++ {
				field := objType.Field(i)
				propertyTag := field.Tag.Get("property")
				arr := strings.Split(propertyTag, ";")
				propName := field.Name
				propDisplayname := propName
				if len(arr) >= 1 && arr[0] != "" {
					propName = arr[0]
				}
				if len(arr) >= 2 {
					propDisplayname = arr[1]
				}

				*supportPlatform.Props = append(*supportPlatform.Props,
					&SupportPlatformProp{
						Name:        propName,
						Displayname: propDisplayname,
					})
			}
			supportPlatforms[mediatype] = append(supportPlatforms[mediatype], supportPlatform)
		}
	}
	return supportPlatforms
}

func FastCheckExternalMedia(media *ExternalMedia) error {
	_, ok := name2PlatformMap[media.Name]
	if !ok {
		return apperror.ErrUnsupportExternalMedia
	}
	return nil
}

func GetExternalPlatform(media *ExternalMedia) (ExternalPlatform, error) {
	externalPlatform, ok := name2PlatformMap[media.Name]
	if !ok {
		return nil, apperror.ErrUnsupportExternalMedia
	}

	externalPlatformClone := deepcopy.MustAnything(externalPlatform).(ExternalPlatform)
	externalPlatformClone.Init(media.Config)
	return externalPlatformClone, nil

}

func RegesterExternalPlatform(externalPlatform ExternalPlatform) {
	mediaType := externalPlatform.GetMediaType()
	if _, ok := platformsMap[string(mediaType)]; !ok {
		platformsMap[string(mediaType)] = []ExternalPlatform{}
	}
	platformsMap[string(mediaType)] = append(platformsMap[string(mediaType)], externalPlatform)
	name2PlatformMap[externalPlatform.GetName()] = externalPlatform

}

type BiliBiliVideo struct {
	BVID string `json:"bvid" property:"bvid;视频ID"`
}

func (*BiliBiliVideo) GetName() string {
	return "bilibili"
}

func (*BiliBiliVideo) GetDisplayName() string {
	return "B站"
}

func (*BiliBiliVideo) GetMediaType() MediaType {
	return VideoMediaType
}

func (b *BiliBiliVideo) Init(config map[string]string) {
	bvid := config["bvid"]
	b.BVID = bvid
}

func (b *BiliBiliVideo) Html(isMobile bool) string {
	return fmt.Sprintf(`<iframe style="width:100%%;aspect-ratio:16/9;" src="//player.bilibili.com/player.html?isOutside=true&bvid=%s&p=1&autoplay=0" scrolling="no" border="0" frameborder="no" framespacing="0" allowfullscreen="true"></iframe>`, b.BVID)
}

type YoukuVideo struct {
	VideoID string `json:"videoid" property:"videoid;视频ID"`
}

func (*YoukuVideo) GetName() string {
	return "youku"
}

func (*YoukuVideo) GetMediaType() MediaType {
	return VideoMediaType
}

func (*YoukuVideo) GetDisplayName() string {
	return "优酷视频"
}

func (y *YoukuVideo) Init(config map[string]string) {
	videoid := config["videoid"]
	y.VideoID = videoid
}

func (y *YoukuVideo) Html(isMobile bool) string {
	return fmt.Sprintf(`<iframe style="width:100%%;aspect-ratio:16/9;" src='https://player.youku.com/embed/%s' frameborder=0 'allowfullscreen'></iframe>`, y.VideoID)
}

type TencentVideo struct {
	VideoID string `json:"videoid" property:"videoid;视频ID"`
}

func (*TencentVideo) GetName() string {
	return "tencentvideo"
}

func (*TencentVideo) GetDisplayName() string {
	return "腾讯视频"
}

func (*TencentVideo) GetMediaType() MediaType {
	return VideoMediaType
}

func (t *TencentVideo) Init(config map[string]string) {
	videoid := config["videoid"]
	t.VideoID = videoid
}

func (t *TencentVideo) Html(isMobile bool) string {
	return fmt.Sprintf(`<iframe style="width:100%%;aspect-ratio:16/9;" frameborder="0" src="https://v.qq.com/txp/iframe/player.html?vid=%s" allowFullScreen="true"></iframe>`, t.VideoID)
}

type YouTuBeVideo struct {
	VideoID string `json:"videoid" property:"videoid;视频ID"`
}

func (*YouTuBeVideo) GetName() string {
	return "youtube"
}

func (*YouTuBeVideo) GetDisplayName() string {
	return "YouTuBe"
}

func (*YouTuBeVideo) GetMediaType() MediaType {
	return VideoMediaType
}

func (y *YouTuBeVideo) Init(config map[string]string) {
	videoid := config["videoid"]
	y.VideoID = videoid
}

func (y *YouTuBeVideo) Html(isMobile bool) string {
	return fmt.Sprintf(`<iframe  style="width:100%%;aspect-ratio:16/9;" src="https://www.youtube.com/embed/%s" frameborder="0" allow="accelerometer; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>`, y.VideoID)
}

type Music163 struct {
	MusicID string `json:"musicid" property:"musicid;音乐ID"`
}

func (*Music163) GetName() string {
	return "music163"
}

func (*Music163) GetDisplayName() string {
	return "网易云音乐"
}

func (*Music163) GetMediaType() MediaType {
	return MusicMediaType
}

func (m *Music163) Init(config map[string]string) {
	musicid := config["musicid"]
	m.MusicID = musicid
}

func (m *Music163) Html(isMobile bool) string {
	if isMobile {
		return fmt.Sprintf(`<iframe frameborder="no" border="0" marginwidth="0" marginheight="0" style="width:100%%;aspect-ratio:310/88" src="https://music.163.com/m/outchain/player?type=2&id=%s&auto=0&height=66"></iframe>`, m.MusicID)
	} else {
		return fmt.Sprintf(`<iframe frameborder="no" border="0" marginwidth="0" marginheight="0" style="width:100%%;aspect-ratio:310/66" src="https://music.163.com/outchain/player?type=2&id=%s&auto=0&height=66"></iframe>`, m.MusicID)
	}

}

type QQMusic struct {
	SongID string `json:"songid" property:"songid;音乐ID"`
}

func (*QQMusic) GetName() string {
	return "QQMusic"
}

func (*QQMusic) GetDisplayName() string {
	return "QQ音乐"
}

func (*QQMusic) GetMediaType() MediaType {
	return MusicMediaType
}

func (m *QQMusic) Init(config map[string]string) {
	songid := config["songid"]
	m.SongID = songid
}

func (m *QQMusic) Html(isMobile bool) string {
	return fmt.Sprintf(`<iframe frameborder="no" border="0" marginwidth="0" marginheight="0" style="width:100%%;aspect-ratio:330/66" src="https://i.y.qq.com/n2/m/outchain/player/index.html?songid=%s&songtype=0"></iframe>`, m.SongID)
}
