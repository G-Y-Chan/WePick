package handlers

import (
	"backend/util"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func (s *Server) GetRoomCode(w http.ResponseWriter, req *http.Request) {
	// Inform client that the response type is JSON
	w.Header().Set("Content-Type", "application/json")
    // Set the HTTP status code (optional, http.StatusOK is 200).
	w.WriteHeader(http.StatusOK)
	var code = s.RoomService.GenerateCode()
	m := util.Message{"Room Code", code}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleRoomJoin(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	roomCode, err := parseRoomCode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		fmt.Print(roomCode)
	}

	// Inform client that the response type is JSON
	w.Header().Set("Content-Type", "application/json")

	joined, err := s.RoomService.JoinRoom(roomCode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(util.ErrorResponse{
			Header: "Join Room Error",
			Body:   err.Error(),
		})
		return
	}

	if !joined {
		w.WriteHeader(http.StatusForbidden) // Room already started
		json.NewEncoder(w).Encode(util.ErrorResponse{
			Header: "Join Room Error",
			Body:   "room already started",
		})
		return
	}

	m := util.Message{"Join Status", strconv.FormatBool(joined)}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) HandleRoomStart(w http.ResponseWriter, req *http.Request) {
	roomCode, err := parseRoomCode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Attempting to start room:", roomCode)
	w.Header().Set("Content-Type", "application/json")

	started, err := s.RoomService.StartRoom(roomCode)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(util.ErrorResponse{
			Header: "Start Room Error",
			Body:   err.Error(),
		})
		return
	}

	m := util.Message{"Start Status", strconv.FormatBool(started)}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func parseRoomCode(req *http.Request) (string, error) {
	defer req.Body.Close()

	var roomCode string
	if err := json.NewDecoder(req.Body).Decode(&roomCode); err != nil {
		return "", err
	}

	if roomCode == "" {
		return "", errors.New("empty room code")
	}

	return roomCode, nil
}
