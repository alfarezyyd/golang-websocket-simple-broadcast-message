package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"golang-websocket-simple-broadcast-message/helper"
	"log"
	"net/http"
	"time"
)

const (
	pongWait      = 10 * time.Second
	pingPeriod    = 5 * time.Second
	writeDeadline = 10 * time.Second
)

var updgraderConn = &websocket.Upgrader{}

type Client struct {
	clientBroadcast *Broadcast
	websocketConn   *websocket.Conn
	sendMessage     chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.clientBroadcast.unregisterClient <- c
		err := c.websocketConn.Close()
		helper.PanicIfError(err)
	}()

	err := c.websocketConn.SetReadDeadline(time.Now().Add(pongWait))
	helper.PanicIfError(err)

	c.websocketConn.SetPongHandler(func(string) error {
		fmt.Println("Received Pong")
		err = c.websocketConn.SetReadDeadline(time.Now().Add(pongWait))
		helper.PanicIfError(err)
		return nil
	})

	for {
		_, _, err = c.websocketConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Error", err)
			}
			log.Println("Error", err)
			break
		}
	}
}

func (c *Client) writePump() {
	newTicker := time.NewTicker(pingPeriod)
	defer func() {
		newTicker.Stop()
		c.websocketConn.Close()
	}()

	for {
		select {
		case messageData, stateData := <-c.sendMessage:
			err := c.websocketConn.SetWriteDeadline(time.Now().Add(pingPeriod))
			helper.PanicIfError(err)
			if !stateData {
				log.Println("Close the channel")
				return
			}
			err = c.websocketConn.WriteMessage(websocket.TextMessage, messageData)
			helper.PanicIfError(err)
			err = c.websocketConn.SetWriteDeadline(time.Time{})
			helper.PanicIfError(err)
		case <-newTicker.C:
			log.Println("Send Ping")
			err := c.websocketConn.SetWriteDeadline(time.Now().Add(writeDeadline))
			helper.PanicIfError(err)

			c.websocketConn.WriteMessage(websocket.PingMessage, nil)
			err = c.websocketConn.SetWriteDeadline(time.Time{})
			helper.PanicIfError(err)
		}
	}
}

func serveWs(broadcast *Broadcast, w http.ResponseWriter, r *http.Request) {
	websocketConn, err := updgraderConn.Upgrade(w, r, nil)
	helper.PanicIfError(err)

	newClient := &Client{
		clientBroadcast: broadcast,
		websocketConn:   websocketConn,
		sendMessage:     make(chan []byte, 256),
	}
	newClient.clientBroadcast.registerClient <- newClient
	go newClient.writePump()
	go newClient.readPump()
}

func messageHandler(broadcast *Broadcast, w http.ResponseWriter, r *http.Request) {
	broadcast.broadcastMessage <- []byte("Send Message")
}
