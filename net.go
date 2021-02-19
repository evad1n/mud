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
	if err != nil {
		serverLog.Fatalf("Error starting server on port %s: %v", port, err)
	}
	defer server.Close()
	serverLog.Printf("Listening for connections on %s:%s\n", serverAddress, port)
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
	clientLog := log.New(conn, "CLIENT: ", log.Ldate|log.Ltime)
	fmt.Fprintln(conn)
	clientLog.Printf("Connected to MUD server on %s:%s\n\n", serverAddress, port)

	// Log connection to server
	serverLog.Printf("Client connected from %s", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)

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
		p, err = createPlayer(name, conn, clientLog)
	}

	// Player object is initialized
	go p.listenMUD()

	p.joinServer()

	fmt.Fprintf(p.conn, "\nHello, %s! Welcome to MUD!\n\n\n", p.name)

	serverLog.Printf("player '%s' joined the MUD from %s", p.name, conn.RemoteAddr().String())

	// Add delay to show welcome msg before entering prompt
	time.Sleep(1000 * time.Millisecond)

	fmt.Fprint(p.conn, "\x1b[2J")

	p.printLocation()

	p.events <- event{
		player:      nil,
		output:      "Type 'help' to see all available commands!",
		unsolicited: true,
	}

	for scanner.Scan() {
		// Send raw input as command to be parsed
		inputs <- input{p, scanner.Text(), false}
	}
	if err := scanner.Err(); err != nil {
		if p.events != nil {
			serverLog.Printf("Client %s connection error: %v", conn.RemoteAddr().String(), err)
		}
		// Ignore
	}
	// Connection has been closed
	inputs <- input{p, "", true}
}

// Have a client listen for mud events
func (p *player) listenMUD() {
	defer p.conn.Close()
	for ev := range p.events {
		if ev.updateMap {
			p.visited[p.room.id] = true
			p.minimap.trace(p.room, p.visited)
		}
		if ev.player != p {
			ev.unsolicited = true
		}
		if ev.command != nil {
			// Color output based on command effect
			switch ev.command.category {
			case emotes:
				ev.output = ansiWrap(ev.output, ansiColors["white"])
			case nav:
				ev.output = ansiWrap(ev.output, ansiColors["cyan"])
			}
		}
		if ev.err {
			ev.output = ansiWrap(ev.output, ansiColors["red"])
		}
		p.eventPrint(ev)
		time.Sleep(time.Duration(ev.delay) * time.Millisecond)
	}
	// Clear screen
	fmt.Fprint(p.conn, "\x1b[2J")
	fmt.Fprintf(p.conn, "Goodbye %s!\nThanks for playing!\n", p.name)
	p.log.Printf("Disconnected from MUD server on %s:%s\n", serverAddress, port)
	playTime := time.Now().Sub(p.beginTime)
	h, m := int(math.Round(playTime.Hours())), int(math.Round(playTime.Minutes()))%60
	p.log.Printf("You played for %s %s and %s %s", ansiWrap(fmt.Sprint(h), ansiColors["green"]), plural(h, "hour"), ansiWrap(fmt.Sprint(m), ansiColors["green"]), plural(m, "minute"))

	serverLog.Printf("player '%s' connection terminated\n", p.name)
}

// Terminate a connection and remove the player from the world data
func (p *player) disconnect() {
	// Notify players in room
	for _, other := range p.room.players {
		if other != p {
			other.events <- event{
				player: p,
				output: fmt.Sprintf("%s has left the room", p.name),
			}
		}
	}
	// Notify players on server of player leaving
	for _, other := range players {
		if other != p {
			other.events <- event{
				player: p,
				output: fmt.Sprintf("%s has left the server", p.name),
			}
		}
	}
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
