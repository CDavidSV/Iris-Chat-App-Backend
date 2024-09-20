package handlers

import (
	"errors"
	"net/http"

	"github.com/CDavidSV/Iris-Chat-App-Backend/internal"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) GetMe(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	user, err := s.Users.FetchUser(userID)
	if err != nil {
		return internal.ServerError(c, err, "Failed to fetch user")
	}

	return c.JSON(user)
}

func (s *Server) GetUser(c *fiber.Ctx) error {
	userID := c.Params("userID")

	user, err := s.Users.FetchUser(userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return internal.ClientError(c, http.StatusNotFound, internal.DefaultError{
				Code:    "USER_NOT_FOUND",
				Message: "User with the given ID does not exist",
			})
		}
		return internal.ServerError(c, err, "Failed to fetch user")
	}

	return c.JSON(map[string]any{
		"userID":            user.UserID,
		"username":          user.Username,
		"displayName":       user.DisplayName,
		"joinedAt":          user.JoinedAt,
		"customStatus":      user.CustomStatus,
		"profilePictureURL": user.ProfilePictureURL,
	})
}

func (s *Server) GetUsersByUsername(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return internal.ClientError(c, http.StatusBadRequest, internal.DefaultError{
			Code:    "MISSING_PARAM",
			Message: "Missing username parameter",
		})
	}

	user, err := s.Users.FetchUsersByUsername(username)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return c.JSON([]models.PublicUserDTO{})
		}

		return internal.ServerError(c, err, "Failed to fetch user")
	}

	return c.JSON(user)
}
