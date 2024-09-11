package internal

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/oklog/ulid/v2"
)

type ClientReqError struct {
	Status string `json:"status"`
	Errors []any  `json:"errors"`
}

type DefaultError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var ErrInvalidContentType = errors.New("invalid Content-Type")

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

func VerifyContentType(c *fiber.Ctx, contentType string) error {
	headers := c.GetReqHeaders()

	contentTypeSlice, ok := headers["Content-Type"]
	if !ok || len(contentTypeSlice) < 1 {
		return ErrInvalidContentType
	}

	if contentTypeSlice[0] != contentType {
		return ErrInvalidContentType
	}

	return nil
}
