// ABOUTME: Defines the typed renderer-neutral cover-image configuration contract.
// ABOUTME: Validates strict local keys while preserving permissive configuration elsewhere.
package config

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var coverHexColourRE = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// CoverImage is the complete effective cover-image value after cascade merge.
type CoverImage struct {
	Path                  string
	Mode                  string
	Placement             string
	Size                  string
	ImageTreatment        string
	ImageOpacity          float64
	Superscale            string
	DodgePadding          string
	OverlayBlur           string
	OverlayWashTreatment  string
	OverlayWashOpacity    float64
	OverlayPadding        string
	BackdropMode          string
	BackdropImageBlur     string
	BackdropSolidColor    string
	BackdropWashTreatment string
	BackdropWashOpacity   float64
}

// CoverImage returns the typed effective cover-image configuration when enabled.
func (c Config) CoverImage() (CoverImage, bool, error) {
	if !c.Bool("folio.title-page.enabled", true) {
		return CoverImage{}, false, nil
	}
	raw, exists := c.Get("folio.cover-image")
	if !exists || raw == nil {
		return CoverImage{}, false, nil
	}
	block, ok := raw.(map[string]any)
	if !ok {
		return CoverImage{}, false, fmt.Errorf("folio.cover-image must be a mapping; got %#v", raw)
	}
	if err := coverKeys("folio.cover-image", block, "enabled", "path", "mode", "placement", "size", "image", "superscale", "dodge", "overlay", "backdrop"); err != nil {
		return CoverImage{}, false, err
	}
	if len(block) == 0 {
		return CoverImage{}, false, nil
	}
	if value, present := block["enabled"]; present {
		enabled, ok := value.(bool)
		if !ok {
			return CoverImage{}, false, fmt.Errorf("folio.cover-image.enabled must be a boolean; got %#v", value)
		}
		if !enabled {
			return CoverImage{}, false, nil
		}
	}
	path, err := coverString(block, "path", "folio.cover-image.path", true)
	if err != nil {
		return CoverImage{}, false, err
	}
	if path == "" {
		return CoverImage{}, false, fmt.Errorf("folio.cover-image.path must be a non-empty local PNG or JPEG path")
	}
	if strings.Contains(path, "://") {
		return CoverImage{}, false, fmt.Errorf("folio.cover-image.path must be a local PNG or JPEG path; got %q", path)
	}
	mode, err := coverString(block, "mode", "folio.cover-image.mode", true)
	if err != nil {
		return CoverImage{}, false, err
	}
	if mode != "background" && mode != "inset" {
		return CoverImage{}, false, fmt.Errorf("folio.cover-image.mode must be background or inset; got %q", mode)
	}
	placement, err := coverStringDefault(block, "placement", "folio.cover-image.placement", "overlay")
	if err != nil {
		return CoverImage{}, false, err
	}
	if placement != "overlay" && placement != "dodge" {
		return CoverImage{}, false, fmt.Errorf("folio.cover-image.placement must be overlay or dodge; got %q", placement)
	}
	if mode == "background" && placement == "dodge" {
		return CoverImage{}, false, fmt.Errorf("folio.cover-image.placement dodge is supported only with folio.cover-image.mode inset")
	}

	size, err := coverStringDefault(block, "size", "folio.cover-image.size", "")
	if err != nil {
		return CoverImage{}, false, err
	}
	if mode == "inset" {
		if _, err := coverLength(size, "folio.cover-image.size", "mm", true); err != nil {
			return CoverImage{}, false, err
		}
	} else if size != "" {
		return CoverImage{}, false, fmt.Errorf("folio.cover-image.size is forbidden when folio.cover-image.mode is background")
	}

	image, err := coverMapping(block, "image", "folio.cover-image.image")
	if err != nil {
		return CoverImage{}, false, err
	}
	if err := coverKeys("folio.cover-image.image", image, "treatment", "opacity"); err != nil {
		return CoverImage{}, false, err
	}
	imageTreatment, err := coverStringDefault(image, "treatment", "folio.cover-image.image.treatment", "light")
	if err != nil {
		return CoverImage{}, false, err
	}
	if imageTreatment != "light" && imageTreatment != "dark" {
		return CoverImage{}, false, fmt.Errorf("folio.cover-image.image.treatment must be light or dark; got %q", imageTreatment)
	}
	imageOpacity, err := coverPercentage(image, "opacity", "folio.cover-image.image.opacity", 1)
	if err != nil {
		return CoverImage{}, false, err
	}

	superscale := ""
	if value, ok := block["superscale"]; ok && value != nil {
		text, ok := value.(string)
		if !ok {
			return CoverImage{}, false, fmt.Errorf("folio.cover-image.superscale must be null, default, 2x, 4x or 8x; got %#v", value)
		}
		superscale = strings.TrimSpace(text)
		if superscale != "default" && superscale != "2x" && superscale != "4x" && superscale != "8x" {
			return CoverImage{}, false, fmt.Errorf("folio.cover-image.superscale must be null, default, 2x, 4x or 8x; got %q", superscale)
		}
	}
	dodge, err := coverMapping(block, "dodge", "folio.cover-image.dodge")
	if err != nil {
		return CoverImage{}, false, err
	}
	if err := coverKeys("folio.cover-image.dodge", dodge, "padding"); err != nil {
		return CoverImage{}, false, err
	}
	if placement != "dodge" && len(dodge) != 0 {
		return CoverImage{}, false, fmt.Errorf("folio.cover-image.dodge is supported only when folio.cover-image.placement is dodge")
	}
	dodgePadding, err := coverStringDefault(dodge, "padding", "folio.cover-image.dodge.padding", "1em")
	if err != nil {
		return CoverImage{}, false, err
	}
	if _, err := coverLength(dodgePadding, "folio.cover-image.dodge.padding", "em|mm", true); err != nil {
		return CoverImage{}, false, err
	}

	overlay, err := coverMapping(block, "overlay", "folio.cover-image.overlay")
	if err != nil {
		return CoverImage{}, false, err
	}
	if err := coverKeys("folio.cover-image.overlay", overlay, "blur", "wash", "padding"); err != nil {
		return CoverImage{}, false, err
	}
	blur, err := coverStringDefault(overlay, "blur", "folio.cover-image.overlay.blur", "0mm")
	if err != nil {
		return CoverImage{}, false, err
	}
	blurValue, err := coverLength(blur, "folio.cover-image.overlay.blur", "mm", false)
	if err != nil {
		return CoverImage{}, false, err
	}
	padding, err := coverStringDefault(overlay, "padding", "folio.cover-image.overlay.padding", "0em")
	if err != nil {
		return CoverImage{}, false, err
	}
	paddingValue, err := coverLength(padding, "folio.cover-image.overlay.padding", "em|mm", false)
	if err != nil {
		return CoverImage{}, false, err
	}
	wash, err := coverMapping(overlay, "wash", "folio.cover-image.overlay.wash")
	if err != nil {
		return CoverImage{}, false, err
	}
	if err := coverKeys("folio.cover-image.overlay.wash", wash, "treatment", "opacity"); err != nil {
		return CoverImage{}, false, err
	}
	washTreatment, err := coverStringDefault(wash, "treatment", "folio.cover-image.overlay.wash.treatment", "light")
	if err != nil {
		return CoverImage{}, false, err
	}
	if washTreatment != "light" && washTreatment != "dark" {
		return CoverImage{}, false, fmt.Errorf("folio.cover-image.overlay.wash.treatment must be light or dark; got %q", washTreatment)
	}
	washOpacity, err := coverPercentage(wash, "opacity", "folio.cover-image.overlay.wash.opacity", 0)
	if err != nil {
		return CoverImage{}, false, err
	}
	if placement == "dodge" && (blurValue != 0 || paddingValue != 0 || washOpacity != 0 || washTreatment != "light") {
		return CoverImage{}, false, fmt.Errorf("folio.cover-image.overlay must retain its defaults when folio.cover-image.placement is dodge")
	}

	backdropMode, backdropImageBlur, backdropSolidColor := "", "", ""
	backdropWashTreatment := ""
	backdropWashOpacity := 0.0
	backdropValue, backdropPresent := block["backdrop"]
	if backdropPresent && backdropValue != nil {
		backdrop, err := coverMapping(block, "backdrop", "folio.cover-image.backdrop")
		if err != nil {
			return CoverImage{}, false, err
		}
		if err := coverKeys("folio.cover-image.backdrop", backdrop, "mode", "image", "solid", "wash"); err != nil {
			return CoverImage{}, false, err
		}
		backdropImage, err := coverMapping(backdrop, "image", "folio.cover-image.backdrop.image")
		if err != nil {
			return CoverImage{}, false, err
		}
		if err := coverKeys("folio.cover-image.backdrop.image", backdropImage, "blur"); err != nil {
			return CoverImage{}, false, err
		}
		backdropSolid, err := coverMapping(backdrop, "solid", "folio.cover-image.backdrop.solid")
		if err != nil {
			return CoverImage{}, false, err
		}
		if err := coverKeys("folio.cover-image.backdrop.solid", backdropSolid, "color"); err != nil {
			return CoverImage{}, false, err
		}
		backdropWash, err := coverMapping(backdrop, "wash", "folio.cover-image.backdrop.wash")
		if err != nil {
			return CoverImage{}, false, err
		}
		if err := coverKeys("folio.cover-image.backdrop.wash", backdropWash, "treatment", "opacity"); err != nil {
			return CoverImage{}, false, err
		}

		backdropMode, err = coverString(backdrop, "mode", "folio.cover-image.backdrop.mode", true)
		if err != nil {
			return CoverImage{}, false, err
		}
		backdropImageBlur, err = coverStringDefault(backdropImage, "blur", "folio.cover-image.backdrop.image.blur", "8mm")
		if err != nil {
			return CoverImage{}, false, err
		}
		backdropSolidColor, err = coverStringDefault(backdropSolid, "color", "folio.cover-image.backdrop.solid.color", "")
		if err != nil {
			return CoverImage{}, false, err
		}
		backdropWashTreatment, err = coverStringDefault(backdropWash, "treatment", "folio.cover-image.backdrop.wash.treatment", "")
		if err != nil {
			return CoverImage{}, false, err
		}
		if value, exists := backdropWash["opacity"]; exists && value != nil {
			if _, ok := value.(string); !ok {
				return CoverImage{}, false, fmt.Errorf("folio.cover-image.backdrop.wash.opacity must be a percentage string; got %#v", value)
			}
			if inactiveValue, inactiveErr := coverPercentage(backdropWash, "opacity", "folio.cover-image.backdrop.wash.opacity", 0); inactiveErr == nil {
				backdropWashOpacity = inactiveValue
			}
		}

		if backdropMode != "none" && backdropMode != "image" && backdropMode != "solid" {
			return CoverImage{}, false, fmt.Errorf("folio.cover-image.backdrop.mode must be none, image or solid; got %q", backdropMode)
		}
		if mode != "inset" {
			return CoverImage{}, false, fmt.Errorf("folio.cover-image.backdrop is supported only when folio.cover-image.mode is inset")
		}
		if backdropMode == "image" {
			if _, err := coverLength(backdropImageBlur, "folio.cover-image.backdrop.image.blur", "mm", true); err != nil {
				return CoverImage{}, false, err
			}
		}
		if backdropMode == "solid" {
			if !coverHexColourRE.MatchString(backdropSolidColor) {
				return CoverImage{}, false, fmt.Errorf("folio.cover-image.backdrop.solid.color must be exactly #RRGGBB; got %q", backdropSolidColor)
			}
		}
		if backdropMode == "image" || backdropMode == "solid" {
			if backdropWashTreatment == "" {
				return CoverImage{}, false, fmt.Errorf("folio.cover-image.backdrop.wash.treatment is required")
			}
			if backdropWashTreatment != "light" && backdropWashTreatment != "dark" {
				return CoverImage{}, false, fmt.Errorf("folio.cover-image.backdrop.wash.treatment must be light or dark; got %q", backdropWashTreatment)
			}
			backdropWashOpacity, err = coverPercentage(backdropWash, "opacity", "folio.cover-image.backdrop.wash.opacity", 0)
			if err != nil {
				return CoverImage{}, false, err
			}
		}
	}

	return CoverImage{
		Path: path, Mode: mode, Placement: placement, Size: size,
		ImageTreatment: imageTreatment, ImageOpacity: imageOpacity, Superscale: superscale,
		DodgePadding: dodgePadding,
		OverlayBlur:  blur, OverlayWashTreatment: washTreatment,
		OverlayWashOpacity: washOpacity, OverlayPadding: padding,
		BackdropMode: backdropMode, BackdropImageBlur: backdropImageBlur,
		BackdropSolidColor: backdropSolidColor, BackdropWashTreatment: backdropWashTreatment,
		BackdropWashOpacity: backdropWashOpacity,
	}, true, nil
}

func coverKeys(path string, values map[string]any, allowed ...string) error {
	known := map[string]bool{}
	for _, key := range allowed {
		known[key] = true
	}
	var unknown []string
	for key := range values {
		if !known[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sort.Strings(unknown)
	return fmt.Errorf("unknown configuration field %s.%s", path, unknown[0])
}

func coverMapping(parent map[string]any, key, path string) (map[string]any, error) {
	value, ok := parent[key]
	if !ok || value == nil {
		return map[string]any{}, nil
	}
	mapping, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a mapping; got %#v", path, value)
	}
	return mapping, nil
}

func coverString(values map[string]any, key, path string, required bool) (string, error) {
	value, ok := values[key]
	if !ok || value == nil {
		if required {
			return "", fmt.Errorf("%s is required", path)
		}
		return "", nil
	}
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string; got %#v", path, value)
	}
	text = strings.TrimSpace(text)
	if required && text == "" {
		return "", fmt.Errorf("%s must not be empty", path)
	}
	return text, nil
}

func coverStringDefault(values map[string]any, key, path, fallback string) (string, error) {
	value, ok := values[key]
	if !ok || value == nil {
		return fallback, nil
	}
	return coverString(values, key, path, false)
}

func coverPercentage(values map[string]any, key, path string, fallback float64) (float64, error) {
	value, ok := values[key]
	if !ok || value == nil {
		return fallback, nil
	}
	text, ok := value.(string)
	if !ok || !strings.HasSuffix(strings.TrimSpace(text), "%") {
		return 0, fmt.Errorf("%s must be a percentage from 0%% through 100%%; got %#v", path, value)
	}
	number, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(text), "%")), 64)
	if err != nil || math.IsNaN(number) || math.IsInf(number, 0) || number < 0 || number > 100 {
		return 0, fmt.Errorf("%s must be a percentage from 0%% through 100%%; got %#v", path, value)
	}
	return number / 100, nil
}

func coverLength(text, path, units string, positive bool) (float64, error) {
	text = strings.TrimSpace(text)
	unit := ""
	for _, candidate := range strings.Split(units, "|") {
		if strings.HasSuffix(text, candidate) {
			unit = candidate
			break
		}
	}
	if unit == "" {
		return 0, fmt.Errorf("%s must be a %s length; got %q", path, strings.ReplaceAll(units, "|", " or "), text)
	}
	number, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(text, unit)), 64)
	if err != nil || math.IsNaN(number) || math.IsInf(number, 0) || number < 0 || (positive && number == 0) {
		qualifier := "non-negative"
		if positive {
			qualifier = "positive"
		}
		return 0, fmt.Errorf("%s must be a %s %s length; got %q", path, qualifier, strings.ReplaceAll(units, "|", " or "), text)
	}
	return number, nil
}
