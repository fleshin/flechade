package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/fleshin/flechade/run"
)

//go:embed data/*
var configFS embed.FS

func main() {

	bi, _ := debug.ReadBuildInfo()
	ver := bi.Settings[11].Value

	fmt.Println("flechade - customize your linux")
	fmt.Println("Version:", ver)
	fmt.Println("")

	if os.Geteuid() != 0 {
		log.Fatal("This tool needs root access. Please use sudo.")
	}

	fmt.Print("Setting up the environment: ")

	set := run.NewSet()
	fmt.Println(set.OSrel)
	set.SetFiles(configFS)
	if err := set.Load(); err != nil {
		// Adding repositories
		set.AddStep("Installing Debian Testing repo sources", run.CopyFile, "sources.list", "/etc/apt")
		set.AddStep("Google Chrome repo sources", run.CopyFile, "google-chrome.list", "/etc/apt/sources.list.d/")
		set.AddStep("Adding Google public keys", run.AddRepoKey, "https://dl.google.com/linux/linux_signing_key.pub", "/etc/apt/keyrings/linux_signing_key.pub")
		set.AddStep("MS VSCode repo sources", run.CopyFile, "vscode.list", "/etc/apt/sources.list.d/")
		set.AddStep("Adding MS public keys", run.AddRepoKey, "https://packages.microsoft.com/keys/microsoft.asc", "/etc/apt/keyrings/packages.microsoft.asc")
		set.AddStep("Updating package repositories", run.UpdateRepos)
		set.AddStep("Upgrading packages", run.UpgradePackages)
		set.AddStep("Creating pipewire directory", run.CreateDir, "/etc/pipewire")
		set.AddStep("Enabling HiFi audio", run.CopyFile, "pipewire.conf", "/etc/pipewire")
	}
	set.Run()
	fmt.Println("Setup complete. Enjoy!")
}
