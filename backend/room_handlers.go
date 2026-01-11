package main

import (
	"encoding/json"
	"main/util"
	"net/http"
	"fmt"
)

func getRoomCode(w http.ResponseWriter, req *http.Request) {
	// Inform client that the response type is JSON
	w.Header().Set("Content-Type", "application/json")
    // Set the HTTP status code (optional, http.StatusOK is 200).
	w.WriteHeader(http.StatusOK)
	var code = generateWrapper()
	m := util.Message{"Room Code", code}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func verifyRoomCode(w http.ResponseWriter, req *http.Request) {
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

	var status = verifyWrapper(roomCode)
	m := util.Message{"Verification Status", status}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    // Set the HTTP status code (optional, http.StatusOK is 200).
	w.WriteHeader(http.StatusOK)
}

func verifyStart(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var roomCode string

	err := json.NewDecoder(req.Body).Decode(&roomCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		fmt.Print("Starting room: " + roomCode)
	}

	// Inform client that the response type is JSON
	w.Header().Set("Content-Type", "application/json")

	var status = verifyWrapper(roomCode)
	m := util.Message{"Verification Status", status}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    // Set the HTTP status code (optional, http.StatusOK is 200).
	w.WriteHeader(http.StatusOK)
}