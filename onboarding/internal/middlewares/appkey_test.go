package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	appmw "github.com/ksendel-coder/onboarding_croquant/onboarding/internal/middlewares"
	platform "github.com/ksendel-coder/onboarding_croquant/platform/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// stubRuntime подменяет только AppByKey: остальные методы интерфейса приходят
// из встроенного nil-интерфейса и в этих тестах не вызываются.
type stubRuntime struct {
	domains.RuntimeDomainInterface
	app *domains.App
	err error
}

func (s stubRuntime) AppByKey(context.Context, string) (*domains.App, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.app, nil
}

// router повторяет порядок мидлварей из server.go: CORS отвечает на preflight
// до проверки ключа, ErrorHandler превращает c.Error в статус ответа.
func router(runtime domains.RuntimeDomainInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger := zerolog.Nop()

	r := gin.New()
	r.Use(platform.RuntimeCORS(), platform.ErrorHandler(&logger))

	v1 := r.Group("/v1")
	v1.Use(appmw.AppKeyAuth(runtime))
	v1.POST("/resolve", func(c *gin.Context) { c.Status(http.StatusOK) })

	return r
}

func do(t *testing.T, r *gin.Engine, method, key, origin string) int {
	t.Helper()

	req := httptest.NewRequest(method, "/v1/resolve", nil)
	if key != "" {
		req.Header.Set("X-App-Key", key)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec.Code
}

func TestAppKeyAuthOrigin(t *testing.T) {
	app := &domains.App{
		Id:             uuid.New(),
		Name:           "onboarding-demo",
		PublicKey:      "pk_test",
		AllowedOrigins: []string{"http://localhost:5173"},
	}
	r := router(stubRuntime{app: app})

	cases := []struct {
		name   string
		method string
		key    string
		origin string
		want   int
	}{
		{"свой origin пропускается", http.MethodPost, "pk_test", "http://localhost:5173", http.StatusOK},
		{"чужой origin отвергается", http.MethodPost, "pk_test", "http://evil.example.com", http.StatusForbidden},
		{"чужой порт отвергается", http.MethodPost, "pk_test", "http://localhost:3000", http.StatusForbidden},
		// Браузер завершающий слэш не шлёт, но список мог быть заполнен руками
		// до нормализации — сравнение обязано быть симметричным.
		{"слэш в origin не мешает", http.MethodPost, "pk_test", "http://localhost:5173/", http.StatusOK},
		// Запрос без Origin — это не браузер, а server-to-server вызов.
		// Он проходит намеренно, это не дыра.
		{"без Origin пропускается", http.MethodPost, "pk_test", "", http.StatusOK},
		{"без ключа 401", http.MethodPost, "", "http://localhost:5173", http.StatusUnauthorized},
		// Preflight обязан отвечать до проверки ключа: браузер шлёт OPTIONS
		// без X-App-Key, и 401 на него убил бы весь рантайм.
		{"preflight отвечает без ключа", http.MethodOptions, "", "http://localhost:5173", http.StatusNoContent},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := do(t, r, c.method, c.key, c.origin); got != c.want {
				t.Fatalf("статус %d, ожидался %d", got, c.want)
			}
		})
	}
}

func TestAppKeyAuthUnknownKey(t *testing.T) {
	r := router(stubRuntime{err: domains.ErrAppNotFound})

	if got := do(t, r, http.MethodPost, "pk_unknown", ""); got != http.StatusUnauthorized {
		t.Fatalf("статус %d, ожидался %d", got, http.StatusUnauthorized)
	}
}

func TestAppKeyAuthEmptyOriginListDeniesBrowser(t *testing.T) {
	app := &domains.App{
		Id:        uuid.New(),
		Name:      "onboarding-demo",
		PublicKey: "pk_test",
	}
	r := router(stubRuntime{app: app})

	if got := do(t, r, http.MethodPost, "pk_test", "http://localhost:5173"); got != http.StatusForbidden {
		t.Fatalf("статус %d, ожидался %d: пустой список обязан запрещать всё", got, http.StatusForbidden)
	}
}
