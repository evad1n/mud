package main

import (
	"fmt"
	"sort"
	"strings"
)

// Command represents a user function with its associated name
type Command struct {
	Name string
	// The linked function
	Run func(*Player, string)
}

// Map of command aliases to commands
var commands map[string]Command

// Direction abbreviations
var dirs map[string]string

// Maps the exit direction string to its place in the Room exit array
var dirRuneToInt map[rune]int

// Maps a place in the Room exit array to a direction abbreviation rune
var dirIntToRune map[int]rune

// Adds all starting commands
func defaultCommands() {
	commands = make(map[string]Command)

	// In order of precedence

	// Navigation
	addCommand("look", Command{"look", doLook})
	addCommand("north", Command{"north", doNorth})
	addCommand("south", Command{"south", doSouth})
	addCommand("east", Command{"east", doEast})
	addCommand("west", Command{"west", doWest})
	addCommand("up", Command{"up", doUp})
	addCommand("down", Command{"down", doDown})
	addCommand("recall", Command{"recall", doRecall})
	// Communication
	addCommand("say", Command{"say", doSay})
	addCommand("tell", Command{"tell", doTell})
	addCommand("laugh", Command{"laugh", doLaugh})
	addCommand("sigh", Command{"sigh", doSigh})
	addCommand("smile", Command{"smile", doSmile})
	addCommand("scowl", Command{"scowl", doScowl})
	addCommand("think", Command{"think", doThink})
	// Special
	addCommand("cmds", Command{"cmds", listCommands})
	addCommand("help", Command{"cmds", listCommands})
	addCommand("quit", Command{"quit", doQuit})
	addCommand("exit", Command{"quit", doQuit})
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

//////////////
// Commands //
//////////////

// Special commands

// Lists known aliases for commands
func listCommands(p *Player, _ string) {
	fmt.Fprintln(p.Conn, "------------------")
	fmt.Fprintln(p.Conn, "All known commands")
	fmt.Fprintln(p.Conn, "------------------")

	// Sort commands primarily by name in alphabetical order, secondarily by alias

	// A reverse of the commands map (from Command.Name -> slice of aliases)
	cmdMap := make(map[string][]string)
	for alias, cmd := range commands {
		cmdMap[cmd.Name] = append(cmdMap[cmd.Name], alias)
	}

	type AliasList struct {
		Name    string
		Aliases []string
	}

	cmdList := []AliasList{}
	for name, aliases := range cmdMap {
		cmdList = append(cmdList, AliasList{name, aliases})
	}

	// Sort individual groups
	for _, a := range cmdList {
		sort.Strings(a.Aliases)
	}

	// Sort whole list
	sort.Slice(cmdList, func(i, j int) bool {
		return cmdList[i].Name < cmdList[j].Name
	})

	for _, cmd := range cmdList {
		for _, alias := range cmd.Aliases {
			fmt.Fprintf(p.Conn, "%-12s => %s\n", alias, cmd.Name)
		}
	}
}

func doQuit(p *Player, _ string) {
	fmt.Fprintf(p.Conn, "Goodbye %s!", p.Name)
}

func doRecall(p *Player, _ string) {
	fmt.Fprint(p.Conn, "You head back to the Temple of Midgard...\n\n")
	p.Location = rooms[3001]
	printLocation(p)
}

// Navigation

func doLook(p *Player, direction string) {
	if len(direction) == 0 {
		printLocation(p)
	} else {
		if fullDir, exists := dirs[direction]; exists {
			lookDirection(p, fullDir)
		} else {
			fmt.Fprintln(p.Conn, "Usage: look <north|south|east|west|up|down>")
		}
	}
}

func lookDirection(p *Player, dir string) {
	if exit := p.Location.Exits[dirRuneToInt[rune(strings.ToLower(dir)[0])]]; exit.To != nil {
		fmt.Fprint(p.Conn, exit.Description)
	} else {
		fmt.Fprintln(p.Conn, "There's nothing there...")
	}

}

func doNorth(p *Player, _ string) {
	moveDirection(p, 0)
}
func doEast(p *Player, _ string) {
	moveDirection(p, 1)
}
func doWest(p *Player, _ string) {
	moveDirection(p, 2)
}
func doSouth(p *Player, _ string) {
	moveDirection(p, 3)
}
func doUp(p *Player, _ string) {
	moveDirection(p, 4)
}
func doDown(p *Player, _ string) {
	moveDirection(p, 5)
}

// Make sure it is a valid direction
func moveDirection(p *Player, dir int) {
	if exit := p.Location.Exits[dir]; exit.To != nil {
		p.Location = exit.To
		printLocation(p)
	} else {
		fmt.Fprintln(p.Conn, "You can't go that way...")
	}
}

// Communication

func doSay(p *Player, msg string) {
	fmt.Fprintf(p.Conn, "You said: '%s'\n", msg)
}

func doTell(p *Player, cmd string) {
	if words := strings.Fields(cmd); len(words) > 1 {
		person := words[0]
		msg := strings.Join(words[1:], " ")
		fmt.Fprintf(p.Conn, "You whisper to %s: '%s'\n", person, msg)
	} else {
		fmt.Fprintln(p.Conn, "Usage: tell <Player Name> <Message>")
	}
}

// Emotes

func doSigh(p *Player, _ string) {
	fmt.Fprintln(p.Conn, "*You sigh heavily*")
}

func doLaugh(p *Player, _ string) {
	fmt.Fprintln(p.Conn, "*You laugh heartily*")
}

func doSmile(p *Player, _ string) {
	fmt.Fprintln(p.Conn, "*You smile happily*")
}

func doScowl(p *Player, _ string) {
	fmt.Fprintln(p.Conn, "*You scowl disapprovingly*")
}

func doThink(p *Player, _ string) {
	fmt.Fprintln(p.Conn, "*You put on your thinking cap*")
}
