package models

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Relationship struct {
	UserA  string
	UserB  string
	Status string
}

type RelationshipUserDTO struct {
	UserID            string     `json:"userID"`
	Username          string     `json:"username"`
	DisplayName       NullString `json:"displayName"`
	ProfilePictureURL NullString `json:"profilePictureURL"`
	Status            string     `json:"status"`
}

type BlockedUserDTO struct {
	UserID            string     `json:"userID"`
	Username          string     `json:"username"`
	DisplayName       NullString `json:"displayName"`
	ProfilePictureURL NullString `json:"profilePictureURL"`
}

type RelationshipModel struct {
	DB *pgxpool.Pool
}

func (m *RelationshipModel) FetchFriends(userID string) ([]RelationshipUserDTO, error) {
	query := `SELECT
				u.userID,
				u.username,
				u.displayName,
				u.profilePictureURL,
				r.status
				FROM relationships r
					JOIN users u ON r.userB = u.userID
					WHERE userA = $1;`

	// Fetch all friends
	rows, err := m.DB.Query(context.Background(), query, userID)
	if err != nil {
		return []RelationshipUserDTO{}, err
	}

	var friends []RelationshipUserDTO
	for rows.Next() {
		var r RelationshipUserDTO
		err := rows.Scan(&r.UserID, &r.Username, &r.DisplayName, &r.ProfilePictureURL, &r.Status)

		if err != nil {
			return []RelationshipUserDTO{}, err
		}

		friends = append(friends, r)
	}

	return friends, nil
}

func (m *RelationshipModel) FetchFriendRequests(userID string) ([]RelationshipUserDTO, error) {
	query := `SELECT
				CASE
					WHEN r.userB = $1 THEN u2.userID
					ELSE u.userID
				END AS userID,
				CASE
					WHEN r.userB = $1 THEN u2.username
					ELSE u.username
				END AS username,
				CASE
					WHEN r.userB = $1 THEN u2.displayName
					ELSE u.displayName
				END AS displayName,
				CASE
					WHEN r.userB = $1 THEN u2.profilePictureURL
					ELSE u.profilePictureURL
				END AS profilePictureURL,
				CASE
					WHEN r.status = 'accepted' THEN r.status
					WHEN r.userB = $1 THEN 'incoming'
					ELSE 'outgoing'
				END AS status
			FROM relationships r
					JOIN users u ON r.userB = u.userID
					JOIN users u2 ON r.userA = u2.userID
					WHERE (userA = $1 OR userB = $1) AND status = 'pending';`

	// Fetch all friend requests
	rows, err := m.DB.Query(context.Background(), query, userID)
	if err != nil {
		return []RelationshipUserDTO{}, err
	}

	var requests []RelationshipUserDTO
	for rows.Next() {
		var r RelationshipUserDTO
		err := rows.Scan(&r.UserID, &r.Username, &r.DisplayName, &r.ProfilePictureURL, &r.Status)

		if err != nil {
			return []RelationshipUserDTO{}, err
		}

		requests = append(requests, r)
	}

	return requests, nil
}

func (m *RelationshipModel) FetchBlockedUsers(userID string) ([]BlockedUserDTO, error) {
	// Fetch blocked users
	query := `SELECT u.userID, u.username, u.displayName, u.profilePictureURL FROM blockedUsers 
				JOIN users u on u.userID = blockedUsers.blockedUserID 
				WHERE blockedUsers.userFromID = $1;`
	rows, err := m.DB.Query(context.Background(), query, userID)
	if err != nil {
		return []BlockedUserDTO{}, err
	}

	blockedUsers := []BlockedUserDTO{}
	for rows.Next() {
		var r BlockedUserDTO
		err := rows.Scan(&r.UserID, &r.Username, &r.DisplayName, &r.ProfilePictureURL)
		if err != nil {
			return []BlockedUserDTO{}, err
		}

		blockedUsers = append(blockedUsers, r)
	}

	return blockedUsers, nil
}

func (m *RelationshipModel) SetRelationship(userA, userB string) (string, error) {
	if userA == userB {
		return "", ErrSameUser
	}

	// First check if the user can add more friends
	var count int
	query := "SELECT COUNT(*) FROM relationships WHERE userA = $1 AND status = 'accepted'"
	err := m.DB.QueryRow(context.Background(), query, userA).Scan(&count)
	if err != nil {
		return "", err
	}

	if count >= 2500 {
		return "", ErrMaxFriends
	}

	query = "SELECT * FROM relationships WHERE userA = $1 AND userB = $2"
	row := m.DB.QueryRow(context.Background(), query, userB, userA)

	status := "accepted" // Default status is accepted because we assume that the user is accepting the request

	var r Relationship
	err = row.Scan(&r.UserA, &r.UserB, &r.Status)
	if err != nil {
		// If a row exists between the recipient (userB) and the sender (userA), we can say that the recipient
		// already sent a request to the sender. In this case, we should accept the request. Otherwise, we should send a request (pending)
		if errors.Is(err, pgx.ErrNoRows) {
			status = "pending"
		} else {
			return "", err
		}
	}

	// Check if the recipient has blocked the sender
	if r.Status == "blocked" {
		return "", ErrRecipientHasBlockedUser
	}

	// Start a transaction
	tx, err := m.DB.Begin(context.Background())
	if err != nil {
		return "", err
	}

	// Insert the new relationship
	query = "INSERT INTO relationships (userA, userB, status) VALUES ($1, $2, $3) ON CONFLICT (userA, userB) DO UPDATE SET status = EXCLUDED.status"
	_, err = tx.Exec(context.Background(), query, userA, userB, status)
	if err != nil {
		tx.Rollback(context.Background())

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // Unique constraint violation because the relationship already exists
				return "REQUEST_ALREADY_SENT", nil
			}

			if pgErr.Code == "23503" { // Foreign key constraint violation due to user not found
				return "", ErrUserNotFound
			}

			if pgErr.Code == "23505" && status == "accepted" { // If a row exists betwen the sender and recipient and the status is accepted it means the relationship already exists
				return "", ErrRelationshipExists
			}
		}
		return "", err
	}

	// If the relationship was accepted, update the other side of the relationship
	// If it was pending we don't need to do anything since the other user will send a request or accept it
	if status == "accepted" {
		query = "UPDATE relationships  SET status = 'accepted' WHERE userA = $1 AND userB = $2"
		_, err = tx.Exec(context.Background(), query, userB, userA)
		if err != nil {
			tx.Rollback(context.Background())
			return "", err
		}

		tx.Commit(context.Background())

		return "REQUEST_ACCEPTED", nil
	}

	tx.Commit(context.Background())

	return "REQUEST_SENT", nil
}

func (m *RelationshipModel) DeleteRelationship(userA, userB string) error {
	if userA == userB {
		return ErrSameUser
	}

	query := "DELETE FROM relationships WHERE (userA = $1 AND userB = $2) OR (userA = $2 AND userB = $1)"
	_, err := m.DB.Exec(context.Background(), query, userA, userB)
	if err != nil {
		return err
	}

	return nil
}

func (m *RelationshipModel) BlockUser(userA, userB string) error {
	if userA == userB {
		return ErrSameUser
	}

	err := m.DeleteRelationship(userA, userB)
	if err != nil {
		return err
	}

	query := "INSERT INTO blockedUsers (userFromID, blockedUserID) VALUES ($1, $2)"
	_, err = m.DB.Exec(context.Background(), query, userA, userB)
	if err != nil {
		return err
	}

	return nil
}
