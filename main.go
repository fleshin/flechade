package main

import (
	"embed"
	"flag"
)

//go:embed data/*
var configFS embed.FS

func main() {

	showVersion()

	dataDir := flag.String("d", "", "directory to load from")
	repoUrl := flag.String("u", "", "GIT repo to load from")
	//help := flag.Bool("h", false, "Print usage help")
	runSet := flag.Bool("l", false, "Run default configuration set")

	flag.Parse()

	switch {
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
