package models

import (
	"context"
	"errors"
	"time"

	"github.com/CDavidSV/Iris-Chat-App-Backend/internal"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserID            string
	Username          string
	Email             string
	Password          string
	JoinedAt          time.Time
	UpdatedAt         time.Time
	CustomStatus      NullString
	ProfilePictureURL NullString
}

type UserDTO struct {
	UserID            string
	Username          string
	Email             string
	JoinedAt          time.Time
	UpdatedAt         time.Time
	CustomStatus      NullString
	ProfilePictureURL NullString
}

type UserModel struct {
	DB *pgxpool.Pool
}

func (m *UserModel) InsertUser(username, email, password string) (string, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}

	userID := internal.GenerateID()
	query := "INSERT INTO users (userID, username, email, password) VALUES ($1, $2, $3, $4)"

	_, err = m.DB.Exec(context.Background(), query, userID, username, email, hashedPassword)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == "23505" && pgError.ConstraintName == "users_email_key" {
				return "", ErrDuplicateEmail
			}

			if pgError.Code == "23505" && pgError.ConstraintName == "users_username_key" {
				return "", ErrDuplicateUsername
			}

			return "", err
		}

		return "", err
	}

	return userID, nil
}

func (m *UserModel) Authenticate(email, password string) (string, error) {
	query := "SELECT userID, password FROM users WHERE email = $1"

	var id string
	var hashedPassword string

	row := m.DB.QueryRow(context.Background(), query, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", ErrInvalidCredentials
		} else {
			return "", err
		}
	}

	return id, nil
}

func (m *UserModel) FetchUser(userID string) (UserDTO, error) {
	query := "SELECT userID, username, email, joinedAt, customStatus, profilePictureURL, updatedAt FROM users WHERE userID = $1"

	user := UserDTO{}
	row := m.DB.QueryRow(context.Background(), query, userID)
	err := row.Scan(&user.UserID, &user.Username, &user.Email, &user.JoinedAt, &user.CustomStatus, &user.ProfilePictureURL, &user.UpdatedAt)
	if err != nil {
		return user, err
	}

	return user, nil
}
