package card

import (
	"os"
	"regexp"
	"strings"
)

func splitByBraces(s string) []string {
	// 定义正则表达式来匹配 { 和 }
	re := regexp.MustCompile(`(\{|\})`)

	// 使用正则表达式分割字符串并保留分隔符
	parts := re.Split(s, -1)
	matches := re.FindAllString(s, -1)

	// 合并结果，确保分割后的字符串包含 { 和 }
	result := make([]string, 0)
	for i, part := range parts {
		if part != "" {
			result = append(result, part)
		}
		if i < len(matches) {
			result = append(result, matches[i])
		}
	}

	return result
}

type CssRule struct {
	IsMedia  bool
	Selector string
	Styles   []*string
	Rules    []*CssRule
}

type CssFile struct {
	Rules []*CssRule
}

type NodeStatus int

const (
	NodeNotStartStatus NodeStatus = 0
	NodeStartStatus    NodeStatus = 1
	NodeEnterStatus    NodeStatus = 2
)

type CssParser struct {
	cssFile *CssFile
}

func parseCSS(cssContent string) CssFile {
	var cssFile CssFile
	var currentRule *CssRule
	var selector string
	mediaNodeStatus := NodeNotStartStatus
	cssNodeStatus := NodeNotStartStatus
	lines := strings.Split(cssContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		for _, line := range splitByBraces(line) {
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "@media") {
				mediaNodeStatus = NodeStartStatus
				selector = line
			} else if line == "{" {
				if mediaNodeStatus == NodeStartStatus {
					currentRule = &CssRule{
						IsMedia:  true,
						Selector: selector,
						Rules:    []*CssRule{},
					}
					mediaNodeStatus = NodeEnterStatus
				} else if cssNodeStatus == NodeStartStatus {
					if mediaNodeStatus == NodeEnterStatus {
						if currentRule != nil {
							newrule := &CssRule{
								IsMedia:  false,
								Selector: selector,
								Styles:   []*string{},
							}
							currentRule.Rules = append(currentRule.Rules, newrule)
						}

					} else {
						currentRule = &CssRule{
							IsMedia:  false,
							Selector: selector,
						}
					}
					cssNodeStatus = NodeEnterStatus
				}
			} else if line == "}" {
				if mediaNodeStatus == NodeEnterStatus {
					if cssNodeStatus == NodeEnterStatus {
						cssNodeStatus = NodeNotStartStatus
					} else {
						mediaNodeStatus = NodeNotStartStatus
						if currentRule != nil {
							cssFile.Rules = append(cssFile.Rules, currentRule)
						}
						currentRule = nil
					}
				} else {
					cssNodeStatus = NodeNotStartStatus
					if currentRule != nil {
						cssFile.Rules = append(cssFile.Rules, currentRule)
					}
					currentRule = nil
				}
			} else {
				if (mediaNodeStatus == NodeNotStartStatus || mediaNodeStatus == NodeEnterStatus) && cssNodeStatus == NodeNotStartStatus {
					cssNodeStatus = NodeStartStatus
					selector = line
				} else {
					if mediaNodeStatus == NodeEnterStatus && cssNodeStatus == NodeEnterStatus {
						if currentRule != nil && currentRule.IsMedia && len(currentRule.Rules) > 0 {
							index := len(currentRule.Rules) - 1
							currentRule.Rules[index].Styles = append(currentRule.Rules[index].Styles, &line)
						}
					} else if mediaNodeStatus == NodeNotStartStatus && cssNodeStatus == NodeEnterStatus {
						if currentRule != nil && !currentRule.IsMedia {
							currentRule.Styles = append(currentRule.Styles, &line)
						}
					}
				}
			}
		}

	}

	return cssFile
}

func ParseCss(csspath string) (*CssParser, error) {
	content, err := os.ReadFile(csspath)
	if err != nil {
		return nil, err
	}
	cssContent := string(content)
	cssFile := parseCSS(cssContent)
	return &CssParser{
		cssFile: &cssFile,
	}, nil
}

func (p *CssParser) SetPrefix(prefix string) {
	if p.cssFile != nil {
		for _, rule := range p.cssFile.Rules {
			if rule.IsMedia {
				for _, item := range rule.Rules {
					item.Selector = prefix + " " + item.Selector
				}
			} else {
				rule.Selector = prefix + " " + rule.Selector
			}
		}
	}
}

func (p *CssParser) WriteFile(path string) error {
	var cssBuilder strings.Builder
	cssBuilder.WriteString(`/* 由程序每次启动时自动生成，请勿修改 */` + "\n\n")
	for _, rule := range p.cssFile.Rules {
		if rule.IsMedia {
			cssBuilder.WriteString(rule.Selector + " {\n")
			for _, item := range rule.Rules {
				cssBuilder.WriteString("  " + item.Selector + " {\n")
				for _, style := range item.Styles {
					cssBuilder.WriteString("    " + *style + "\n")
				}
				cssBuilder.WriteString("  }\n")
			}
			cssBuilder.WriteString("}\n")
		} else {
			cssBuilder.WriteString(rule.Selector + " {\n")
			for _, style := range rule.Styles {
				cssBuilder.WriteString("  " + *style + "\n")
			}
			cssBuilder.WriteString("}\n")
		}
	}

	return os.WriteFile(path, []byte(cssBuilder.String()), 0644)

}
