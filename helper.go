package main

import (
	"net"
	"sort"
	"strings"
)

// Return the index of an element that satisfies the predicate.
// If none can be found then returns -1
func index(length int, predicate func(idx int) bool) int {
	for i := 0; i < length; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

// Returns whether the given slice has an element that satisfies the predicate
func contain(length int, predicate func(idx int) bool) bool {
	for i := 0; i < length; i++ {
		if predicate(i) {
			return true
		}
	}
	return false
}

// Centers text in the middle of a column of size {size}
func centerText(text string, size int, fill rune) string {
	if &fill == nil {
		fill = ' '
	}
	size -= len(text)
	front := size / 2
	return strings.Repeat(string(fill), front) + text + strings.Repeat(string(fill), size-front)
}

// Remove a player from the room
func (r *room) removePlayer(p *player) {
	if i := index(len(r.players), func(idx int) bool { return r.players[idx] == p }); i != -1 {
		r.players = append(r.players[:i], r.players[i+1:]...)
	}
}

// Sort players in alphabetical order
func (r *room) sortPlayers() {
	sort.Slice(r.players, func(i, j int) bool {
		return r.players[i].name < r.players[j].name
	})
}

// Remove a player from the zone
func (z *zone) removePlayer(p *player) {
	if i := index(len(z.players), func(idx int) bool { return z.players[idx] == p }); i != -1 {
		z.players = append(z.players[:i], z.players[i+1:]...)
	}
}

// Sort players in alphabetical order
func (z *zone) sortPlayers() {
	sort.Slice(z.players, func(i, j int) bool {
		return z.players[i].name < z.players[j].name
	})
}

// Formats plural or singular form
func plural(num int, unit string) string {
	if num == 1 {
		return unit
	}
	return unit + "s"
}

// Returns current IPv4 address
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
