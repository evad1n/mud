package main

import (
	"fmt"
	"sort"
	"strings"
)

// Map of command aliases to commands
var commands map[string]Command

// Direction abbreviations
var dirs map[string]string

// Maps the exit direction string to its place in the Room exit array
var dirRuneToInt map[rune]int

// Maps a place in the Room exit array to a direction abbreviation rune
var dirIntToRune map[int]rune

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
	commands = make(map[string]Command)

	// In order of precedence

	// Navigation
	addCommand("north", Command{"north", "navigation", "Move north", (*Player).doNorth})
	addCommand("south", Command{"south", "navigation", "Move south", (*Player).doSouth})
	addCommand("east", Command{"east", "navigation", "Move east", (*Player).doEast})
	addCommand("west", Command{"west", "navigation", "Move west", (*Player).doWest})
	addCommand("up", Command{"up", "navigation", "Move up", (*Player).doUp})
	addCommand("down", Command{"down", "navigation", "Move down", (*Player).doDown})
	addCommand("recall", Command{"recall", "navigation", "Return to the Temple of Midgaard", (*Player).doRecall})
	// Information
	addCommand("look", Command{"look", "information", "Look around or in a specific direction", (*Player).doLook})
	addCommand("where", Command{"where", "information", "Display names and locations of all players in current zone", (*Player).doWhere})
	addCommand("help", Command{"help", "information", "List all commands", (*Player).doListCommands})
	// Communication
	addCommand("gossip", Command{"gossip", "communication", "Speak to all players on the server", (*Player).doGossip})
	addCommand("shout", Command{"shout", "communication", "Speak to all players in the current zone", (*Player).doShout})
	addCommand("say", Command{"say", "communication", "Speak to all players in the current room", (*Player).doSay})
	addCommand("tell", Command{"tell", "communication", "Speak privately to a specific player", (*Player).doTell})
	// Emotes
	addCommand("poke", Command{"poke", "communication", "Poke a player", (*Player).doPoke})
	addCommand("laugh", Command{"laugh", "emotes", "Laugh at a player, or in general", (*Player).doLaugh})
	addCommand("sigh", Command{"sigh", "emotes", "Sigh at a player, or in general", (*Player).doSigh})
	addCommand("smile", Command{"smile", "emotes", "Smile at a player, or in general", (*Player).doSmile})
	addCommand("scowl", Command{"scowl", "emotes", "Scowl at a player, or in general", (*Player).doScowl})
	addCommand("think", Command{"think", "emotes", "Put on your thinking cap", (*Player).doThink})
	// Special
	c := Command{"quit", "special", "Leave the MUD", (*Player).doQuit}
	addCommand("quit", c)
	addCommand("exit", c)
}

/* Auto adds all prefixes of alias.
Will not overwrite existing alias mappings.
Add commands in order of importance for alias precedence. */
func addCommand(alias string, cmd Command) {
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

func (p *Player) doRecall(_ string) {
	fmt.Fprint(p.Conn, "You head back to the Temple of Midgard...\n\n")
	p.moveToRoom(rooms[3001])
}

// Navigation

func (p *Player) doNorth(_ string) {
	p.moveDirection(0)
}
func (p *Player) doEast(_ string) {
	p.moveDirection(1)
}
func (p *Player) doWest(_ string) {
	p.moveDirection(2)
}
func (p *Player) doSouth(_ string) {
	p.moveDirection(3)
}
func (p *Player) doUp(_ string) {
	p.moveDirection(4)
}
func (p *Player) doDown(_ string) {
	p.moveDirection(5)
}

// Make sure it is a valid direction
func (p *Player) moveDirection(dir int) {
	if exit := p.Room.Exits[dir]; exit.To != nil {
		p.moveToRoom(exit.To)
	} else {
		fmt.Fprintln(p.Conn, "You can't go that way...")
	}
}

func (p *Player) moveToRoom(r *Room) {
	//  Remove from old room/zone
	p.Room.removePlayer(p)
	p.Zone.removePlayer(p)

	// Notify other players in room
	for _, other := range r.Players {
		other.Chan <- Output{p, fmt.Sprintf("%s has entered the room.", p.Name)}
	}

	p.Room = r
	p.Room.Players = append(r.Players, p)
	p.Zone = r.Zone
	p.Zone.Players = append(r.Zone.Players, p)

	p.Room.sortPlayers()

	p.printLocation()
}

// Information

func (p *Player) doLook(direction string) {
	if len(direction) == 0 {
		p.printLocation()
	} else {
		if fullDir, exists := dirs[direction]; exists {
			p.lookDirection(fullDir)
		} else {
			fmt.Fprintln(p.Conn, "Usage: look <north|south|east|west|up|down>")
		}
	}
}

// Prints current room description and available exits
func (p *Player) printLocation() {
	fmt.Fprintln(p.Conn, p.Room.Name+"\n")
	fmt.Fprintln(p.Conn, p.Room.Description)
	// Print exits
	fmt.Fprintf(p.Conn, "EXITS: [ ")
	for i, exit := range p.Room.Exits {
		if exit.To != nil {
			fmt.Fprintf(p.Conn, "%c ", dirIntToRune[i])
		}
	}
	fmt.Fprintf(p.Conn, "]\n\n")

	fmt.Fprintf(p.Conn, "PLAYERS: [ ")
	for _, other := range p.Room.Players {
		if other != p {
			fmt.Fprintf(p.Conn, "%s ", other.Name)
		}
	}
	fmt.Fprintf(p.Conn, "]\n\n")
}

func (p *Player) lookDirection(dir string) {
	if exit := p.Room.Exits[dirRuneToInt[rune(strings.ToLower(dir)[0])]]; exit.To != nil {
		fmt.Fprint(p.Conn, exit.Description)
	} else {
		fmt.Fprintln(p.Conn, "There's nothing there...")
	}

}

// Print all players in the current zone and their corresponding room
func (p *Player) doWhere(_ string) {
	fmt.Fprintf(p.Conn, "%s\n+%s+\n", centerText(p.Zone.Name, 60, ' '), strings.Repeat("-", 61))
	fmt.Fprintf(p.Conn, "|%s|%s|\n", centerText("PLAYER", 20, ' '), centerText("ROOM", 40, ' '))
	fmt.Fprintf(p.Conn, "+%s+\n", strings.Repeat("-", 61))

	p.Zone.sortPlayers()

	for _, other := range p.Zone.Players {
		fmt.Fprintf(p.Conn, "|%s|%s|\n", centerText(other.Name, 20, ' '), centerText(other.Room.Name, 40, ' '))
	}

	fmt.Fprintf(p.Conn, "+%s+\n", strings.Repeat("-", 61))
}

// Lists known aliases for commands
func (p *Player) doListCommands(_ string) {
	fmt.Fprintf(p.Conn, "+%s+\n", strings.Repeat("-", 30))
	fmt.Fprintf(p.Conn, "|%s|\n", centerText("COMMANDS LIST", 30, ' '))
	fmt.Fprintf(p.Conn, "+%s+\n", strings.Repeat("-", 30))

	// Sort commands alphabetically by category, command name, then alias, in that order

	// This seems unecessarily complex
	categoryMap := make(map[string]map[string][]string)

	for alias, cmd := range commands {
		if _, exists := categoryMap[cmd.Category]; !exists {
			categoryMap[cmd.Category] = make(map[string][]string)
		}
		categoryMap[cmd.Category][cmd.Name] = append(categoryMap[cmd.Category][cmd.Name], alias)
	}

	type AliasList struct {
		Command Command
		Aliases []string
	}

	type Category struct {
		Name     string
		Commands []AliasList
	}

	categoryList := []Category{}

	for typeName, cmdMap := range categoryMap {
		cmdList := Category{typeName, []AliasList{}}
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
			return cmdList.Commands[i].Command.Name < cmdList.Commands[j].Command.Name
		})
	}

	// Sort by type
	sort.Slice(categoryList, func(i, j int) bool {
		return categoryList[i].Name < categoryList[j].Name
	})

	for _, category := range categoryList {
		fmt.Fprintf(p.Conn, "\n+%s+\n", centerText(strings.ToUpper(" "+category.Name+" "), 30, '-'))
		for _, aliasList := range category.Commands {
			for i, alias := range aliasList.Aliases {
				fmt.Fprintf(p.Conn, "| %-10s -->     %-9s |", alias, aliasList.Command.Name)
				if i == 0 {
					fmt.Fprintf(p.Conn, " %-40s\n", aliasList.Command.Description)
				} else {
					fmt.Fprintf(p.Conn, " %s\n", strings.Repeat(" ", 40))
				}
			}
		}
		fmt.Fprintf(p.Conn, "+%s+\n", strings.Repeat("-", 30))
	}
}

// Communication

// Speak to all players on server
func (p *Player) doGossip(msg string) {
	p.serverCommand(
		fmt.Sprintf("%s gossips: %s", p.Name, msg),
		fmt.Sprintf("You gossip: %s", msg),
	)
}

// Speak to all players in a zone
func (p *Player) doShout(msg string) {
	p.zoneCommand(
		fmt.Sprintf("%s shouts: %s", p.Name, msg),
		fmt.Sprintf("You shout: %s", msg),
	)
}

// Speak to all players in a room
func (p *Player) doSay(msg string) {
	p.roomCommand(
		fmt.Sprintf("%s says: %s", p.Name, msg),
		fmt.Sprintf("You say: %s", msg),
	)
}

// Sends a message to specific player, regardless of where they are
func (p *Player) doTell(cmd string) {
	if words := strings.Fields(cmd); len(words) > 1 {
		name := words[0]
		msg := strings.Join(words[1:], " ")
		p.targetedServerCommand(
			name,
			fmt.Sprintf("%s tells you: %s\n", p.Name, msg),
			fmt.Sprintf("You tell %s: %s\n", name, msg),
			"You know talking to yourself is a sign of insanity, right?",
		)
	} else {
		fmt.Fprintln(p.Conn, "Usage: tell <Player Name> <Message>")
	}
}

func (p *Player) doPoke(cmd string) {
	if words := strings.Fields(cmd); len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			name,
			fmt.Sprintf("%s poked you!", p.Name),
			fmt.Sprintf("You poke %s", name),
			"Why are you poking yourself...",
		)
	} else {
		fmt.Fprintln(p.Conn, "Usage: poke <Player Name>")
	}
}

// Emotes

func (p *Player) doSmile(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(fmt.Sprintf("%s smiles happily", p.Name), "You smile happily")
	} else if len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			name,
			fmt.Sprintf("%s smiles at you", p.Name),
			fmt.Sprintf("You smile at %s", name),
			"You smile ... at yourself?",
		)
	} else {
		fmt.Fprintln(p.Conn, "Usage: smile <?Player Name>")
	}
}

func (p *Player) doScowl(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(fmt.Sprintf("%s scowls angrily.", p.Name), "You scowl angrily")
	} else if len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			name,
			fmt.Sprintf("%s scowls at you", p.Name),
			fmt.Sprintf("You scowl at %s", name),
			"You must really hate yourself...",
		)
	} else {
		fmt.Fprintln(p.Conn, "Usage: scowl <?Player Name>")
	}
}

func (p *Player) doSigh(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(fmt.Sprintf("%s sighs heavily", p.Name), "You sigh heavily")
	} else if len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			name,
			fmt.Sprintf("%s sighs at you", p.Name),
			fmt.Sprintf("You sigh at %s", name),
			"Rough day, huh?",
		)
	} else {
		fmt.Fprintln(p.Conn, "Usage: sigh <?Player Name>")
	}
}

func (p *Player) doLaugh(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(fmt.Sprintf("%s laughs heartily", p.Name), "You laugh heartily")
	} else if len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			name,
			fmt.Sprintf("%s laughs at you", p.Name),
			fmt.Sprintf("You laugh at %s", name),
			"It's always good to be able to laugh at yourself",
		)
	} else {
		fmt.Fprintln(p.Conn, "Usage: laugh <?Player Name>")
	}
}

func (p *Player) doThink(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(
			fmt.Sprintf("%s is in deep thought", p.Name),
			"You are in deep thought",
		)
	} else {
		fmt.Fprintln(p.Conn, "Usage: think")
	}
}

// Special

func (p *Player) doQuit(_ string) {
	fmt.Fprintf(p.Conn, "Goodbye %s!\nThanks for playing!\n\n", p.Name)
	p.disconnect()
}

// Helper functions

// Represents a command that targets another player in the room
func (p *Player) targetedRoomCommand(name string, outMsg string, selfMsg string, errSelf string) {
	if idx := index(len(p.Room.Players), func(i int) bool { return p.Room.Players[i].Name == name }); idx != -1 {
		other := p.Room.Players[idx]
		if other != p {
			other.Chan <- Output{p, outMsg}
			fmt.Fprintln(p.Conn, selfMsg)
		} else {
			fmt.Fprintln(p.Conn, errSelf)
		}
	} else {
		fmt.Fprintln(p.Conn, "No such player!")
	}
}

// Represents a command that targets another player cross-server
func (p *Player) targetedServerCommand(name string, outMsg string, selfMsg string, errSelf string) {
	if other, exists := players[name]; exists {
		if other != p {
			other.Chan <- Output{p, outMsg}
			fmt.Fprintln(p.Conn, selfMsg)
		} else {
			fmt.Fprintln(p.Conn, errSelf)
		}
	} else {
		fmt.Fprintln(p.Conn, "No such player!")
	}
}

// A command that affects everyone in the room
func (p *Player) roomCommand(outMsg string, selfMsg string) {
	for _, other := range p.Room.Players {
		if other != p {
			if ch := other.Chan; ch != nil {
				ch <- Output{p, outMsg}
			}
		} else {
			fmt.Fprintln(p.Conn, selfMsg)
		}
	}
}

// A command that affects everyone in the zone
func (p *Player) zoneCommand(outMsg string, selfMsg string) {
	for _, other := range p.Zone.Players {
		if other != p {
			if ch := other.Chan; ch != nil {
				ch <- Output{p, outMsg}
			}
		} else {
			fmt.Fprintln(p.Conn, selfMsg)
		}
	}
}

// A command that affects everyone in the server
func (p *Player) serverCommand(outMsg string, selfMsg string) {
	for _, other := range players {
		if other != p {
			if ch := other.Chan; ch != nil {
				ch <- Output{p, outMsg}
			}
		} else {
			fmt.Fprintln(p.Conn, selfMsg)
		}
	}
}
