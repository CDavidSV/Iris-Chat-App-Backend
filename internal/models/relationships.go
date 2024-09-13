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

type RelationshipModel struct {
	DB *pgxpool.Pool
}

func (m *RelationshipModel) FetchRelationships(userID string) ([]Relationship, error) {
	query := "SELECT * FROM relationships WHERE userA = $1 OR (userB = $1 AND status = 'pending')"

	rows, err := m.DB.Query(context.Background(), query, userID)
	if err != nil {
		return []Relationship{}, err
	}

	var relationships []Relationship
	for rows.Next() {
		var r Relationship
		err := rows.Scan(&r.UserA, &r.UserB, &r.Status)

		if err != nil {
			return []Relationship{}, err
		}

		relationships = append(relationships, r)
	}

	return relationships, nil
}

func (m *RelationshipModel) SetRelationship(userA, userB string) (bool, error) {
	if userA == userB {
		return false, ErrSameUser
	}

	query := "SELECT * FROM relationships WHERE (userA = $1 AND userB = $2) OR (userA = $2 AND userB = $1)"
	rows, err := m.DB.Query(context.Background(), query, userA, userB)
	if err != nil {
		return false, err
	}

	var friendship []Relationship
	var r Relationship
	for rows.Next() {
		err := rows.Scan(&r.UserA, &r.UserB, &r.Status)
		if err != nil {
			return false, err
		}

		friendship = append(friendship, r)
	}

	switch len(friendship) {
	case 1:
		// Update the existing relationship
		if friendship[0].UserA == userA {
			return true, nil
		} else { // If the relationship is pending, accept it
			// Start a transaction
			tx, err := m.DB.Begin(context.Background())
			if err != nil {
				return false, err
			}

			// Insert the new relationship
			query := "INSERT INTO relationships (userA, userB, status) VALUES ($1, $2, 'accepted')"
			res, err := tx.Exec(context.Background(), query, userA, userB)
			if err != nil || res.RowsAffected() == 0 {
				tx.Rollback(context.Background())
				return false, err
			}

			// Update the existing one if exists, else insert it again
			query = "INSERT INTO relationships (userA, userB, status) VALUES ($1, $2, 'accepted') ON CONFLICT (userA, userB) DO UPDATE SET status = EXCLUDED.status"
			res, err = tx.Exec(context.Background(), query, userB, userA)
			if err != nil || res.RowsAffected() == 0 {
				tx.Rollback(context.Background())
				return false, err
			}

			tx.Commit(context.Background())

			// TODO: Send a notification to the user that the friend request has been accepted
			// TODO: Create a new channel for the user

			return true, nil
		}
	case 2: // If the query yields 2 results, the relationship already exists
		return false, ErrRelationshipExists
	default:
		// Insert the new relationship
		query = "INSERT INTO relationships (userA, userB, status) VALUES ($1, $2, 'pending')"
		res, err := m.DB.Exec(context.Background(), query, userA, userB)
		if err != nil || res.RowsAffected() == 0 {
			return false, err
		}

		// TODO: Send a notification to the user

		return true, nil
	}
}

func (m *RelationshipModel) DeleteRelationship(userA, userB string) (bool, error) {
	if userA == userB {
		return false, ErrSameUser
	}

	query := "DELETE FROM friends WHERE LEAST(userA, userB) = LEAST($1, $2) AND GREATEST(userA, userB) = GREATEST($1, $2)"
	res, err := m.DB.Exec(context.Background(), query, userA, userB)
	if err != nil || res.RowsAffected() == 0 {
		return false, err
	}

	return true, nil
}
