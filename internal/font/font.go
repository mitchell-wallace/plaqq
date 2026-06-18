package font

import (
	"strings"
)

// Height is the row height of all font patterns
const Height = 5

// letters maps runes to their 5-row chunky representation.
var letters = map[rune][5]string{
	'A': {
		" ▄▄▄▄ ",
		"█    █",
		"█▀▀▀▀█",
		"█    █",
		"█    █",
	},
	'B': {
		"████▄ ",
		"█   █ ",
		"████▀ ",
		"█   █ ",
		"████▀ ",
	},
	'C': {
		" ▄████",
		"█▀    ",
		"█     ",
		"█▄    ",
		" ▀████",
	},
	'D': {
		"████▄ ",
		"█   █ ",
		"█   █ ",
		"█   █ ",
		"████▀ ",
	},
	'E': {
		"██████",
		"█     ",
		"█████ ",
		"█     ",
		"██████",
	},
	'F': {
		"██████",
		"█     ",
		"█████ ",
		"█     ",
		"█     ",
	},
	'G': {
		" ▄████",
		"█▀    ",
		"█  ███",
		"█▄   █",
		" ▀████",
	},
	'H': {
		"█    █",
		"█    █",
		"██████",
		"█    █",
		"█    █",
	},
	'I': {
		"█████",
		"  █  ",
		"  █  ",
		"  █  ",
		"█████",
	},
	'J': {
		"   ███",
		"     █",
		"     █",
		"█    █",
		" ▀███▀",
	},
	'K': {
		"█   █",
		"█  █ ",
		"███  ",
		"█  █ ",
		"█   █",
	},
	'L': {
		"█     ",
		"█     ",
		"█     ",
		"█     ",
		"██████",
	},
	'M': {
		"█   █",
		"██ ██",
		"█ █ █",
		"█   █",
		"█   █",
	},
	'N': {
		"█    █",
		"██   █",
		"█ █  █",
		"█  █ █",
		"█   ██",
	},
	'O': {
		" ▄██▄ ",
		"█    █",
		"█    █",
		"█    █",
		" ▀██▀ ",
	},
	'P': {
		"████▄ ",
		"█   █ ",
		"████▀ ",
		"█     ",
		"█     ",
	},
	'Q': {
		" ▄██▄ ",
		"█    █",
		"█  █ █",
		"█   ██",
		" ▀██▀▀",
	},
	'R': {
		"████▄ ",
		"█   █ ",
		"████▀ ",
		"█  █  ",
		"█   █ ",
	},
	'S': {
		" ▄████",
		"▀█▄▄▄ ",
		" ▀▀▀█▄",
		"▄▄▄▄█▀",
		"████▀ ",
	},
	'T': {
		"███████",
		"   █   ",
		"   █   ",
		"   █   ",
		"   █   ",
	},
	'U': {
		"█    █",
		"█    █",
		"█    █",
		"█    █",
		" ▀██▀ ",
	},
	'V': {
		"█    █",
		"█    █",
		"█    █",
		" █  █ ",
		"  ▀▀  ",
	},
	'W': {
		"█      █",
		"█      █",
		"█  ▄▄  █",
		"█ █  █ █",
		" ▀    ▀ ",
	},
	'X': {
		"█    █",
		" █  █ ",
		"  ██  ",
		" █  █ ",
		"█    █",
	},
	'Y': {
		"█    █",
		" █  █ ",
		"  ██  ",
		"  ██  ",
		"  ██  ",
	},
	'Z': {
		"██████",
		"    ▄▀",
		"  ▄▀  ",
		"▄▀    ",
		"██████",
	},
	'0': {
		"█████",
		"█   █",
		"█   █",
		"█   █",
		"█████",
	},
	'1': {
		" ▄█ ",
		"  █ ",
		"  █ ",
		"  █ ",
		"████",
	},
	'2': {
		"█████",
		"    █",
		"█████",
		"█    ",
		"█████",
	},
	'3': {
		"█████",
		"    █",
		"█████",
		"    █",
		"█████",
	},
	'4': {
		"█   █",
		"█   █",
		"█████",
		"    █",
		"    █",
	},
	'5': {
		"█████",
		"█    ",
		"█████",
		"    █",
		"█████",
	},
	'6': {
		"█████",
		"█    ",
		"█████",
		"█   █",
		"█████",
	},
	'7': {
		"█████",
		"    █",
		"   █ ",
		"  █  ",
		" █   ",
	},
	'8': {
		"█████",
		"█   █",
		"█████",
		"█   █",
		"█████",
	},
	'9': {
		"█████",
		"█   █",
		"█████",
		"    █",
		"█████",
	},
	' ': {
		"    ",
		"    ",
		"    ",
		"    ",
		"    ",
	},
	'!': {
		"█",
		"█",
		"█",
		" ",
		"█",
	},
	'.': {
		" ",
		" ",
		" ",
		" ",
		"█",
	},
	',': {
		" ",
		" ",
		" ",
		"▄",
		"█",
	},
	'?': {
		"█████",
		"    █",
		"  ██▀",
		"     ",
		"  █  ",
	},
	'-': {
		"     ",
		"     ",
		"█████",
		"     ",
		"     ",
	},
	'\'': {
		"█",
		"█",
		" ",
		" ",
		" ",
	},
	'"': {
		"█ █",
		"█ █",
		"   ",
		"   ",
		"   ",
	},
	':': {
		" ",
		"▄",
		" ",
		"▄",
		" ",
	},
	'/': {
		"    █",
		"   █ ",
		"  █  ",
		" █   ",
		"█    ",
	},
	'(': {
		" ▄█",
		"█  ",
		"█  ",
		"█  ",
		" ▀█",
	},
	')': {
		"█▄ ",
		"  █",
		"  █",
		"  █",
		"▀█ ",
	},
}

// RuneWidth returns the display column width of a single character in the font.
func RuneWidth(r rune) int {
	pattern, ok := letters[r]
	if !ok {
		pattern = letters['?']
	}
	return len([]rune(pattern[0]))
}

// StringWidth returns the total display column width of a string rendered in chunky font.
func StringWidth(s string) int {
	if len(s) == 0 {
		return 0
	}
	w := 0
	runes := []rune(strings.ToUpper(s))
	for i, r := range runes {
		if i > 0 {
			w += 1 // spacing between letters
		}
		w += RuneWidth(r)
	}
	return w
}

// RenderString returns the 5-row representation of a single-line string.
func RenderString(s string) [Height]string {
	var result [Height]string
	runes := []rune(strings.ToUpper(s))
	for i, r := range runes {
		pattern, ok := letters[r]
		if !ok {
			pattern = letters['?']
		}
		for row := 0; row < Height; row++ {
			if i > 0 {
				result[row] += " " // spacing between letters
			}
			result[row] += pattern[row]
		}
	}
	return result
}

// WrapText splits a long message into multiple lines, where each line's rendered chunky width
// does not exceed maxWidth.
func WrapText(text string, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	var currentLine []string

	for _, word := range words {
		joined := word
		if len(currentLine) > 0 {
			joined = strings.Join(currentLine, " ") + " " + word
		}
		if StringWidth(joined) <= maxWidth {
			currentLine = append(currentLine, word)
		} else {
			if len(currentLine) > 0 {
				lines = append(lines, strings.Join(currentLine, " "))
				currentLine = []string{word}
			} else {
				// Word itself is wider than maxWidth, place on its own line
				lines = append(lines, word)
				currentLine = nil
			}
		}
	}
	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, " "))
	}
	return lines
}
