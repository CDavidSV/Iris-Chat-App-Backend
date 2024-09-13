package handlers

import (
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal"
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
