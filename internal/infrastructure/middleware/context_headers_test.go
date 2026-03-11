package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	reqctx "lending-hub-service/internal/shared/context"
)

func TestContextHeaders_MissingAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ContextHeaders(map[string]bool{}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	// Parse JSON response
	assert.Contains(t, w.Body.String(), "PAYIN3_INVALID_CONTEXT")
	assert.Contains(t, w.Body.String(), "x-platform")
	assert.Contains(t, w.Body.String(), "x-user-ip")
	_ = response
}

func TestContextHeaders_MissingOne(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ContextHeaders(map[string]bool{}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Missing x-user-ip
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("x-platform", "WEB")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "PAYIN3_INVALID_CONTEXT")
	assert.Contains(t, w.Body.String(), "x-user-ip")
}

func TestContextHeaders_AllPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ContextHeaders(map[string]bool{}))

	router.GET("/test", func(c *gin.Context) {
		rc, exists := c.Get("requestContext")
		assert.True(t, exists)
		reqCtx, ok := rc.(*reqctx.RequestContext)
		assert.True(t, ok)
		assert.Equal(t, "WEB", reqCtx.Platform)
		assert.Equal(t, "127.0.0.1", reqCtx.UserIP)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("x-platform", "WEB")
	req.Header.Set("x-user-ip", "127.0.0.1")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContextHeaders_SkipPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ContextHeaders(map[string]bool{
		"/health": true,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	// No headers provided
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJoinMissing(t *testing.T) {
	tests := []struct {
		name     string
		headers  []string
		expected string
	}{
		{"empty", []string{}, ""},
		{"single", []string{"x-platform"}, "x-platform"},
		{"two", []string{"x-platform", "x-user-ip"}, "x-platform, x-user-ip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinMissing(tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}
