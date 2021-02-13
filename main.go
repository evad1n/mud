package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	port = "9001"
)

var (
	serverAddress string
	serverLog     *log.Logger
	eventLog      *log.Logger
	players       map[string]*player // All players on the server
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
	players = make(map[string]*player)

	// Client input channel
	inputs := make(chan input)

	serverLog.Println("Listening for connections...")
	go listenConnections(inputs)

	// Create event log
	eventLog = log.New(os.Stdout, "EVENT: ", log.Ldate|log.Ltime)

	for {
		// Get input
		ev := <-inputs
		// Check for closed connection
		if ev.end {
			if ev.player.events != nil {
				ev.player.disconnect()
			} else {
				// Already shutting down -> ignore
				serverLog.Printf("player '%s' connection already closed\n", ev.player.name)
			}
			continue
		}
		// Otherwise process commands
		if words := strings.Fields(ev.text); len(words) > 0 {
			// Check if cmd exists
			if validCmd, exists := commands[strings.ToLower(words[0])]; exists {
				params := strings.Join(words[1:], " ")
				// Log to server
				eventLog.Printf("PLAYER: %s | COMMAND: %s | PARAMS: %s\n", ev.player.name, validCmd.name, params)
				// Actually run the command
				validCmd.run(ev.player, params)
			} else {
				ev.player.events <- event{
					player: ev.player,
					output: "Unrecognized command!",
					err:    true,
				}
			}
		}
	}
}

// Initialize world, commands, and lookup tables
func initWorld() {
	createMaps()
	defaultCommands()
	if err := loadWorld(); err != nil {
		serverLog.Fatal(fmt.Errorf("loading world from database: %v", err))
	}
}

// Create a player and add it to the global players map
func createPlayer(name string, conn net.Conn, log *log.Logger, out chan event) (*player, error) {
	if _, exists := players[name]; exists {
		return nil, fmt.Errorf("That username is taken")
	}

	p := &player{
		name:      name,
		conn:      conn,
		log:       log,
		events:    out,
		beginTime: time.Now(),
		zone:      nil,
		room:      nil,
		minimap:   newMapBuilder(4),
		visited:   make(map[int]bool),
	}
	// Add to data
	players[p.name] = p

	return p, nil
}

// Move to starting room (Temple of Midgaard)
// Notify other players
func (p *player) joinServer() {
	r := rooms[3001]

	// Notify players on server of new join
	for _, other := range players {
		if other != p {
			other.events <- event{
				player: p,
				output: fmt.Sprintf("%s has joined the server", p.name),
			}
		}
	}

	// Notify other players in starting room
	for _, other := range r.players {
		other.events <- event{
			player: p,
			output: fmt.Sprintf("%s has entered the room", p.name),
		}
	}

	p.room = r
	p.room.players = append(r.players, p)
	p.zone = r.zone
	p.zone.players = append(r.zone.players, p)

	p.room.sortPlayers()

	// Update map
	p.visited[p.room.id] = true
	p.minimap.trace(p.room, p.visited)
}
