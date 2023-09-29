package run

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func LoadCommands() {
	SetCommand("CreateDir", execCreateDir)
	SetCommand("AppendFile", execAppendFile)
	SetCommand("AddGroup", execAddGroup)
	SetCommand("AssignGroups", execAssignGroups)
	SetCommand("PrimaryGroup", execPrimaryGroup)
	SetCommand("Replace", execReplace)
	SetCommand("ChangeOwner", execChangeOwner)
	SetCommand("ChangePerm", execChangePerm)
	SetCommand("ReloadSysctl", execReloadSysctl)
	SetCommand("UpdateRepos", execUpdateRepos)
	SetCommand("UpgradePackages", execUpgradePackages)
	SetCommand("AddArch", execAddArch)
	SetCommand("ReloadUnits", execReloadUnits)
	SetCommand("InstallPackages", execInstallPackages)
	SetCommand("InstallFlatpaks", execInstallFlatpaks)
	SetCommand("EnableFlatpak", execEnableFlatpak)
	SetCommand("InstallPip", execInstallPip)
	SetCommand("EnableAptFile", execEnableAptFile)
	SetCommand("EnableService", execEnableService)
	SetCommand("UnzipFile", execUnzipFile)
	SetCommand("Untar", execUntar)
	SetCommand("AddUser", execAddUser)
	SetCommand("CloneRepo", execCloneRepo)
	SetCommand("CloneAndRun", execCloneAndRun)
	SetCommand("CloneAndRunAsUser", execCloneAndRunAsUser)
	SetCommand("InstallGnomeExt", execInstallGnomeExt)
	SetCommand("EnableGnomeExt", execEnableGnomeExt)
	SetCommand("InstallZshPlugin", execInstallZshPlugin)
	SetCommand("EnableZsh", execEnableZsh)
	SetCommand("InstallGnomeSettings", execInstallGnomeSettings)
	SetCommand("Run", execRun)
	SetCommand("Download", execDownload)
	SetCommand("AddRepoKey", execAddRepoKey)
	SetCommand("SetPass", execSetPass)
	SetCommand("CopyFile", execCopyFile)
	SetCommand("InstallUserConfig", execInstallUserConfig)
}

func execCreateDir(s *Set, param ...string) (string, error) {
	dirName := param[0]
	if _, err := os.Stat(dirName); !os.IsNotExist(err) {
		var ok error
		return "", ok
	}
	err := os.Mkdir(dirName, 0755)
	return "", err
}

func execAppendFile(s *Set, param ...string) (string, error) {
	cfg := param[0]
	dst := param[1]
	arg := []string{"-q", "flechade", dst}
	Cmd := exec.Command("grep", arg...)
	out, err := Cmd.CombinedOutput()
	if err == nil {
		return string(out), err
	}

	dstFile, err := os.OpenFile(dst, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()
	cfgFile, err := s.files.Open(cfg)
	if err != nil {
		return "", err
	}
	defer cfgFile.Close()
	fmt.Fprintln(dstFile, "# flechade START")
	_, err = io.Copy(dstFile, cfgFile)
	fmt.Fprintln(dstFile, "# flechade END")
	return "", err
}

func execAddGroup(s *Set, param ...string) (string, error) {
	groupName := param[0]
	argGroup := []string{groupName}
	groupCmd := exec.Command("groupadd", argGroup...)
	out, err := groupCmd.CombinedOutput()
	if err == nil {
		return string(out), err
	}
	if err.Error() == "exit status 9" {
		var e error
		return "", e
	}
	//fmt.Println(out)
	return string(out), err
}

func execAssignGroups(s *Set, param ...string) (string, error) {
	groups := param[0]
	args := []string{"-aG", groups, s.user}
	Cmd := exec.Command("usermod", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execPrimaryGroup(s *Set, param ...string) (string, error) {
	user := param[0]
	pg := param[1]
	args := []string{"-g", pg, user}
	Cmd := exec.Command("usermod", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execReplace(s *Set, param ...string) (string, error) {
	subReg := param[0]
	file := param[1]
	argSed := []string{"-Ei", "-e", subReg, file}
	sedCmd := exec.Command("sed", argSed...)
	out, err := sedCmd.CombinedOutput()
	return string(out), err
}

func execChangeOwner(s *Set, param ...string) (string, error) {
	owner := param[0]
	file := param[1]
	argSed := []string{"-R", owner, file}
	sedCmd := exec.Command("chown", argSed...)
	out, err := sedCmd.CombinedOutput()
	return string(out), err
}

func execChangePerm(s *Set, param ...string) (string, error) {
	mode := param[0]
	file := param[1]
	argChmod := []string{"-R", mode, file}
	chmodCmd := exec.Command("chmod", argChmod...)
	out, err := chmodCmd.CombinedOutput()
	return string(out), err
}

func execReloadSysctl(s *Set, param ...string) (string, error) {
	args := []string{"-p"}
	Cmd := exec.Command("sysctl", args...)
	out, err := Cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	args = []string{"-a"}
	Cmd = exec.Command("sysctl", args...)
	out, err = Cmd.CombinedOutput()
	return string(out), err
}

func execUpdateRepos(s *Set, param ...string) (string, error) {
	args := []string{"update", "-y", "-o", "Dpkg::Options::=--force-confdef"}
	Cmd := exec.Command("apt", args...)
	Cmd.Env = os.Environ()
	Cmd.Env = append(Cmd.Env, "DEBIAN_FRONTEND=noninteractive")
	Cmd.Env = append(Cmd.Env, "DEBCONF_NONINTERACTIVE_SEEN=true")
	Cmd.Env = append(Cmd.Env, "APT_LISTCHANGES_FRONTEND=none")
	Cmd.Env = append(Cmd.Env, "NEEDRESTART_MODE=a")
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execUpgradePackages(s *Set, param ...string) (string, error) {
	args := []string{"upgrade", "-y", "-o", "Dpkg::Options::=--force-confnew"}
	Cmd := exec.Command("apt", args...)
	Cmd.Env = os.Environ()
	Cmd.Env = append(Cmd.Env, "DEBCONF_NONINTERACTIVE_SEEN=true")
	Cmd.Env = append(Cmd.Env, "APT_LISTCHANGES_FRONTEND=none")
	Cmd.Env = append(Cmd.Env, "NEEDRESTART_MODE=a")
	Cmd.Env = append(Cmd.Env, "DEBIAN_FRONTEND=noninteractive")
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execAddArch(s *Set, param ...string) (string, error) {
	arch := param[0]
	arg := []string{"--add-architecture", arch}
	Cmd := exec.Command("dpkg", arg...)
	Cmd.Env = os.Environ()
	Cmd.Env = append(Cmd.Env, "DEBCONF_NONINTERACTIVE_SEEN=true")
	Cmd.Env = append(Cmd.Env, "DEBIAN_FRONTEND=noninteractive")
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execReloadUnits(s *Set, param ...string) (string, error) {
	args := []string{"daemon-reload"}
	Cmd := exec.Command("systemctl", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execInstallPackages(s *Set, param ...string) (string, error) {
	pkgs := param[0]
	arg := []string{"install", "-y", "-o", "Dpkg::Options::=--force-confnew"}
	plist := strings.Split(pkgs, " ")
	arg = append(arg, plist...)
	Cmd := exec.Command("apt", arg...)
	Cmd.Env = os.Environ()
	Cmd.Env = append(Cmd.Env, "DEBCONF_NONINTERACTIVE_SEEN=true")
	Cmd.Env = append(Cmd.Env, "APT_LISTCHANGES_FRONTEND=none")
	Cmd.Env = append(Cmd.Env, "NEEDRESTART_MODE=a")
	Cmd.Env = append(Cmd.Env, "DEBIAN_FRONTEND=noninteractive")
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execInstallFlatpaks(s *Set, param ...string) (string, error) {
	pkgs := param[0]
	arg := []string{"install", "--noninteractive", "--assumeyes", "-v"}
	plist := strings.Split(pkgs, " ")
	arg = append(arg, plist...)
	Cmd := exec.Command("flatpak", arg...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execInstallPip(s *Set, param ...string) (string, error) {
	pkgs := param[0]
	args := []string{"-m", "pip", "install", "--break-system-packages"}
	plist := strings.Split(pkgs, " ")
	args = append(args, plist...)
	Cmd := exec.Command("python3", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execEnableAptFile(s *Set, param ...string) (string, error) {
	out, err := execInstallPackages(s, "apt-file")
	if err != nil {
		return out, err
	}
	Cmd := exec.Command("apt-file", "update")
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func execEnableFlatpak(s *Set, param ...string) (string, error) {
	//Installing flatpak and plugin for software manager
	out, err := execInstallPackages(s, "flatpak gnome-software-plugin-flatpak")
	if err != nil {
		return out, err
	}
	//Adding flathub repo
	args := []string{"remote-add", "--if-not-exists", "flathub", "https://flathub.org/repo/flathub.flatpakrepo"}
	Cmd := exec.Command("flatpak", args...)
	output, err := Cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	//Pulling available packages
	args = []string{"update", "--noninteractive", "--assumeyes"}
	Cmd = exec.Command("flatpak", args...)
	output, err = Cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	//Prividing access to themes
	args = []string{"override", "--filesystem=~/.themes", "--filesystem=~/.icons", "--filesystem=xdg-config/gtk-4.0"}
	Cmd = exec.Command("flatpak", args...)
	output, err = Cmd.CombinedOutput()
	return string(output), err
}

func execEnableService(s *Set, param ...string) (string, error) {
	svc := param[0]
	arg := []string{"enable", svc}
	Cmd := exec.Command("systemctl", arg...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execUnzipFile(s *Set, param ...string) (string, error) {
	file := param[0]
	dir := param[1]
	arg := []string{"-n", file, "-d", dir}
	Cmd := exec.Command("unzip", arg...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execUntar(s *Set, param ...string) (string, error) {
	file := param[0]
	dir := param[1]
	arg := []string{"xf", file, "-C", dir, "--strip-components=1"}
	Cmd := exec.Command("tar", arg...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execAddUser(s *Set, param ...string) (string, error) {
	name := param[0]
	arg := []string{"-m", name}
	Cmd := exec.Command("useradd", arg...)
	out, err := Cmd.CombinedOutput()
	if err == nil {
		return string(out), err
	}
	if err.Error() == "exit status 9" {
		var e error
		return "", e
	}
	return string(out), err
}

func execCloneRepo(s *Set, param ...string) (string, error) {
	repo := param[0]
	dir := param[1]
	arg := []string{"clone", "--depth", "1", repo, dir}
	Cmd := exec.Command("git", arg...)
	Cmd.Env = os.Environ()
	Cmd.Env = append(Cmd.Env, "GIT_SSL_NO_VERIFY=true")
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execCloneAndRun(s *Set, param ...string) (string, error) {
	repo := param[0]
	command := param[1]
	rlen := len(repo)
	last := strings.LastIndex(repo, "/")
	rname := repo[last : rlen-4]
	clist := strings.Split(command, " ")
	xfile := clist[0]
	if _, err := os.Stat("/tmp/" + rname + "/" + xfile); errors.Is(err, os.ErrNotExist) {
		out, err := execCloneRepo(s, repo, "/tmp/"+rname)
		if err != nil {
			return out, err
		}
	}
	args := clist[1:]
	Cmd := exec.Command("/tmp/"+rname+"/"+xfile, args...)
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func execCloneAndRunAsUser(s *Set, param ...string) (string, error) {
	repo := param[0]
	command := param[1]
	rlen := len(repo)
	last := strings.LastIndex(repo, "/")
	rname := repo[last:rlen-4] + ".usr"
	clist := strings.Split(command, " ")
	xfile := clist[0]
	if _, err := os.Stat("/tmp/" + rname + "/" + xfile); errors.Is(err, os.ErrNotExist) {
		out, err := execCloneRepo(s, repo, "/tmp/"+rname)
		if err != nil {
			return out, err
		}
	}
	chownArgs := []string{"-R", s.user, "/tmp/" + rname}
	chownCmd := exec.Command("chown", chownArgs...)
	chownout, err := chownCmd.CombinedOutput()
	if err != nil {
		return string(chownout), err
	}
	args := clist[1:]
	concParms := strings.Join(args, " ")
	concCmd := "/tmp/" + rname + "/" + xfile + " " + concParms
	flags := append([]string{s.user, "-c"}, concCmd)
	Cmd := exec.Command("su", flags...)
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func execInstallGnomeExt(s *Set, param ...string) (string, error) {
	extid := param[0]
	version := param[1]
	//last := strings.LastIndex(url, "/")
	//file := url[last:]

	file := strings.ReplaceAll(extid, "@", "")
	url := "https://extensions.gnome.org/extension-data/" + file + ".v" + version + ".shell-extension.zip"

	out, err := execDownload(s, url, "/tmp/"+file)
	if err != nil {
		return out, err
	}
	Cmd := exec.Command("gnome-extensions", "install", "--force", "/tmp/"+file)
	output, err := Cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	//Activating the extension in session
	Cmd = exec.Command("su", "-", s.user, "-c",
		"DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/"+s.uid+"/bus busctl --user call org.gnome.Shell.Extensions /org/gnome/Shell/Extensions org.gnome.Shell.Extensions InstallRemoteExtension s "+extid)
	time.Sleep(2 * time.Second)
	output, err = Cmd.CombinedOutput()
	time.Sleep(2 * time.Second)
	if err != nil {
		if err.Error() == "exit status 1" {
			//ignore disconnect
			var e error
			return "", e
		}
	}
	return string(output), err
}

func execEnableGnomeExt(s *Set, param ...string) (string, error) {
	ext := param[0]

	Cmd := exec.Command("su", s.user, "-c", "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/"+s.uid+"/bus gnome-extensions enable "+ext)
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func execInstallZshPlugin(s *Set, param ...string) (string, error) {
	repo := param[0]

	rlen := len(repo)
	last := strings.LastIndex(repo, "/")
	rname := repo[last : rlen-4]
	Cmd := exec.Command("su", s.user, "-c", "git clone --depth=1 "+repo+" ~/.oh-my-zsh/custom/plugins/"+rname)
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func execEnableZsh(s *Set, param ...string) (string, error) {
	Cmd := exec.Command("usermod", "-s", "/bin/zsh", s.user)
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func execInstallGnomeSettings(s *Set, param ...string) (string, error) {
	cfg := param[0]

	cfgFile, err := s.files.Open(cfg)
	if err != nil {
		return "", err
	}
	Cmd := exec.Command("su", s.user, "-c", "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/"+s.uid+"/bus dconf load /")
	buf, _ := io.ReadAll(cfgFile)
	Cmd.Stdin = strings.NewReader(string(buf))
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func execRun(s *Set, param ...string) (string, error) {
	cmd := param[0]

	args := []string{}
	Cmd := exec.Command(cmd, args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execDownload(s *Set, param ...string) (string, error) {
	url := param[0]
	file := param[1]

	args := []string{"--continue", url, "-O", file}
	Cmd := exec.Command("wget", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execAddRepoKey(s *Set, param ...string) (string, error) {
	URL := param[0]
	file := param[1]

	args := []string{"--continue", URL, "-O", file}
	Cmd := exec.Command("wget", args...)
	out, err := Cmd.CombinedOutput()

	if err != nil {
		return string(out), err
	}
	args = []string{"--batch", "--yes", "--dearmor", file}
	Cmd = exec.Command("gpg", args...)
	out, err = Cmd.CombinedOutput()

	return string(out), err
}

func execSetPass(s *Set, param ...string) (string, error) {
	user := param[0]
	pass := param[1]

	args := []string{}
	Cmd := exec.Command("chpasswd", args...)
	Cmd.Stdin = strings.NewReader(user + ":" + pass)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func execCopyFile(s *Set, param ...string) (string, error) {
	fileName := param[0]
	dstDir := param[1]

	dstFile, err := os.OpenFile(dstDir+"/"+fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()
	cfgFile, err := s.files.Open(fileName)
	if err != nil {
		return "", err
	}
	defer cfgFile.Close()
	_, err = io.Copy(dstFile, cfgFile)
	return "", err
}

func execInstallUserConfig(s *Set, param ...string) (string, error) {
	fileName := param[0]
	relDir := param[1]

	dstDir := "/home/" + s.user + "/" + relDir
	dstName := dstDir + "/" + fileName

	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return "", err
	}

	if relDir != "" {
		parts := strings.Split(relDir, "/")
		out, err := execChangeOwner(s, s.user, "/home/"+s.user+"/"+parts[0])
		if err != nil {
			return out, err
		}
	}

	dstFile, err := os.OpenFile(dstName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()
	cfgFile, err := s.files.Open(fileName)
	if err != nil {
		return "", err
	}
	defer cfgFile.Close()
	_, err = io.Copy(dstFile, cfgFile)
	if err != nil {
		return "", err
	}
	out, err := execChangeOwner(s, s.user, dstName)
	return out, err
}
