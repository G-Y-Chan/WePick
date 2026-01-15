package main

import (
	"encoding/json"
	"backend/util"
	"net/http"
	"fmt"
	"strconv"
)

func (s *Server) getRoomCode(w http.ResponseWriter, req *http.Request) {
	// Inform client that the response type is JSON
	w.Header().Set("Content-Type", "application/json")
    // Set the HTTP status code (optional, http.StatusOK is 200).
	w.WriteHeader(http.StatusOK)
	var code = s.roomService.GenerateCode()
	m := util.Message{"Room Code", code}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleRoomStatus(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var roomCode string

	err := json.NewDecoder(req.Body).Decode(&roomCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		fmt.Print(roomCode)
	}

	// Inform client that the response type is JSON
	w.Header().Set("Content-Type", "application/json")

	var status = s.roomService.VerifyCode(roomCode)
	if status {
		intCode, _ := strconv.Atoi(roomCode)
		var joined = strconv.FormatBool(s.roomService.JoinRoomLocked(intCode))
		m := util.Message{"Verification Status", joined}
		if err := json.NewEncoder(w).Encode(m); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the HTTP status code (optional, http.StatusOK is 200).
		w.WriteHeader(http.StatusOK)
	} else {
		m := util.Message{"Verification Status", "false"}
		if err := json.NewEncoder(w).Encode(m); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleStart(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var roomCode string

	err := json.NewDecoder(req.Body).Decode(&roomCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		fmt.Println("Attempting to start room: " + roomCode)
	}

	// Inform client that the response type is JSON
	w.Header().Set("Content-Type", "application/json")

	var status = s.roomService.VerifyCode(roomCode)
	if status {
		// Verify code ensures valid code, thus error variable is redundant
		intCode, _ := strconv.Atoi(roomCode)
		var started = s.roomService.StartRoomLocked(intCode)
		if started {
			fmt.Println("Host of room " + roomCode + " has started room.")
			stringStatus := strconv.FormatBool(status)
			m := util.Message{"Verification Status", stringStatus}
			if err := json.NewEncoder(w).Encode(m); err == nil {
				w.WriteHeader(http.StatusOK)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// Room was not started successfully.
			http.Error(w, "Room not started successfully.", http.StatusInternalServerError)
		}
	} else {
		// Invalid room code provided in request
		http.Error(w, "Invalid room code provided.", http.StatusInternalServerError)
	}
}