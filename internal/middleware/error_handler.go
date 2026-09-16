package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ranji/clothing-erp/internal/apperr"
	"github.com/ranji/clothing-erp/internal/response"
)

// HTTPErrorHandler translates any error returned by a handler into the
// standard { success, message, data, errors } envelope, using apperr.Error's
// Kind to pick the HTTP status when present.
func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var ae *apperr.Error
	if errors.As(err, &ae) {
		// For unexpected (KindInternal) failures, surface the wrapped cause in
		// `errors` so it's visible from the browser/API client, not just server
		// logs -- validation/conflict/etc. messages are already self-explanatory
		// and don't need this.
		var detail any
		if ae.Kind == apperr.KindInternal && ae.Err != nil {
			detail = ae.Err.Error()
		}
		_ = response.Fail(c, ae.HTTPStatus(), ae.Message, detail)
		return
	}

	var he *echo.HTTPError
	if errors.As(err, &he) {
		_ = response.Fail(c, he.Code, http.StatusText(he.Code), he.Message)
		return
	}

	_ = response.Fail(c, http.StatusInternalServerError, "internal server error", err.Error())
}
