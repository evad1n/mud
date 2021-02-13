package main

import (
	"fmt"
	"sort"
	"strings"
)

// Command category enums
// This also defines precedence when listing in 'help' command
const (
	nav commandCategory = iota
	info
	comm
	emotes
	special
)

var (
	commands           map[string]*command        // Map of command aliases to commands
	commandCategoryMap map[commandCategory]string // Maps commandCategory enums to string descriptions
	dirs               map[string]string          // Direction abbreviations
	dirRuneToInt       map[rune]int               // Maps the exit direction string to its place in the room exit array
	dirIntToRune       map[int]rune               // Maps a place in the room exit array to a direction abbreviation rune
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

	// Command categories
	commandCategoryMap = make(map[commandCategory]string)
	commandCategoryMap[nav] = "navigation"
	commandCategoryMap[info] = "information"
	commandCategoryMap[comm] = "communication"
	commandCategoryMap[emotes] = "emotes"
	commandCategoryMap[special] = "special"
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
	commands = make(map[string]*command)

	// In order of precedence

	// Navigation
	addCommand("north", command{
		name:        "north",
		category:    nav,
		description: "Move north",
		run:         (*player).doNorth,
	})
	addCommand("south", command{
		name:        "south",
		category:    nav,
		description: "Move south",
		run:         (*player).doSouth,
	})
	addCommand("east", command{
		name:        "east",
		category:    nav,
		description: "Move east",
		run:         (*player).doEast,
	})
	addCommand("west", command{
		name:        "west",
		category:    nav,
		description: "Move west",
		run:         (*player).doWest,
	})
	addCommand("up", command{
		name:        "up",
		category:    nav,
		description: "Move up",
		run:         (*player).doUp,
	})
	addCommand("down", command{
		name:        "down",
		category:    nav,
		description: "Move down",
		run:         (*player).doDown,
	})
	addCommand("recall", command{
		name:        "recall",
		category:    nav,
		description: "Return to the Temple of Midgaard",
		run:         (*player).doRecall,
	})
	// Information
	addCommand("look", command{
		name:        "look",
		category:    info,
		description: "Look around or in a specific direction",
		run:         (*player).doLook,
	})
	addCommand("where", command{
		name:        "where",
		category:    info,
		description: "Display names and locations of all players in current zone",
		run:         (*player).doWhere,
	})
	addCommand("help", command{
		name:        "help",
		category:    info,
		description: "List all commands",
		run:         (*player).doListCommands,
	})
	// Communication
	addCommand("gossip", command{
		name:        "gossip",
		category:    comm,
		description: "Speak to all players on the server",
		run:         (*player).doGossip,
	})
	addCommand("shout", command{
		name:        "shout",
		category:    comm,
		description: "Speak to all players in the current zone",
		run:         (*player).doShout,
	})
	addCommand("say", command{
		name:        "say",
		category:    comm,
		description: "Speak to all players in the current room",
		run:         (*player).doSay,
	})
	addCommand("tell", command{
		name:        "tell",
		category:    comm,
		description: "Speak privately to a specific player",
		run:         (*player).doTell,
	})
	// Emotes
	addCommand("poke", command{
		name:        "poke",
		category:    comm,
		description: "Poke a player",
		run:         (*player).doPoke,
	})
	addCommand("laugh", command{
		name:        "laugh",
		category:    emotes,
		description: "Laugh at a player, or in general",
		run:         (*player).doLaugh,
	})
	addCommand("sigh", command{
		name:        "sigh",
		category:    emotes,
		description: "Sigh at a player, or in general",
		run:         (*player).doSigh,
	})
	addCommand("smile", command{
		name:        "smile",
		category:    emotes,
		description: "Smile at a player, or in general",
		run:         (*player).doSmile,
	})
	addCommand("scowl", command{
		name:        "scowl",
		category:    emotes,
		description: "Scowl at a player, or in general",
		run:         (*player).doScowl,
	})
	addCommand("think", command{
		name:        "think",
		category:    emotes,
		description: "Put on your thinking cap",
		run:         (*player).doThink,
	})
	// Special
	c := command{
		name:        "quit",
		category:    special,
		description: "Leave the MUD",
		run:         (*player).doQuit,
	}
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
			commands[prefix] = &cmd
		}
	}
	commands[alias] = &cmd
}

//////////////
// Commands //
//////////////

// Navigation

func (p *player) doRecall(_ string) {
	p.events <- event{
		player:  p,
		output:  "You head back to the Temple of Midgard...\n",
		command: commands["recall"],
		delay:   1000,
	}
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
		p.events <- event{
			player: p,
			output: "You can't go that way...",
		}
	}
}

func (p *player) moveToRoom(r *room) {
	//  Remove from old room/zone
	p.room.removePlayer(p)
	p.zone.removePlayer(p)

	// Notify other players in old room
	for _, other := range p.room.players {
		if other != p {
			other.events <- event{
				player: p,
				output: fmt.Sprintf("%s has left the room", p.name),
			}
		}
	}

	// Notify other players in new room
	for _, other := range r.players {
		other.events <- event{
			player: p,
			output: fmt.Sprintf("%s has entered the room", p.name),
		}
	}

	p.room = r
	p.room.players = append(r.players, p)
	p.zone = r.zone
	p.zone.players = append(r.zone.players, p)

	p.room.sortPlayers()

	// Update map
	p.events <- event{
		player:    p,
		output:    "",
		updateMap: true,
	}

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
			p.events <- event{
				player: p,
				output: "Usage: look <north|south|east|west|up|down>",
				err:    true,
			}
		}
	}
}

// Prints current room description and available exits
func (p *player) printLocation() {
	output := ""
	output += (p.room.name + "\n\n")
	output += p.room.description

	// Show exits
	output += "\nEXITS: [ "
	for i, exit := range p.room.exits {
		if exit.to != nil {
			output += fmt.Sprintf("%s ", ansiWrap(string(dirIntToRune[i]), "\x1b[36m"))
		}
	}
	output += "]\n\n"
	// Show players
	output += "PLAYERS: [ "
	for _, other := range p.room.players {
		if other != p {
			output += fmt.Sprintf("%s ", ansiWrap(other.name, "\x1b[32m"))
		}
	}
	output += "]"
	// Send formatted output to player
	p.events <- event{
		player: p,
		output: output,
	}
}

func (p *player) lookDirection(dir string) {
	if exit := p.room.exits[dirRuneToInt[rune(strings.ToLower(dir)[0])]]; exit.to != nil {
		p.events <- event{
			player: p,
			output: exit.description,
		}
	} else {
		p.events <- event{
			player: p,
			output: "There's nothing there...",
		}
	}
}

// Print all players in the current zone and their corresponding room
func (p *player) doWhere(_ string) {
	output := ""
	output += fmt.Sprintf("%s\n+%s+\n", centerText(p.zone.name, 60, ' '), strings.Repeat("-", 61))
	output += fmt.Sprintf("|%s|%s|\n", centerText("PLAYER", 20, ' '), centerText("ROOM", 40, ' '))
	output += fmt.Sprintf("+%s+\n", strings.Repeat("-", 61))

	p.zone.sortPlayers()

	for _, other := range p.zone.players {
		output += fmt.Sprintf("|%s|%s|\n", centerText(other.name, 20, ' '), centerText(other.room.name, 40, ' '))
	}

	output += fmt.Sprintf("+%s+", strings.Repeat("-", 61))
	// Send formatted output to player
	p.events <- event{
		player: p,
		output: output,
	}
}

// Lists known aliases for commands
func (p *player) doListCommands(_ string) {
	output := ""
	output += fmt.Sprintf("+%s+\n", strings.Repeat("-", 30))
	output += fmt.Sprintf("|%s|\n", centerText("COMMANDS LIST", 30, ' '))
	output += fmt.Sprintf("+%s+\n", strings.Repeat("-", 30))

	// Sort commands alphabetically by category, command name, then alias, in that order

	// This seems unecessarily complex
	categoryMap := make(map[commandCategory]map[string][]string)

	for alias, cmd := range commands {
		if _, exists := categoryMap[cmd.category]; !exists {
			categoryMap[cmd.category] = make(map[string][]string)
		}
		categoryMap[cmd.category][cmd.name] = append(categoryMap[cmd.category][cmd.name], alias)
	}

	// All aliases associated with a command
	type aliasList struct {
		cmd     *command
		aliases []string
	}

	// A list of commands belonging to a specific category
	type commandList struct {
		category commandCategory
		commands []aliasList
	}

	categoryList := []commandList{}

	for category, cmdMap := range categoryMap {
		cmdList := commandList{
			category: category,
			commands: []aliasList{},
		}
		for _, aliases := range cmdMap {
			cmdList.commands = append(cmdList.commands, aliasList{commands[aliases[0]], aliases})
		}
		categoryList = append(categoryList, cmdList)
	}

	for _, cmdList := range categoryList {
		// Sort individual groups
		for _, aliasList := range cmdList.commands {
			sort.Strings(aliasList.aliases)
		}
		// Sort command list
		sort.Slice(cmdList.commands, func(i, j int) bool {
			return cmdList.commands[i].cmd.name < cmdList.commands[j].cmd.name
		})
	}

	// Sort by type
	sort.Slice(categoryList, func(i, j int) bool {
		return categoryList[i].category < categoryList[j].category
	})

	for _, category := range categoryList {
		output += fmt.Sprintf(
			"\n+%s+\n",
			centerText(
				strings.ToUpper(" "+commandCategoryMap[category.category]+" "),
				30,
				'-',
			),
		)
		for _, aliasList := range category.commands {
			for i, alias := range aliasList.aliases {
				output += fmt.Sprintf("| %-10s -->     %-9s |", alias, aliasList.cmd.name)
				if i == 0 {
					output += fmt.Sprintf(" %-40s\n", aliasList.cmd.description)
				} else {
					output += fmt.Sprintf(" %s\n", strings.Repeat(" ", 40))
				}
			}
		}
		output += fmt.Sprintf("+%s+", strings.Repeat("-", 30))
	}
	// Send formatted output to player
	p.events <- event{
		player: p,
		output: output,
	}
}

// Communication

// Speak to all players on server
func (p *player) doGossip(msg string) {
	p.serverCommand(
		commands["gossip"],
		fmt.Sprintf("%s gossips: %s", p.name, msg),
		fmt.Sprintf("You gossip: %s", msg),
	)
}

// Speak to all players in a zone
func (p *player) doShout(msg string) {
	p.zoneCommand(
		commands["shout"],
		fmt.Sprintf("%s shouts: %s", p.name, msg),
		fmt.Sprintf("You shout: %s", msg),
	)
}

// Speak to all players in a room
func (p *player) doSay(msg string) {
	p.roomCommand(
		commands["say"],
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
			commands["tell"],
			name,
			fmt.Sprintf("%s tells you: %s", p.name, msg),
			fmt.Sprintf("You tell %s: %s", name, msg),
			"You know talking to yourself is a sign of insanity, right?",
		)
	} else {
		p.events <- event{
			player: p,
			output: "Usage: tell <player name> <Message>",
			err:    true,
		}
	}
}

func (p *player) doPoke(cmd string) {
	if words := strings.Fields(cmd); len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			commands["poke"],
			name,
			fmt.Sprintf("%s poked you!", p.name),
			fmt.Sprintf("You poke %s", name),
			"Why are you poking yourself...",
		)
	} else {
		p.events <- event{
			player: p,
			output: "Usage: poke <?player name>",
			err:    true,
		}
	}
}

// Emotes

func (p *player) doSmile(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(
			commands["smile"],
			fmt.Sprintf("%s smiles happily", p.name),
			"You smile happily",
		)
	} else if len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			commands["smile"],
			name,
			fmt.Sprintf("%s smiles at you", p.name),
			fmt.Sprintf("You smile at %s", name),
			"You smile ... at yourself?",
		)
	} else {
		p.events <- event{
			player: p,
			output: "Usage: smile <?player name>",
			err:    true,
		}

	}
}

func (p *player) doScowl(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(
			commands["scowl"],
			fmt.Sprintf("%s scowls angrily", p.name),
			"You scowl angrily",
		)
	} else if len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			commands["scowl"],
			name,
			fmt.Sprintf("%s scowls at you", p.name),
			fmt.Sprintf("You scowl at %s", name),
			"You must really hate yourself...",
		)
	} else {
		p.events <- event{
			player: p,
			output: "Usage: scowl <?player name>",
			err:    true,
		}
	}
}

func (p *player) doSigh(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(
			commands["sigh"],
			fmt.Sprintf("%s sighs heavily", p.name),
			"You sigh heavily",
		)
	} else if len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			commands["sigh"],
			name,
			fmt.Sprintf("%s sighs at you", p.name),
			fmt.Sprintf("You sigh at %s", name),
			"Rough day, huh?",
		)
	} else {
		p.events <- event{
			player: p,
			output: "Usage: sigh <?player name>",
			err:    true,
		}
	}
}

func (p *player) doLaugh(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(commands["laugh"],
			fmt.Sprintf("%s laughs heartily",
				p.name),
			"You laugh heartily",
		)
	} else if len(words) == 1 {
		name := words[0]
		p.targetedRoomCommand(
			commands["laugh"],
			name,
			fmt.Sprintf("%s laughs at you", p.name),
			fmt.Sprintf("You laugh at %s", name),
			"It's always good to be able to laugh at yourself",
		)
	} else {
		p.events <- event{
			player: p,
			output: "Usage: laugh <?player name>",
			err:    true,
		}
	}
}

func (p *player) doThink(cmd string) {
	if words := strings.Fields(cmd); len(words) == 0 {
		p.roomCommand(
			commands["think"],
			fmt.Sprintf("%s is in deep thought", p.name),
			"You are in deep thought",
		)
	} else {
		p.events <- event{
			player: p,
			output: "Usage: think",
			err:    true,
		}
	}
}

// Special

// Disconnect the player gracefully
func (p *player) doQuit(_ string) {
	p.disconnect()
}

// Helper functions

// Represents a command that targets another player in the room
func (p *player) targetedRoomCommand(cmd *command, name string, outMsg string, selfMsg string, errSelf string) {
	if idx := index(len(p.room.players), func(i int) bool { return p.room.players[i].name == name }); idx != -1 {
		other := p.room.players[idx]
		if other != p {
			other.events <- event{
				player:  p,
				output:  outMsg,
				command: cmd,
			}
			p.events <- event{
				player:  p,
				output:  selfMsg,
				command: cmd,
			}
		} else {
			p.events <- event{
				player:  p,
				output:  errSelf,
				command: cmd,
			}
		}
	} else {
		p.events <- event{
			player: p,
			output: "No such player in this room!",
			err:    true,
		}
	}
}

// Represents a command that targets another player cross-server
func (p *player) targetedServerCommand(cmd *command, name string, outMsg string, selfMsg string, errSelf string) {
	if other, exists := players[name]; exists {
		if other != p {
			other.events <- event{
				player:  p,
				output:  outMsg,
				command: cmd,
			}
			p.events <- event{
				player:  p,
				output:  selfMsg,
				command: cmd,
			}
		} else {
			p.events <- event{
				player:  p,
				output:  errSelf,
				command: cmd,
			}
		}
	} else {
		p.events <- event{
			player: p,
			output: "No such player!",
			err:    true,
		}
	}
}

// A command that affects everyone in the room
func (p *player) roomCommand(cmd *command, outMsg string, selfMsg string) {
	for _, other := range p.room.players {
		if other != p {
			if ch := other.events; ch != nil {
				ch <- event{
					player:  p,
					output:  outMsg,
					command: cmd,
				}
			}
		} else {
			p.events <- event{
				player:  p,
				output:  selfMsg,
				command: cmd,
			}
		}
	}
}

// A command that affects everyone in the zone
func (p *player) zoneCommand(cmd *command, outMsg string, selfMsg string) {
	for _, other := range p.zone.players {
		if other != p {
			if ch := other.events; ch != nil {
				ch <- event{
					player:  p,
					output:  outMsg,
					command: cmd,
				}
			}
		} else {
			p.events <- event{
				player:  p,
				output:  selfMsg,
				command: cmd,
			}
		}
	}
}

// A command that affects everyone in the server
func (p *player) serverCommand(cmd *command, outMsg string, selfMsg string) {
	for _, other := range players {
		if other != p {
			if ch := other.events; ch != nil {
				ch <- event{
					player:  p,
					output:  outMsg,
					command: cmd,
				}
			}
		} else {
			p.events <- event{
				player:  p,
				output:  selfMsg,
				command: cmd,
			}
		}
	}
}
