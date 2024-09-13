package handlers

import (
	"errors"
	"net/http"

	"github.com/CDavidSV/Iris-Chat-App-Backend/internal"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/models"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/websocket"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) GetRelationships(c *fiber.Ctx) error {
	rs, err := s.Relationships.FetchRelationships(c.Locals("userID").(string))
	if err != nil {
		return internal.ServerError(c, err, "Failed to fetch relationships")
	}

	return c.JSON(rs)
}

func (s *Server) CreateRelationship(c *fiber.Ctx) error {
	clientID := c.Locals("userID").(string)
	username := c.Params("username")

	if username == "" {
		return internal.ClientError(c, http.StatusBadRequest, internal.DefaultError{
			Code:    "MISSING_PARAM",
			Message: "Missing username in request",
		})
	}

	recipientUser, err := s.Users.FetchUserByUsername(username)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return internal.ClientError(c, http.StatusNotFound, internal.DefaultError{
				Code:    "USER_NOT_FOUND",
				Message: "User with that username not found",
			})
		}

		return internal.ServerError(c, err, "Failed to fetch user")
	}

	res, err := s.Relationships.SetRelationship(clientID, recipientUser.UserID)
	if err != nil {
		if errors.Is(err, models.ErrRelationshipExists) {
			return internal.ClientError(c, http.StatusConflict, internal.DefaultError{
				Code:    "RELATIONSHIP_EXISTS",
				Message: "Relationship already exists",
			})
		}

		if errors.Is(err, models.ErrSameUser) {
			return internal.ClientError(c, http.StatusConflict, internal.DefaultError{
				Code:    "SAME_USER",
				Message: "Cannot create relationship with yourself",
			})
		}

		return internal.ServerError(c, err, "Failed to create relationship")
	}

	// Send the notification if necessary
	if res == "REQUEST_ACCEPTED" {
		s.Websocket.UpdateFriendStatus(recipientUser.UserID, websocket.FriendStatus{
			Type:   "ACCEPTED",
			UserID: clientID,
		})
	} else if res == "REQUEST_SENT" {
		s.Websocket.UpdateFriendStatus(recipientUser.UserID, websocket.FriendStatus{
			Type:   "REQUEST",
			UserID: clientID,
		})
	}

	// TODO: Create new channel if one doesn't exist

	return nil
}

func (s *Server) DelRelationship(c *fiber.Ctx) error {
	clientID := c.Locals("userID").(string)
	recipientID := c.Params("userID")

	err := s.Relationships.DeleteRelationship(clientID, recipientID)
	if err != nil {
		if errors.Is(err, models.ErrSameUser) {
			return internal.ClientError(c, http.StatusConflict, internal.DefaultError{
				Code:    "SAME_USER",
				Message: "You can't have a relationship with yourself :/",
			})
		}
	}

	// Send the notification to remove the friend
	s.Websocket.UpdateFriendStatus(recipientID, websocket.FriendStatus{
		Type:   "REMOVED",
		UserID: clientID,
	})

	return nil
}
