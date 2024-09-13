package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/CDavidSV/Iris-Chat-App-Backend/internal"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Authorize(c *fiber.Ctx) error {
	headers := c.GetReqHeaders()
	authorizationValue, ok := headers["Authorization"]
	if !ok || len(authorizationValue) == 0 {
		return c.SendStatus(http.StatusUnauthorized)
	}

	if !strings.HasPrefix(authorizationValue[0], "Bearer") {
		return c.SendStatus(http.StatusUnauthorized)
	}

	accessToken := strings.Split(authorizationValue[0], " ")[1]

	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(accessTokenSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return internal.ClientError(c, http.StatusUnauthorized, internal.DefaultError{
				Code:    "TOKEN_EXPIRED",
				Message: "Access token has expired",
			})
		}

		return c.SendStatus(http.StatusUnauthorized)
	}

	if !token.Valid {
		return c.SendStatus(http.StatusUnauthorized)
	}

	c.Locals("userID", claims["userID"])
	c.Locals("sessionID", claims["sessionID"])
	c.Locals("accessToken", token)

	return c.Next()
}
