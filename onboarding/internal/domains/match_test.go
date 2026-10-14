package domains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{"звёздочка ловит всё", "/*", "/anything/deep", true},
		{"точное совпадение", "/dashboard", "/dashboard", true},
		{"точное не ловит вложенное", "/dashboard", "/dashboard/stats", false},
		{"хвостовой слеш игнорируется", "/dashboard", "/dashboard/", true},
		{"префикс со звёздочкой", "/additem*", "/additem/new", true},
		{"префикс не ловит чужое", "/additem*", "/profile", false},
		{"суффикс со звёздочкой", "*/edit", "/items/42/edit", true},
		{"звёздочка в середине", "/items/*/edit", "/items/42/edit", true},
		{"query игнорируется, если не упомянут", "/dashboard", "/dashboard?tab=stats", true},
		{"query учитывается, если упомянут", "/dashboard?tab=stats", "/dashboard?tab=stats", true},
		{"пустой паттерн не ловит ничего", "", "/dashboard", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(
				t,
				tt.want,
				MatchPath(tt.pattern, tt.path),
			)
		})
	}
}

func TestPathFromURL(t *testing.T) {
	assert.Equal(
		t,
		"/additem?cat=1",
		PathFromURL("https://demo.local/additem?cat=1"),
	)

	assert.Equal(
		t,
		"/dashboard",
		PathFromURL("http://localhost:3000/dashboard"),
	)

	assert.Equal(
		t,
		"/",
		PathFromURL("https://demo.local"),
	)
}
