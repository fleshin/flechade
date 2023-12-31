package main

import (
	"embed"
	"flag"
)

//go:embed data/*
var configFS embed.FS

func main() {

	showVersion()

	dataDir := flag.String("d", "", "Load customizations from local directory")
	repoUrl := flag.String("r", "", "Load customizations from GIT repository")
	runSet := flag.Bool("l", false, "Run default customizations")
	cont := flag.Bool("c", false, "Continue previous execution from the last successful step")

	flag.Parse()

	switch {
	case *cont:
		contPrevRun()
	case *runSet:
		runFromLocal(configFS)
	case *dataDir != "":
		runFromDir(*dataDir)
	case *repoUrl != "":
		runFromUrl(*repoUrl)
	default:
		flag.Usage()
	}
}
