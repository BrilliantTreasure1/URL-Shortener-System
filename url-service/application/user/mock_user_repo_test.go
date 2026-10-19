package user

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	entities "url-shortener/entities/user"
	userRepo "url-shortener/repository/user"
)

type mockUserRepo struct {
	findByEmailUser      *entities.User
	findByEmailErr       error
	calledFindByEmail    bool
	callFindByEmailEmail string

	registerUser     *entities.User
	registerErr      error
	calledRegister   bool
	callRegisterUser *entities.User
}

func (m *mockUserRepo) FindByEmail(email string) (*entities.User, error) {
	m.calledFindByEmail = true
	m.callFindByEmailEmail = email
	return m.findByEmailUser, m.findByEmailErr
}

func (m *mockUserRepo) Register(user *entities.User) (*entities.User, error) {
	m.calledRegister = true
	m.callRegisterUser = user
	return m.registerUser, m.registerErr
}

var _ userRepo.UserRepository = (*mockUserRepo)(nil)

func intPtr(i int) *int {
	return &i
}

func mustUser(t *testing.T, id *int, username, email, phone, password string) *entities.User {
	t.Helper()

	user, err := entities.NewUser(id, username, email, phone, password)
	if err != nil {
		t.Fatalf("failed to build user: %v", err)
	}
	return user
}

func mustUserWithHash(t *testing.T, id *int, username, email, phone, plainPassword string) *entities.User {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return mustUser(t, id, username, email, phone, string(hash))
}
