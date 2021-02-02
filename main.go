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
	players map[string]*Player
)

// Erase player's old prompt
func (p *Player) erasePrompt() {
	// Move up 3 lines and then clear to bottom
	fmt.Fprint(p.Conn, "\x1b[3A\x1b[1000D\x1b[0J")
	// Move back to bottom
	fmt.Fprint(p.Conn, "\x1b[1B")
}

// Display player command prompt
func (p *Player) prompt() {
	fmt.Fprintf(p.Conn, "%s\n>>> ", strings.Repeat("_", 40))
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
		// Erase player's old prompt
		ev.Player.erasePrompt()
		// Check for closed connection
		if ev.End {
			if ev.Player.Chan != nil {
				ev.Player.disconnect()
			}
			// Already shutting down -> ignore
			continue
		}
		// Otherwise process commands
		if words := strings.Fields(ev.Text); len(words) > 0 {
			// Check if cmd exists
			if validCmd, exists := commands[strings.ToLower(words[0])]; exists {
				params := strings.Join(words[1:], " ")
				// Log to server
				eventLog.Printf("PLAYER: %s | COMMAND: %s | PARAMS: %s\n", ev.Player.Name, validCmd.Name, params)
				// Actually run the command
				validCmd.Run(ev.Player, params)
			} else {
				fmt.Fprintln(ev.Player.Conn, "Unrecognized command!")
			}
		}
		// Show another prompt if player is still playing
		if ev.Player.Chan != nil {
			ev.Player.prompt()
		}

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

// Create a player and add it to the global players map
func createPlayer(name string, conn net.Conn, log *log.Logger, out chan Output) (*Player, error) {
	if _, exists := players[name]; exists {
		return nil, fmt.Errorf("That username is taken")
	}
	player := &Player{
		Name:  name,
		Conn:  conn,
		Log:   log,
		Chan:  out,
		Begin: time.Now(),
		Zone:  rooms[3001].Zone,
		Room:  rooms[3001],
	}
	// Add to data
	players[player.Name] = player
	player.Zone.Players = append(player.Zone.Players, player)
	player.Room.Players = append(player.Room.Players, player)

	return player, nil
}
