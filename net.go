package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net"
	"strings"
	"time"
)

// Listen for incoming client connections
func listenConnections(in chan Input) {
	server, err := net.Listen("tcp", ":"+port)
	defer server.Close()
	if err != nil {
		serverLog.Fatalf("Error starting server on port %s: %v", port, err)
	}
	for {
		conn, err := server.Accept()
		if err != nil {
			serverLog.Fatalf("Error accepting connection: %v", err)
		}
		go handleConnection(conn, in)
	}
}

// Handle a client connection with their own command loop
func handleConnection(conn net.Conn, in chan Input) {
	clientLog := log.New(conn, "CLIENT: ", log.Ldate|log.Ltime)
	fmt.Fprintln(conn)
	clientLog.Printf("Connected to MUD server on %s:%s\n\n", serverAddress, port)

	// Log connection to server
	serverLog.Printf("Client connected from %s", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)

	fmt.Fprint(conn, "Please enter your name: ")
	scanner.Scan()
	name := scanner.Text()
	fmt.Fprintf(conn, "\nHello, %s! Welcome to MUD!\n\n\n", name)

	serverLog.Printf("Player '%s' joined the MUD from %s", name, conn.RemoteAddr().String())

	time.Sleep(1 * time.Second)

	// Init player
	out := make(chan Output)
	player := &Player{
		Name:  name,
		Conn:  conn,
		Log:   clientLog,
		Chan:  out,
		Begin: time.Now(),
		Zone:  rooms[3001].Zone,
		Room:  rooms[3001],
	}
	// Add to data
	players[player.Name] = player
	player.Zone.Players = append(player.Zone.Players, player)
	player.Room.Players = append(player.Room.Players, player)
	printLocation(player)
	fmt.Fprintln(conn, "Type 'help' to see all available commands!")

	go player.listenMUD()

	// FIX: network failure needs to cause dc
	// Initial prompt
	fmt.Fprintf(conn, "\n>>> ")
	for scanner.Scan() {
		// Add a newline after commands
		fmt.Fprintln(conn)

		if words := strings.Fields(scanner.Text()); len(words) > 0 {
			// Check if cmd exists
			if validCmd, exists := commands[strings.ToLower(words[0])]; exists {
				in <- Input{player, validCmd, strings.Join(words[1:], " ")}
			} else {
				fmt.Fprintln(conn, "Unrecognized command!")
			}
		}
	}
	if err := scanner.Err(); err != nil {
		// Connection has been closed
		in <- Input{player, Command{"END_CONN", "END_CONN", "Terminates the connection", nil}, "Connection successfully terminated"}
	}
}

// Have a client listen for mud events
func (p *Player) listenMUD() {
	for ev := range p.Chan {
		fmt.Fprintln(p.Conn)
		fmt.Fprintln(p.Conn, ev.Effect)
		// Prompt
		fmt.Fprintf(p.Conn, "\n>>> ")
	}
	fmt.Fprintf(p.Conn, "\n\n")
	p.Log.Printf("Disconnected from MUD server on %s:%s\n", serverAddress, port)
	playTime := time.Now().Sub(p.Begin)
	h, m := int(math.Round(playTime.Hours())), int(math.Round(playTime.Minutes()))%60
	p.Log.Printf("You played for %d %s and %d %s", h, plural(h, "hour"), m, plural(m, "minute"))
	// Log to server
	serverLog.Printf("Player '%s' disconnected from %s\n", p.Name, p.Conn.RemoteAddr().String())
	// Close connection
	p.Conn.Close()
}
