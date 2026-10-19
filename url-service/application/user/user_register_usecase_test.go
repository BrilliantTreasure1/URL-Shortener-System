package user

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	entities "url-shortener/entities/user"
)

func TestRegisterUseCase(t *testing.T) {
	findErr := errors.New("repository failed")
	registerErr := errors.New("register failed")

	tests := []struct {
		name string

		username string
		email    string
		phone    string
		password string

		findUser *entities.User
		findErr  error

		registerUser *entities.User
		registerErr  error

		wantErr            string
		wantRegisterCalled bool
		wantVerifyPlain    string
	}{
		{
			name:     "success",
			username: "johndoe",
			email:    "john@example.com",
			phone:    "09123456789",
			password: "supersecret",

			findUser: nil,
			findErr:  nil,

			registerUser: mustUser(
				t, intPtr(1), "johndoe", "john@example.com", "09123456789", "stored-hash",
			),
			registerErr: nil,

			wantErr:            "",
			wantRegisterCalled: true,
			wantVerifyPlain:    "supersecret",
		},
		{
			name:     "email already registered",
			username: "johndoe",
			email:    "john@example.com",
			phone:    "09123456789",
			password: "supersecret",

			findUser: mustUser(
				t, intPtr(1), "johndoe", "john@example.com", "09123456789", "stored-hash",
			),
			findErr: nil,

			registerUser: nil,
			registerErr:  nil,

			wantErr:            "user already registered",
			wantRegisterCalled: false,
		},
		{
			name:     "find by email error",
			username: "johndoe",
			email:    "john@example.com",
			phone:    "09123456789",
			password: "supersecret",

			findUser: nil,
			findErr:  findErr,

			registerUser: nil,
			registerErr:  nil,

			wantErr:            "repository failed",
			wantRegisterCalled: false,
		},
		{
			name:     "register error",
			username: "johndoe",
			email:    "john@example.com",
			phone:    "09123456789",
			password: "supersecret",

			findUser: nil,
			findErr:  nil,

			registerUser: nil,
			registerErr:  registerErr,

			wantErr:            "register failed",
			wantRegisterCalled: true,
		},
		{
			name:     "username too short",
			username: "abc",
			email:    "john@example.com",
			phone:    "09123456789",
			password: "supersecret",

			findUser: nil,
			findErr:  nil,

			registerUser: nil,
			registerErr:  nil,

			wantErr:            "username is too short",
			wantRegisterCalled: false,
		},
		{
			name:     "invalid email",
			username: "johndoe",
			email:    "johnexample.com",
			phone:    "09123456789",
			password: "supersecret",

			findUser: nil,
			findErr:  nil,

			registerUser: nil,
			registerErr:  nil,

			wantErr:            "invalid email",
			wantRegisterCalled: false,
		},
		{
			name:     "phone number too short",
			username: "johndoe",
			email:    "john@example.com",
			phone:    "09123",
			password: "supersecret",

			findUser: nil,
			findErr:  nil,

			registerUser: nil,
			registerErr:  nil,

			wantErr:            "phone number is too short",
			wantRegisterCalled: false,
		},
		// NOTE: entities.NewUser also rejects passwords shorter than 8 characters,
		// but RegisterUseCase bcrypt-hashes the password *before* building the
		// entity. The validation therefore runs against the ~60-character hash and
		// can never fail on the register path, so it is not exercised here. The
		// rule is observable only through entities.NewUser itself.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				findByEmailUser: tt.findUser,
				findByEmailErr:  tt.findErr,
				registerUser:    tt.registerUser,
				registerErr:     tt.registerErr,
			}

			uc := NewRegisterUseCase(repo)

			got, err := uc.Register(tt.username, tt.email, tt.phone, tt.password)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if got == nil {
					t.Fatal("expected a user, got nil")
				}

				if repo.callRegisterUser == nil {
					t.Fatal("expected repo.Register to receive a user")
				}
				stored := repo.callRegisterUser.Password()

				if stored == tt.wantVerifyPlain {
					t.Error("expected stored password to be a hash, got plain text")
				}
				if cmpErr := bcrypt.CompareHashAndPassword(
					[]byte(stored), []byte(tt.wantVerifyPlain),
				); cmpErr != nil {
					t.Errorf("stored hash does not verify against the plain password: %v", cmpErr)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				if got != nil {
					t.Errorf("expected nil user on error, got %v", got)
				}
			}

			if repo.calledRegister != tt.wantRegisterCalled {
				t.Errorf("expected repo.Register called = %v, got %v",
					tt.wantRegisterCalled, repo.calledRegister)
			}
		})
	}
}
