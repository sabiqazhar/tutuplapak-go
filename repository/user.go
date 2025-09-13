package repository

import (
	"database/sql"
	"errors"
	"tutuplapak-go/provider"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository() *UserRepository {
	return &UserRepository{DB: provider}
}

func (r *UserRepository) CreateUser(email, password string) (*User, error) {
	query := `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id`
	user := &User{Email: email, Password: password}

	err := r.DB.QueryRow(query, email, password).Scan(&user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) FindByEmail(email string) (*User, error) {
	query := `SELECT id, email, password FROM users WHERE email = $1`
	user := &User{}

	err := r.DB.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}

	return user, nil
}