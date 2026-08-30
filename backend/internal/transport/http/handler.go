package http

import (
	"context"
	"encoding/json"
	"log/slog"
	nethttp "net/http"
	"strconv"

	"backend/internal/apperr"
	"backend/internal/domain/room"
)

// Handler is the HTTP transport layer. It depends only on room.Service and
// never imports the repository, places, or websocket adapters directly.
type Handler struct {
	svc           room.Service
	logger        *slog.Logger
	healthChecker func(ctx context.Context) error
}

// NewHandler constructs the HTTP transport Handler.
func NewHandler(svc room.Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		svc:    svc,
		logger: logger,
	}
}

// WithHealthChecker attaches an optional dependency-inspection callback (e.g. a
// Redis PING). When nil the /healthz endpoint reports a simple liveness check.
func (h *Handler) WithHealthChecker(checker func(ctx context.Context) error) *Handler {
	h.healthChecker = checker
	return h
}

// Routes returns a fully-populated *http.ServeMux and is the single source of
// route truth for the HTTP API. Middleware is applied by the composition root.
func (h *Handler) Routes() nethttp.Handler {
	mux := nethttp.NewServeMux()

	mux.HandleFunc("GET /rooms", h.CreateRoom)
	mux.HandleFunc("POST /rooms/{code}/join", h.JoinRoom)
	mux.HandleFunc("POST /rooms/{code}/start", h.StartRoom)
	mux.HandleFunc("GET /rooms/{code}/cards", h.GetCards)
	mux.HandleFunc("GET /image", h.GetPhoto)
	mux.HandleFunc("GET /healthz", h.Healthz)

	return mux
}

// CreateRoom handles GET /rooms.
func (h *Handler) CreateRoom(w nethttp.ResponseWriter, r *nethttp.Request) {
	code, err := h.svc.CreateRoom(r.Context())
	if err != nil {
		h.writeErrorHeader(w, err, "Room Code Error")
		return
	}

	body := code.String()
	h.writeJSON(w, nethttp.StatusOK, messageDTO{
		Header: "Room Code",
		Body:   &body,
	})
}

// JoinRoom handles POST /rooms/{code}/join.
func (h *Handler) JoinRoom(w nethttp.ResponseWriter, r *nethttp.Request) {
	code, err := h.roomCodeFromPath(w, r, "Join Room Error")
	if err != nil {
		return
	}

	joined, err := h.svc.JoinRoom(r.Context(), code)
	if err != nil {
		h.writeErrorHeader(w, err, "Join Room Error")
		return
	}

	body := strconv.FormatBool(joined)
	h.writeJSON(w, nethttp.StatusOK, messageDTO{
		Header: "Join Status",
		Body:   &body,
	})
}

// StartRoom handles POST /rooms/{code}/start.
func (h *Handler) StartRoom(w nethttp.ResponseWriter, r *nethttp.Request) {
	code, err := h.roomCodeFromPath(w, r, "Start Room Error")
	if err != nil {
		return
	}

	var payload startRoomPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		_ = r.Body.Close()
		h.writeErrorHeader(w, apperr.New(apperr.CodeInvalid, "invalid JSON payload"), "Start Room Error")
		return
	}
	_ = r.Body.Close()

	filters, err := payload.Filters.toDomain()
	if err != nil {
		h.writeErrorHeader(w, err, "Start Room Error")
		return
	}

	if err := h.svc.StartRoom(r.Context(), code, filters); err != nil {
		h.writeErrorHeader(w, err, "Start Room Error")
		return
	}

	started := strconv.FormatBool(true)
	h.writeJSON(w, nethttp.StatusOK, messageDTO{
		Header: "Start Status",
		Body:   &started,
	})
}

// GetCards handles GET /rooms/{code}/cards.
func (h *Handler) GetCards(w nethttp.ResponseWriter, r *nethttp.Request) {
	code, err := h.roomCodeFromPath(w, r, "Get Card Data Error")
	if err != nil {
		return
	}

	cards, err := h.svc.Cards(r.Context(), code)
	if err != nil {
		h.writeErrorHeader(w, err, "Get Card Data Error")
		return
	}

	dtos := cardDTOsFromDomain(cards)
	if dtos == nil {
		// Preserve legacy behaviour: the client always receives a JSON array,
		// never null, when no cards have been stored yet.
		dtos = []cardDTO{}
	}

	h.writeJSON(w, nethttp.StatusOK, messageDTO{
		Header: "CARD_DATA",
		Cards:  dtos,
	})
}

// GetPhoto handles GET /image?photoName=... by redirecting to the resolved
// Google photo URL.
func (h *Handler) GetPhoto(w nethttp.ResponseWriter, r *nethttp.Request) {
	photoName := r.URL.Query().Get("photoName")
	if photoName == "" {
		h.writeErrorHeader(w, apperr.New(apperr.CodeInvalid, "missing photo name parameter"), "Image Error")
		return
	}

	photoURI, err := h.svc.PhotoURL(r.Context(), photoName)
	if err != nil {
		h.writeErrorHeader(w, err, "Image Error")
		return
	}

	nethttp.Redirect(w, r, photoURI, nethttp.StatusFound)
}

// Healthz handles GET /healthz. When a health checker is configured (e.g. a
// Redis PING) it reports the dependency status alongside the service status.
func (h *Handler) Healthz(w nethttp.ResponseWriter, r *nethttp.Request) {
	status := map[string]string{"status": "ok"}

	if h.healthChecker != nil {
		if err := h.healthChecker(r.Context()); err != nil {
			status["redis"] = "unhealthy"
			status["redis_error"] = err.Error()
			h.writeJSON(w, nethttp.StatusServiceUnavailable, status)
			return
		}
		status["redis"] = "healthy"
	}

	h.writeJSON(w, nethttp.StatusOK, status)
}

// roomCodeFromPath extracts and validates the {code} path parameter, returning
// the legacy "missing room code in url path" body when the path value is empty
// and the domain-level CodeInvalid error for malformed codes.
func (h *Handler) roomCodeFromPath(w nethttp.ResponseWriter, r *nethttp.Request, header string) (room.Code, error) {
	raw := r.PathValue("code")
	if raw == "" {
		h.writeErrorHeader(w, apperr.New(apperr.CodeInvalid, "missing room code in url path"), header)
		return "", apperr.New(apperr.CodeInvalid, "missing room code in url path")
	}

	code, err := room.ParseCode(raw)
	if err != nil {
		h.writeErrorHeader(w, err, header)
		return "", err
	}
	return code, nil
}
