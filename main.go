package main

import (
	"golang-websocket-simple-broadcast-message/helper"
	"net/http"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./template/home.gohtml")
}

func main() {
	newBroadcast := NewBroadcast()
	go newBroadcast.Run()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		serveWs(newBroadcast, writer, request)
	})
	http.HandleFunc("/send", func(writer http.ResponseWriter, request *http.Request) {
		messageHandler(newBroadcast, writer, request)
	})
	err := http.ListenAndServe("localhost:3000", nil)
	helper.PanicIfError(err)
}
