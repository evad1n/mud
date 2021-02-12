package main

import (
	"fmt"
	"strings"
)

const (
	fullWidth = 140
)

var ()

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
			zeroCol(p)
			fmt.Fprintf(p.conn, "\x1b[1A\x1b[2D%c\x1b[1B", '║')
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
			zeroCol(p)
			fmt.Fprintf(p.conn, "\x1b[1A\x1b[2D%c\x1b[1B", '║')
			col = 0
			continue
		}
		col++
	}
	// Print remaining text
	zeroCol(p)
	fmt.Fprintf(p.conn, "%s\n\n", text)

	// Make space for prompt (so map fits snugly)
	fmt.Fprintf(p.conn, "\n")

	p.drawMap()

	// New prompt
	p.prompt()

	p.drawDivider()

	// Move cursor to correct position
	fmt.Fprintf(p.conn, "\x1b[1000B")
	zeroCol(p)
	fmt.Fprintf(p.conn, "\x1b[4C")
}

// Go to the 0 column for the event display
func zeroCol(p *player) {
	fmt.Fprintf(p.conn, "\x1b[%dG", p.minimap.width+3)
}

// Go to the 0 column for the event display
func dividerCol(p *player) {
	fmt.Fprintf(p.conn, "\x1b[%dG", p.minimap.width+1)
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
	// Back to print location
	fmt.Fprint(p.conn, "\x1b[1B")
	zeroCol(p)
}

// Display player command prompt
func (p *player) prompt() {
	// Go to bottom fo screen in events channel
	fmt.Fprintf(p.conn, "\x1b[10000B")
	zeroCol(p)
	fmt.Fprintf(p.conn, "\x1b[1A")
	zeroCol(p)
	fmt.Fprintf(p.conn, "%s\x1b[1B", strings.Repeat("_", fullWidth-p.minimap.width)) // Separator
	zeroCol(p)
	fmt.Fprintf(p.conn, ">>> ")
}

// Draws the vertical divider for the visible screen
func (p *player) drawDivider() {
	// Cursor top left
	fmt.Fprintf(p.conn, "\x1b[1000A")
	dividerCol(p)
	for i := 0; i <= 100; i++ {
		fmt.Fprintf(p.conn, "%c\x1b[1B\x1b[1D", '║')
	}

}

// Wrap some text in an ansi code
func ansiWrap(text string, code string) string {
	return fmt.Sprintf("%s%s\x1b[0m", code, text)
}
