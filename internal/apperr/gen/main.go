// Command gen writes the frontend list of error codes to the path given as
// its only argument.
package main

import (
	"fmt"
	"os"

	"github.com/olymp-software/chiron/internal/apperr"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: gen <output.ts>")
		os.Exit(2)
	}
	if err := os.WriteFile(os.Args[1], apperr.TypeScript(), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
