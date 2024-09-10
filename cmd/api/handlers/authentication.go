package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CDavidSV/Iris-Chat-App-Backend/internal"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/jwt"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/models"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/validator"
	"github.com/gofiber/fiber/v2"
)

type registerUserDTO struct {
	Email    string `validate:"email,req"`
	Username string `validate:"min=1,max=30,req"`
	Password string `validate:"min=8,max=50,req"`
	Device   string `validate:"req"`
	OS       string `validate:"req"`
}

type loginUserDTO struct {
	Username string `validate:"email,req"`
	Password string `validate:"min=8,max=50,req"`
	Device   string `validate:"req"`
	OS       string `validate:"req"`
}

type refreshTokenDTO struct {
	SessionID string `validate:"req"`
	Token     string `validate:"req"`
}

func (s *Server) Register(c *fiber.Ctx) error {
	var registeruserDTO registerUserDTO
	err := c.BodyParser(&registeruserDTO)
	if err != nil {
		return err
	}

	result, err := validator.Validate(registeruserDTO)
	if err != nil {
		return internal.ServerError(c, err, "Failed to validate request body")
	}

	if !result.IsValid {
		return result.SendValidationError(c)
	}

	// Attempt to create user
	userID, err := s.Users.InsertUser(
		registeruserDTO.Username,
		registeruserDTO.Email,
		registeruserDTO.Password,
	)

	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			return internal.ClientError(c, http.StatusUnprocessableEntity, "Email already in use")
		}

		if errors.Is(err, models.ErrDuplicateUsername) {
			return internal.ClientError(c, http.StatusUnprocessableEntity, "Username already in use")
		}

		return err
	}

	// Generate access and refresh tokens
	session, err := s.Sessions.NewSession(userID, registeruserDTO.Device, registeruserDTO.OS, c.IP())
	if err != nil {
		return internal.ServerError(c, err, "Failed to establish user session")
	}

	accessToken, err := jwt.GenerateAccessToken(userID, session.SessionID)
	if err != nil {
		fmt.Println(err)
		return internal.ServerError(c, err, "Failed to generate access token")
	}

	user, err := s.Users.FetchUser(userID)
	if err != nil {
		return internal.ServerError(c, err, "Failed to fetch user data")
	}

	return c.JSON(map[string]any{
		"msg":       "user created successfully",
		"sessionID": session.SessionID,
		"tokens": map[string]string{
			"accessToken":  accessToken,
			"refreshToken": session.RefreshToken,
		},
		"tokenType": "Bearer",
		"expiresIn": jwt.AccessTokenExpirationDelta / time.Millisecond,
		"user":      user,
	})
}

func (s *Server) Login(c *fiber.Ctx) error {
	var loginuserDTO loginUserDTO
	err := c.BodyParser(&loginuserDTO)
	if err != nil {
		return err
	}

	result, err := validator.Validate(loginuserDTO)
	if err != nil {
		return internal.ServerError(c, err, "Failed to validate request body")
	}

	if !result.IsValid {
		return result.SendValidationError(c)
	}

	userID, err := s.Users.Authenticate(loginuserDTO.Username, loginuserDTO.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			return internal.ClientError(c, http.StatusUnauthorized, "Invalid email address or password")
		}

		return internal.ServerError(c, err, "Failed to authenticate user")
	}

	session, err := s.Sessions.NewSession(userID, loginuserDTO.Device, loginuserDTO.OS, c.IP())
	if err != nil {
		return internal.ServerError(c, err, "Failed to establish user session")
	}

	accessToken, err := jwt.GenerateAccessToken(userID, session.SessionID)
	if err != nil {
		return internal.ServerError(c, err, "Failed to generate access token")
	}

	user, err := s.Users.FetchUser(userID)
	if err != nil {
		return internal.ServerError(c, err, "Failed to fetch user data")
	}

	return c.JSON(map[string]any{
		"msg":       "login successful",
		"sessionID": session.SessionID,
		"tokens": map[string]string{
			"accessToken":  accessToken,
			"refreshToken": session.RefreshToken,
		},
		"tokenType": "Bearer",
		"expiresIn": jwt.AccessTokenExpirationDelta / time.Millisecond,
		"user":      user,
	})
}

func (s *Server) Logout(c *fiber.Ctx) error {
	var logoutuserDTO refreshTokenDTO
	err := c.BodyParser(&logoutuserDTO)
	if err != nil {
		return err
	}

	result, err := validator.Validate(logoutuserDTO)
	if err != nil {
		return internal.ServerError(c, err, "Failed to validate request body")
	}

	if !result.IsValid {
		return result.SendValidationError(c)
	}

	err = s.Sessions.DeleteSession(c.Locals("sessionID").(string), logoutuserDTO.Token)
	if err != nil {
		if errors.Is(err, models.ErrNoSessionsFound) {
			return internal.ClientError(c, http.StatusUnprocessableEntity, "No sessions found")
		}
	}

	return c.JSON(map[string]string{
		"msg": "Session closed",
	})
}

func (s *Server) Token(c *fiber.Ctx) error {
	var refreshDTO refreshTokenDTO
	err := c.BodyParser(&refreshDTO)
	if err != nil {
		return err
	}

	result, err := validator.Validate(refreshDTO)
	if err != nil {
		return internal.ServerError(c, err, "Failed to validate request body")
	}

	if !result.IsValid {
		return result.SendValidationError(c)
	}

	// Attempt to revalidate the user's session
	newSession, err := s.Sessions.RevalidateSession(refreshDTO.SessionID, refreshDTO.Token)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidSession):
			return internal.ClientError(c, http.StatusUnauthorized, "session data is invalid")
		case errors.Is(err, models.ErrSessionExpired):
			return internal.ClientError(c, http.StatusUnauthorized, "session has expired")
		default:
			return internal.ServerError(c, err, "Unable to revalidate session")
		}
	}

	accessToken, err := jwt.GenerateAccessToken(newSession.UserID, newSession.SessionID)
	if err != nil {
		return internal.ServerError(c, err, "Unable to revalidate session")
	}

	return c.JSON(map[string]any{
		"msg": "session revalidated",
		"tokens": map[string]string{
			"accessToken":  accessToken,
			"refreshToken": newSession.RefreshToken,
		},
		"tokenType": "Bearer",
		"expiresIn": jwt.AccessTokenExpirationDelta / time.Millisecond,
	})
}
