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
	// All players on the server
	players map[string]*player
)

// Erase player's old prompt
func (p *player) erasePrompt() {
	// Move up 3 lines and then clear to bottom
	fmt.Fprint(p.conn, "\x1b[3A\x1b[1000D\x1b[0J")
	// Move back to bottom
	fmt.Fprint(p.conn, "\x1b[1B")
}

// Display player command prompt
func (p *player) prompt() {
	fmt.Fprintf(p.conn, "%s\n>>> ", strings.Repeat("_", 40))
}

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
		// Erase player's old prompt
		ev.player.erasePrompt()
		// Check for closed connection
		if ev.end {
			if ev.player.events != nil {
				ev.player.disconnect()
			}
			// Already shutting down -> ignore
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
				fmt.Fprintln(ev.player.conn, "Unrecognized command!")
			}
		}
		// Show another prompt if player is still playing
		if ev.player.events != nil {
			ev.player.prompt()
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
		zone:      rooms[3001].zone,
		room:      rooms[3001],
	}
	// Add to data
	players[p.name] = p
	p.zone.players = append(p.zone.players, p)
	p.room.players = append(p.room.players, p)

	return p, nil
}
