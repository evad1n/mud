package main

import (
	"fmt"
	"log"
	"os"
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
		// Check for closed connection
		if ev.Command.Category == "END_CONN" {
			if ev.Player.Chan != nil {
				ev.Player.disconnect()
			}
			// Already shutting down -> ignore
			continue
		}
		// serverLog to server
		eventLog.Printf("PLAYER: %s | COMMAND: %s | PARAMS: %s\n", ev.Player.Name, ev.Command.Name, ev.Params)
		// Run command
		ev.Command.Run(ev.Player, ev.Params)
		// Prompt
		fmt.Fprintf(ev.Player.Conn, "\n>>> ")
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
