package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Zone represents a an area of the world
type Zone struct {
	ID    int
	Name  string
	Rooms []*Room
}

// Room represents a room in a zone
type Room struct {
	ID          int
	Zone        *Zone
	Name        string
	Description string
	Exits       [6]Exit
}

// Exit represents a connection between rooms
type Exit struct {
	To          *Room
	Description string
}

var (
	db    *sql.DB
	zones map[int]*Zone
	rooms map[int]*Room
)

// Load all rooms, zones, exits and link them appropriately
func loadDB() error {
	// the path to the database--this could be an absolute path
	path := "world.db"
	options :=
		"?" + "_busy_timeout=10000" +
			"&" + "_foreign_keys=ON" +
			"&" + "_journal_mode=WAL" +
			"&" + "mode=rw" +
			"&" + "_synchronous=NORMAL"

	var err error
	db, err = sql.Open("sqlite3", path+options)
	if err != nil {
		return err
	}
	defer db.Close()

	// Read zones
	if err := readTransaction(readZones); err != nil {
		return err
	}
	// Read rooms
	if err := readTransaction(readRooms); err != nil {
		return err
	}
	// Read exits
	if err := readTransaction(readExits); err != nil {
		return err
	}

	return nil
}

// A wrapper function for a read transaction
func readTransaction(f func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := f(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// Reads all zones into the 'zones' map
func readZones(tx *sql.Tx) error {
	rows, err := tx.Query("SELECT * FROM zones")
	if err != nil {
		return fmt.Errorf("reading a zone from the database: %v", err)
	}
	defer rows.Close()

	zones = make(map[int]*Zone)
	for rows.Next() {
		var (
			ID   int
			name string
		)
		if err := rows.Scan(&ID, &name); err != nil {
			return fmt.Errorf("reading a zone from the database: %v", err)
		}
		// Store zone
		zones[ID] = &Zone{ID: ID, Name: name, Rooms: []*Room{}}
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading a zone from the database: %v", err)
	}

	return nil
}

// Reads rooms and links them to zones. Rooms are stored in the 'rooms' map
func readRooms(tx *sql.Tx) error {
	rows, err := tx.Query("SELECT * FROM rooms")
	if err != nil {
		return fmt.Errorf("reading a room from the database: %v", err)
	}
	defer rows.Close()

	rooms = make(map[int]*Room)
	for rows.Next() {
		var (
			ID     int
			zoneID int
			name   string
			desc   string
		)
		if err = rows.Scan(&ID, &zoneID, &name, &desc); err != nil {
			return fmt.Errorf("reading a room from the database: %v", err)
		}

		// Store room
		rooms[ID] = &Room{ID: ID, Zone: zones[zoneID], Name: name, Description: desc}
		// Link to zone
		zones[zoneID].Rooms = append(zones[zoneID].Rooms, rooms[ID])
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading a room from the database: %v", err)
	}

	return nil
}

// Reads exits and links them to rooms
func readExits(tx *sql.Tx) error {
	rows, err := tx.Query("SELECT * FROM exits")
	if err != nil {
		return fmt.Errorf("reading an exit from the database: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			fromID int
			toID   int
			dir    string
			desc   string
		)
		if err := rows.Scan(&fromID, &toID, &dir, &desc); err != nil {
			return fmt.Errorf("reading an exit from the database: %v", err)
		}
		// Link exit to room
		rooms[fromID].Exits[dirRuneToInt[rune(dir[0])]] = Exit{To: rooms[toID], Description: desc}
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading an exit from the database: %v", err)
	}

	return nil
}
