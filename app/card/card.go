package card

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"

	apperror "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/app/models"
	"github.com/lixiaofei123/daily/app/utils"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func init() {
	loadCardTemplates()
}

func loadCardTemplates() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	templatesDir := path.Join(currentDir, "public/card/")
	templatesDirFile, err := os.Open(templatesDir)
	if err != nil {
		return err
	}

	templateDirs, err := templatesDirFile.ReadDir(-1)
	if err != nil {
		return err
	}

	for _, templateDir := range templateDirs {
		if templateDir.IsDir() {
			loadCardTemplate(currentDir, path.Join(templatesDir, templateDir.Name()))
		}
	}

	return nil

}

var cardMap map[string]*Card = map[string]*Card{}

func loadCardTemplate(workpath, templateDir string) error {

	cardName := path.Base(templateDir)
	configFile := path.Join(templateDir, "card.yaml")
	stat, err := os.Stat(configFile)
	if err == nil && !stat.IsDir() {
		viperConfig := viper.New()
		viperConfig.SetConfigFile(configFile)
		if err := viperConfig.ReadInConfig(); err == nil {
			displayName := viperConfig.GetString("displayName")
			cssPath := viperConfig.GetString("css")
			modelPath := viperConfig.GetString("model")
			templatePath := viperConfig.GetString("template")
			javascriptPath := viperConfig.GetString("javascript")
			globalJavascriptPath := viperConfig.GetString("globalJavascript")
			jsDataScopes := viperConfig.GetStringSlice("jsDataScopes")

			if displayName == "" {
				displayName = cardName
			}

			if templatePath == "" {
				return nil
			}

			tmpl, err := template.ParseFiles(path.Join(templateDir, templatePath))
			if err != nil {
				return nil
			}

			card := &Card{
				template: tmpl,
			}

			if cssPath != "" {
				classname := "card-" + utils.Md5(cardName)
				cssParser, err := ParseCss(path.Join(templateDir, cssPath))
				if err == nil {
					cssParser.SetPrefix("." + classname)
					newcsspath := path.Join(workpath, "static/css/card", cardName+".css")
					err = cssParser.WriteFile(newcsspath)
					if err == nil {
						card.cssFile = cardName + ".css"
					}
				}
			}

			if modelPath != "" {
				modelViperConfig := viper.New()
				modelViperConfig.SetConfigFile(path.Join(templateDir, modelPath))
				if err := modelViperConfig.ReadInConfig(); err == nil {
					cardModel := &CardModel{
						Name:        cardName,
						Displayname: displayName,
						Props:       []*CardProp{},
					}

					var parsedData yaml.MapSlice
					data, _ := os.ReadFile(path.Join(templateDir, modelPath))
					yaml.Unmarshal(data, &parsedData)

					for _, item := range parsedData {
						key := item.Key.(string)
						value := modelViperConfig.Get(key)
						valuedata, _ := json.Marshal(value)
						cardProp := &CardProp{}
						err = json.Unmarshal(valuedata, cardProp)
						if err == nil {
							cardProp.Name = key
							if cardProp.Displayname == "" {
								cardProp.Displayname = key
							}
							cardModel.Props = append(cardModel.Props, cardProp)
						}
					}

					CardModels[cardName] = cardModel
				}

			}

			if javascriptPath != "" {
				data, err := os.ReadFile(path.Join(templateDir, javascriptPath))
				if err == nil {
					card.javascript = string(data)
				}
			}

			if globalJavascriptPath != "" {
				data, err := os.ReadFile(path.Join(templateDir, globalJavascriptPath))
				if err == nil {
					card.globalJavascript = string(data)
				}
			}

			if len(jsDataScopes) > 0 {
				card.dataScopes = jsDataScopes
			}

			log.Printf("加载模板[%s]成功 \n", cardName)

			cardMap[cardName] = card

		}
	}

	return nil
}

type Image struct {
	Url string `json:"url"`
}

type Link struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

type Option struct {
	Value string `json:"value"`
}

type Card struct {
	template         *template.Template
	cssFile          string
	javascript       string
	globalJavascript string
	dataScopes       []string
}

type PropType string

const (
	NumberPropType   PropType = "number"
	StringPropType   PropType = "string"
	ImagePropType    PropType = "image"
	LinkPropType     PropType = "link"
	OptionPropType   PropType = "option"
	BooleanPropType  PropType = "boolean"
	ArrayPropType    PropType = "array"
	DatePropType     PropType = "date"
	TimePropType     PropType = "time"
	DateTimePropType PropType = "datetime"
)

type OptionItem struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type CardProp struct {
	Name        string        `json:"name"`
	Displayname string        `json:"displayname"`
	Type        PropType      `json:"type"`
	Long        bool          `json:"long"`
	ApectRadio  *string       `json:"aspectRadio"`
	OptionItems []*OptionItem `json:"options"`
	Required    bool          `json:"required"`
	Min         int           `json:"min"`
	Max         int           `json:"max"`
}

type CardModel struct {
	Name        string      `json:"name"`
	Displayname string      `json:"displayname"`
	Props       []*CardProp `json:"props"`
}

var CardModels map[string]*CardModel = map[string]*CardModel{}

func Html(cardname string, modeldata interface{}, post *models.PostDetail) (string, error) {
	card, ok := cardMap[cardname]
	if !ok {
		return "", apperror.ErrNoSuchCard
	}

	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	templateData := map[string]interface{}{}
	templateData["model"] = modeldata
	templateData["post"] = post

	err := card.template.Execute(buffer, templateData)
	if err != nil {
		return "", err
	}

	classname := "card-" + utils.Md5(cardname)
	cardid := fmt.Sprintf("post_card_%d", post.Post.ID)

	js := card.javascript
	scopes := card.dataScopes
	localscript := ""
	if js != "" {
		modelbytes, _ := json.Marshal(modeldata)
		jsmodel := string(modelbytes)
		jspost := "{}"
		jsuser := "{}"
		jsemail := ""
		for _, scope := range scopes {
			if scope == "post" {
				simplePost := &models.Post{
					Content: post.Post.Content,
					Address: post.Post.Address,
				}
				simplePost.ID = post.Post.ID
				simplePost.CreatedAt = post.Post.CreatedAt

				jspost0, _ := json.Marshal(simplePost)
				jspost = string(jspost0)
			}
			if scope == "user" {
				simpleUser := &models.UserDetail{
					ID: post.UserDetail.ID,
					User: &models.User{
						Role:     post.UserDetail.User.Role,
						Username: post.UserDetail.User.Username,
					},
					Profile: &models.UserProfile{
						Avatar:    post.UserDetail.Profile.Avatar,
						Cover:     post.UserDetail.Profile.Cover,
						Signature: post.UserDetail.Profile.Signature,
					},
				}

				jsuser0, _ := json.Marshal(simpleUser)
				jsuser = string(jsuser0)
			}

			if scope == "email" {
				jsemail = post.UserDetail.User.Email
			}
		}

		localscript = fmt.Sprintf(`<script>
			(function initcard(){
				let carddiv = $("#%s")
				let model = %s
				let post = %s
				let user = %s
				let email = "%s"
				%s
			})()
		</script>`, cardid, jsmodel, jspost, jsuser, jsemail, js)
	}

	return fmt.Sprintf(`<div class="%s" id="%s" >%s</div> %s`, classname, cardid, buffer.String(), localscript), nil
}

func GetCardCssPathes() []string {
	csses := []string{}
	for _, card := range cardMap {
		if card.cssFile != "" {
			csses = append(csses, card.cssFile)
		}
	}
	return csses
}

func CardGlobalJavascript() string {
	javascript := ""
	for _, card := range cardMap {
		if card.globalJavascript != "" {
			javascript = javascript + "\n" + card.globalJavascript
		}
	}
	return javascript

}
