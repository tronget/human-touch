package ws

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/tronget/human-touch/communication-service/internal/constants"
	"github.com/tronget/human-touch/communication-service/internal/models"
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

func WsHandler(w http.ResponseWriter, r *http.Request) {
	uid, err := UserIDFromRequest(r)
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

func UserIDFromRequest(r *http.Request) (int64, error) {
	if user, ok := r.Context().Value(constants.CtxUserKey).(models.User); ok && user.ID != 0 {
		return user.ID, nil
	}

	return 0, errors.New("user not found in request's context")
}
