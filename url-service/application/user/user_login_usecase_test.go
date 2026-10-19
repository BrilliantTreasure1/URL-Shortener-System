package user

import (
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"url-shortener/config"
	entities "url-shortener/entities/user"
)

func TestLoginUseCase(t *testing.T) {
	findErr := errors.New("repository failed")

	const (
		successEmail    = "john@example.com"
		successPassword = "supersecret"
		successUserID   = 7
	)

	tests := []struct {
		name string

		email    string
		password string

		findUser *entities.User
		findErr  error

		wantErr   string
		wantToken bool
	}{
		{
			name:     "success",
			email:    successEmail,
			password: successPassword,

			findUser: mustUserWithHash(
				t, intPtr(successUserID), "johndoe", successEmail, "09123456789", successPassword,
			),
			findErr: nil,

			wantErr:   "",
			wantToken: true,
		},
		{
			name:     "unknown email",
			email:    "unknown@example.com",
			password: successPassword,

			findUser: nil,
			findErr:  nil,

			wantErr:   "invalid email or password",
			wantToken: false,
		},
		{
			name:     "wrong password",
			email:    successEmail,
			password: "not-the-right-password",

			findUser: mustUserWithHash(
				t, intPtr(successUserID), "johndoe", successEmail, "09123456789", successPassword,
			),
			findErr: nil,

			wantErr:   "invalid email or password",
			wantToken: false,
		},
		{
			name:     "repository error",
			email:    successEmail,
			password: successPassword,

			findUser: nil,
			findErr:  findErr,

			wantErr:   "repository failed",
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				findByEmailUser: tt.findUser,
				findByEmailErr:  tt.findErr,
			}

			uc := NewLoginUseCase(repo)

			token, err := uc.Login(tt.email, tt.password)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				if token != "" {
					t.Errorf("expected an empty token on error, got %q", token)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if token == "" {
				t.Fatal("expected a token, got an empty string")
			}

			parsed, parseErr := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(config.GetJWTSecret()), nil
			})
			if parseErr != nil {
				t.Fatalf("returned token failed to parse/verify: %v", parseErr)
			}
			if !parsed.Valid {
				t.Fatal("expected returned token to be valid")
			}
			if parsed.Method.Alg() != "HS256" {
				t.Errorf("expected signing method HS256, got %q", parsed.Method.Alg())
			}

			claims, ok := parsed.Claims.(jwt.MapClaims)
			if !ok {
				t.Fatalf("expected jwt.MapClaims, got %T", parsed.Claims)
			}

			if got := claims["email"]; got != tt.email {
				t.Errorf("expected email claim %q, got %v", tt.email, got)
			}
			if got := claims["user_id"]; got != float64(successUserID) {
				t.Errorf("expected user_id claim %v, got %v (type %T)",
					float64(successUserID), got, got)
			}
		})
	}
}
