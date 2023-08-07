package main

import (
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/fleshin/flechade/run"
)

//go:embed data/*
var configFS embed.FS

func main() {

	ver := "0.0.1"

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
		set.AddStep("Installing Debian Testing sources", "CopyFile", "sources.list", "/etc/apt")
		set.AddStep("Google Chrome repo sources", "CopyFile", "google-chrome.list", "/etc/apt/sources.list.d/")
		set.AddStep("Adding Google public keys", "AddRepoKey", "https://dl.google.com/linux/linux_signing_key.pub", "/etc/apt/keyrings/linux_signing_key.pub")
		set.AddStep("MS VSCode repo sources", "CopyFile", "vscode.list", "/etc/apt/sources.list.d/")
		set.AddStep("Adding MS public keys", "AddRepoKey", "https://packages.microsoft.com/keys/microsoft.asc", "/etc/apt/keyrings/packages.microsoft.asc")
		set.AddStep("Enabling 32bit packages", "AddArch", "i386")
		set.AddStep("Updating package repositories", "UpdateRepos")
		set.AddStep("Upgrading packages", "UpgradePackages")

		set.AddStep("Enabling Flatpaks", "EnableFlatpak")

		set.AddStep("Installing system tools", "InstallPackages", "zsh nala lsd fonts-font-awesome neofetch mc tmux curl plocate libvirt-clients virt-manager sassc dbus-x11")
		set.AddStep("Making user member of virt groups", "AssignGroups", "kvm,libvirt")
		set.AddStep("Installing basic development env", "InstallPackages", "git build-essential golang libgl1-mesa-dev xorg-dev libglib2.0-dev-bin")
		set.AddStep("Enabling apt-file", "EnableAptFile")

		set.AddStep("Installing Flatpak apps", "InstallFlatpaks", "com.github.tchx84.Flatseal com.usebottles.bottles com.github.wwmm.easyeffects net.davidotek.pupgui2 com.slack.Slack org.gnome.Geary")

		set.AddStep("Installing Google Chrome", "InstallPackages", "google-chrome-stable")
		set.AddStep("Installing VS Code", "InstallPackages", "code")
		set.AddStep("Installing Steam", "InstallPackages", "mesa-vulkan-drivers libglx-mesa0:i386 mesa-vulkan-drivers:i386 libgl1-mesa-dri:i386 steam-installer")

		set.AddStep("Installing Nerd fonts", "CloneAndRun", "https://github.com/ryanoasis/nerd-fonts.git", "install.sh --install-to-system-path")

		set.AddStep("Installing Gnome Extension Dash to Dock", "InstallGnomeExt", "dash-to-dock@micxgx.gmail.com", "84")
		set.AddStep("Installing Gnome Extension OpenWheater", "InstallGnomeExt", "openweather-extension@jenslody.de", "121")
		set.AddStep("Installing Gnome Extension Tray Icons", "InstallGnomeExt", "trayIconsReloaded@selfmade.pl", "26")
		set.AddStep("Installing Gnome Extension Blur My Shell", "InstallGnomeExt", "blur-my-shell@aunetx", "47")

		set.AddStep("Enabling Gnome Extension User Themes", "EnableGnomeExt", "user-theme@gnome-shell-extensions.gcampax.github.com")

		set.AddStep("Installing WhiteSur Gnome theme", "CloneAndRunAsUser", "https://github.com/vinceliuice/WhiteSur-gtk-theme.git", "install.sh -l -c Light")
		set.AddStep("Installing WhiteSur Nautilus theme", "CloneAndRun", "https://github.com/vinceliuice/WhiteSur-gtk-theme.git", "install.sh -N mojave")
		set.AddStep("Installing WhiteSur GDM tweaks", "CloneAndRun", "https://github.com/vinceliuice/WhiteSur-gtk-theme.git", "tweaks.sh -g")
		set.AddStep("Installing WhiteSur Flatpak tweaks", "CloneAndRun", "https://github.com/vinceliuice/WhiteSur-gtk-theme.git", "tweaks.sh -F")
		set.AddStep("Installing WhiteSur Icons", "CloneAndRun", "https://github.com/vinceliuice/WhiteSur-icon-theme.git", "install.sh -a -b")
		set.AddStep("Installing Grub Theme", "CloneAndRun", "https://github.com/vinceliuice/grub2-themes.git", "install.sh -t whitesur")

		set.AddStep("Loading Gnome Settings", "InstallGnomeSettings", "dconf.toml")

		set.AddStep("Creating pipewire directory", "CreateDir", "/etc/pipewire")
		set.AddStep("Enabling HiFi audio", "CopyFile", "pipewire.conf", "/etc/pipewire")

		set.AddStep("Installing Oh My Zsh", "CloneAndRunAsUser", "https://github.com/ohmyzsh/ohmyzsh.git", "tools/install.sh --unattended")
		set.AddStep("Enabling Zsh", "EnableZsh")
		set.AddStep("Installing Zsh Highlighting plugin", "InstallZshPlugin", "https://github.com/zsh-users/zsh-syntax-highlighting.git")
		set.AddStep("Installing Zsh Autosuggestions plugin", "InstallZshPlugin", "https://github.com/zsh-users/zsh-autosuggestions.git")
		set.AddStep("Installing Zsh settings", "InstallUserConfig", ".zshrc")

		set.AddStep("Installing WhiteSur Dock tweaks", "CloneAndRun", "https://github.com/vinceliuice/WhiteSur-gtk-theme.git", "tweaks.sh -d")

	}
	set.Run()
	fmt.Println("Setup complete. Enjoy!")
}
