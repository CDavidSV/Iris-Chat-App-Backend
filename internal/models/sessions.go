package models

import (
	"context"
	"errors"
	"time"

	"github.com/CDavidSV/Iris-Chat-App-Backend/internal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Session struct {
	SessionID    string
	UserID       string
	RefreshToken string
	IpAddress    string
	Device       string
	OS           string
	ExpiresAt    time.Time
	UpdatedAt    time.Time
}

type NewSession struct {
	UserID       string
	SessionID    string
	RefreshToken string
}

type SessionsModel struct {
	DB *pgxpool.Pool
}

var RefreshTokenExpirationDelta time.Duration = time.Hour * 24 * 30

func generateRefreshToken() (string, time.Time) {
	return uuid.New().String(), time.Now().Add(RefreshTokenExpirationDelta)
}

func (m *SessionsModel) NewSession(userID, device, os, ip string) (NewSession, error) {
	sessionID := internal.GenerateID()

	newSession := NewSession{
		UserID:    userID,
		SessionID: sessionID,
	}

	refreshToken, expiresAt := generateRefreshToken()
	newSession.RefreshToken = refreshToken

	// Insert session in db
	query := "INSERT INTO sessions (sessionID, userID, refreshToken, ipAddress, device, os, expiresAt) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	_, err := m.DB.Exec(context.Background(), query, sessionID, userID, refreshToken, ip, device, os, expiresAt)
	if err != nil {
		return NewSession{}, err
	}

	return newSession, nil
}

func (m *SessionsModel) RevalidateSession(sessionID string, refreshToken string) (NewSession, error) {
	query := "SELECT * FROM sessions WHERE sessionID = $1"

	row := m.DB.QueryRow(context.Background(), query, sessionID)

	session := Session{}
	err := row.Scan(&session.SessionID, &session.UserID, &session.RefreshToken, &session.IpAddress, &session.Device, &session.OS, &session.ExpiresAt, &session.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return NewSession{}, ErrSessionExpired
		}

		return NewSession{}, err
	}

	if session.ExpiresAt.Unix() <= time.Now().Unix() {
		return NewSession{}, ErrSessionExpired
	}

	// Verify if the refresh token matches the one store of the sessions table
	if refreshToken != session.RefreshToken {
		// Clear all session for the user as a security measure
		m.DeleteAllSessions(session.UserID)
		return NewSession{}, ErrInvalidSession
	}

	// Generate new refresh token and save
	newRefreshToken, expiresAt := generateRefreshToken()
	query = "UPDATE sessions SET refreshToken = $1, expiresAt = $2 WHERE sessionID = $3"
	_, err = m.DB.Exec(context.Background(), query, newRefreshToken, expiresAt, sessionID)
	if err != nil {
		return NewSession{}, err
	}

	return NewSession{
		UserID:       session.UserID,
		SessionID:    session.SessionID,
		RefreshToken: newRefreshToken,
	}, nil
}

func (m *SessionsModel) DeleteSession(sessionID, refreshToken string) error {
	query := "DELETE FROM sessions WHERE sessionID = $1 AND refreshToken = $2"

	res, err := m.DB.Exec(context.Background(), query, sessionID, refreshToken)
	if err != nil {
		return err
	}

	if res.RowsAffected() < 1 {
		return ErrNoSessionsFound
	}

	return nil
}

func (m *SessionsModel) DeleteAllSessions(userID string) error {
	query := "DELETE FROM sessions WHERE userID = $1"

	_, err := m.DB.Exec(context.Background(), query, userID)
	if err != nil {
		return err
	}

	return nil
}
