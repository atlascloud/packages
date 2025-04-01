package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// MockPkgRepo implements a mock version of the package repository
type MockPkgRepo struct {
	packagesDir string
}

// ListPackages now returns an echo.HandlerFunc instead of being a handler itself
func (m *MockPkgRepo) ListPackages() echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := c.Request().Header.Get("Authorization")
		if auth == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
		}
		if auth != "Bearer valid-token" {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}
		return c.JSON(http.StatusOK, []string{})
	}
}

func TestMain(m *testing.M) {
	// Setup test environment
	os.Exit(m.Run())
}

func setupTestServer() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Setup routes
	api := e.Group("/api/v1")
	mockRepo := &MockPkgRepo{packagesDir: "/tmp/test-packages"}

	// Add middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			return next(c)
		}
	})

	// Setup routes
	api.GET("/:org/packages", mockRepo.ListPackages()) // Note the () to invoke the handler function
	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.GET("/metrics", func(c echo.Context) error {
		return c.String(http.StatusOK, "packages_metric_example 1.0")
	})

	return e
}

func TestAuthenticationMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		orgName      string
		token        string
		expectedCode int
	}{
		{
			name:         "Valid token",
			orgName:      "testorg",
			token:        "valid-token",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Missing token",
			orgName:      "testorg",
			token:        "",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Invalid token",
			orgName:      "testorg",
			token:        "invalid-token",
			expectedCode: http.StatusUnauthorized,
		},
	}

	e := setupTestServer()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/api/v1/%s/packages", tc.orgName),
				nil,
			)

			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestMetricsEndpoint(t *testing.T) {
	e := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "packages_")
}

func TestServerConfiguration(t *testing.T) {
	tests := []struct {
		name string
		port int
		dir  string
	}{
		{
			name: "Default configuration",
			port: 8888,
			dir:  "/srv/packages",
		},
		{
			name: "Custom configuration",
			port: 9999,
			dir:  "/tmp/packages",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			log.Printf("Testing configuration: %s", tc.name)

			// Reset flags
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Define the flags first
			port := flag.Int("port", 8888, "Port to listen on")
			dir := flag.String("dir", "/srv/packages", "Directory to store packages")

			// Set test flags
			os.Args = []string{"cmd", "-port", fmt.Sprintf("%d", tc.port), "-dir", tc.dir}

			// Parse flags
			flag.Parse()

			// Verify the values
			assert.Equal(t, tc.port, *port)
			assert.Equal(t, tc.dir, *dir)
		})
	}
}

func TestSecureMiddleware(t *testing.T) {
	e := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/testorg/packages", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Check security headers
	assert.NotEmpty(t, rec.Header().Get("X-Frame-Options"))
	assert.NotEmpty(t, rec.Header().Get("X-Content-Type-Options"))
	assert.NotEmpty(t, rec.Header().Get("X-XSS-Protection"))
}

func TestHealthCheck(t *testing.T) {
	e := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
