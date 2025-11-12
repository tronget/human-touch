package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// return r.Header.Get("Origin") == "http://gateway:8000"
		return true
	},
}

type client struct {
	conn   *websocket.Conn
	userID int64
}

var clients = struct {
	sync.RWMutex
	m map[int64][]*client
}{
	m: make(map[int64][]*client),
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	uid, err := ValidateToken(token)

	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		log.Println("upgrading to websocket: ", err)
		return
	}

	c := &client{conn: conn, userID: uid}

	clients.Lock()
	clients.m[uid] = append(clients.m[uid], c)
	clients.Unlock()

	log.Printf("User %d connected via WebSocket\n", uid)
	defer log.Printf("User %d disconnected from WebSocket\n", uid)

	// read loop
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("error reading from websocket: ", err)
			break
		}
	}
	conn.Close()

	// remove client from the list
	clients.Lock()
	arr := clients.m[uid]
	for i, cl := range arr {
		if cl == c {
			clients.m[uid] = append(arr[:i], arr[i+1:]...)
			break
		}
	}
	if len(clients.m[uid]) == 0 {
		delete(clients.m, uid)
	}
	clients.Unlock()
}

func BroadcastToUser(userID int64, payload any) {
	clients.RLock()
	arr := clients.m[userID]
	clients.RUnlock()

	for _, c := range arr {
		if err := c.conn.WriteJSON(payload); err != nil {
			log.Println("error writing to websocket: ", err)
			c.conn.Close()
		}
	}
}

func ValidateToken(token string) (int64, error) {
	if token == "" {
		return 0, errors.New("missing token")
	}

	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		return 0, errors.New("missing `JWT_SECRET` in environment variables")
	}

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return secret, nil
	})

	if err != nil || !parsedToken.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	uid, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user id")
	}

	return int64(uid), nil
}
