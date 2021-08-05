package main

import (
	// why do this?, is to tell initialization on matcher, even though we are not using in main
	_ "github.com/jonnyGit81/chat/goInaction/chapter2/matchers"
	"github.com/jonnyGit81/chat/goInaction/chapter2/search"

	"log"
	"os"
)

// init is called prior to main.
func init() {
	// Change the device for logging to stdout.
	log.SetOutput(os.Stdout)
}

// main is the entry point for the program.
func main() {
	// Perform the search for the specified term.
	search.Run("president")
}
