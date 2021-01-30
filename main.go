package main

import (
	"log"
	"net"
	"os"
	"strings"
)

// Player represents the user and all associated state
type Player struct {
	// Username
	Name string
	// Connection
	Conn net.Conn
	// MUD event channel
	Chan chan Output
	// The current zone
	Zone *Zone
	// The current room
	Room *Room
}

const (
	port = "9001"
)

var (
	serverAddress string
	serverLog     *log.Logger
	eventLog      *log.Logger
	// All players on the server
	players map[string]*Player
)

func main() {
	// Get local IP
	serverAddress = getLocalAddress()

	// Create server log
	serverLog = log.New(os.Stdout, "SERVER: ", log.Ldate|log.Ltime)
	serverLog.SetPrefix("SERVER: ")
	serverLog.Printf("Starting server on %s:%s...\n", serverAddress, port)

	serverLog.Println("Initializing world...")
	initWorld()

	// Create global players list
	players = make(map[string]*Player)

	// Client input channel
	in := make(chan Input)

	serverLog.Println("Listening for connections...")
	go listenConnections(in)

	// Create event log
	eventLog = log.New(os.Stdout, "EVENT: ", log.Ldate|log.Ltime)

	for {
		// Get input
		ev := <-in
		// Check for closed connection to ignore
		if ev.Command.Category == "END" {
			if ev.Player.Chan != nil {
				close(ev.Player.Chan)
			} else {
				serverLog.Printf("Player '%s' connection terminated\n", ev.Player.Name)
			}
			continue
		}
		// serverLog to server
		eventLog.Printf("PLAYER: %s | COMMAND: %s | PARAMS: %s\n", ev.Player.Name, ev.Command.Name, ev.Params)
		// Run command
		ev.Command.Run(ev.Player, ev.Params)
	}
}

// Initialize world, commands, and lookup tables
func initWorld() {
	createMaps()
	defaultCommands()
	if err := loadDB(); err != nil {
		serverLog.Fatal(err)
	}
}

// Return the index of an element that satisfies the predicate.
// If none can be found then returns -1
func index(length int, predicate func(idx int) bool) int {
	for i := 0; i < length; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

// // Removes a player from all data
// func removePlayer(p *Player) {
// 	// Remove from global list
// 	if i := index(len(players), func(idx int) bool { return players[idx] == p }); i != -1 {
// 		players = append(players[:i], players[i+1:]...)
// 	}
// }

// Centers text in the middle of a column of size {size}
func centerText(text string, size int, fill rune) string {
	if &fill == nil {
		fill = ' '
	}
	size -= len(text)
	front := size / 2
	return strings.Repeat(string(fill), front) + text + strings.Repeat(string(fill), size-front)
}
