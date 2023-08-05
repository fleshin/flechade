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
		set.AddStep("Installing Debian Testing sources", run.CopyFile, "sources.list", "/etc/apt")
		set.AddStep("Google Chrome repo sources", run.CopyFile, "google-chrome.list", "/etc/apt/sources.list.d/")
		set.AddStep("Adding Google public keys", run.AddRepoKey, "https://dl.google.com/linux/linux_signing_key.pub", "/etc/apt/keyrings/linux_signing_key.pub")
		set.AddStep("MS VSCode repo sources", run.CopyFile, "vscode.list", "/etc/apt/sources.list.d/")
		set.AddStep("Adding MS public keys", run.AddRepoKey, "https://packages.microsoft.com/keys/microsoft.asc", "/etc/apt/keyrings/packages.microsoft.asc")
		set.AddStep("Enabling 32bit packages", run.AddArch, "i386")
		set.AddStep("Updating package repositories", run.UpdateRepos)
		set.AddStep("Upgrading packages", run.UpgradePackages)

		set.AddStep("Enabling Flatpaks", run.EnableFlatpak)

		set.AddStep("Installing system tools", run.InstallPackages, "zsh nala lsd fonts-font-awesome neofetch mc tmux curl plocate libvirt-clients virt-manager sassc dbus-x11")
		set.AddStep("Making user member of virt groups", run.AssignGroups, "kvm,libvirt")
		set.AddStep("Installing basic development env", run.InstallPackages, "git build-essential golang libgl1-mesa-dev xorg-dev libglib2.0-dev-bin")
		set.AddStep("Enabling apt-file", run.EnableAptFile)

		set.AddStep("Installing Flatpak apps", run.InstallFlatpaks, "com.github.tchx84.Flatseal com.usebottles.bottles com.github.wwmm.easyeffects net.davidotek.pupgui2 com.slack.Slack org.gnome.Geary")

		set.AddStep("Installing Google Chrome", run.InstallPackages, "google-chrome-stable")
		set.AddStep("Installing VS Code", run.InstallPackages, "code")
		set.AddStep("Installing Steam", run.InstallPackages, "mesa-vulkan-drivers libglx-mesa0:i386 mesa-vulkan-drivers:i386 libgl1-mesa-dri:i386 steam-installer")

		set.AddStep("Installing Nerd fonts", run.CloneAndRun, "https://github.com/ryanoasis/nerd-fonts.git", "install.sh --install-to-system-path")

		set.AddStep("Installing Gnome Extension Dash to Dock", run.InstallGnomeExt, "dash-to-dock@micxgx.gmail.com", "84")
		set.AddStep("Installing Gnome Extension OpenWheater", run.InstallGnomeExt, "openweather-extension@jenslody.de", "121")
		set.AddStep("Installing Gnome Extension Tray Icons", run.InstallGnomeExt, "trayIconsReloaded@selfmade.pl", "26")
		set.AddStep("Installing Gnome Extension Blur My Shell", run.InstallGnomeExt, "blur-my-shell@aunetx", "47")

		set.AddStep("Enabling Gnome Extension User Themes", run.EnableGnomeExt, "user-theme@gnome-shell-extensions.gcampax.github.com")

		//set.AddStep("Installing WhiteSur Gnome theme", run.CloneAndRun, "https://github.com/vinceliuice/WhiteSur-gtk-theme.git", "install.sh -l -c Light")
		set.AddStep("Installing WhiteSur Gnome theme", run.CloneAndRunAsUser, "https://github.com/vinceliuice/WhiteSur-gtk-theme.git", "install.sh -l -c Light")
		set.AddStep("Installing WhiteSur Nautilus theme", run.CloneAndRun, "https://github.com/vinceliuice/WhiteSur-gtk-theme.git", "install.sh -N mojave")
		set.AddStep("Installing WhiteSur GDM tweaks", run.CloneAndRun, "https://github.com/vinceliuice/WhiteSur-gtk-theme.git", "tweaks.sh -g")
		set.AddStep("Installing WhiteSur Flatpak tweaks", run.CloneAndRun, "https://github.com/vinceliuice/WhiteSur-gtk-theme.git", "tweaks.sh -F")
		set.AddStep("Installing WhiteSur Dock tweaks", run.CloneAndRun, "https://github.com/vinceliuice/WhiteSur-gtk-theme.git", "tweaks.sh -d")
		set.AddStep("Installing WhiteSur Icons", run.CloneAndRun, "https://github.com/vinceliuice/WhiteSur-icon-theme.git", "install.sh -a -b")

		set.AddStep("Loading Gnome Settings", run.InstallGnomeSettings, "dconf.toml")

		set.AddStep("Creating pipewire directory", run.CreateDir, "/etc/pipewire")
		set.AddStep("Enabling HiFi audio", run.CopyFile, "pipewire.conf", "/etc/pipewire")

	}
	set.Run()
	fmt.Println("Setup complete. Enjoy!")
}
