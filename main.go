package main

import (
	"log"

	"github.com/chanzuckerberg/czid-cli/cmd"
)

func main() {
	log.SetFlags(0)
	cmd.Execute()
}
