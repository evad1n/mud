package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db    *sql.DB
	zones map[int]*zone
	rooms map[int]*room
)

const (
	path    = "world.db" // the path to the database--this could be an absolute path
	options = "?" + "_busy_timeout=10000" +
		"&" + "_foreign_keys=ON" +
		"&" + "_journal_mode=WAL" +
		"&" + "mode=rw" +
		"&" + "_synchronous=NORMAL"
)

// Load all rooms, zones, exits and link them appropriately
func loadWorld() error {
	var err error
	db, err = sql.Open("sqlite3", path+options)
	if err != nil {
		return fmt.Errorf("opening database: %v", err)
	}
	defer db.Close()

	// Read zones
	if err := readTransaction(readZones); err != nil {
		return fmt.Errorf("reading zones: %v", err)
	}
	// Read rooms
	if err := readTransaction(readRooms); err != nil {
		return fmt.Errorf("reading rooms: %v", err)
	}
	// Read exits
	if err := readTransaction(readExits); err != nil {
		return fmt.Errorf("reading exits: %v", err)
	}

	return nil
}

// A wrapper function for a read transaction
func readTransaction(f func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %v", err)
	}

	if err := f(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("committing transaction: %v", err)
	}

	return nil
}

// Reads all zones into the 'zones' map
func readZones(tx *sql.Tx) error {
	rows, err := tx.Query("SELECT * FROM zones")
	if err != nil {
		return fmt.Errorf("querying zones: %v", err)
	}
	defer rows.Close()

	zones = make(map[int]*zone)
	for rows.Next() {
		var (
			id   int
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			return fmt.Errorf("reading a zone: %v", err)
		}
		// Store zone
		zones[id] = &zone{
			id:      id,
			name:    name,
			rooms:   []*room{},
			players: []*player{},
		}
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating over zones: %v", err)
	}

	return nil
}

// Reads rooms and links them to zones. Rooms are stored in the 'rooms' map
func readRooms(tx *sql.Tx) error {
	rows, err := tx.Query("SELECT * FROM rooms")
	if err != nil {
		return fmt.Errorf("querying rooms: %v", err)
	}
	defer rows.Close()

	rooms = make(map[int]*room)
	for rows.Next() {
		var (
			id     int
			zoneID int
			name   string
			desc   string
		)
		if err = rows.Scan(&id, &zoneID, &name, &desc); err != nil {
			return fmt.Errorf("reading a room: %v", err)
		}

		// Store room
		rooms[id] = &room{
			id:          id,
			zone:        zones[zoneID],
			name:        name,
			description: desc,
			players:     []*player{},
		}
		// Link to zone
		zones[zoneID].rooms = append(zones[zoneID].rooms, rooms[id])
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating over rooms: %v", err)
	}

	return nil
}

// Reads exits and links them to rooms
func readExits(tx *sql.Tx) error {
	rows, err := tx.Query("SELECT * FROM exits")
	if err != nil {
		return fmt.Errorf("querying exits: %v", err)
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
			return fmt.Errorf("reading an exit: %v", err)
		}
		// Link exit to room
		rooms[fromID].exits[dirRuneToInt[rune(dir[0])]] = exit{
			to:          rooms[toID],
			description: desc,
		}
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating over exits: %v", err)
	}

	return nil
}
