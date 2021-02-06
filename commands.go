package main

import (
	"fmt"
	"sort"
	"strings"
)

var (
	commands     map[string]command // Map of command aliases to commands
	dirs         map[string]string  // Direction abbreviations
	dirRuneToInt map[rune]int       // Maps the exit direction string to its place in the room exit array
	dirIntToRune map[int]rune       // Maps a place in the room exit array to a direction abbreviation rune
)

// Initialize and populate lookup tables
func createMaps() {
	// Directions
	dirs = make(map[string]string)
	addMapPrefix("north", dirs)
	addMapPrefix("south", dirs)
	addMapPrefix("east", dirs)
	addMapPrefix("west", dirs)
	addMapPrefix("up", dirs)
	addMapPrefix("down", dirs)

	// Exit directions
	dirRuneToInt = make(map[rune]int)
	dirRuneToInt['n'] = 0
	dirRuneToInt['e'] = 1
	dirRuneToInt['w'] = 2
	dirRuneToInt['s'] = 3
	dirRuneToInt['u'] = 4
	dirRuneToInt['d'] = 5

	dirIntToRune = make(map[int]rune)
	dirIntToRune[0] = 'n'
	dirIntToRune[1] = 'e'
	dirIntToRune[2] = 'w'
	dirIntToRune[3] = 's'
	dirIntToRune[4] = 'u'
	dirIntToRune[5] = 'd'
}

/* Maps prefixes to full name for a map */
func addMapPrefix(full string, m map[string]string) {
	for i := range full {
		if i == 0 {
			continue
		}
		prefix := full[:i]
		if _, exists := m[prefix]; !exists {
			m[prefix] = full
		}
	}
	m[full] = full
}

// Adds all starting commands
func defaultCommands() {
	commands = make(map[string]command)

	// In order of precedence

	// Navigation
	addCommand("north", command{"north", "navigation", "Move north", (*player).doNorth})
	addCommand("south", command{"south", "navigation", "Move south", (*player).doSouth})
	addCommand("east", command{"east", "navigation", "Move east", (*player).doEast})
	addCommand("west", command{"west", "navigation", "Move west", (*player).doWest})
	addCommand("up", command{"up", "navigation", "Move up", (*player).doUp})
	addCommand("down", command{"down", "navigation", "Move down", (*player).doDown})
	addCommand("recall", command{"recall", "navigation", "Return to the Temple of Midgaard", (*player).doRecall})
	// Information
	addCommand("look", command{"look", "information", "Look around or in a specific direction", (*player).doLook})
	addCommand("where", command{"where", "information", "Display names and locations of all players in current zone", (*player).doWhere})
	addCommand("help", command{"help", "information", "List all commands", (*player).doListCommands})
	// Communication
	addCommand("gossip", command{"gossip", "communication", "Speak to all players on the server", (*player).doGossip})
	addCommand("shout", command{"shout", "communication", "Speak to all players in the current zone", (*player).doShout})
	addCommand("say", command{"say", "communication", "Speak to all players in the current room", (*player).doSay})
	addCommand("tell", command{"tell", "communication", "Speak privately to a specific player", (*player).doTell})
	// Emotes
	addCommand("poke", command{"poke", "communication", "Poke a player", (*player).doPoke})
	addCommand("laugh", command{"laugh", "emotes", "Laugh at a player, or in general", (*player).doLaugh})
	addCommand("sigh", command{"sigh", "emotes", "Sigh at a player, or in general", (*player).doSigh})
	addCommand("smile", command{"smile", "emotes", "Smile at a player, or in general", (*player).doSmile})
	addCommand("scowl", command{"scowl", "emotes", "Scowl at a player, or in general", (*player).doScowl})
	addCommand("think", command{"think", "emotes", "Put on your thinking cap", (*player).doThink})
	// Special
	c := command{"quit", "special", "Leave the MUD", (*player).doQuit}
	addCommand("quit", c)
	addCommand("exit", c)
}

/* Auto adds all prefixes of alias.
Will not overwrite existing alias mappings.
Add commands in order of importance for alias precedence. */
func addCommand(alias string, cmd command) {
	for i := range alias {
		if i == 0 {
			continue
		}
		prefix := alias[:i]
		if _, exists := commands[prefix]; !exists {
			commands[prefix] = cmd
		}
	}
	commands[alias] = cmd
}

//////////////
// Commands //
//////////////

// Navigation

func (p *player) doRecall(_ string) {
	fmt.Fprint(p.conn, "You head back to the Temple of Midgard...\n\n")
	p.moveToRoom(rooms[3001])
}

// Navigation

func (p *player) doNorth(_ string) {
	p.moveDirection(0)
}
func (p *player) doEast(_ string) {
	p.moveDirection(1)
}
func (p *player) doWest(_ string) {
	p.moveDirection(2)
}
func (p *player) doSouth(_ string) {
	p.moveDirection(3)
}
func (p *player) doUp(_ string) {
	p.moveDirection(4)
}
func (p *player) doDown(_ string) {
	p.moveDirection(5)
}

// Make sure it is a valid direction
func (p *player) moveDirection(dir int) {
	if exit := p.room.exits[dir]; exit.to != nil {
		p.moveToRoom(exit.to)
	} else {
		fmt.Fprintln(p.conn, "You can't go that way...")
	}
}

func (p *player) moveToRoom(r *room) {
	//  Remove from old room/zone
	p.room.removePlayer(p)
	p.zone.removePlayer(p)

	// Notify other players in room
	for _, other := range r.players {
		other.events <- event{p, fmt.Sprintf("%s has entered the room.", p.name)}
	}

	p.room = r
	p.room.players = append(r.players, p)
	p.zone = r.zone
	p.zone.players = append(r.zone.players, p)

	p.room.sortPlayers()

	p.printLocation()
}

// Information

func (p *player) doLook(direction string) {
	if len(direction) == 0 {
		p.printLocation()
	} else {
		if fullDir, exists := dirs[direction]; exists {
			p.lookDirection(fullDir)
		} else {
			fmt.Fprintln(p.conn, "Usage: look <north|south|east|west|up|down>")
		}
	}
}

// Prints current room description and available exits
func (p *player) printLocation() {
	fmt.Fprintln(p.conn, p.room.name+"\n")
	fmt.Fprintln(p.conn, p.room.description)
	// Print exits
	fmt.Fprintf(p.conn, "EXITS: [ ")
	for i, exit := range p.room.exits {
		if exit.to != nil {
			fmt.Fprintf(p.conn, "%c ", dirIntToRune[i])
		}
	}
	fmt.Fprintf(p.conn, "]\n\n")

	fmt.Fprintf(p.conn, "PLAYERS: [ ")
	for _, other := range p.room.players {
		if other != p {
			fmt.Fprintf(p.conn, "%s ", other.name)
		}
	}
	fmt.Fprintf(p.conn, "]\n\n")
}

func (p *player) lookDirection(dir string) {
	if exit := p.room.exits[dirRuneToInt[rune(strings.ToLower(dir)[0])]]; exit.to != nil {
		fmt.Fprint(p.conn, exit.description)
	} else {
		fmt.Fprintln(p.conn, "There's nothing there...")
	}

}

// Print all players in the current zone and their corresponding room
func (p *player) doWhere(_ string) {
	fmt.Fprintf(p.conn, "%s\n+%s+\n", centerText(p.zone.name, 60, ' '), strings.Repeat("-", 61))
	fmt.Fprintf(p.conn, "|%s|%s|\n", centerText("PLAYER", 20, ' '), centerText("ROOM", 40, ' '))
	fmt.Fprintf(p.conn, "+%s+\n", strings.Repeat("-", 61))

	p.zone.sortPlayers()

	for _, other := range p.zone.players {
		fmt.Fprintf(p.conn, "|%s|%s|\n", centerText(other.name, 20, ' '), centerText(other.room.name, 40, ' '))
	}

	fmt.Fprintf(p.conn, "+%s+\n", strings.Repeat("-", 61))
}

// Lists known aliases for commands
func (p *player) doListCommands(_ string) {
	fmt.Fprintf(p.conn, "+%s+\n", strings.Repeat("-", 30))
	fmt.Fprintf(p.conn, "|%s|\n", centerText("COMMANDS LIST", 30, ' '))
	fmt.Fprintf(p.conn, "+%s+\n", strings.Repeat("-", 30))

	// Sort commands alphabetically by category, command name, then alias, in that order

	// This seems unecessarily complex
	categoryMap := make(map[string]map[string][]string)

	for alias, cmd := range commands {
		if _, exists := categoryMap[cmd.category]; !exists {
			categoryMap[cmd.category] = make(map[string][]string)
		}
		categoryMap[cmd.category][cmd.name] = append(categoryMap[cmd.category][cmd.name], alias)
	}

	type AliasList struct {
		command command
		Aliases []string
	}

	type category struct {
		name     string
		Commands []AliasList
	}

	categoryList := []category{}

	for typeName, cmdMap := range categoryMap {
		cmdList := category{typeName, []AliasList{}}
		for _, aliases := range cmdMap {
			cmdList.Commands = append(cmdList.Commands, AliasList{commands[aliases[0]], aliases})
		}
		categoryList = append(categoryList, cmdList)
	}

	for _, cmdList := range categoryList {
		// Sort individual groups
		for _, aliasList := range cmdList.Commands {
			sort.Strings(aliasList.Aliases)
		}
		// Sort command list
		sort.Slice(cmdList.Commands, func(i, j int) bool {
			return cmdList.Commands[i].command.name < cmdList.Commands[j].command.name
		})
	}

	// Sort by type
	sort.Slice(categoryList, func(i, j int) bool {
		return categoryList[i].name < categoryList[j].name
	})

	for _, category := range categoryList {
		fmt.Fprintf(p.conn, "\n+%s+\n", centerText(strings.ToUpper(" "+category.name+" "), 30, '-'))
		for _, aliasList := range category.Commands {
			for i, alias := range aliasList.Aliases {
				fmt.Fprintf(p.conn, "| %-10s -->     %-9s |", alias, aliasList.command.name)
				if i == 0 {
					fmt.Fprintf(p.conn, " %-40s\n", aliasList.command.description)
				} else {
					fmt.Fprintf(p.conn, " %s\n", strings.Repeat(" ", 40))
				}
			}
		}
		fmt.Fprintf(p.conn, "+%s+\n", strings.Repeat("-", 30))
	}
}

// Communication

// Speak to all players on server
func (p *player) doGossip(msg string) {
	p.serverCommand(
		fmt.Sprintf("%s gossips: %s", p.name, msg),
		fmt.Sprintf("You gossip: %s", msg),
	)
}

// Speak to all players in a zone
func (p *player) doShout(msg string) {
	p.zoneCommand(
		fmt.Sprintf("%s shouts: %s", p.name, msg),
		fmt.Sprintf("You shout: %s", msg),
	)
}

// Speak to all players in a room
func (p *player) doSay(msg string) {
	p.roomCommand(
		fmt.Sprintf("%s says: %s", p.name, msg),
		fmt.Sprintf("You say: %s", msg),
	)
}

// Sends a message to specific player, regardless of where they are
func (p *player) doTell(cmd string) {
	if words := strings.Fields(cmd); len(words) > 1 {
		name := words[0]
		msg := strings.Join(words[1:], " ")
		p.targetedServerCommand(
			name,
			fmt.Sprintf("%s tells you: %s\n", p.name, msg),
			fmt.Sprintf("You tell %s: %s\n", name, msg),
			"You know talking to yourself is a sign of insanity, right?",
		)
	} else {
		fmt.Fprintln(p.conn, "Usage: tell <player name> <Message>")
	}
}

func (p *player) doPoke(cmd string) {
	if words := strings.Fields(cmd); len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			name,
			fmt.Sprintf("%s poked you!", p.name),
			fmt.Sprintf("You poke %s", name),
			"Why are you poking yourself...",
		)
	} else {
		fmt.Fprintln(p.conn, "Usage: poke <player name>")
	}
}

// Emotes

func (p *player) doSmile(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(fmt.Sprintf("%s smiles happily", p.name), "You smile happily")
	} else if len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			name,
			fmt.Sprintf("%s smiles at you", p.name),
			fmt.Sprintf("You smile at %s", name),
			"You smile ... at yourself?",
		)
	} else {
		fmt.Fprintln(p.conn, "Usage: smile <?player name>")
	}
}

func (p *player) doScowl(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(fmt.Sprintf("%s scowls angrily.", p.name), "You scowl angrily")
	} else if len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			name,
			fmt.Sprintf("%s scowls at you", p.name),
			fmt.Sprintf("You scowl at %s", name),
			"You must really hate yourself...",
		)
	} else {
		fmt.Fprintln(p.conn, "Usage: scowl <?player name>")
	}
}

func (p *player) doSigh(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(fmt.Sprintf("%s sighs heavily", p.name), "You sigh heavily")
	} else if len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			name,
			fmt.Sprintf("%s sighs at you", p.name),
			fmt.Sprintf("You sigh at %s", name),
			"Rough day, huh?",
		)
	} else {
		fmt.Fprintln(p.conn, "Usage: sigh <?player name>")
	}
}

func (p *player) doLaugh(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(fmt.Sprintf("%s laughs heartily", p.name), "You laugh heartily")
	} else if len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			name,
			fmt.Sprintf("%s laughs at you", p.name),
			fmt.Sprintf("You laugh at %s", name),
			"It's always good to be able to laugh at yourself",
		)
	} else {
		fmt.Fprintln(p.conn, "Usage: laugh <?player name>")
	}
}

func (p *player) doThink(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(
			fmt.Sprintf("%s is in deep thought", p.name),
			"You are in deep thought",
		)
	} else {
		fmt.Fprintln(p.conn, "Usage: think")
	}
}

// Special

func (p *player) doQuit(_ string) {
	fmt.Fprintf(p.conn, "Goodbye %s!\nThanks for playing!\n\n", p.name)
	p.disconnect()
}

// Helper functions

// Represents a command that targets another player in the room
func (p *player) targetedRoomCommand(name string, outMsg string, selfMsg string, errSelf string) {
	if idx := index(len(p.room.players), func(i int) bool { return p.room.players[i].name == name }); idx != -1 {
		other := p.room.players[idx]
		if other != p {
			other.events <- event{p, outMsg}
			fmt.Fprintln(p.conn, selfMsg)
		} else {
			fmt.Fprintln(p.conn, errSelf)
		}
	} else {
		fmt.Fprintln(p.conn, "No such player!")
	}
}

// Represents a command that targets another player cross-server
func (p *player) targetedServerCommand(name string, outMsg string, selfMsg string, errSelf string) {
	if other, exists := players[name]; exists {
		if other != p {
			other.events <- event{p, outMsg}
			fmt.Fprintln(p.conn, selfMsg)
		} else {
			fmt.Fprintln(p.conn, errSelf)
		}
	} else {
		fmt.Fprintln(p.conn, "No such player!")
	}
}

// A command that affects everyone in the room
func (p *player) roomCommand(outMsg string, selfMsg string) {
	for _, other := range p.room.players {
		if other != p {
			if ch := other.events; ch != nil {
				ch <- event{p, outMsg}
			}
		} else {
			fmt.Fprintln(p.conn, selfMsg)
		}
	}
}

// A command that affects everyone in the zone
func (p *player) zoneCommand(outMsg string, selfMsg string) {
	for _, other := range p.zone.players {
		if other != p {
			if ch := other.events; ch != nil {
				ch <- event{p, outMsg}
			}
		} else {
			fmt.Fprintln(p.conn, selfMsg)
		}
	}
}

// A command that affects everyone in the server
func (p *player) serverCommand(outMsg string, selfMsg string) {
	for _, other := range players {
		if other != p {
			if ch := other.events; ch != nil {
				ch <- event{p, outMsg}
			}
		} else {
			fmt.Fprintln(p.conn, selfMsg)
		}
	}
}
