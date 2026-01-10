package main

import (
	"encoding/json"
	"main/util"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"fmt"
)

var intervals []util.Interval
var max = 1000000

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

func generateWrapper() string {
	paddedLength := 6
	var code int
	code = generateRoomCode()
	insertRoomCode(code)
	if code >= 100000 {
		return strconv.Itoa(code)
	} else {
		var temp = strconv.Itoa(code)
		for len(temp) < paddedLength {
			temp = "0" + temp
		}
		return temp
	}
}

func generateRoomCode() int {
	var code int
	code = rand.Intn(max)
	if len(intervals) == 0 {
		return code
	}
	for i := 0; i < len(intervals); i++ {
		s := intervals[i].Start
		e := intervals[i].End
		if code >= e {
			code += (e-s+1)
		} else if code >= s {
			code := e + (code-s+1)
			return code
		} else {
			return code
		}
	}
	max := max - 1
	_ = max
	return code
}

// Logic behind generating new room code guarantees validity of code.
// Thus, no checks required before inserting new code into intervals.
func insertRoomCode(code int) {
	var newIntervalIndex = -1
	for i := 0; i < len(intervals); i++ {
		s := intervals[i].Start
		e := intervals[i].End
		if code >= e {
			if code == e+1 {
				intervals[i].End = code
				return
			}
		} else {
			if code+1 == s {
				intervals[i].Start = code
				return
			} else {
				newIntervalIndex = i
				break
			}
		}
	}
	// New code larger than all previous ones
	if newIntervalIndex == -1 {
		intervals = append(intervals, util.Interval{code, code})
	} else {
		newInterval := util.Interval{code, code}
		intervals = slices.Insert(intervals, newIntervalIndex, newInterval)
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

func verifyWrapper(roomCode string) string {
	var intCode, err = strconv.Atoi(roomCode)
	if err != nil {
		return "false"
	}
	for i := 0; i < len(intervals); i++ {
		if intCode >= intervals[i].Start && intCode <= intervals[i].End {
			return "true"
		}
	}
	return "false"
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