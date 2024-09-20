package handlers

import (
	"errors"
	"net/http"

	"github.com/CDavidSV/Iris-Chat-App-Backend/internal"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/models"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func (s *Server) GetFriendships(c *fiber.Ctx) error {
	rs, err := s.Relationships.FetchFriends(c.Locals("userID").(string))
	if err != nil {
		return internal.ServerError(c, err, "Failed to fetch relationships")
	}

	if rs == nil {
		return c.JSON([]models.RelationshipUserDTO{})
	}

	return c.JSON(rs)
}

func (s *Server) GetBlockedUsers(c *fiber.Ctx) error {
	blocked, err := s.Relationships.FetchBlockedUsers(c.Locals("userID").(string))
	if err != nil {
		return internal.ServerError(c, err, "Failed to fetch blocked users")
	}

	if blocked == nil {
		return c.JSON([]models.BlockedUserDTO{})
	}

	return c.JSON(blocked)
}

func (s *Server) GetFriendRequests(c *fiber.Ctx) error {
	rs, err := s.Relationships.FetchFriendRequests(c.Locals("userID").(string))
	if err != nil {
		return internal.ServerError(c, err, "Failed to fetch friend requests")
	}

	if rs == nil {
		return c.JSON([]models.RelationshipUserDTO{})
	}

	return c.JSON(rs)
}

func (s *Server) CreateRelationship(c *fiber.Ctx) error {
	clientID := c.Locals("userID").(string)
	recipientID := c.Params("userID")

	res, err := s.Relationships.SetRelationship(clientID, recipientID)
	if err != nil {
		// Check if the error is a constraint error
		if errors.Is(err, models.ErrUserNotFound) {
			return internal.ClientError(c, http.StatusNotFound, internal.DefaultError{
				Code:    "USER_NOT_FOUND",
				Message: "User with the given ID does not exist",
			})
		}

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

	go func() {
		// Fetch the user data
		user, err := s.Users.FetchUser(clientID)
		if err != nil {
			log.Error("Failed to fetch user data for friend request: ", err)
			return
		}

		message := websocket.FriendStatus{
			Type:              "ACCEPTED",
			UserID:            clientID,
			Username:          user.Username,
			ProfilePictureURL: string(user.ProfilePictureURL),
		}

		// Send the notification if necessary
		if res == "REQUEST_ACCEPTED" {
			message.Type = "ACCEPTED"
			s.Websocket.UpdateFriendStatus(recipientID, message)
		} else if res == "REQUEST_SENT" {
			message.Type = "REQUEST"
			s.Websocket.UpdateFriendStatus(recipientID, message)
		}
	}()

	// TODO: Create new channel if one doesn't exist

	return c.JSON(map[string]string{
		"message": "Friend request sent",
	})
}

func (s *Server) DelRelationship(c *fiber.Ctx) error {
	clientID := c.Locals("userID").(string)
	recipientID := c.Params("userID")

	err := s.Relationships.DeleteRelationship(clientID, recipientID)
	if err != nil {
		if errors.Is(err, models.ErrSameUser) {
			return internal.ClientError(c, http.StatusConflict, internal.DefaultError{
				Code:    "SAME_USER",
				Message: "You can't unfriend yourself :/",
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

func (s *Server) BlockUser(c *fiber.Ctx) error {
	clientID := c.Locals("userID").(string)
	recipientID := c.Params("userID")

	err := s.Relationships.BlockUser(clientID, recipientID)
	if err != nil {
		if errors.Is(err, models.ErrSameUser) {
			return internal.ClientError(c, http.StatusConflict, internal.DefaultError{
				Code:    "SAME_USER",
				Message: "You can't block yourself",
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
