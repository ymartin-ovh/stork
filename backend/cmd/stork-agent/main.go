package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	flags "github.com/jessevdk/go-flags"

	"isc.org/stork/agent"
)


func main() {
	storkAgent := agent.StorkAgent{}

	// Setup logging
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// Prepare parse for command line flags.
	parser := flags.NewParser(&storkAgent.Settings, flags.Default)
	parser.ShortDescription = "Stork Agent"
	parser.LongDescription = "Stork Agent"

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	storkAgent.Serve()
}