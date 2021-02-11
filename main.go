package main

import (
	"fmt"
	"io"
	"log"
	"mud/screen"
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
	var (
		s     *screen.Screen
		left  *screen.Section
		right *screen.Section
		bot   *screen.Section
		err   error
	)
	s, err = screen.CreateScreen(101, 40, os.Stdout)
	left, err = s.NewSection(0, 20, 50, 20)
	if err != nil {
		fmt.Println(err)
	}
	_, err = s.NewStaticSection(50, 20, 1, 20, strings.Repeat("|", 20))
	if err != nil {
		fmt.Println(err)
	}
	right, err = s.NewSection(51, 20, 50, 20)
	if err != nil {
		fmt.Println(err)
	}
	bot, err = s.NewSection(0, 0, 100, 20)
	if err != nil {
		fmt.Println(err)
	}

	left.Write(strings.Repeat("L", 900))
	right.Write(strings.Repeat("R", 900))
	bot.Write(strings.Repeat("B", 900))

	s.Render()

	left.Write(strings.Repeat("l", 900))
	right.Write(strings.Repeat("r", 900))
	bot.Write(strings.Repeat("b", 900))

	s.Render()

	// // Get local IP
	// serverAddress = getLocalAddress()

	// // Create server log
	// serverLog = log.New(os.Stdout, "SERVER: ", log.Ldate|log.Ltime)
	// serverLog.SetPrefix("SERVER: ")
	// serverLog.Printf("Starting server on %s:%s...\n", serverAddress, port)

	// serverLog.Println("Initializing world...")
	// initWorld()

	// // Create global players list
	// players = make(map[string]*player)

	// // Client input channel
	// inputs := make(chan input)

	// serverLog.Println("Listening for connections...")
	// go listenConnections(inputs)

	// // Create event log
	// eventLog = log.New(os.Stdout, "EVENT: ", log.Ldate|log.Ltime)

	// for {
	// 	// Get input
	// 	ev := <-inputs
	// 	// Erase player's old prompt
	// 	ev.player.erasePrompt(ev.player)
	// 	// Check for closed connection
	// 	if ev.end {
	// 		if ev.player.events != nil {
	// 			ev.player.disconnect()
	// 		}
	// 		// Already shutting down -> ignore
	// 		serverLog.Printf("player '%s' disconnecting...\n", ev.player.name)
	// 		continue
	// 	}
	// 	// Otherwise process commands
	// 	if words := strings.Fields(ev.text); len(words) > 0 {
	// 		// Check if cmd exists
	// 		if validCmd, exists := commands[strings.ToLower(words[0])]; exists {
	// 			params := strings.Join(words[1:], " ")
	// 			// Log to server
	// 			eventLog.Printf("PLAYER: %s | COMMAND: %s | PARAMS: %s\n", ev.player.name, validCmd.name, params)
	// 			// Actually run the command
	// 			validCmd.run(ev.player, params)
	// 		} else {
	// 			ev.player.events <- event{
	// 				player: ev.player,
	// 				output: "Unrecognized command!",
	// 				err:    true,
	// 			}
	// 		}
	// 	}
	// 	// Show another prompt if player is still playing
	// 	if ev.player.events != nil {
	// 		ev.player.prompt()
	// 	}
	// }
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
	// Create display
	var (
		display *mudDisplay
		err     error
	)
	if display, err = createMUDDisplay(conn); err != nil {
		serverLog.Fatalf("creating mud display: %v", err)
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
		display:   *display,
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
				output: fmt.Sprintf("%s has join the server.", p.name),
			}
		}
	}

	// Notify other players in room
	for _, other := range r.players {
		other.events <- event{
			player: p,
			output: fmt.Sprintf("%s has entered the room.", p.name),
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

func createMUDDisplay(w io.WriteCloser) (*mudDisplay, error) {
	const (
		width  = 100
		height = 40
	)

	var (
		s        *screen.Screen
		minimap  *screen.Section
		chat     *screen.Section
		location *screen.Section
		err      error
	)

	if s, err = screen.CreateScreen(width+1, height+1, w); err != nil {
		return nil, err
	}
	if minimap, err = s.NewSection(0, height/2, width/2, height/2); err != nil {
		return nil, err
	}
	// Horizontal divider
	if _, err = s.NewStaticSection(0, height/2, width/2, 1, strings.Repeat("-", width/2)); err != nil {
		return nil, err
	}
	if location, err = s.NewSection(0, 0, width/2, height/2); err != nil {
		return nil, err
	}
	// Vertical divider
	if _, err = s.NewStaticSection(width/2, 0, 1, height, strings.Repeat("|", height)); err != nil {
		return nil, err
	}
	if chat, err = s.NewSection(width/2, 0, width/2, height); err != nil {
		return nil, err
	}

	return &mudDisplay{
		screen:   s,
		minimap:  minimap,
		location: location,
		chat:     chat,
	}, nil
}
