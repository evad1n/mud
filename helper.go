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
func (r *Room) removePlayer(p *Player) {
	if i := index(len(r.Players), func(idx int) bool { return r.Players[idx] == p }); i != -1 {
		r.Players = append(r.Players[:i], r.Players[i+1:]...)
	}
}

// Sort players in alphabetical order
func (r *Room) sortPlayers() {
	sort.Slice(r.Players, func(i, j int) bool {
		return r.Players[i].Name < r.Players[j].Name
	})
}

// Remove a player from the zone
func (z *Zone) removePlayer(p *Player) {
	if i := index(len(z.Players), func(idx int) bool { return z.Players[idx] == p }); i != -1 {
		z.Players = append(z.Players[:i], z.Players[i+1:]...)
	}
}

// Sort players in alphabetical order
func (z *Zone) sortPlayers() {
	sort.Slice(z.Players, func(i, j int) bool {
		return z.Players[i].Name < z.Players[j].Name
	})
}

// Formats plural or singular form
func plural(num int, unit string) string {
	if num == 1 {
		return unit
	}
	return unit + "s"
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
