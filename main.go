package main

import (
	"log"

	"github.com/chanzuckerberg/idseq-cli-v2/cmd"
)

func main() {
	log.SetFlags(0)
	cmd.Execute()
}
