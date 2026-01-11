package main

import (
    "net/http"
)

func main() {
	http.HandleFunc("/test", test)
	http.HandleFunc("/headers", headers)
	http.HandleFunc("/post-email", postEmail)
	http.HandleFunc("/get-room-code", getRoomCode)
	http.HandleFunc("/verify-room-code", verifyRoomCode)
	http.HandleFunc("/start-room", verifyStart)
    http.ListenAndServe(":8090", nil)
}