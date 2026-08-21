package api

import (
	"backend/internal/room"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Warn("failed to encode response", "error", err)
	}
}

// GET /rooms
func (s *Server) GetRoomCode(w http.ResponseWriter, req *http.Request) {
	code, err := s.RoomManager.AddRoom(req.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{
			Header: "Room Code Error",
			Body:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, Message{
		Header: "Room Code",
		Body:   &code,
	})
}

// POST /rooms/{code}/join
func (s *Server) HandleRoomJoin(w http.ResponseWriter, req *http.Request) {
	code := req.PathValue("code")
	if code == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Header: "Join Room Error",
			Body:   "missing room code in url path",
		})
		return
	}

	slog.Info("received request to join room", "code", code)

	joined, err := s.RoomManager.ValidateRoomJoin(req.Context(), code)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Header: "Join Room Error",
			Body:   err.Error(),
		})
		return
	}

	if !joined {
		respondJSON(w, http.StatusForbidden, ErrorResponse{
			Header: "Join Room Error",
			Body:   "room already started",
		})
		return
	}

	strJoined := strconv.FormatBool(joined)
	respondJSON(w, http.StatusOK, Message{
		Header: "Join Status",
		Body:   &strJoined,
	})
}

// POST /rooms/{code}/start
func (s *Server) HandleRoomStart(w http.ResponseWriter, req *http.Request) {
	code := req.PathValue("code")
	if code == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Header: "Start Room Error",
			Body:   "missing room code in url path",
		})
		return
	}

	var payload StartRoomPayload
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Header: "Start Room Error",
			Body:   "invalid JSON payload",
		})
		return
	}
	defer req.Body.Close()

	filters := room.Filters{
		Latitude:  payload.Filters.Latitude,
		Longitude: payload.Filters.Longitude,
		Radius:    payload.Filters.Radius,
		MaxPrice:  payload.Filters.MaxPrice,
		Category:  payload.Filters.Category,
		OpenNow:   payload.Filters.OpenNow,
	}

	slog.Info("attempting to start room", "code", code, "filters", filters)

	started, err := s.RoomManager.StartRoom(req.Context(), code, filters)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Header: "Start Room Error",
			Body:   err.Error(),
		})
		return
	}

	strStarted := strconv.FormatBool(started)
	respondJSON(w, http.StatusOK, Message{
		Header: "Start Status",
		Body:   &strStarted,
	})
}

// GET /rooms/{code}/cards
func (s *Server) HandleGetCardData(w http.ResponseWriter, req *http.Request) {
	code := req.PathValue("code")
	if code == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Header: "Get Card Data Error",
			Body:   "missing room code in url path",
		})
		return
	}

	cards, err := s.RoomManager.GetRoomCards(req.Context(), code)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{
			Header: "Get Card Data Error",
			Body:   "Failed to fetch cards: " + err.Error(),
		})
		return
	}

	if cards == nil {
		cards = []room.Card{}
	}

	// Translate domain cards to DTO cards
	dtoCards := make([]Card, len(cards))
	for i, c := range cards {
		dtoCards[i] = Card{
			ID:          c.ID,
			Title:       c.Title,
			Category:    c.Category,
			PriceLevel:  c.PriceLevel,
			Rating:      c.Rating,
			ReviewCount: c.ReviewCount,
			OpenNow:     c.OpenNow,
			Summary:     c.Summary,
			Address:     c.Address,
			PhotoName:   c.PhotoName,
		}
	}

	respondJSON(w, http.StatusOK, Message{
		Header: "CARD_DATA",
		Cards:  dtoCards,
	})
}

// GET /image
func (s *Server) HandleGetImage(w http.ResponseWriter, req *http.Request) {
	photoName := req.URL.Query().Get("photoName")
	if photoName == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{
			Header: "Image Error",
			Body:   "missing photo name parameter",
		})
		return
	}

	photoUri, err := s.RoomManager.GetPhotoURL(photoName)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{
			Header: "Image Error",
			Body:   "failed to fetch image",
		})
		return
	}

	http.Redirect(w, req, photoUri, http.StatusFound)
}