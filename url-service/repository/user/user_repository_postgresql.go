package user

import (
	"database/sql"

	"url-shortener/entities/user"
)

type UserRepositoryPostgresql struct {
	db *sql.DB
}

func NewUserRepositoryPostgresql(db *sql.DB) *UserRepositoryPostgresql {
	return &UserRepositoryPostgresql{
		db: db,
	}
}

func (r *UserRepositoryPostgresql) FindByEmail(
	email string,
) (*entities.User, error) {

	query := `
		SELECT
			id,
			username,
			email,
			phonenumber,
			password
		FROM users
		WHERE email = $1
	`

	var (
		id          int
		username    string
		userEmail   string
		phoneNumber string
		password    string
	)

	err := r.db.QueryRow(query, email).Scan(
		&id,
		&username,
		&userEmail,
		&phoneNumber,
		&password,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return entities.NewUser(
		&id,
		username,
		userEmail,
		phoneNumber,
		password,
	)
}

func (r *UserRepositoryPostgresql) Register(
	user *entities.User,
) (*entities.User, error) {

	query := `
		INSERT INTO users (
			username,
			email,
			phonenumber,
			password
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			username,
			email,
			phonenumber,
			password
	`

	var (
		id          int
		username    string
		email       string
		phoneNumber string
		password    string
	)

	err := r.db.QueryRow(
		query,
		user.Username(),
		user.Email(),
		user.Phonenumber(),
		user.Password(),
	).Scan(
		&id,
		&username,
		&email,
		&phoneNumber,
		&password,
	)

	if err != nil {
		return nil, err
	}

	return entities.NewUser(
		&id,
		username,
		email,
		phoneNumber,
		password,
	)
}