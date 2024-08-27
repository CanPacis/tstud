package main

import (
	"os"

	"github.com/CanPacis/tstud-core/cli"
	"github.com/CanPacis/tstud-core/proto"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	if len(os.Args) < 2 {
		proto.Run()
	} else {
		cli.Run()
	}
}
