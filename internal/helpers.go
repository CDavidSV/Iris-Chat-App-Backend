package internal

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/oklog/ulid/v2"
)

type ClientReqError struct {
	Status string
	Errors []any
}

func GetIrisEpoch() time.Time {
	epoch, _ := time.Parse("2006-01-02T15:04:05.000Z", IrisEpoch)

	return epoch
}

func GenerateID() string {
	id := ulid.Make()

	return id.String()
}

func ServerError(c *fiber.Ctx, err error, errorMsg string) error {
	log.Error(err)

	errorJson := map[string]any{
		"status": http.StatusInternalServerError,
		"error":  errorMsg,
	}

	return c.Status(http.StatusInternalServerError).JSON(errorJson)
}

func ClientError(c *fiber.Ctx, status int, errors ...any) error {
	errorRes := ClientReqError{
		Status: strconv.Itoa(status),
		Errors: errors,
	}

	return c.Status(status).JSON(errorRes)
}
