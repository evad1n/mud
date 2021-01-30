package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Input represents an event going from the player to MUD
type Input struct {
	Player  *Player
	Command Command
	Params  string
}

// Output represents an event going from MUD to the player
type Output struct {
	Player *Player
	Effect string
}

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
	fmt.Println(conn)
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
	player := Player{name, conn, out, rooms[3001].Zone, rooms[3001]}
	// Add to data
	players[player.Name] = &player
	player.Zone.Players = append(player.Zone.Players, &player)
	player.Room.Players = append(player.Room.Players, &player)
	printLocation(&player)
	fmt.Fprintln(conn, "Type 'cmds' or 'help' to see all available commands!")

	go listenMUD(&player, clientLog)

	// FIX: network failure needs to cause dc
	// Initial prompt
	fmt.Fprintf(conn, "\n>>> ")
	for scanner.Scan() {
		// Add a newline after commands
		fmt.Fprintln(conn)

		if words := strings.Fields(scanner.Text()); len(words) > 0 {
			// Check if cmd exists
			if validCmd, exists := commands[strings.ToLower(words[0])]; exists {
				in <- Input{&player, validCmd, strings.Join(words[1:], " ")}
			} else {
				fmt.Fprintln(conn, "Unrecognized command!")
			}
		}

		// Prompt
		fmt.Fprintf(conn, "\n>>> ")
	}
	if err := scanner.Err(); err != nil {
		serverLog.Println("BAD")
		// Connection has been closed
		in <- Input{&player, Command{"END_CONN", "END", nil}, "Connection successfully terminated"}
	}
}

// Have a client listen for mud events
func listenMUD(p *Player, clientLog *log.Logger) {
	for ev := range p.Chan {
		fmt.Fprintln(p.Conn, ev.Effect)
	}
	clientLog.Printf("Disconnected from MUD server on %s:%s\n", serverAddress, port)
	// Log to server
	serverLog.Printf("Player '%s' disconnected from %s\n", p.Name, p.Conn.RemoteAddr().String())
	// Close connection
	p.Conn.Close()
}

func getLocalAddress() string {
	var localaddress string

	ifaces, err := net.Interfaces()
	if err != nil {
		panic("getLocalAddress: failed to find network interfaces")
	}

	// find the first non-loopback interface with an IPv4 address
	for _, elt := range ifaces {
		if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
			addrs, err := elt.Addrs()
			if err != nil {
				panic("getLocalAddress: failed to get addresses for network interface")
			}

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
						localaddress = ip4.String()
						break
					}
				}
			}
		}
	}
	if localaddress == "" {
		panic("localaddress: failed to find non-loopback interface with valid IPv4 address")
	}

	return localaddress
}
