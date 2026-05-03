package main

import (
	"fmt"
	"os"
)

const usage = `notes — take and find markdown notes

Usage:
  notes new <title>    create a new note (issue #3)
  notes find <query>   search notes (issue #4)
  notes ls             list notes by recency (issue #5)
  notes help           print this message

Configuration: see README (issue #10).
`

func main() {
	fmt.Print(usage)
	if len(os.Args) > 1 && os.Args[1] != "help" {
		os.Exit(2)
	}
}
