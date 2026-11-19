package domains

import (
	"errors"
	"regexp"
	"strings"
)

type Theme struct {
	PrimaryColor    string `json:"primary_color"`
	BackgroundColor string `json:"background_color"`
	TextColor       string `json:"text_color"`
}

type ThemePatch struct {
	PrimaryColor    *string `json:"primary_color"`
	BackgroundColor *string `json:"background_color"`
	TextColor       *string `json:"text_color"`
}

var ErrInvalidThemeColor = errors.New("theme color must be #RRGGBB")
var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func NormalizeThemePatch(p ThemePatch) (ThemePatch, error) {
	var result ThemePatch
	for _, pair := range []struct {
		src *string
		dst **string
	}{
		{p.PrimaryColor, &result.PrimaryColor},
		{p.BackgroundColor, &result.BackgroundColor},
		{p.TextColor, &result.TextColor},
	} {
		if pair.src == nil {
			continue
		}
		if !hexColor.MatchString(*pair.src) {
			return ThemePatch{}, ErrInvalidThemeColor
		}
		value := strings.ToUpper(*pair.src)
		*pair.dst = &value
	}
	return result, nil
}
