package main

import (
	"fmt"
	"strings"
)

const (
	fullWidth = 140
)

func (p *player) eventPrint(ev event) {
	width := fullWidth - p.minimap.width
	text := ev.output

	// Erase old prompt
	p.erasePrompt(ev)

	col := 0
	for len(text) > col {
		if text[col] == '\n' {
			line := text[:col]
			text = text[col+1:]
			zeroCol(p)
			fmt.Fprintf(p.conn, "%s\n", line)
			col = 0
			continue
		}
		if col > width {
			rawLine := text[:col]
			truncateIdx := strings.LastIndex(rawLine, " ")
			line := text[:truncateIdx]
			text = text[truncateIdx:]
			zeroCol(p)
			fmt.Fprintf(p.conn, "%s\n", line)
			col = 0
			continue
		}
		col++
	}
	// Print remaining text
	zeroCol(p)
	fmt.Fprintf(p.conn, "%s\n\n", text)

	p.drawMap()

	// New prompt
	p.prompt()
}

// Go to the 0 column for the event display
func zeroCol(p *player) {
	fmt.Fprintf(p.conn, "\x1b[%dG", p.minimap.width+3)
}

// Erase player's old prompt
func (p *player) erasePrompt(ev event) {
	// Clear last 2 lines
	if !ev.unsolicited {
		// Account for newline from user pressing enter
		fmt.Fprintf(p.conn, "\x1b[1A")
	}
	for i := 0; i < 2; i++ {
		fmt.Fprintf(p.conn, "\x1b[%dG\x1b[0K\x1b[1A", p.minimap.width)
	}
	// Back to bottom
	// fmt.Fprintf(p.conn, "\x1b[3B")
	zeroCol(p)
}

// Display player command prompt
func (p *player) prompt() {
	zeroCol(p)
	fmt.Fprintf(p.conn, "\n")
	zeroCol(p)
	fmt.Fprintf(p.conn, "%s\n", strings.Repeat("_", fullWidth-p.minimap.width)) // Separator
	zeroCol(p)
	fmt.Fprintf(p.conn, ">>> ")
}

// Wrap some text in an ansi code
func ansiWrap(text string, code string) string {
	return fmt.Sprintf("%s%s\x1b[0m", code, text)
}
