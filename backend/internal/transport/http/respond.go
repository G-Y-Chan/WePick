package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"backend/internal/apperr"
)

// writeJSON writes payload as a JSON response with the supplied status code.
func (h *Handler) writeJSON(w nethttp.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// The response is already committed; there is nothing useful we can
		// send to the client here. Logging is the only recovery option.
		h.logger.Error("failed to encode JSON response", "err", err)
	}
}

// writeError maps err to an errorResponseDTO and the authoritative HTTP status
// code from apperr.HTTPStatus. It is the single source of truth for
// Code -> HTTP mapping in the transport layer.
func (h *Handler) writeError(w nethttp.ResponseWriter, err error) {
	h.writeErrorHeader(w, err, "Error")
}

// writeErrorHeader is the same as writeError but preserves the legacy
// handler-specific error envelope Header values (e.g. "Join Room Error").
// Status mapping is still centralized through apperr.Error.HTTPStatus.
func (h *Handler) writeErrorHeader(w nethttp.ResponseWriter, err error, header string) {
	appErr := asAppError(err)
	if header == "" {
		header = "Error"
	}

	h.writeJSON(w, appErr.HTTPStatus(), errorResponseDTO{
		Header: header,
		Body:   appErr.Message,
	})
}

// asAppError normalizes any error to an *apperr.Error so handlers never have
// to string-match or hand-roll status codes. Unknown errors become CodeInternal
// and their details are kept in the wrapped cause, never exposed to clients.
func asAppError(err error) *apperr.Error {
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		return appErr
	}

	return apperr.Wrap(apperr.CodeInternal, "internal server error", err)
}
