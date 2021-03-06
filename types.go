package main

import (
	"log"
	"net"
	"time"
)

type (
	player struct {
		name      string       // Username
		conn      net.Conn     // Connection
		log       *log.Logger  // Client log
		events    chan event   // MUD outgoing event channel
		beginTime time.Time    // The beginning of the session
		zone      *zone        // The current zone
		room      *room        // The current room
		minimap   *mapBuilder  // The displayed minimap
		visited   map[int]bool // Visited rooms for the map
	}

	// A command with all it's info, including linked function
	command struct {
		name        string
		category    commandCategory // The type of command
		description string          // Short description of command
		run         commandFunc     // The linked function
	}

	commandFunc func(*player, string) // A command run by a player

	commandCategory int // A type of command

	// Input represents an event going from the player to MUD
	input struct {
		player *player // The sending player
		text   string  // The raw text entered
		end    bool    // Signals the connection should be terminated
	}

	// Output represents an event going from MUD to the player
	event struct {
		player      *player  // The player who initiated the effect
		output      string   // The string output to be printed to the recieving player
		command     *command // The command tat caused this event
		delay       int      // An optional delay (in milliseconds) after this prompt
		unsolicited bool     // Whether the user pressed enter
		noPrompt    bool     // Whether to draw the prompt again
		updateMap   bool     // Whether to retrace the map
		err         bool     // Prints in red
	}

	// An area of the world
	zone struct {
		id      int
		name    string
		rooms   []*room   // All rooms in this zone
		players []*player // All players currently in the zone
	}

	// A room in a zone
	room struct {
		id          int
		zone        *zone // The zone this room is part of
		name        string
		description string
		exits       [6]exit   // The connections from this room to others
		players     []*player // All players currently in the room
	}

	// A connection between rooms
	exit struct {
		to          *room // Where this exit leads to
		description string
	}
)
