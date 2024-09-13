package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Relationship struct {
	UserA  string
	UserB  string
	Status string
}

type RelationshipUserDTO struct {
	UserID            string `json:"userID"`
	Username          string `json:"username"`
	DisplayName       string `json:"displayName"`
	ProfilePictureURL string `json:"profilePictureURL"`
}

type RelationshipModel struct {
	DB *pgxpool.Pool
}

func (m *RelationshipModel) FetchRelationships(userID string) ([]RelationshipUserDTO, error) {
	query := `SELECT u.userID, u.username, u.displayName, u.profilePictureURL FROM relationships r 
				JOIN users u ON r.userB = u.userID
				WHERE userA = $1 OR (userB = $1 AND status = 'pending')`

	rows, err := m.DB.Query(context.Background(), query, userID)
	if err != nil {
		return []RelationshipUserDTO{}, err
	}

	var relationships []RelationshipUserDTO
	for rows.Next() {
		var r RelationshipUserDTO
		err := rows.Scan(&r.UserID, &r.Username, &r.DisplayName, &r.ProfilePictureURL)

		if err != nil {
			return []RelationshipUserDTO{}, err
		}

		relationships = append(relationships, r)
	}

	return relationships, nil
}

func (m *RelationshipModel) SetRelationship(userA, userB string) (string, error) {
	if userA == userB {
		return "", ErrSameUser
	}

	query := "SELECT * FROM relationships WHERE (userA = $1 AND userB = $2) OR (userA = $2 AND userB = $1)"
	rows, err := m.DB.Query(context.Background(), query, userA, userB)
	if err != nil {
		return "", err
	}

	var friendship []Relationship
	var r Relationship
	for rows.Next() {
		err := rows.Scan(&r.UserA, &r.UserB, &r.Status)
		if err != nil {
			return "", err
		}

		friendship = append(friendship, r)
	}

	switch len(friendship) {
	case 1:
		// Update the existing relationship
		if friendship[0].UserA == userA {
			return "REQUEST_ALREADY_SENT", nil
		} else { // If the relationship is pending, accept it
			// Start a transaction
			tx, err := m.DB.Begin(context.Background())
			if err != nil {
				return "", err
			}

			// Insert the new relationship
			query := "INSERT INTO relationships (userA, userB, status) VALUES ($1, $2, 'accepted')"
			res, err := tx.Exec(context.Background(), query, userA, userB)
			if err != nil || res.RowsAffected() == 0 {
				tx.Rollback(context.Background())
				return "", err
			}

			// Update the existing one if exists, else insert it again
			query = "INSERT INTO relationships (userA, userB, status) VALUES ($1, $2, 'accepted') ON CONFLICT (userA, userB) DO UPDATE SET status = EXCLUDED.status"
			res, err = tx.Exec(context.Background(), query, userB, userA)
			if err != nil || res.RowsAffected() == 0 {
				tx.Rollback(context.Background())
				return "", err
			}

			tx.Commit(context.Background())

			return "REQUEST_ACCEPTED", nil
		}
	case 2: // If the query yields 2 results, the relationship already exists
		return "", ErrRelationshipExists
	default:
		// Insert the new relationship
		query = "INSERT INTO relationships (userA, userB, status) VALUES ($1, $2, 'pending')"
		res, err := m.DB.Exec(context.Background(), query, userA, userB)
		if err != nil || res.RowsAffected() == 0 {
			return "", err
		}

		return "REQUEST_SENT", nil
	}
}

func (m *RelationshipModel) DeleteRelationship(userA, userB string) error {
	if userA == userB {
		return ErrSameUser
	}

	query := "DELETE FROM friends WHERE (userA = $1 AND userB = $2) OR (userA = $2 AND userB = $1)"
	res, err := m.DB.Exec(context.Background(), query, userA, userB)
	if err != nil || res.RowsAffected() == 0 {
		return err
	}

	return nil
}
