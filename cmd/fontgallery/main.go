package main

import (
	"fmt"

	"github.com/mitchell-wallace/plaqq/internal/font"
)

func main() {
	names := font.Names()
	for i, name := range names {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("Font: %s\n", name)

		fmt.Println("Sample:")
		f := font.Get(name)
		for _, row := range f.Render("SHIP IT! BUILD FAILED 0123") {
			fmt.Println(row)
		}

		fmt.Println("Charset:")
		for _, row := range f.Render(font.Charset) {
			fmt.Println(row)
		}
	}
}
