// ABOUTME: Defines and validates the uniform six-property font contract.
// ABOUTME: Keeps font inheritance confined to ordinary same-path configuration merging.
package config

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Font is the complete effective value shared by every configurable text role.
type Font struct {
	Family        string `yaml:"family"`
	Size          string `yaml:"size"`
	Weight        string `yaml:"weight"`
	Stretch       string `yaml:"stretch"`
	Style         string `yaml:"style"`
	LetterSpacing string `yaml:"letter-spacing"`
}

// FontRolePaths lists every public role governed by the uniform font contract.
var FontRolePaths = []string{
	"folio.font",
	"folio.heading.font",
	"folio.title-page.title.font",
	"folio.title-page.subtitle.font",
	"folio.title-page.author.font",
	"folio.title-page.date.font",
	"folio.title-page.version.font",
	"folio.positioning.speech.speaker.font",
	"folio.positioning.speech.speech-instruction.font",
	"folio.positioning.speech.dialogue.font",
	"folio.positioning.stage-direction.font",
	"folio.positioning.transition.font",
	"folio.positioning.frontmatter.header.font",
	"folio.positioning.act-header.font",
	"folio.positioning.scene-header.font",
	"folio.letter.font",
	"folio.manuscript.font",
	"folio.manuscript.heading.font",
	"folio.manuscript.mono.font",
	"folio.manuscript.quoted-block.font",
	"folio.manuscript.title-page.title.font",
	"folio.manuscript.title-page.subtitle.font",
	"folio.manuscript.title-page.author.font",
	"folio.manuscript.title-page.date.font",
	"folio.manuscript.title-page.wordcount.font",
	"folio.manuscript.title-page.version.font",
	"folio.manuscript.title-page.contact.font",
	"folio.manuscript.page-header.font",
	"folio.manuscript.page-footer.font",
	"folio.manuscript.toc.font",
	"folio.manuscript.toc.heading.font",
	"folio.manuscript.copyright.font",
	"folio.manuscript.copyright.label.font",
}

var (
	fontLengthRE = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]+)?|\.[0-9]+)(?:pt|mm|cm|in|em)$`)
	fontSizeRE   = regexp.MustCompile(`^(?:[0-9]+(?:\.[0-9]+)?|\.[0-9]+)(?:pt|mm|cm|in|em)$`)
	stretchRE    = regexp.MustCompile(`^(?:[0-9]+(?:\.[0-9]+)?|\.[0-9]+)%?$`)
)

var validFontWeights = map[string]bool{
	"thin": true, "extralight": true, "light": true, "regular": true,
	"medium": true, "semibold": true, "bold": true, "extrabold": true, "black": true,
}

var fontProperties = map[string]bool{
	"family": true, "size": true, "weight": true, "stretch": true,
	"style": true, "letter-spacing": true,
}

// Font returns a validated effective font block.
func (c Config) Font(path string) (Font, error) {
	value, ok := c.Get(path)
	if !ok {
		return Font{}, fmt.Errorf("%s is missing", path)
	}
	block, ok := value.(map[string]any)
	if !ok {
		return Font{}, fmt.Errorf("%s must be a mapping", path)
	}
	property := func(name string) (string, error) {
		value, ok := block[name].(string)
		if !ok {
			return "", fmt.Errorf("%s.%s must be a validated scalar value", path, name)
		}
		return value, nil
	}
	values := make([]string, 0, 6)
	for _, name := range []string{"family", "size", "weight", "stretch", "style", "letter-spacing"} {
		value, err := property(name)
		if err != nil {
			return Font{}, err
		}
		values = append(values, value)
	}
	return Font{
		Family: values[0], Size: values[1], Weight: values[2],
		Stretch: values[3], Style: values[4], LetterSpacing: values[5],
	}, nil
}

func validateFonts(data map[string]any) error {
	if err := rejectRetiredFontKeys(data); err != nil {
		return err
	}
	cfg := Config{data: data}
	for _, path := range FontRolePaths {
		value, ok := cfg.Get(path)
		if !ok {
			return fmt.Errorf("%s is missing", path)
		}
		block, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%s must be a mapping", path)
		}
		if err := validateFontBlock(path, block); err != nil {
			return err
		}
	}
	return nil
}

func validateFontBlock(path string, block map[string]any) error {
	for key := range block {
		if !fontProperties[key] {
			return fmt.Errorf("%s.%s is not supported", path, key)
		}
	}
	for property := range fontProperties {
		value, ok := block[property]
		if !ok {
			return fmt.Errorf("%s.%s is missing", path, property)
		}
		normalized, err := validateFontProperty(path+"."+property, property, value)
		if err != nil {
			return err
		}
		block[property] = normalized
	}
	return nil
}

func validateFontProperty(path, property string, value any) (string, error) {
	if value == nil {
		return "", fmt.Errorf("%s must not be null", path)
	}
	text, isString := value.(string)
	switch property {
	case "family":
		if !isString || strings.TrimSpace(text) == "" {
			return "", fmt.Errorf("%s must be a non-empty string", path)
		}
		return strings.TrimSpace(text), nil
	case "size":
		text = strings.TrimSpace(text)
		if !isString || !fontSizeRE.MatchString(text) {
			return "", fmt.Errorf("%s %q must be a positive length in pt, mm, cm, in, or em", path, value)
		}
		if numericPart(text) <= 0 {
			return "", fmt.Errorf("%s %q must be greater than zero", path, value)
		}
		return text, nil
	case "weight":
		return normalizeWeight(path, value)
	case "stretch":
		return normalizeStretch(path, value)
	case "style":
		if !isString {
			return "", fmt.Errorf("%s %q must be regular, italic, or oblique", path, value)
		}
		text = strings.ToLower(strings.TrimSpace(text))
		switch text {
		case "regular", "italic", "oblique":
			return text, nil
		default:
			return "", fmt.Errorf("%s %q must be regular, italic, or oblique", path, value)
		}
	case "letter-spacing":
		if !isString || !fontLengthRE.MatchString(strings.TrimSpace(text)) {
			return "", fmt.Errorf("%s %q must be a signed length in pt, mm, cm, in, or em", path, value)
		}
		return strings.TrimSpace(text), nil
	default:
		return "", fmt.Errorf("%s is not supported", path)
	}
}

func normalizeWeight(path string, value any) (string, error) {
	if text, ok := value.(string); ok {
		text = strings.ToLower(strings.TrimSpace(text))
		if validFontWeights[text] {
			return text, nil
		}
		weight, err := strconv.Atoi(text)
		if err == nil && weight >= 100 && weight <= 900 {
			return strconv.Itoa(weight), nil
		}
	} else if weight, ok := integerValue(value); ok && weight >= 100 && weight <= 900 {
		return strconv.FormatInt(weight, 10), nil
	}
	return "", fmt.Errorf("%s %q must be a Typst weight name or an integer from 100 to 900", path, value)
}

func normalizeStretch(path string, value any) (string, error) {
	var number float64
	switch typed := value.(type) {
	case string:
		text := strings.TrimSpace(typed)
		if !stretchRE.MatchString(text) {
			return "", fmt.Errorf("%s %q must be a positive percentage or plain number", path, value)
		}
		parsed, err := strconv.ParseFloat(strings.TrimSuffix(text, "%"), 64)
		if err != nil {
			return "", fmt.Errorf("%s %q must be a positive percentage or plain number", path, value)
		}
		number = parsed
	case int:
		number = float64(typed)
	case int64:
		number = float64(typed)
	case uint64:
		number = float64(typed)
	case float64:
		number = typed
	default:
		return "", fmt.Errorf("%s %q must be a positive percentage or plain number", path, value)
	}
	if math.IsNaN(number) || math.IsInf(number, 0) || number <= 0 {
		return "", fmt.Errorf("%s %q must be a positive percentage or plain number", path, value)
	}
	return strconv.FormatFloat(number, 'f', -1, 64) + "%", nil
}

func integerValue(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int64:
		return typed, true
	case uint64:
		if typed <= math.MaxInt64 {
			return int64(typed), true
		}
	}
	return 0, false
}

func numericPart(value string) float64 {
	for _, suffix := range []string{"pt", "mm", "cm", "in", "em"} {
		if strings.HasSuffix(value, suffix) {
			number, _ := strconv.ParseFloat(strings.TrimSuffix(value, suffix), 64)
			return number
		}
	}
	return 0
}

func rejectRetiredFontKeys(data map[string]any) error {
	fontParents := make(map[string]bool, len(FontRolePaths))
	for _, path := range FontRolePaths {
		fontParents[strings.TrimSuffix(path, ".font")] = true
	}
	var walk func(map[string]any, string) error
	walk = func(node map[string]any, path string) error {
		for key, value := range node {
			full := key
			if path != "" {
				full = path + "." + key
			}
			if fontParents[path] {
				switch key {
				case "font-size", "font-weight", "font-stretch", "font-style", "letter-spacing", "bold", "italic":
					return fmt.Errorf("%s is a retired font property; use %s.font", full, path)
				}
			}
			if isRetiredPrefixedFontPath(full) {
				return fmt.Errorf("%s is a retired font property", full)
			}
			if child, ok := value.(map[string]any); ok {
				if err := walk(child, full); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(data, "")
}

func isRetiredPrefixedFontPath(path string) bool {
	if path == "folio.heading-font" || strings.HasPrefix(path, "folio.heading-font-") {
		return true
	}
	const manuscript = "folio.manuscript."
	if !strings.HasPrefix(path, manuscript) {
		return false
	}
	name := strings.TrimPrefix(path, manuscript)
	for _, prefix := range []string{"heading", "mono", "title", "subtitle", "author", "date", "version", "wordcount", "contact"} {
		if name == prefix+"-font" || strings.HasPrefix(name, prefix+"-font-") {
			return true
		}
	}
	return name == "toc.heading-font" || strings.HasPrefix(name, "toc.heading-font-") || name == "copyright.heading-font-weight"
}
