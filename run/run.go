package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/theckman/yacspin"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

var Commands []string
var version string

func init() {
	version = "0.0.4"
	Commands = []string{
		//Basic file ops
		"CreateDir",
		"AppendFile",
		"CopyFile",
		"ChangeOwner",
		"ChangePerm",
		"Replace",
		"Download",

		// Basic user ops
		"AddUser",
		"AddGroup",
		"SetPass",
		"AssignGroups",
		"PrimaryGroup",
		"InstallUserConfig",

		// Compression
		"Untar",
		"UnzipFile",

		// Packages
		"AddRepoKey",
		"UpdateRepos",
		"InstallPackages",
		"InstallFlatpaks",
		"InstallPip",
		"UpgradePackages",
		"AddArch",
		"EnableFlatpak",
		"EnableAptFile",

		// Zsh
		"InstallZshPlugin",
		"EnableZsh",

		//Gnome
		"InstallGnomeExt",
		"EnableGnomeExt",
		"InstallGnomeSettings",

		//Services
		"ReloadUnits",
		"ReloadSysctl",
		"EnableService",

		// GIT
		"CloneRepo",
		"CloneAndRun",
		"CloneAndRunAsUser",

		// Ops
		"Run",
	}

}

type stepStat struct {
	ErrLvl  int
	Message string
}

type step struct {
	//Id       int
	Command  string
	Params   []string `yaml:"params,omitempty"`
	Desc     string
	Status   stepStat `yaml:"status,omitempty"`
	Complete bool     `yaml:"complete,omitempty"`
}

type Set struct {
	Ver         string
	osRel       string
	configFile  string
	files       fs.FS
	Name        string
	Description string
	user        string
	uid         string
	Steps       []step
}

func GetVer() string {
	return version
}

func NewSet(name, description string) *Set {
	var s Set
	home, _ := os.UserHomeDir()
	s.configFile = home + "/.flechade"
	s.Ver = GetVer()
	s.Name = name
	s.Description = description
	args := []string{"-s", "-d"}
	cmd := exec.Command("lsb_release", args...)
	out, _ := cmd.CombinedOutput()
	s.osRel = string(out)
	// Get non root username
	s.user = os.Getenv("USER")
	sudoUser, ok := os.LookupEnv("SUDO_USER")
	if ok {
		s.user = sudoUser
	}
	s.uid = os.Getenv("UID")
	sudoUid, ok := os.LookupEnv("SUDO_UID")
	if ok {
		s.uid = sudoUid
	}
	return &s
}

func (ds *Set) GetOS() string {
	return ds.osRel
}

func LoadSetFromDir(dir fs.FS) (*Set, error) {
	var s Set
	s.files = dir
	//data, err := fs.ReadDir(dir, "flachade.yaml")
	yfile, err := dir.Open("flechade.yaml")
	if err != nil {
		return &s, err
	}
	data, err := io.ReadAll(yfile)
	if err != nil {
		return &s, err
	}
	//fmt.Println(string(data))
	err = yaml.Unmarshal(data, &s)
	if err != nil {
		return &s, err
	}
	home, _ := os.UserHomeDir()
	s.configFile = home + "/.flechade"
	if !s.checkVersion() {
		err := errors.New("file version not compatible")
		return &s, err
	}
	args := []string{"-s", "-d"}
	cmd := exec.Command("lsb_release", args...)
	out, _ := cmd.CombinedOutput()
	s.osRel = string(out)
	// Get non root username
	s.user = os.Getenv("USER")
	sudoUser, ok := os.LookupEnv("SUDO_USER")
	if ok {
		s.user = sudoUser
	}
	s.uid = os.Getenv("UID")
	sudoUid, ok := os.LookupEnv("SUDO_UID")
	if ok {
		s.uid = sudoUid
	}
	return &s, err
}

func (ds *Set) checkVersion() bool {
	return true
}

func (ds *Set) AddStep(desc string, cmdId string, args ...string) error {
	var err error
	var stp step
	if !slices.Contains(Commands, cmdId) {
		return errors.New("command not found")
	}
	stp.Params = args
	stp.Command = cmdId
	stp.Desc = desc
	ds.Steps = append(ds.Steps, stp)
	err = ds.saveStats()
	return err
}

func (s *Set) Load() error {
	file, err := os.Open(s.configFile)
	if err != nil {
		//log.Println("Config file does not exist.")
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(s)
	return err
}

func (s *Set) saveStats() error {
	file, err := os.Create(s.configFile)
	if err != nil {
		log.Fatal("Unable to save config file:", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(s)
	return err
}

func (ds *Set) execCreateDir(dirName string) (string, error) {
	if _, err := os.Stat(dirName); !os.IsNotExist(err) {
		var ok error
		return "", ok
	}
	err := os.Mkdir(dirName, 0755)
	return "", err
}

func (ds *Set) execAppendFile(cfg string, dst string) (string, error) {

	args := []string{"-q", "flechade", dst}
	Cmd := exec.Command("grep", args...)
	out, err := Cmd.CombinedOutput()
	if err == nil {
		return string(out), err
	}

	dstFile, err := os.OpenFile(dst, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()
	cfgFile, err := ds.files.Open(cfg)
	if err != nil {
		return "", err
	}
	defer cfgFile.Close()
	fmt.Fprintln(dstFile, "# flechade START")
	_, err = io.Copy(dstFile, cfgFile)
	fmt.Fprintln(dstFile, "# flechade END")
	return "", err
}

func (ds *Set) execAddGroup(groupName string) (string, error) {
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

func (ds *Set) execAssignGroups(groups string) (string, error) {
	args := []string{"-aG", groups, ds.user}
	Cmd := exec.Command("usermod", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execPrimaryGroup(user string, pg string) (string, error) {
	args := []string{"-g", pg, user}
	Cmd := exec.Command("usermod", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execReplace(subReg string, file string) (string, error) {
	argSed := []string{"-Ei", "-e", subReg, file}
	sedCmd := exec.Command("sed", argSed...)
	out, err := sedCmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execChangeOwner(owner string, file string) (string, error) {
	argSed := []string{"-R", owner, file}
	sedCmd := exec.Command("chown", argSed...)
	out, err := sedCmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execChangePerm(mode string, file string) (string, error) {
	argChmod := []string{"-R", mode, file}
	chmodCmd := exec.Command("chmod", argChmod...)
	out, err := chmodCmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execReloadSysctl() (string, error) {
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

func (ds *Set) execUpdateRepos() (string, error) {
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

func (ds *Set) execUpgradePackages() (string, error) {
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

func (ds *Set) execAddArch(arch string) (string, error) {
	args := []string{"--add-architecture", arch}
	Cmd := exec.Command("dpkg", args...)
	Cmd.Env = os.Environ()
	Cmd.Env = append(Cmd.Env, "DEBCONF_NONINTERACTIVE_SEEN=true")
	Cmd.Env = append(Cmd.Env, "DEBIAN_FRONTEND=noninteractive")
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execReloadUnits() (string, error) {
	args := []string{"daemon-reload"}
	Cmd := exec.Command("systemctl", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execInstallPackages(pkgs string) (string, error) {
	args := []string{"install", "-y", "-o", "Dpkg::Options::=--force-confnew"}
	plist := strings.Split(pkgs, " ")
	args = append(args, plist...)
	Cmd := exec.Command("apt", args...)
	Cmd.Env = os.Environ()
	Cmd.Env = append(Cmd.Env, "DEBCONF_NONINTERACTIVE_SEEN=true")
	Cmd.Env = append(Cmd.Env, "APT_LISTCHANGES_FRONTEND=none")
	Cmd.Env = append(Cmd.Env, "NEEDRESTART_MODE=a")
	Cmd.Env = append(Cmd.Env, "DEBIAN_FRONTEND=noninteractive")
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execInstallFlatpaks(pkgs string) (string, error) {
	args := []string{"install", "--noninteractive", "--assumeyes", "-v"}
	plist := strings.Split(pkgs, " ")
	args = append(args, plist...)
	Cmd := exec.Command("flatpak", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execInstallPip(pkgs string) (string, error) {
	args := []string{"-m", "pip", "install", "--break-system-packages"}
	plist := strings.Split(pkgs, " ")
	args = append(args, plist...)
	Cmd := exec.Command("python3", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execEnableAptFile() (string, error) {
	out, err := ds.execInstallPackages("apt-file")
	if err != nil {
		return out, err
	}
	Cmd := exec.Command("apt-file", "update")
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func (ds *Set) execEnableFlatpak() (string, error) {
	//Installing flatpak and plugin for software manager
	out, err := ds.execInstallPackages("flatpak gnome-software-plugin-flatpak")
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

func (ds *Set) execEnableService(svc string) (string, error) {
	args := []string{"enable", svc}
	Cmd := exec.Command("systemctl", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execUnzipFile(file string, dir string) (string, error) {
	args := []string{"-n", file, "-d", dir}
	Cmd := exec.Command("unzip", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execUntar(file string, dir string) (string, error) {
	args := []string{"xf", file, "-C", dir, "--strip-components=1"}
	Cmd := exec.Command("tar", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execAddUser(name string) (string, error) {
	args := []string{"-m", name}
	Cmd := exec.Command("useradd", args...)
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

func (ds *Set) execCloneRepo(repo string, dir string) (string, error) {
	args := []string{"clone", "--depth", "1", repo, dir}
	Cmd := exec.Command("git", args...)
	Cmd.Env = os.Environ()
	Cmd.Env = append(Cmd.Env, "GIT_SSL_NO_VERIFY=true")
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execCloneAndRun(repo, command string) (string, error) {
	rlen := len(repo)
	last := strings.LastIndex(repo, "/")
	rname := repo[last : rlen-4]
	clist := strings.Split(command, " ")
	xfile := clist[0]
	if _, err := os.Stat("/tmp/" + rname + "/" + xfile); errors.Is(err, os.ErrNotExist) {
		out, err := ds.execCloneRepo(repo, "/tmp/"+rname)
		if err != nil {
			return out, err
		}
	}
	args := clist[1:]
	Cmd := exec.Command("/tmp/"+rname+"/"+xfile, args...)
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func (ds *Set) execCloneAndRunAsUser(repo, command string) (string, error) {
	rlen := len(repo)
	last := strings.LastIndex(repo, "/")
	rname := repo[last:rlen-4] + ".usr"
	clist := strings.Split(command, " ")
	xfile := clist[0]
	if _, err := os.Stat("/tmp/" + rname + "/" + xfile); errors.Is(err, os.ErrNotExist) {
		out, err := ds.execCloneRepo(repo, "/tmp/"+rname)
		if err != nil {
			return out, err
		}
	}
	chownArgs := []string{"-R", ds.user, "/tmp/" + rname}
	chownCmd := exec.Command("chown", chownArgs...)
	chownout, err := chownCmd.CombinedOutput()
	if err != nil {
		return string(chownout), err
	}
	args := clist[1:]
	concParms := strings.Join(args, " ")
	concCmd := "/tmp/" + rname + "/" + xfile + " " + concParms
	flags := append([]string{ds.user, "-c"}, concCmd)
	Cmd := exec.Command("su", flags...)
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func (ds *Set) execInstallGnomeExt(extid, version string) (string, error) {
	//last := strings.LastIndex(url, "/")
	//file := url[last:]

	file := strings.ReplaceAll(extid, "@", "")
	url := "https://extensions.gnome.org/extension-data/" + file + ".v" + version + ".shell-extension.zip"

	out, err := ds.execDownload(url, "/tmp/"+file)
	if err != nil {
		return out, err
	}
	Cmd := exec.Command("gnome-extensions", "install", "--force", "/tmp/"+file)
	output, err := Cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	//Activating the extension in session
	Cmd = exec.Command("su", "-", ds.user, "-c",
		"DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/"+ds.uid+"/bus busctl --user call org.gnome.Shell.Extensions /org/gnome/Shell/Extensions org.gnome.Shell.Extensions InstallRemoteExtension s "+extid)
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

func (ds *Set) execEnableGnomeExt(ext string) (string, error) {
	Cmd := exec.Command("su", ds.user, "-c", "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/"+ds.uid+"/bus gnome-extensions enable "+ext)
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func (ds *Set) execInstallZshPlugin(repo string) (string, error) {
	rlen := len(repo)
	last := strings.LastIndex(repo, "/")
	rname := repo[last : rlen-4]
	Cmd := exec.Command("su", ds.user, "-c", "git clone --depth=1 "+repo+" ~/.oh-my-zsh/custom/plugins/"+rname)
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func (ds *Set) execEnableZsh() (string, error) {
	Cmd := exec.Command("usermod", "-s", "/bin/zsh", ds.user)
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func (ds *Set) execInstallGnomeSettings(cfg string) (string, error) {
	cfgFile, err := ds.files.Open(cfg)
	if err != nil {
		return "", err
	}
	Cmd := exec.Command("su", ds.user, "-c", "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/"+ds.uid+"/bus dconf load /")
	buf, _ := io.ReadAll(cfgFile)
	Cmd.Stdin = strings.NewReader(string(buf))
	output, err := Cmd.CombinedOutput()
	return string(output), err
}

func (ds *Set) execRun(cmd string) (string, error) {
	args := []string{}
	Cmd := exec.Command(cmd, args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execDownload(url string, file string) (string, error) {
	args := []string{"--continue", url, "-O", file}
	Cmd := exec.Command("wget", args...)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execAddRepoKey(URL string, file string) (string, error) {
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

func (ds *Set) execSetPass(user string, pass string) (string, error) {
	args := []string{}
	Cmd := exec.Command("chpasswd", args...)
	Cmd.Stdin = strings.NewReader(user + ":" + pass)
	out, err := Cmd.CombinedOutput()
	return string(out), err
}

func (ds *Set) execCopyFile(fileName string, dstDir string) (string, error) {
	dstFile, err := os.OpenFile(dstDir+"/"+fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()
	cfgFile, err := ds.files.Open(fileName)
	if err != nil {
		return "", err
	}
	defer cfgFile.Close()
	_, err = io.Copy(dstFile, cfgFile)
	return "", err
}

func (ds *Set) execInstallUserConfig(fileName, relDir string) (string, error) {
	dstDir := "/home/" + ds.user + "/" + relDir
	dstName := dstDir + "/" + fileName

	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return "", err
	}

	if relDir != "" {
		parts := strings.Split(relDir, "/")
		out, err := ds.execChangeOwner(ds.user, "/home/"+ds.user+"/"+parts[0])
		if err != nil {
			return out, err
		}
	}

	dstFile, err := os.OpenFile(dstName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()
	cfgFile, err := ds.files.Open(fileName)
	if err != nil {
		return "", err
	}
	defer cfgFile.Close()
	_, err = io.Copy(dstFile, cfgFile)
	if err != nil {
		return "", err
	}
	out, err := ds.execChangeOwner(ds.user, dstName)
	return out, err
}

func (ds *Set) Run() {

	if os.Geteuid() != 0 {
		log.Fatal("This tool needs root access. Please use sudo.")
	}

	fmt.Print("Setting up the environment: " + ds.Name)

	var err error
	var spinner *yacspin.Spinner
	var cfg yacspin.Config
	var out string

	for i, step := range ds.Steps {

		if step.Complete {
			continue
		}
		m := fmt.Sprintf("%-40s", step.Desc)[:40]
		cfg = yacspin.Config{
			Frequency:         100 * time.Millisecond,
			CharSet:           yacspin.CharSets[78],
			Suffix:            " ",
			Prefix:            " ",
			Colors:            []string{"fgYellow"},
			StopMessage:       m + "	[OK]",
			StopFailMessage:   step.Desc + "	[Failed]",
			SuffixAutoColon:   true,
			Message:           m,
			StopCharacter:     "✓",
			StopColors:        []string{"fgGreen"},
			StopFailCharacter: "✗",
			StopFailColors:    []string{"fgRed"},
		}
		spinner, _ = yacspin.New(cfg)
		_ = spinner.Start()

		switch step.Command {
		case "CreateDir":
			out, err = ds.execCreateDir(step.Params[0])
		case "AppendFile":
			out, err = ds.execAppendFile(step.Params[0], step.Params[1])
		case "AddGroup":
			out, err = ds.execAddGroup(step.Params[0])
		case "Replace":
			out, err = ds.execReplace(step.Params[0], step.Params[1])
		case "CopyFile":
			out, err = ds.execCopyFile(step.Params[0], step.Params[1])
		case "ChangeOwner":
			out, err = ds.execChangeOwner(step.Params[0], step.Params[1])
		case "ChangePerm":
			out, err = ds.execChangePerm(step.Params[0], step.Params[1])
		case "AssignGroups":
			out, err = ds.execAssignGroups(step.Params[0])
		case "ReloadSysctl":
			out, err = ds.execReloadSysctl()
		case "UpdateRepos":
			out, err = ds.execUpdateRepos()
		case "UpgradePackages":
			out, err = ds.execUpgradePackages()
		case "InstallPackages":
			out, err = ds.execInstallPackages(step.Params[0])
		case "EnableService":
			out, err = ds.execEnableService(step.Params[0])
		case "UnzipFile":
			out, err = ds.execUnzipFile(step.Params[0], step.Params[1])
		case "Run":
			out, err = ds.execRun(step.Params[0])
		case "Untar":
			out, err = ds.execUntar(step.Params[0], step.Params[1])
		case "CloneRepo":
			out, err = ds.execCloneRepo(step.Params[0], step.Params[1])
		case "AddUser":
			out, err = ds.execAddUser(step.Params[0])
		case "SetPass":
			out, err = ds.execSetPass(step.Params[0], step.Params[1])
		case "Download":
			out, err = ds.execDownload(step.Params[0], step.Params[1])
		case "PrimaryGroup":
			out, err = ds.execPrimaryGroup(step.Params[0], step.Params[1])
		case "ReloadUnits":
			out, err = ds.execReloadUnits()
		case "AddRepoKey":
			out, err = ds.execAddRepoKey(step.Params[0], step.Params[1])
		case "AddArch":
			out, err = ds.execAddArch(step.Params[0])
		case "EnableFlatpak":
			out, err = ds.execEnableFlatpak()
		case "EnableAptFile":
			out, err = ds.execEnableAptFile()
		case "InstallFlatpaks":
			out, err = ds.execInstallFlatpaks(step.Params[0])
		case "InstallPip":
			out, err = ds.execInstallPip(step.Params[0])
		case "CloneAndRun":
			out, err = ds.execCloneAndRun(step.Params[0], step.Params[1])
		case "CloneAndRunAsUser":
			out, err = ds.execCloneAndRunAsUser(step.Params[0], step.Params[1])
		case "InstallGnomeExt":
			out, err = ds.execInstallGnomeExt(step.Params[0], step.Params[1])
		case "EnableGnomeExt":
			out, err = ds.execEnableGnomeExt(step.Params[0])
		case "InstallGnomeSettings":
			out, err = ds.execInstallGnomeSettings(step.Params[0])
		case "InstallZshPlugin":
			out, err = ds.execInstallZshPlugin(step.Params[0])
		case "EnableZsh":
			out, err = ds.execEnableZsh()
		case "InstallUserConfig":
			out, err = ds.execInstallUserConfig(step.Params[0], step.Params[1])
		}
		if err != nil {
			step.Status.ErrLvl = 1
			step.Status.Message = err.Error()
			ds.Steps[i] = step
			_ = ds.saveStats()
			spinner.StopFailMessage(step.Desc + ": " + err.Error())
			_ = spinner.StopFail()
			log.Fatal(out)
		} else {
			step.Complete = true
			ds.Steps[i] = step
			_ = ds.saveStats()
			_ = spinner.Stop()
		}
	}
}
