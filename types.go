package main

import (
	"log"
	"net"
	"time"
)

// Player represents the user and all associated state
type Player struct {
	// Username
	Name string
	// Connection
	Conn net.Conn
	// Client log
	Log *log.Logger
	// MUD event channel
	Chan chan Output
	// The beginning of the session
	Begin time.Time
	// The current zone
	Zone *Zone
	// The current room
	Room *Room
}

// Command represents a user function with its associated name
type Command struct {
	Name string
	// The family of the command
	Category string
	// Short description of command
	Description string
	// The linked function
	Run func(*Player, string)
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
	Effect string
}

// Zone represents a an area of the world
type Zone struct {
	ID      int
	Name    string
	Rooms   []*Room
	Players []*Player
}

// Room represents a room in a zone
type Room struct {
	ID          int
	Zone        *Zone
	Name        string
	Description string
	Exits       [6]Exit
	Players     []*Player
}

// Exit represents a connection between rooms
type Exit struct {
	To          *Room
	Description string
}
