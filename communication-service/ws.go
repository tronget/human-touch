package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type client struct {
	conn   *websocket.Conn
	userID int64
}

var clients = struct {
	sync.RWMutex
	m map[int64][]*client
}{m: make(map[int64][]*client)}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// в реале проверить JWT из query/header
	uid := r.Context().Value("user_id").(int64)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &client{conn: conn, userID: uid}
	clients.Lock()
	clients.m[uid] = append(clients.m[uid], c)
	clients.Unlock()
	// read loop
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
	conn.Close()
}

func BroadcastToUser(userID int64, payload any) {
	clients.RLock()
	arr := clients.m[userID]
	clients.RUnlock()
	for _, c := range arr {
		c.conn.WriteJSON(payload)
	}
}
