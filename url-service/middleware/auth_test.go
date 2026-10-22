package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"url-shortener/config"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const (
		wantUserID = 123
		wantEmail  = "john@example.com"
	)

	sign := func(claims jwt.MapClaims, secret string) string {
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(
			[]byte(secret),
		)
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}
		return token
	}

	validClaims := jwt.MapClaims{
		"user_id": wantUserID,
		"email":   wantEmail,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	validToken := sign(validClaims, config.GetJWTSecret())
	wrongSignatureToken := sign(validClaims, "wrong-secret")

	expiredClaims := jwt.MapClaims{
		"user_id": wantUserID,
		"email":   wantEmail,
		"exp":     time.Now().Add(-time.Hour).Unix(),
	}
	expiredToken := sign(expiredClaims, config.GetJWTSecret())

	tests := []struct {
		name          string
		authHeader    string
		setHeader     bool
		wantStatus    int
		wantUserIDSet bool
	}{
		{
			name:          "valid token",
			authHeader:    "Bearer " + validToken,
			setHeader:     true,
			wantStatus:    http.StatusOK,
			wantUserIDSet: true,
		},
		{
			name:          "missing authorization header",
			authHeader:    "",
			setHeader:     false,
			wantStatus:    http.StatusUnauthorized,
			wantUserIDSet: false,
		},
		{
			name:          "invalid format: bare bearer",
			authHeader:    "Bearer",
			setHeader:     true,
			wantStatus:    http.StatusUnauthorized,
			wantUserIDSet: false,
		},
		{
			name:          "invalid format: wrong scheme",
			authHeader:    "Token " + validToken,
			setHeader:     true,
			wantStatus:    http.StatusUnauthorized,
			wantUserIDSet: false,
		},
		{
			name:          "wrong signature",
			authHeader:    "Bearer " + wrongSignatureToken,
			setHeader:     true,
			wantStatus:    http.StatusUnauthorized,
			wantUserIDSet: false,
		},
		{
			name:          "expired token",
			authHeader:    "Bearer " + expiredToken,
			setHeader:     true,
			wantStatus:    http.StatusUnauthorized,
			wantUserIDSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedUserID interface{}
			var capturedUserIDSet bool

			router := gin.New()
			router.Use(AuthMiddleware())
			router.GET("/protected", func(c *gin.Context) {
				if value, exists := c.Get("user_id"); exists {
					capturedUserID = value
					capturedUserIDSet = true
				}
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.setHeader {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if capturedUserIDSet != tt.wantUserIDSet {
				t.Fatalf("expected user_id set = %v, got %v", tt.wantUserIDSet, capturedUserIDSet)
			}

			if tt.wantUserIDSet && capturedUserID != float64(wantUserID) {
				t.Errorf("expected user_id %v, got %v (type %T)",
					float64(wantUserID), capturedUserID, capturedUserID)
			}
		})
	}
}
