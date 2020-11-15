package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
)

type player struct {
	Y      int    // Y position of the player
	X      int    // X position
	ID     string // a unique id to identify the player by the frontend
	Online bool
	socket *websocket.Conn // websocket connection of the player
}

var players = make(map[string]*player)
var mutex sync.RWMutex

func remoteHandler(res http.ResponseWriter, req *http.Request) {
	var err error

	//when someone requires a ws connection we create a new player and store a
	// pointer to the connection inside player.Socket
	ws, err := websocket.Upgrade(res, req, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(res, "not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}

	log.Printf("got websocket conn from %v\n", ws.RemoteAddr())
	id := uuid.New()
	player := new(player)
	player.socket = ws
	player.ID = id
	player.Online = true

	mutex.Lock()
	players[id] = player
	mutex.Unlock()

	mutex.RLock()
	for k := range players {
		if k != player.ID {
			if err = player.socket.WriteJSON(players[k]); err != nil {
				log.Println(err)
			}
		}
	}
	mutex.RUnlock()

	go func() {
		for {
			if err = player.socket.ReadJSON(&player); err != nil {
				log.Println("player Disconnected waiting", err)
				disconnectHandler(player)
				return
			}

			mutex.RLock()
			for k := range players {
				if k != player.ID {
					if err = players[k].socket.WriteJSON(player); err != nil {
						log.Println(err)
					}
				}
			}
			mutex.RUnlock()
		}
	}()
}

func disconnectHandler(p *player) {
	p.Online = false
	mutex.RLock()
	for k := range players {
		if k != p.ID {
			if err := players[k].socket.WriteJSON(p); err != nil {
				log.Println(err)
			}
		}
	}
	mutex.RUnlock()

	mutex.Lock()
	delete(players, p.ID)
	mutex.Unlock()
	log.Println("number of players still connected ...", len(players))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ws", remoteHandler)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	http.ListenAndServe(":3000", r)
}
