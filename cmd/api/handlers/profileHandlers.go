package handlers

import (
	"errors"
	"net/http"

	"github.com/CDavidSV/Iris-Chat-App-Backend/internal"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/models"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/validator"
	"github.com/gofiber/fiber/v2"
)

type changePasswordDTO struct {
	OldPassword string `validate:"required,min=1"`
	NewPassword string `validate:"required,min=8,max=50"`
}

func (s *Server) UpdateProfile(c *fiber.Ctx) error {
	var updateData map[string]any

	err := c.BodyParser(updateData)
	if err != nil {
		return internal.ClientError(c, fiber.StatusBadRequest, internal.DefaultError{
			Code:    "INVALID_BODY",
			Message: "Invalid request body",
		})
	}

	var newDisplayName, newBio string

	clientUser, err := s.Users.FetchUser(c.Locals("userID").(string))
	if err != nil {
		return internal.ServerError(c, err, "Failed to fetch user")
	}

	if _, ok := updateData["displayName"]; ok {
		newDisplayName = updateData["displayName"].(string)
	} else {
		newDisplayName = string(clientUser.DisplayName)
	}

	if _, ok := updateData["bio"]; ok {
		newBio = updateData["bio"].(string)
	} else {
		newBio = string(clientUser.Bio)
	}

	updatedUser, err := s.Users.UpdateProfileInfo(c.Locals("userID").(string), newDisplayName, newBio)
	if err != nil {
		return internal.ServerError(c, err, "Failed to update user profile")
	}

	return c.JSON(updatedUser)
}

func (s *Server) ChangePassword(c *fiber.Ctx) error {
	err := internal.VerifyContentType(c, "application/x-www-form-urlencoded")
	if err != nil {
		return internal.ClientError(c, http.StatusUnsupportedMediaType, internal.DefaultError{
			Code:    "UNSUPPORTED_MEDIA_TYPE",
			Message: "Content-Type header must be application/x-www-form-urlencoded",
		})
	}

	var changePassDTO changePasswordDTO
	err = c.BodyParser(&changePassDTO)
	if err != nil {
		return err
	}

	result, err := validator.Validate(changePassDTO)
	if err != nil {
		return internal.ServerError(c, err, "Failed to validate request body")
	}

	if !result.IsValid {
		return result.SendValidationError(c)
	}

	err = s.Users.UpdatePassword(c.Locals("userID").(string), changePassDTO.OldPassword, changePassDTO.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			return internal.ClientError(c, fiber.StatusUnauthorized, internal.DefaultError{
				Code:    "INVALID_CREDENTIALS",
				Message: "Invalid old password",
			})
		}

		return internal.ServerError(c, err, "Failed to update password")
	}

	return nil
}
