package services

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sitehub/service/rbac/model/entity"
	"strings"
	"unicode/utf8"
)

const maxSystemIconSVGBytes = 64 * 1024

var (
	systemIconNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)
	allowedSVGElements    = map[string]bool{
		"svg": true, "g": true, "path": true, "rect": true, "circle": true,
		"ellipse": true, "line": true, "polyline": true, "polygon": true,
		"title": true, "desc": true,
	}
	allowedSVGAttributes = map[string]bool{
		"xmlns": true, "viewBox": true, "width": true, "height": true,
		"fill": true, "fill-rule": true, "fill-opacity": true,
		"stroke": true, "stroke-width": true, "stroke-linecap": true,
		"stroke-linejoin": true, "stroke-opacity": true, "opacity": true,
		"transform": true, "d": true, "x": true, "y": true, "x1": true,
		"y1": true, "x2": true, "y2": true, "cx": true, "cy": true,
		"r": true, "rx": true, "ry": true, "points": true,
		"preserveAspectRatio": true, "role": true, "aria-hidden": true,
	}
)

type SystemIconInput struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Category string `json:"category"`
	SVG      string `json:"svg"`
}

func normalizeSystemIconInput(input SystemIconInput) (*entity.SystemIcon, error) {
	name := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(input.Name, "custom:")))
	if !systemIconNamePattern.MatchString(name) {
		return nil, fmt.Errorf("图标标识只能包含小写字母、数字和连字符，并且必须以字母开头")
	}
	label := strings.TrimSpace(input.Label)
	if label == "" {
		label = name
	}
	if utf8.RuneCountInString(label) > 80 {
		return nil, fmt.Errorf("图标名称不能超过 80 个字符")
	}
	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = "自定义"
	}
	if utf8.RuneCountInString(category) > 40 {
		return nil, fmt.Errorf("图标分类不能超过 40 个字符")
	}
	svg, err := sanitizeSystemIconSVG(input.SVG)
	if err != nil {
		return nil, err
	}
	return &entity.SystemIcon{Name: "custom:" + name, Label: label, Category: category, SVG: svg}, nil
}

func sanitizeSystemIconSVG(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("SVG 内容不能为空")
	}
	if len(raw) > maxSystemIconSVGBytes {
		return "", fmt.Errorf("SVG 内容不能超过 64KB")
	}

	decoder := xml.NewDecoder(strings.NewReader(raw))
	decoder.Strict = true
	var output bytes.Buffer
	encoder := xml.NewEncoder(&output)
	depth, elements := 0, 0
	rootSeen := false
	textElementDepth := 0

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("SVG XML 格式无效: %w", err)
		}

		switch value := token.(type) {
		case xml.StartElement:
			name := value.Name.Local
			if !allowedSVGElements[name] {
				return "", fmt.Errorf("SVG 不允许使用 <%s> 元素", name)
			}
			if !rootSeen {
				if name != "svg" {
					return "", fmt.Errorf("SVG 根元素必须是 <svg>")
				}
				rootSeen = true
			} else if depth == 0 || name == "svg" {
				return "", fmt.Errorf("SVG 只能包含一个根元素")
			}
			depth++
			elements++
			if depth > 32 || elements > 512 {
				return "", fmt.Errorf("SVG 结构过于复杂")
			}
			if name == "title" || name == "desc" {
				textElementDepth = depth
			}

			clean := xml.StartElement{Name: xml.Name{Local: name}}
			if depth == 1 {
				clean.Attr = append(clean.Attr, xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: "http://www.w3.org/2000/svg"})
			}
			for _, attribute := range value.Attr {
				attributeName := attribute.Name.Local
				if attributeName == "xmlns" || attribute.Name.Space == "xmlns" {
					continue
				}
				if attribute.Name.Space != "" || !allowedSVGAttributes[attributeName] {
					return "", fmt.Errorf("SVG 不允许使用 %s 属性", attributeName)
				}
				lowerValue := strings.ToLower(attribute.Value)
				if strings.Contains(lowerValue, "url(") || strings.Contains(lowerValue, "javascript:") ||
					strings.Contains(lowerValue, "data:") || strings.Contains(lowerValue, "http:") ||
					strings.Contains(lowerValue, "https:") || strings.Contains(lowerValue, "//") {
					return "", fmt.Errorf("SVG 属性不能引用外部资源")
				}
				clean.Attr = append(clean.Attr, xml.Attr{Name: xml.Name{Local: attributeName}, Value: attribute.Value})
			}
			if err := encoder.EncodeToken(clean); err != nil {
				return "", err
			}
		case xml.EndElement:
			if depth <= 0 {
				return "", fmt.Errorf("SVG 元素闭合顺序无效")
			}
			if err := encoder.EncodeToken(xml.EndElement{Name: xml.Name{Local: value.Name.Local}}); err != nil {
				return "", err
			}
			if textElementDepth == depth {
				textElementDepth = 0
			}
			depth--
		case xml.CharData:
			if strings.TrimSpace(string(value)) != "" && textElementDepth == 0 {
				return "", fmt.Errorf("SVG 仅允许在 title 或 desc 中使用文本")
			}
			if textElementDepth > 0 {
				if err := encoder.EncodeToken(value); err != nil {
					return "", err
				}
			}
		case xml.Comment:
			// Comments are removed from stored resources.
		case xml.ProcInst:
			if strings.ToLower(value.Target) != "xml" {
				return "", fmt.Errorf("SVG 不允许处理指令")
			}
		case xml.Directive:
			return "", fmt.Errorf("SVG 不允许 DOCTYPE 或实体声明")
		}
	}

	if !rootSeen || depth != 0 {
		return "", fmt.Errorf("SVG 根元素不完整")
	}
	if err := encoder.Flush(); err != nil {
		return "", err
	}
	return output.String(), nil
}
