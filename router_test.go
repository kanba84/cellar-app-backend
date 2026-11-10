package main

import (
	"cellar-app/handler"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthDoesNotUseMCPAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mcpAuth := func(c *gin.Context) {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	router := SetupRouter(&handler.Handler{}, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), mcpAuth)

	for _, test := range []struct {
		path       string
		wantStatus int
	}{
		{path: "/health", wantStatus: http.StatusOK},
		{path: "/mcp", wantStatus: http.StatusUnauthorized},
	} {
		request := httptest.NewRequest(http.MethodGet, test.path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != test.wantStatus {
			t.Errorf("%s status = %d, want %d", test.path, response.Code, test.wantStatus)
		}
	}
}
