package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// Player represents the user and all associated state
type Player struct {
	// Username
	Name string
	// Connection
	Conn net.Conn
	// MUD event channel
	Chan chan Output
	// The current room
	Location *Room
}

// Input represents an event going from the player to MUD
type Input struct {
	Player  *Player
	Command Command
	Params  string
}

// Output represents an event going from MUD to the player
type Output struct {
	Player *Player
	Params string
}

const (
	port = "9001"
)

var (
	serverAddress string
	serverLog     *log.Logger
	eventLog      *log.Logger
	players       []Player
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

	// Create players slice
	players = []Player{}

	// Client input channel
	in := make(chan Input)

	serverLog.Println("Listening for connections...")
	go listenConnections(in)

	// Create event log
	eventLog = log.New(os.Stdout, "EVENT: ", log.Ldate|log.Ltime)

	for {
		// Get input
		ev := <-in
		// serverLog to server
		eventLog.Printf("PLAYER: %s | COMMAND: %s | PARAMS: %s\n", ev.Player.Name, ev.Command.Name, ev.Params)
		// Run command
		ev.Command.Run(ev.Player, ev.Params)
		// Send to all players
		for _, p := range players {
			if ch := p.Chan; ch != nil {
				ch <- Output{ev.Player, ev.Params}
			} else {
				// Shut down player
			}
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

// Listen for incoming client connections
func listenConnections(ch chan Input) {
	server, err := net.Listen("tcp", ":"+port)
	if err != nil {
		serverLog.Fatalf("Error starting server on port %s: %v", port, err)
	}
	for {
		conn, err := server.Accept()
		if err != nil {
			serverLog.Fatalf("Error accepting connection: %v", err)
		}
		go handleConnection(conn, ch)
	}
}

// Handle a client connection with their own command loop
func handleConnection(conn net.Conn, ch chan Input) {
	clientLog := log.New(conn, "CLIENT: ", log.Ldate|log.Ltime)
	clientLog.Printf("Connected to MUD server on %s:%s\n", serverAddress, port)

	scanner := bufio.NewScanner(conn)

	fmt.Fprint(conn, "Please enter your name: ")
	scanner.Scan()
	name := scanner.Text()
	fmt.Fprintf(conn, "Hello, %s! Welcome to MUD!\n\n", name)
	time.Sleep(2 * time.Second)

	// Log connection to server
	serverLog.Printf("Player '%s' connected from %s", name, conn.RemoteAddr().String())

	// Init player
	out := make(chan Output)
	player := Player{name, conn, out, rooms[3001]}
	players = append(players, player)
	printLocation(&player)
	fmt.Fprintln(conn, "Type 'cmds' or 'help' to see all available commands!")

	go listenMUD(out, conn)

	// Initial prompt
	fmt.Fprintf(conn, "\n>>> ")
	for scanner.Scan() {
		// Add a newline after commands
		fmt.Fprintln(conn)

		if words := strings.Fields(scanner.Text()); len(words) > 0 {
			// Check if cmd exists
			if validCmd, exists := commands[strings.ToLower(words[0])]; exists {
				ch <- Input{&player, validCmd, strings.Join(words[1:], " ")}
			} else {
				fmt.Fprintln(conn, "Unrecognized command!")
			}
		}

		// Prompt
		fmt.Fprintf(conn, "\n>>> ")
	}
	if err := scanner.Err(); err != nil {
		clientLog.Fatalf("Error processing commands: %v", err)
	}
}

// Have a client listen for mud events
func listenMUD(ch chan Output, conn net.Conn) {
	for {
		ev := <-ch
		fmt.Fprintln(conn, ev.Params)
	}
}

func commandLoop() error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Welcome to MUD!\nPlease enter your name: ")
	name := scanner.Text()
	fmt.Printf("Hello, %s!", name)
	// Set start room to the Temple of Midgard
	player := Player{name, nil, nil, rooms[3001]}
	printLocation(&player)
	fmt.Println("Type 'cmds' or 'help' to see all available commands!")

	// Initial prompt
	fmt.Print("\n>>> ")
	for scanner.Scan() {
		// Add a newline after commands
		fmt.Println()

		if words := strings.Fields(scanner.Text()); len(words) > 0 {
			// Check if cmd exists
			if validCmd, exists := commands[strings.ToLower(words[0])]; exists {
				validCmd.Run(&player, strings.Join(words[1:], " "))
			} else {
				fmt.Println("Unrecognized command!")
			}
		}

		// Prompt
		fmt.Print("\n>>> ")
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("in main command loop: %v", err)
	}

	return nil
}

// Prints current room description and available exits
func printLocation(p *Player) {
	fmt.Fprintln(p.Conn, p.Location.Name+"\n")
	fmt.Fprintln(p.Conn, p.Location.Description)
	// Print exits
	fmt.Fprintf(p.Conn, "EXITS: [ ")
	for i, exit := range p.Location.Exits {
		if exit.To != nil {
			fmt.Fprintf(p.Conn, "%c ", dirIntToRune[i])
		}
	}
	fmt.Fprintf(p.Conn, "]\n")
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
