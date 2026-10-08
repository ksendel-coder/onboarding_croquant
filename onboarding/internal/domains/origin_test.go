package domains

import (
	"errors"
	"testing"
)

func TestNormalizeOrigin(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"обычный origin", "http://localhost:5173", "http://localhost:5173"},
		{"завершающий слэш", "http://localhost:5173/", "http://localhost:5173"},
		{"пробелы по краям", "  http://localhost:5173  ", "http://localhost:5173"},
		{"верхний регистр хоста", "HTTP://LocalHost:5173", "http://localhost:5173"},
		{"ip с портом", "http://72.56.92.137:5173", "http://72.56.92.137:5173"},
		{"https без порта", "https://cluer.example.com", "https://cluer.example.com"},
		{"путь не является частью origin", "http://localhost:5173/admin", ""},
		{"query не является частью origin", "http://localhost:5173?a=1", ""},
		{"без схемы", "localhost:5173", ""},
		{"чужая схема", "ftp://localhost:5173", ""},
		{"пустая строка", "", ""},
		{"только пробелы", "   ", ""},
		{"звёздочка не origin", "*", ""},
		{"мусор", "не-origin", ""},
		{"креды в url", "http://user:pass@localhost:5173", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := NormalizeOrigin(c.in); got != c.want {
				t.Fatalf("NormalizeOrigin(%q) = %q, ожидалось %q", c.in, got, c.want)
			}
		})
	}
}

func TestValidateOrigins(t *testing.T) {
	t.Run("пустой список отвергается", func(t *testing.T) {
		if _, err := ValidateOrigins(nil); !errors.Is(err, ErrOriginRequired) {
			t.Fatalf("ожидалась ErrOriginRequired, получено %v", err)
		}
		if _, err := ValidateOrigins([]string{}); !errors.Is(err, ErrOriginRequired) {
			t.Fatalf("ожидалась ErrOriginRequired, получено %v", err)
		}
	})

	t.Run("нормализует и дедуплицирует", func(t *testing.T) {
		got, err := ValidateOrigins([]string{
			"http://localhost:5173/",
			"HTTP://localhost:5173",
			"https://cluer.example.com",
		})
		if err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}
		want := []string{"http://localhost:5173", "https://cluer.example.com"}
		if len(got) != len(want) {
			t.Fatalf("получено %v, ожидалось %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("получено %v, ожидалось %v", got, want)
			}
		}
	})

	t.Run("мусор отвергается целиком", func(t *testing.T) {
		if _, err := ValidateOrigins([]string{"http://localhost:5173", "не-origin"}); !errors.Is(err, ErrInvalidOrigin) {
			t.Fatalf("ожидалась ErrInvalidOrigin, получено %v", err)
		}
	})

	// Вайлдкард сохраняется как есть: OriginAllowed на него полагается, и без
	// этой ветки его нельзя было бы записать через API — только руками в базу.
	t.Run("вайлдкард сохраняется", func(t *testing.T) {
		got, err := ValidateOrigins([]string{"*"})
		if err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}
		if len(got) != 1 || got[0] != "*" {
			t.Fatalf("получено %v, ожидалось [*]", got)
		}
	})
}

func TestOriginAllowed(t *testing.T) {
	list := []string{"http://localhost:5173", "https://cluer.example.com"}

	cases := []struct {
		name    string
		allowed []string
		origin  string
		want    bool
	}{
		{"точное совпадение", list, "http://localhost:5173", true},
		{"второй в списке", list, "https://cluer.example.com", true},
		{"чужой порт", list, "http://localhost:3000", false},
		{"чужая схема", list, "https://localhost:5173", false},
		{"чужой хост", list, "http://evil.example.com", false},
		{"пустой список запрещает всё", nil, "http://localhost:5173", false},
		{"вайлдкард разрешает всё", []string{"*"}, "http://evil.example.com", true},
		// Браузер слэш не шлёт, но список мог быть заполнен до нормализации.
		{"слэш во входящем origin", list, "http://localhost:5173/", true},
		{"регистр во входящем origin", list, "HTTP://LOCALHOST:5173", true},
		{"мусор во входящем origin", list, "не-origin", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := OriginAllowed(c.allowed, c.origin); got != c.want {
				t.Fatalf("OriginAllowed(%v, %q) = %v, ожидалось %v", c.allowed, c.origin, got, c.want)
			}
		})
	}
}
