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
	defer conn.Close()
	clientLog := log.New(conn, "CLIENT: ", log.Ldate|log.Ltime)
	fmt.Fprintln(conn)
	clientLog.Printf("Connected to MUD server on %s:%s\n\n", serverAddress, port)

	// Log connection to server
	serverLog.Printf("Client connected from %s", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)

	// Player event channel
	out := make(chan Output)

	// Validate username loop
	var (
		player *Player
		err    error
	)
	for badName := true; badName; badName = (err != nil) {
		if err != nil {
			fmt.Fprintln(conn, err)
		}
		fmt.Fprint(conn, "Please enter your name: ")
		scanner.Scan()
		name := scanner.Text()
		if len(strings.Fields(name)) > 1 {
			err = fmt.Errorf("Username must be one word")
			continue
		}
		if len(name) > 20 {
			err = fmt.Errorf("Username must be less than 21 characters")
			continue
		}
		player, err = createPlayer(name, conn, clientLog, out)
	}
	fmt.Fprintf(conn, "\nHello, %s! Welcome to MUD!\n\n\n", player.Name)

	serverLog.Printf("Player '%s' joined the MUD from %s", player.Name, conn.RemoteAddr().String())

	time.Sleep(1 * time.Second)

	player.printLocation()
	fmt.Fprintln(conn, "Type 'help' to see all available commands!")

	go player.listenMUD()

	// Initial prompt
	player.prompt()
	for scanner.Scan() {
		// Add a newline after commands
		fmt.Fprintln(conn)

		in <- Input{player, scanner.Text(), false}
	}
	// FIX: idk this doesn't look good
	if err := scanner.Err(); err != nil {
		// Ignore
		// serverLog.Printf("Client (%s) connection error: %v", conn.RemoteAddr().String(), err)
	}
	// Connection has been closed
	in <- Input{player, "", true}
}

// Have a client listen for mud events
func (p *Player) listenMUD() {
	defer p.Conn.Close()
	for ev := range p.Chan {
		// Erase old prompt
		p.erasePrompt()
		fmt.Fprintln(p.Conn, ev.Effect)
		// New prompt
		p.prompt()
	}
	p.Log.Printf("Disconnected from MUD server on %s:%s\n", serverAddress, port)
	playTime := time.Now().Sub(p.Begin)
	h, m := int(math.Round(playTime.Hours())), int(math.Round(playTime.Minutes()))%60
	p.Log.Printf("You played for %d %s and %d %s", h, plural(h, "hour"), m, plural(m, "minute"))
	// Close connection
	serverLog.Printf("Player '%s' connection terminated\n", p.Name)
}

// Terminate a connection
func (p *Player) disconnect() {
	// Shut down channel
	close(p.Chan)
	p.Chan = nil
	// Remove player
	p.Room.removePlayer(p)
	p.Zone.removePlayer(p)
	delete(players, p.Name)
	// Log to server
	serverLog.Printf("Player '%s' disconnected from %s\n", p.Name, p.Conn.RemoteAddr().String())
	// Connection will automatically close after channel is closed
}
