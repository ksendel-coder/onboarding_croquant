package domains

import (
	"net/url"
	"slices"
	"strings"
)

// NormalizeOrigin приводит значение к тому виду, в котором origin присылает
// браузер: схема и хост в нижнем регистре, порт как есть, ничего после хоста.
// Пустая строка на выходе значит, что это не origin.
func NormalizeOrigin(raw string) string {
	s := strings.TrimSuffix(strings.TrimSpace(raw), "/")
	if s == "" {
		return ""
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" || u.User != nil {
		return ""
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return ""
	}
	if u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return ""
	}
	return scheme + "://" + strings.ToLower(u.Host)
}

// ValidateOrigins вызывается перед любой записью в базу:
// дедуп, нормализация, ошибка на первом мусоре.
// Пустой список отвергается – приложение без origin не может работать.
func ValidateOrigins(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return nil, ErrOriginRequired
	}
	seen := make(map[string]struct{}, len(raw))
	result := make([]string, 0, len(raw))
	for _, r := range raw {
		// Вайлдкард не origin и нормализации не поддаётся, но OriginAllowed на
		// него опирается. Без этой ветки записать "*" можно было бы только
		// руками в базу, а ветка в проверке осталась бы недостижимой.
		norm := strings.TrimSpace(r)
		if norm != "*" {
			norm = NormalizeOrigin(r)
		}
		if norm == "" {
			return nil, ErrInvalidOrigin
		}
		if _, ok := seen[norm]; ok {
			continue
		}
		seen[norm] = struct{}{}
		result = append(result, norm)
	}
	return result, nil
}

// originAllowed проверяет, разрешён ли origin для списка.
func OriginAllowed(allowed []string, origin string) bool {
	if slices.Contains(allowed, "*") {
		return true
	}
	return slices.Contains(allowed, NormalizeOrigin(origin))
}
