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
		set.AddStep("Enabling 32bit packages", run.AddArch, "i386")
		set.AddStep("Updating package repositories", run.UpdateRepos)
		set.AddStep("Upgrading packages", run.UpgradePackages)

		set.AddStep("Enabling Flatpaks", run.EnableFlatpak)

		set.AddStep("Installing system tools", run.InstallPackages, "zsh nala lsd fonts-font-awesome neofetch mc tmux curl plocate libvirt-clients virt-manager sassc")
		set.AddStep("Making user member of virt groups", run.AssignGroups, "kvm,libvirt")
		set.AddStep("Installing basic development env", run.InstallPackages, "git build-essential golang libgl1-mesa-dev xorg-dev libglib2.0-dev-bin")
		set.AddStep("Enabling apt-file", run.EnableAptFile)

		set.AddStep("Installing Google Chrome", run.InstallPackages, "google-chrome-stable")
		set.AddStep("Installing VS Code", run.InstallPackages, "code")
		set.AddStep("Installing Steam", run.InstallPackages, "mesa-vulkan-drivers libglx-mesa0:i386 mesa-vulkan-drivers:i386 libgl1-mesa-dri:i386 steam-installer")

		set.AddStep("Creating pipewire directory", run.CreateDir, "/etc/pipewire")
		set.AddStep("Enabling HiFi audio", run.CopyFile, "pipewire.conf", "/etc/pipewire")

	}
	set.Run()
	fmt.Println("Setup complete. Enjoy!")
}
