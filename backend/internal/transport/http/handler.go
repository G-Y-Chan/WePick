package http

import (
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
	svc    room.Service
	logger *slog.Logger
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
		writeErrorHeader(w, err, "Room Code Error")
		return
	}

	body := code.String()
	writeJSON(w, nethttp.StatusOK, messageDTO{
		Header: "Room Code",
		Body:   &body,
	})
}

// JoinRoom handles POST /rooms/{code}/join.
func (h *Handler) JoinRoom(w nethttp.ResponseWriter, r *nethttp.Request) {
	code, err := roomCodeFromPath(w, r, "Join Room Error")
	if err != nil {
		return
	}

	joined, err := h.svc.JoinRoom(r.Context(), code)
	if err != nil {
		writeErrorHeader(w, err, "Join Room Error")
		return
	}

	body := strconv.FormatBool(joined)
	writeJSON(w, nethttp.StatusOK, messageDTO{
		Header: "Join Status",
		Body:   &body,
	})
}

// StartRoom handles POST /rooms/{code}/start.
func (h *Handler) StartRoom(w nethttp.ResponseWriter, r *nethttp.Request) {
	code, err := roomCodeFromPath(w, r, "Start Room Error")
	if err != nil {
		return
	}

	var payload startRoomPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		_ = r.Body.Close()
		writeErrorHeader(w, apperr.New(apperr.CodeInvalid, "invalid JSON payload"), "Start Room Error")
		return
	}
	_ = r.Body.Close()

	filters, err := payload.Filters.toDomain()
	if err != nil {
		writeErrorHeader(w, err, "Start Room Error")
		return
	}

	if err := h.svc.StartRoom(r.Context(), code, filters); err != nil {
		writeErrorHeader(w, err, "Start Room Error")
		return
	}

	started := strconv.FormatBool(true)
	writeJSON(w, nethttp.StatusOK, messageDTO{
		Header: "Start Status",
		Body:   &started,
	})
}

// GetCards handles GET /rooms/{code}/cards.
func (h *Handler) GetCards(w nethttp.ResponseWriter, r *nethttp.Request) {
	code, err := roomCodeFromPath(w, r, "Get Card Data Error")
	if err != nil {
		return
	}

	cards, err := h.svc.Cards(r.Context(), code)
	if err != nil {
		writeErrorHeader(w, err, "Get Card Data Error")
		return
	}

	dtos := cardDTOsFromDomain(cards)
	if dtos == nil {
		// Preserve legacy behaviour: the client always receives a JSON array,
		// never null, when no cards have been stored yet.
		dtos = []cardDTO{}
	}

	writeJSON(w, nethttp.StatusOK, messageDTO{
		Header: "CARD_DATA",
		Cards:  dtos,
	})
}

// GetPhoto handles GET /image?photoName=... by redirecting to the resolved
// Google photo URL.
func (h *Handler) GetPhoto(w nethttp.ResponseWriter, r *nethttp.Request) {
	photoName := r.URL.Query().Get("photoName")
	if photoName == "" {
		writeErrorHeader(w, apperr.New(apperr.CodeInvalid, "missing photo name parameter"), "Image Error")
		return
	}

	photoURI, err := h.svc.PhotoURL(r.Context(), photoName)
	if err != nil {
		writeErrorHeader(w, err, "Image Error")
		return
	}

	nethttp.Redirect(w, r, photoURI, nethttp.StatusFound)
}

// Healthz handles GET /healthz (new, additive). Phase 7 wires this to a live
// Redis PING; for the Phase 6 cutover it returns a simple liveness response.
func (h *Handler) Healthz(w nethttp.ResponseWriter, r *nethttp.Request) {
	writeJSON(w, nethttp.StatusOK, map[string]string{"status": "ok"})
}

// roomCodeFromPath extracts and validates the {code} path parameter, returning
// the legacy "missing room code in url path" body when the path value is empty
// and the domain-level CodeInvalid error for malformed codes.
func roomCodeFromPath(w nethttp.ResponseWriter, r *nethttp.Request, header string) (room.Code, error) {
	raw := r.PathValue("code")
	if raw == "" {
		writeErrorHeader(w, apperr.New(apperr.CodeInvalid, "missing room code in url path"), header)
		return "", apperr.New(apperr.CodeInvalid, "missing room code in url path")
	}

	code, err := room.ParseCode(raw)
	if err != nil {
		writeErrorHeader(w, err, header)
		return "", err
	}
	return code, nil
}
