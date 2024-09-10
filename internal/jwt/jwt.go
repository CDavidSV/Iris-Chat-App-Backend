package jwt

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var AccessTokenExpirationDelta time.Duration = time.Minute * 15

func GenerateAccessToken(userID, sessionID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    userID,
		"sessionID": sessionID,
		"exp":       time.Now().Add(AccessTokenExpirationDelta).Unix(), // Token expires after 15 minutes
	})

	accessTokenString, err := token.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))
	if err != nil {
		return "", err
	}

	return accessTokenString, nil
}
