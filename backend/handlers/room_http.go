package handlers

import (
	"backend/util"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// ==========================================
// Request DTO Structs
// ==========================================

type StartRoomPayload struct {
	Filters util.Filters `json:"filters"`
}

// ==========================================
// Helper Function
// ==========================================

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Printf("Failed to encode response: %v\n", err)
	}
}

// ==========================================
// HTTP Handlers
// ==========================================

// Route: GET /rooms
func (s *Server) GetRoomCode(w http.ResponseWriter, req *http.Request) {
	code, err := s.RoomManager.AddRoom()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, util.ErrorResponse{
			Header: "Room Code Error",
			Body:   err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, util.Message{
		Header: "Room Code",
		Body:   &code,
	})
}

// Route: POST /rooms/{code}/join
func (s *Server) HandleRoomJoin(w http.ResponseWriter, req *http.Request) {
	code := req.PathValue("code")
	if code == "" {
		respondJSON(w, http.StatusBadRequest, util.ErrorResponse{
			Header: "Join Room Error",
			Body:   "missing room code in url path",
		})
		return
	}

	fmt.Println("Received request to join room:", code)

	joined, err := s.RoomManager.ValidateRoomJoin(code)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, util.ErrorResponse{
			Header: "Join Room Error",
			Body:   err.Error(),
		})
		return
	}

	if !joined {
		respondJSON(w, http.StatusForbidden, util.ErrorResponse{
			Header: "Join Room Error",
			Body:   "room already started",
		})
		return
	}

	strJoined := strconv.FormatBool(joined)
	respondJSON(w, http.StatusOK, util.Message{
		Header: "Join Status",
		Body:   &strJoined,
	})
}

func (s *Server) HandleRoomStart(w http.ResponseWriter, req *http.Request) {
	code := req.PathValue("code")
	if code == "" {
		respondJSON(w, http.StatusBadRequest, util.ErrorResponse{
			Header: "Start Room Error",
			Body:   "missing room code in url path",
		})
		return
	}

	var payload StartRoomPayload

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		respondJSON(w, http.StatusBadRequest, util.ErrorResponse{
			Header: "Start Room Error",
			Body:   "invalid JSON payload",
		})
		return
	}
	defer req.Body.Close()

	fmt.Printf("Attempting to start room: %s with filters: %+v\n", code, payload.Filters)

	started, err := s.RoomManager.StartRoom(code, payload.Filters)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, util.ErrorResponse{
			Header: "Start Room Error",
			Body:   err.Error(),
		})
		return
	}

	strStarted := strconv.FormatBool(started)
	respondJSON(w, http.StatusOK, util.Message{
		Header: "Start Status",
		Body:   &strStarted,
	})
}

func (s *Server) HandleGetCardData(w http.ResponseWriter, req *http.Request) {
	code := req.PathValue("code")
	if code == "" {
		respondJSON(w, http.StatusBadRequest, util.ErrorResponse{
			Header: "Get Card Data Error",
			Body:   "missing room code in url path",
		})
		return
	}

	cards, err := s.RoomManager.GetRoomCards(req.Context(), code)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, util.ErrorResponse{
			Header: "Get Card Data Error",
			Body:   "Failed to fetch cards: " + err.Error(),
		})
		return
	}

	if cards == nil {
		cards = []util.Card{}
	}

	respondJSON(w, http.StatusOK, util.Message{
		Header: "CARD_DATA",
		Cards:  cards,
	})
}

func (s *Server) HandleGetImage(w http.ResponseWriter, req *http.Request) {
	photoName := req.URL.Query().Get("name")
	if photoName == "" {
		respondJSON(w, http.StatusBadRequest, util.ErrorResponse{
			Header: "Image Error",
			Body:   "missing photo name parameter",
		})
		return
	}

	photoUri, err := s.RoomManager.GetPhotoURL(photoName)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, util.ErrorResponse{
			Header: "Image Error",
			Body:   "failed to fetch image",
		})
		return
	}

	http.Redirect(w, req, photoUri, http.StatusFound)
}
