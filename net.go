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
func listenConnections(inputs chan input) {
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
		go handleConnection(conn, inputs)
	}
}

// Handle a client connection with their own command loop
func handleConnection(conn net.Conn, inputs chan input) {
	defer conn.Close()
	clientLog := log.New(conn, "CLIENT: ", log.Ldate|log.Ltime)
	fmt.Fprintln(conn)
	clientLog.Printf("Connected to MUD server on %s:%s\n\n", serverAddress, port)

	// Log connection to server
	serverLog.Printf("Client connected from %s", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)

	// player event channel
	out := make(chan event)

	// Validate username loop
	var (
		p   *player
		err error
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
		p, err = createPlayer(name, conn, clientLog, out)
	}

	fmt.Fprintf(conn, "\nHello, %s! Welcome to MUD!\n\n\n", p.name)

	serverLog.Printf("player '%s' joined the MUD from %s", p.name, conn.RemoteAddr().String())

	time.Sleep(1 * time.Second)

	p.printLocation()
	fmt.Fprintln(conn, "Type 'help' to see all available commands!")

	go p.listenMUD()

	// Initial prompt
	p.prompt()
	for scanner.Scan() {
		// Add a newline after commands
		fmt.Fprintln(conn)

		inputs <- input{p, scanner.Text(), false}
	}
	if err := scanner.Err(); err != nil {
		serverLog.Printf("Client (%s) connection error: %v", conn.RemoteAddr().String(), err)
		// Ignore
	}
	// Connection has been closed
	inputs <- input{p, "", true}
}

// Have a client listen for mud events
func (p *player) listenMUD() {
	defer p.conn.Close()
	for ev := range p.events {
		// Erase old prompt
		p.erasePrompt()
		fmt.Fprintln(p.conn, ev.effect)
		// New prompt
		p.prompt()
	}
	p.log.Printf("Disconnected from MUD server on %s:%s\n", serverAddress, port)
	playTime := time.Now().Sub(p.beginTime)
	h, m := int(math.Round(playTime.Hours())), int(math.Round(playTime.Minutes()))%60
	p.log.Printf("You played for %d %s and %d %s", h, plural(h, "hour"), m, plural(m, "minute"))
	// Close connection
	serverLog.Printf("player '%s' connection terminated\n", p.name)
}

// Terminate a connection
func (p *player) disconnect() {
	// Shut down channel
	close(p.events)
	p.events = nil
	// Remove player
	p.room.removePlayer(p)
	p.zone.removePlayer(p)
	delete(players, p.name)
	// Log to server
	serverLog.Printf("player '%s' disconnected from %s\n", p.name, p.conn.RemoteAddr().String())
	// Connection will automatically close after channel is closed
}
