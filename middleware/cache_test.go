package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// 回归测试：升级服务端后浏览器必须重新请求 index.html 与 /api/*，
// 否则会一直加载缓存里的旧前端，出现页脚版本不一致与「新版本可用」反复弹窗。
func TestCacheControl(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		path string
		want string
	}{
		{"/", "no-cache, no-store, must-revalidate"},
		{"/index.html", "no-cache, no-store, must-revalidate"},
		{"/api/status", "no-cache, no-store, must-revalidate"},
		{"/message/demo", "no-cache, no-store, must-revalidate"},
		{"/static/js/main.abc123.js", "public, max-age=604800, immutable"},
		{"/public/static/logo.png", "no-cache, no-store, must-revalidate"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, tc.path, nil)
			Cache()(c)
			got := w.Header().Get("Cache-Control")
			if got != tc.want {
				t.Fatalf("Cache-Control = %q, want %q", got, tc.want)
			}
			if tc.want == "no-cache, no-store, must-revalidate" {
				if w.Header().Get("Pragma") != "no-cache" {
					t.Fatal("dynamic response must set Pragma: no-cache")
				}
			}
		})
	}
}
