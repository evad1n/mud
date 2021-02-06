package main

import (
	"log"
	"net"
	"time"
)

type (
	player struct {
		name      string      // Username
		conn      net.Conn    // Connection
		log       *log.Logger // Client log
		events    chan event  // MUD outgoing event channel
		beginTime time.Time   // The beginning of the session
		zone      *zone       // The current zone
		room      *room       // The current room
	}

	// A command with all it's info, including linked function
	command struct {
		name        string
		category    string      // The type of command
		description string      // Short description of command
		run         commandFunc // The linked function
	}

	commandFunc func(*player, string) // A command run by a player

	// Input represents an event going from the player to MUD
	input struct {
		player *player // The sending player
		text   string  // The raw text entered
		end    bool    // Signals the connection should be terminated
	}

	// Output represents an event going from MUD to the player
	event struct {
		player *player // The player who initiated the effect
		effect string  // The effect to be printed to the recieving player
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
