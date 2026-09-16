package response

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Envelope is the standard API response shape:
// { "success": boolean, "message": string, "data": any, "errors": any }
type Envelope struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

func OK(c echo.Context, code int, message string, data any) error {
	return c.JSON(code, Envelope{Success: true, Message: message, Data: data})
}

func Fail(c echo.Context, code int, message string, errs any) error {
	return c.JSON(code, Envelope{Success: false, Message: message, Errors: errs})
}

// PDF streams a generated PDF inline (rendered in-browser, not force-downloaded).
func PDF(c echo.Context, filename string, data []byte) error {
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf(`inline; filename="%s"`, filename))
	return c.Blob(http.StatusOK, "application/pdf", data)
}
