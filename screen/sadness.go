package screen

// import (
// 	"io"
// 	"mud/screen"
// 	"strings"
// )

// const (
// 	locationWidth  = 60
// 	locationHeight = 20

// 	chatHeight = fullHeight
// 	chatWidth  = 60

// 	mapWidth  = 60
// 	mapHeight = 24

// 	fullHeight = mapHeight + locationHeight + 1

// 	fullWidth = mapWidth + chatWidth + 1
// )

// func createMUDDisplay(w io.WriteCloser) (*mudDisplay, error) {
// 	var (
// 		s        *screen.Screen
// 		minimap  *screen.Section
// 		chat     *screen.Section
// 		location *screen.Section
// 		err      error
// 	)

// 	if s, err = screen.CreateScreen(fullWidth, fullHeight, w); err != nil {
// 		return nil, err
// 	}
// 	if minimap, err = s.NewSection(0, locationHeight+1, mapWidth, mapHeight); err != nil {
// 		return nil, err
// 	}
// 	// Horizontal divider
// 	if _, err = s.NewStaticSection(0, locationHeight, locationWidth, 1, strings.Repeat("-", locationWidth)); err != nil {
// 		return nil, err
// 	}
// 	if location, err = s.NewSection(0, 0, locationWidth, locationHeight); err != nil {
// 		return nil, err
// 	}
// 	// Vertical divider
// 	if _, err = s.NewStaticSection(mapWidth, 0, 1, fullHeight, strings.Repeat("|", fullHeight)); err != nil {
// 		return nil, err
// 	}
// 	if chat, err = s.NewSection(mapWidth+1, 0, chatWidth, chatHeight); err != nil {
// 		return nil, err
// 	}

// 	return &mudDisplay{
// 		screen:   s,
// 		minimap:  minimap,
// 		location: location,
// 		chat:     chat,
// 	}, nil
// }
