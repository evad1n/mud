// Package screen provides a way to multiplex (divide the screen into sections) in a single terminal screen
package screen

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type (
	// Screen represents the desired dimensions to display
	Screen struct {
		Width      int               // Maximum width of the screen to display text on
		Height     int               // Maximum height of the screen to display text on
		Writer     io.Writer         // The writer to print the screen to
		Sections   []*Section        // The independent sections of the screen
		SectionMap map[pair]*Section // Map coords to the corresponding section
	}

	// Section is an independent part of the screen
	Section struct {
		X      int // Starting x position in screen coords
		Y      int // Starting y position in screen coords. Y increases upwards.
		Width  int
		Height int
		Text   map[pair]rune // The actual drawn characters
		static bool          // Whether it can be written to
	}

	pair struct {
		x, y int
	}
)

// CreateScreen creates a new Screen with the specified dimensions
func CreateScreen(width int, height int, writer io.Writer) (*Screen, error) {
	if width < 0 || height < 0 {
		return nil, errors.New("no negative values")
	}

	return &Screen{
		Width:      width,
		Height:     height,
		Writer:     writer,
		Sections:   []*Section{},
		SectionMap: make(map[pair]*Section),
	}, nil
}

// NewSection creates and returns a new section of the screen
// Uses screen coordinates
func (s *Screen) NewSection(x int, y int, width int, height int) (*Section, error) {
	if err := s.validateDimensions(x, y, width, height); err != nil {
		return nil, fmt.Errorf("validating dimensions: %v", err)
	}

	section := &Section{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		Text:   make(map[pair]rune),
	}

	s.Sections = append(s.Sections, section)

	for y := section.Y; y < section.Y+section.Height; y++ {
		for x := section.X; x < section.X+section.Width; x++ {
			s.SectionMap[pair{x, y}] = section
			section.Text[pair{x, y}] = ' '
		}
	}

	return section, nil
}

// NewStaticSection creates a section that can't be written to. Instead it accepts a starting string that will always be displayed. Userful for borders
func (s *Screen) NewStaticSection(x int, y int, width int, height int, text string) (*Section, error) {
	if err := s.validateDimensions(x, y, width, height); err != nil {
		return nil, fmt.Errorf("validating dimensions: %v", err)
	}

	section := &Section{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		Text:   make(map[pair]rune),
		static: true,
	}

	s.Sections = append(s.Sections, section)

	for y := section.Y; y < section.Y+section.Height; y++ {
		for x := section.X; x < section.X+section.Width; x++ {
			s.SectionMap[pair{x, y}] = section
			section.Text[pair{x, y}] = ' '
		}
	}

	// Fill with static text
	section.staticWrite(text)

	return section, nil
}

// RemoveSection removes the requested section from the screen
func (s *Screen) RemoveSection(section *Section) error {
	if i := index(len(s.Sections), func(idx int) bool { return s.Sections[idx] == section }); i != -1 {
		s.Sections = append(s.Sections[:i], s.Sections[i+1:]...)

		for y := section.Y; y < section.Y+section.Height; y++ {
			for x := section.X; x < section.X+section.Width; x++ {
				s.SectionMap[pair{x, y}] = nil
			}
		}
		return nil
	}

	return errors.New("no such section")
}

// Render renders all the individual sections of the screen
func (s *Screen) Render() {
	var (
		b strings.Builder
	)
	for y := s.Height - 1; y >= 0; y-- {
		for x := 0; x < s.Width; x++ {
			if section, mapped := s.SectionMap[pair{x, y}]; mapped {
				if ch, written := section.Text[pair{x, y}]; written {
					b.WriteRune(ch)
				} else {
					b.WriteRune('#')
				}
			} else {
				b.WriteRune('@')
			}
		}
		if y != 0 {
			b.WriteRune('\n')
		}
	}

	fmt.Fprint(s.Writer, "\x1b[2J")
	fmt.Fprintf(s.Writer, b.String())
}

func (s *Section) Write(text string) error {
	if s.static {
		return errors.New("can't write to static section")
	}

	var lines []string
	// Split into lines
	for len(text) > s.Width {
		line := text[:s.Width]
		text = text[s.Width:]
		// Split on newlines
		for newLineIdx := strings.IndexRune(line, '\n'); newLineIdx != -1; newLineIdx = strings.IndexRune(line, '\n') {
			l := line[:newLineIdx]
			line = line[newLineIdx+1:]

			// Pad lines with spaces
			l += strings.Repeat(" ", s.Width-len(l))
			lines = append(lines, l)
			fill := s.Width - len(line)
			if fill > len(text) {
				line += text
				text = ""
			} else {
				line += text[:fill]
				text = text[fill:]
			}
		}
		line += strings.Repeat(" ", s.Width-len(line))
		lines = append(lines, line)
	}
	lines = append(lines, text)

	if len(lines) > s.Height {
		lines = lines[len(lines)-s.Height:]
	}

	var oldLines []string
	var b strings.Builder
	// Move old text up by number of added lines
	for y := s.Y; y < s.Height+s.Y; y++ {
		for x := s.X; x < s.Width+s.X; x++ {
			b.WriteRune(s.Text[pair{x, y}])
		}
		oldLines = append(oldLines, b.String())
		b.Reset()
	}
	// Truncate old lines to get space for new lines
	oldLines = oldLines[:s.Height-len(lines)]

	lines = append(lines, oldLines...)

	// Now map lines to text
	for y, line := range lines {
		for x, ch := range line {
			s.Text[pair{s.X + x, s.Y + s.Height - y}] = ch
		}
	}

	return nil
}

func (s *Section) staticWrite(text string) {
	var lines []string
	// Split into lines
	for len(text) > s.Width {
		lines = append(lines, text[:s.Width])
		text = text[s.Width:]
	}
	lines = append(lines, text)

	if len(lines) > s.Height {
		lines = lines[len(lines)-s.Height:]
	}

	var oldLines []string
	var b strings.Builder
	// Move old text up by number of added lines
	for y := s.Y; y < s.Height+s.Y; y++ {
		for x := s.X; x < s.Width+s.X; x++ {
			b.WriteRune(s.Text[pair{x, y}])
		}
		oldLines = append(oldLines, b.String())
		b.Reset()
	}
	// Truncate old lines to get space for new lines
	oldLines = oldLines[:s.Height-len(lines)]

	lines = append(lines, oldLines...)

	// Now map lines to text
	for y, line := range lines {
		for x, ch := range line {
			s.Text[pair{s.X + x, s.Y + y}] = ch
		}
	}
}
