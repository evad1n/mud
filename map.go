package main

import (
	"strings"
)

type (
	pair struct {
		x, y int
	}

	mapBuilder struct {
		depth  int
		width  int // The width of the drawn map in characters
		height int // The height of the drawn map in characters
		grid   map[pair]*room
		text   map[pair]rune
	}
)

const (
	xScale = 6
	yScale = 4
)

var (
	mapBox = [][]rune{
		{'╔', '═', '═', '═', '╗'},
		{'║', ' ', ' ', ' ', '║'},
		{'╚', '═', '═', '═', '╝'},
	}
	cross            = 'X'
	biArrows         = []rune{'⭥', '⇆', '⇄', '⭥', '⮥', '⮦', '⮦', '⮥'}
	inZoneArrows     = []rune{'↑', '→', '←', '↓', '⮭', '⮮'}
	unknownArrows    = []rune{'⇧', '⇨', '⇦', '⇩', '⮭', '⮮'}
	outZoneArrows    = []rune{'⇑', '⇒', '⇐', '⇓', '⇗', '⇙'}
	oppositeDirction = []int{3, 2, 1, 0, 5, 4}
	dxByIndex        = []int{0, 1, -1, 0}
	dyByIndex        = []int{1, 0, 0, -1}
)

func newMapBuilder(depth int) *mapBuilder {
	return &mapBuilder{
		depth:  depth,
		width:  (2 * depth * xScale) + 6,
		height: (2 * depth * yScale) + 4,
		grid:   make(map[pair]*room),
		text:   make(map[pair]rune),
	}
}

func (m *mapBuilder) trace(start *room, visited map[int]bool) {
	// Clear
	m.text = make(map[pair]rune)
	m.grid = make(map[pair]*room)
	m.grid[pair{0, 0}] = start
	q := []pair{{0, 0}}

	for len(q) > 0 {
		here := q[0]
		r := m.grid[here]
		// Pop off queue
		q = q[1:]

		// The limits of the drawn map
		if here.x > m.depth || here.x < -m.depth || here.y > m.depth || here.y < -m.depth {
			continue
		}
		m.drawBox(here)

		// Draw exits and follow
		for forward := 0; forward < 6; forward++ {
			backward := oppositeDirction[forward]
			var target *room
			if target = r.exits[forward].to; target == nil {
				continue
			}

			_, seen := visited[target.id]

			var (
				existing, back *room
				dx, dy         int
			)
			if forward < 4 {
				dx, dy = dxByIndex[forward], dyByIndex[forward]
				existing = m.grid[pair{here.x + dx, here.y + dy}]
			}
			back = target.exits[backward].to

			switch {
			case r.zone != target.zone:
				m.drawExit(here, outZoneArrows[forward], forward)
			case !seen:
				m.drawExit(here, unknownArrows[forward], forward)
			case forward >= 4:
				if r == back {
					m.drawExit(here, biArrows[forward], forward)
					m.drawExit(here, biArrows[forward+2], forward+2)
				} else {
					m.drawExit(here, inZoneArrows[forward], forward)
				}
			case existing == nil:
				loc := pair{here.x + dx, here.y + dy}
				m.grid[loc] = target
				q = append(q, loc)

				fallthrough
			case existing == target:
				if r == back {
					m.drawExit(here, biArrows[forward], forward)
				} else {
					m.drawExit(here, inZoneArrows[forward], forward)
				}
			default:
				m.drawExit(here, inZoneArrows[forward], forward)
			}
		}
	}
}

func (m *mapBuilder) render() []string {
	var (
		w     strings.Builder
		lines []string
	)
	for y := m.depth*yScale + 2; y >= -m.depth*yScale-2; y-- {
		for x := -m.depth*xScale - 3; x <= m.depth*xScale+3; x++ {
			if ch, present := m.text[pair{x, y}]; present {
				if ch == cross {
					w.WriteString(ansiWrap(string(ch), ansiColors["red"]))
				} else if contain(len(unknownArrows), func(idx int) bool { return unknownArrows[idx] == ch }) {
					w.WriteString(ansiWrap(string(ch), ansiColors["cyan"]))
				} else if contain(len(outZoneArrows), func(idx int) bool { return outZoneArrows[idx] == ch }) {
					w.WriteString(ansiWrap(string(ch), ansiColors["magenta"]))
				} else {
					w.WriteRune(ch)
				}
			} else {
				w.WriteRune(' ')
			}
		}
		lines = append(lines, w.String())
		w.Reset()
	}
	return lines
}

func (m *mapBuilder) drawBox(center pair) {
	x, y := textCoords(center)

	for yy, row := range mapBox {
		for xx, elt := range row {
			m.text[pair{x - 2 + xx, y + 1 - yy}] = elt
		}
	}

	if center.x == 0 && center.y == 0 {
		m.text[pair{x, y}] = cross
	}
}

func (m *mapBuilder) drawExit(center pair, arrow rune, dir int) {
	x, y := textCoords(center)
	switch dir {
	case 0: // north
		m.text[pair{x, y + 2}] = arrow
	case 1: // east
		m.text[pair{x + 3, y}] = arrow
	case 2: // west
		m.text[pair{x - 3, y}] = arrow
	case 3: // south
		m.text[pair{x, y - 2}] = arrow
	case 4: // up
		m.text[pair{x + 3, y + 2}] = arrow
	case 5: // down
		m.text[pair{x - 3, y - 2}] = arrow
	case 6: // up bi
		m.text[pair{x + 2, y + 2}] = arrow
	case 7: // down bi
		m.text[pair{x - 2, y - 2}] = arrow
	default:
		m.text[pair{x - 3, y + 2}] = arrow
	}
}

func textCoords(center pair) (int, int) {
	return center.x * xScale, center.y * yScale
}
