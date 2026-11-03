package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoginRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/login", LoginRateLimit(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	hit := func(ip string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/login", nil)
		req.RemoteAddr = ip + ":1234"
		r.ServeHTTP(w, req)
		return w.Code
	}

	// 10x rapat dari 1 IP: lolos (burst).
	for i := 0; i < 10; i++ {
		if code := hit("1.2.3.4"); code != http.StatusOK {
			t.Fatalf("attempt %d: got %d, want 200", i+1, code)
		}
	}
	// ke-11 langsung: 429 + Retry-After.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("attempt 11: got %d, want 429", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("missing Retry-After header")
	}
	// IP lain tidak ikut terblokir.
	if code := hit("5.6.7.8"); code != http.StatusOK {
		t.Fatalf("other IP: got %d, want 200", code)
	}
}
