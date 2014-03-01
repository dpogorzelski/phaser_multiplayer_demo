package main

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// a slice of *Players which will store the list of connected players
var Players = make([]*Player, 0)

type Player struct {
	Y      int             // Y position of the player
	X      int             // X position
	Id     string          // a unique id to identify the player by the frontend
	Socket *websocket.Conn // pointer to websocket connection to the player
}

// each time the position of a player changes the backend will receive
// one of these
type Message struct {
	Y   int    // current Y position
	X   int    // as above
	Id  string // the id of the player that sent the message
	New bool   // true if this player just connected so we know when to
	// spawn a new sprite on the screens of other players. for all subsequent
	// messages is false
	Online bool // if the player is no longer connected we let the frontend
	// know to remove the sprite
}

func (p *Player) position(new bool) Message {
	return Message{X: p.X, Y: p.Y, Id: p.Id, New: new, Online: true}
}

// the core. it upgrades connections to websocket and handles the exchange of
// players positions
func remoteHandler(res http.ResponseWriter, req *http.Request) {
	var err error

	//when someone requires a ws connection we create a new player and store a
	// pointer to the connection inside player.Socket
	ws, _ := websocket.Upgrade(res, req, nil, 1024, 1024)
	log.Printf("got websocket conn from %v\n", ws.RemoteAddr())
	player := new(Player)
	player.Id = uuid.New()
	player.Socket = ws

	// we broadcast the position of the new player to alredy connected
	// players (if any) and viceversa, we tell the player where to spawn already
	// existing players
	log.Println("Publishing positions")
	for _, p := range Players {
		if p.Socket.RemoteAddr() != player.Socket.RemoteAddr() {
			player.Socket.WriteJSON(p.position(true))
			p.Socket.WriteJSON(player.position(true))
		}
	}

	// we append the new player to Players slice
	Players = append(Players, player)
	for {
		// if a network error occurs (aka someone closed the game) we let
		// the other players know to despawn his sprite (Online: false) and
		// remove him from the slice so no further updates will be sent to him
		if err = player.Socket.ReadJSON(&player); err != nil {
			log.Println("Player Disconnected waiting", err.Error())
			for i, p := range Players {
				if p.Socket.RemoteAddr() == player.Socket.RemoteAddr() {
					Players = append(Players[:i], Players[i+1:]...)
				} else {
					log.Println("destroy player", player)
					p.Socket.WriteJSON(Message{Online: false, Id: player.Id})
				}
			}
			log.Println("Number of players still connected ...", len(Players))
			return
		}

		// a regular broadcast to inform all the players about a player's
		// position update
		for _, p := range Players {
			if p.Socket.RemoteAddr() != player.Socket.RemoteAddr() {
				p.Socket.WriteJSON(player.position(false))
			}

		}
	}
}

// it all starts here
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ws", remoteHandler)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	http.ListenAndServe(":3000", r)
}
