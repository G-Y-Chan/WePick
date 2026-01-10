package main

import (
    "fmt"
    "net/http"
	"encoding/json"
	"main/util"
)

func test(w http.ResponseWriter, req *http.Request) {
	// Inform client that the response type is JSON
	w.Header().Set("Content-Type", "application/json")
    // Set the HTTP status code (optional, http.StatusOK is 200).
	w.WriteHeader(http.StatusOK)
	m := util.Message{"testing", "This is a placeholder body."}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func headers(w http.ResponseWriter, req *http.Request) {

    for name, headers := range req.Header {
        for _, h := range headers {
            fmt.Fprintf(w, "%v: %v\n", name, h)
        }
    }
}

func postEmail(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var email string

	// Use json.NewDecoder to decode the stream directly into the struct
	err := json.NewDecoder(req.Body).Decode(&email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		fmt.Print(email)
	}

	// Inform client that the response type is JSON
	w.Header().Set("Content-Type", "application/json")
    // Set the HTTP status code (optional, http.StatusOK is 200).
	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/test", test)
	http.HandleFunc("/headers", headers)
	http.HandleFunc("/post-email", postEmail)
	http.HandleFunc("/get-room-code", getRoomCode)
	http.HandleFunc("/verify-room-code", verifyRoomCode)
	http.HandleFunc("/start-room", verifyStart)
    http.ListenAndServe(":8090", nil)
}