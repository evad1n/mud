package screen

import "errors"

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

// Simble AABB collision
func isColliding(x int, y int, width int, height int, other *Section) bool {

	return x < other.X+other.Width && x+width > other.X &&
		y < other.Y+other.Height && y+height > other.Y
}

// Returns an error of the section dimensions are invalid
func (s *Screen) validateDimensions(x int, y int, width int, height int) error {
	if x < 0 || y < 0 || width < 0 || height < 0 {
		return errors.New("no negative values")
	}
	if x > s.Width-1 || x+width > s.Width {
		return errors.New("section x coordinates out of screen bounds")
	}
	if y > s.Height-1 || y+height > s.Height {
		return errors.New("section y coordinates out of screen bounds")
	}

	// Check for overlapping
	for _, section := range s.Sections {
		if isColliding(x, y, width, height, section) {
			return errors.New("overlapping with already defined section")
		}
	}

	return nil
}
