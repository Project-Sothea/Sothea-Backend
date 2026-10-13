package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestErrorLoggerFileResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	file := filepath.Join(dir, "index.html")
	if err := os.WriteFile(file, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name       string
		path       string
		rangeValue string
		wantStatus int
		wantErrors int
	}{
		{"missing file", "/missing", "", http.StatusNotFound, 1},
		{"invalid range", "/file", "bytes=100-200", http.StatusRequestedRangeNotSatisfiable, 1},
		{"successful file", "/file", "", http.StatusOK, 0},
		{"partial file", "/file", "bytes=0-1", http.StatusPartialContent, 0},
		{"unknown API route", "/api/missing", "", http.StatusNotFound, 0},
		{"recorded error", "/failed", "", http.StatusInternalServerError, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			var recorded []*gin.Error
			router.Use(func(c *gin.Context) {
				c.Next()
				recorded = c.Errors
			}, ErrorLogger())
			router.StaticFile("/file", file)
			router.StaticFile("/missing", filepath.Join(dir, "missing.html"))
			originalErr := errors.New("original failure")
			router.GET("/failed", func(c *gin.Context) {
				c.Error(originalErr)
				c.Status(http.StatusInternalServerError)
			})
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			if test.rangeValue != "" {
				request.Header.Set("Range", test.rangeValue)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != test.wantStatus || len(recorded) != test.wantErrors {
				t.Fatalf("status=%d errors=%v; want status=%d errors=%d", response.Code, recorded, test.wantStatus, test.wantErrors)
			}
			if test.name == "recorded error" && recorded[0].Err != originalErr {
				t.Fatal("original error was replaced")
			}
		})
	}
}
