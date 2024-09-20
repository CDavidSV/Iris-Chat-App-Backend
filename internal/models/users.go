package models

import (
	"context"
	"errors"
	"time"

	"github.com/CDavidSV/Iris-Chat-App-Backend/internal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserID            string
	Username          string
	DisplayName       NullString
	Email             string
	Password          string
	JoinedAt          time.Time
	UpdatedAt         time.Time
	Verified          bool
	CustomStatus      NullString
	ProfilePictureURL NullString
	Bio               NullString
}

type UserDTO struct {
	UserID            string     `json:"userID"`
	Username          string     `json:"username"`
	DisplayName       NullString `json:"displayName"`
	Email             string     `json:"email"`
	JoinedAt          time.Time  `json:"joinedAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	CustomStatus      NullString `json:"customStatus"`
	ProfilePictureURL NullString `json:"profilePictureURL"`
	Bio               NullString `json:"bio"`
}

type PublicUserDTO struct {
	UserID            string     `json:"userID"`
	Username          string     `json:"username"`
	DisplayName       NullString `json:"displayName"`
	JoinedAt          time.Time  `json:"joinedAt"`
	CustomStatus      NullString `json:"customStatus"`
	Bio               NullString `json:"bio"`
	ProfilePictureURL NullString `json:"profilePictureURL"`
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
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrInvalidCredentials
		}

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
	query := "SELECT userID, username, email, joinedAt, customStatus, profilePictureURL, updatedAt, displayName, bio FROM users WHERE userID = $1"

	user := UserDTO{}
	row := m.DB.QueryRow(context.Background(), query, userID)
	err := row.Scan(&user.UserID, &user.Username, &user.Email, &user.JoinedAt, &user.CustomStatus, &user.ProfilePictureURL, &user.UpdatedAt, &user.DisplayName, &user.Bio)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, ErrUserNotFound
		}
		return user, err
	}

	return user, nil
}

func (m *UserModel) FetchUsersByUsername(username string) ([]PublicUserDTO, error) {
	query := "SELECT userID, username, joinedAt, customStatus, profilePictureURL, displayName, bio FROM users WHERE username LIKE $1"

	users := []PublicUserDTO{}
	rows, err := m.DB.Query(context.Background(), query, username+"%")
	if err != nil {
		return users, err
	}

	for rows.Next() {
		var user PublicUserDTO
		err = rows.Scan(&user.UserID, &user.Username, &user.JoinedAt, &user.CustomStatus, &user.ProfilePictureURL, &user.DisplayName, &user.Bio)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return users, ErrUserNotFound
			}
			return users, err
		}

		users = append(users, user)
	}

	if len(users) == 0 {
		return users, ErrUserNotFound
	}

	return users, nil
}

func (m *UserModel) UpdateProfileInfo(userID, displayName, bio string) (UserDTO, error) {
	query := "UPDATE users SET displayName = $1, bio = $2, updatedAt = NOW() WHERE userID = $4 RETURNING userID, username, email, joinedAt, customStatus, profilePictureURL, updatedAt, displayName, bio"

	row := m.DB.QueryRow(context.Background(), query, displayName, bio, userID)

	user := UserDTO{}
	err := row.Scan(&user.UserID, &user.Username, &user.Email, &user.JoinedAt, &user.CustomStatus, &user.ProfilePictureURL, &user.UpdatedAt, &user.DisplayName, &user.Bio)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, ErrUserNotFound
		}

		return user, err
	}

	return user, nil
}

func (m *UserModel) UpdatePassword(userID, oldPassword, newPassword string) error {
	query := "SELECT password FROM users WHERE userID = $1"

	var hashedOldPassword string
	row := m.DB.QueryRow(context.Background(), query, userID)
	err := row.Scan(&hashedOldPassword)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedOldPassword), []byte(oldPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}

		return err
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	query = "UPDATE users SET password = $1 WHERE userID = $2"
	_, err = m.DB.Exec(context.Background(), query, newHashedPassword, userID)
	if err != nil {
		return err
	}

	return nil
}
