package run

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"github.com/theckman/yacspin"
)

type cmd int

const (
	//Basic file ops
	CreateDir cmd = iota
	AppendFile
	CopyFile
	ChangeOwner
	ChangePerm
	Replace
	Download

	// Basic user ops
	AddUser
	AddGroup
	SetPass
	AssignGroups
	PrimaryGroup

	// Compression
	Untar
	UnzipFile

	// Packages
	AddRepoKey
	UpdateRepos
	InstallPackage
	UpgradePackages

	//Services
	ReloadUnits
	ReloadSysctl
	EnableService

	// GIT
	CloneRepo

	// Ops
	Run
)

type stepStat struct {
	ErrLvl  int
	Message string
}

type step struct {
	//Id       int
	Command  cmd
	Params   []string
	Desc     string
	Status   stepStat
	Complete bool
}

type Set struct {
	Version    string
	OSrel      string
	configFile string
	files      embed.FS
	Steps      []step
}

func NewSet() *Set {
	var s Set
	bi, _ := debug.ReadBuildInfo()
	home, _ := os.UserHomeDir()
	s.configFile = home + "/.flechade"
	s.Version = bi.Settings[11].Value
	args := []string{"-s", "-d"}
	cmd := exec.Command("lsb_release", args...)
	out, _ := cmd.Output()
	s.OSrel = string(out)
	return &s
}

func (ds *Set) SetFiles(fs embed.FS) {
	ds.files = fs
}

func (ds *Set) AddStep(desc string, cmdId cmd, args ...string) error {
	var err error
	var stp step
	if cmdId < CreateDir || cmdId > Run {
		return errors.New("command not found")
	}
	stp.Params = args
	stp.Command = cmdId
	stp.Desc = desc
	ds.Steps = append(ds.Steps, stp)
	ds.saveStats()
	return err
}

func (s *Set) Load() error {
	//var s Set
	//home, _ := os.UserHomeDir()
	file, err := os.Open(s.configFile)
	defer file.Close()
	if err != nil {
		//log.Println("Config file does not exist.")
		return err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(s)
	return err
}

func (s *Set) saveStats() error {
	//home, _ := os.UserHomeDir()
	file, err := os.Create(s.configFile)
	defer file.Close()
	if err != nil {
		log.Fatal("Unable to save config file:", err)
	}
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
	out, err := Cmd.Output()
	if err == nil {
		return string(out), err
	}

	dstFile, err := os.OpenFile(dst, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()
	cfgFile, err := ds.files.Open("data/" + cfg)
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
	out, err := groupCmd.Output()
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

func (ds *Set) execAssignGroups(user string, groups string) (string, error) {
	args := []string{"-aG", groups, user}
	Cmd := exec.Command("usermod", args...)
	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execPrimaryGroup(user string, pg string) (string, error) {
	args := []string{"-g", pg, user}
	Cmd := exec.Command("usermod", args...)
	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execReplace(subReg string, file string) (string, error) {
	argSed := []string{"-Ei", "-e", subReg, file}
	sedCmd := exec.Command("sed", argSed...)
	out, err := sedCmd.Output()
	return string(out), err
}

func (ds *Set) execChangeOwner(owner string, file string) (string, error) {
	argSed := []string{"-R", owner, file}
	sedCmd := exec.Command("chown", argSed...)
	out, err := sedCmd.Output()
	return string(out), err
}

func (ds *Set) execChangePerm(mode string, file string) (string, error) {
	argChmod := []string{"-R", mode, file}
	chmodCmd := exec.Command("chmod", argChmod...)
	out, err := chmodCmd.Output()
	return string(out), err
}

func (ds *Set) execReloadSysctl() (string, error) {
	args := []string{"-p"}
	Cmd := exec.Command("sysctl", args...)
	out, err := Cmd.Output()
	args = []string{"-a"}
	Cmd = exec.Command("sysctl", args...)
	out, err = Cmd.Output()
	return string(out), err
}

func (ds *Set) execUpdateRepos() (string, error) {
	args := []string{"update", "-y", "-o", "Dpkg::Options::=\"--force-confdef\""}
	Cmd := exec.Command("apt", args...)
	Cmd.Env = os.Environ()
	Cmd.Env = append(Cmd.Env, "DEBIAN_FRONTEND=noninteractive")
	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execUpgradePackages() (string, error) {
	args := []string{"upgrade", "-y", "-fuy", "--force-yes", "-o", "Dpkg::Options::='--force-confnew'", "-o", "Dpkg::Options::='--force-confdef'", "-o", "quiet=2"}
	//args := []string{"upgrade", "-y", "Dpkg::Options::='--force-confnew'"}
	Cmd := exec.Command("apt", args...)
	Cmd.Env = os.Environ()
	Cmd.Env = append(Cmd.Env, "DEBIAN_FRONTEND=noninteractive")
	//Cmd.Env = append(Cmd.Env, "DEBCONF_NONINTERACTIVE_SEEN=true")
	Cmd.Env = append(Cmd.Env, "APT_LISTCHANGES_FRONTEND=none")
	Cmd.Env = append(Cmd.Env, "NEEDRESTART_MODE=a")

	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execReloadUnits() (string, error) {
	args := []string{"daemon-reload"}
	Cmd := exec.Command("systemctl", args...)
	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execInstallPackage(pkg string) (string, error) {
	args := []string{"install", pkg, "-y"}
	Cmd := exec.Command("apt", args...)
	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execEnableService(svc string) (string, error) {
	args := []string{"enable", svc}
	Cmd := exec.Command("systemctl", args...)
	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execUnzipFile(file string, dir string) (string, error) {
	args := []string{"-n", file, "-d", dir}
	Cmd := exec.Command("unzip", args...)
	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execUntar(file string, dir string) (string, error) {
	args := []string{"xf", file, "-C", dir, "--strip-components=1"}
	Cmd := exec.Command("tar", args...)
	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execAddUser(name string) (string, error) {
	args := []string{"-m", name}
	Cmd := exec.Command("useradd", args...)
	out, err := Cmd.Output()
	if err == nil {
		return string(out), err
	}
	if err.Error() == "exit status 9" {
		var e error
		return "", e
	}
	return string(out), err
}

func (ds *Set) execCloneRepo(user string, repo string, dir string) (string, error) {
	//os.Setenv("GIT_SSL_NO_VERIFY", "true")
	args := []string{"-u", user, "git", "clone", repo, dir}
	Cmd := exec.Command("sudo", args...)
	Cmd.Env = os.Environ()
	Cmd.Env = append(Cmd.Env, "GIT_SSL_NO_VERIFY=true")
	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execRun(cmd string) (string, error) {
	args := []string{}
	Cmd := exec.Command(cmd, args...)
	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execDownload(url string, file string) (string, error) {
	args := []string{"--continue", url, "-O", file}
	Cmd := exec.Command("wget", args...)
	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execAddRepoKey(URL string, file string) (string, error) {
	args := []string{"--continue", URL, "-O", file}
	Cmd := exec.Command("wget", args...)
	out, err := Cmd.Output()

	if err != nil {
		return string(out), err
	}
	args = []string{"--dearmor", file}
	Cmd = exec.Command("gpg", args...)
	out, err = Cmd.Output()

	return string(out), err
}

func (ds *Set) execSetPass(user string, pass string) (string, error) {
	args := []string{}
	Cmd := exec.Command("chpasswd", args...)
	Cmd.Stdin = strings.NewReader(user + ":" + pass)
	out, err := Cmd.Output()
	return string(out), err
}

func (ds *Set) execCopyFile(fileName string, dstDir string) (string, error) {
	dstFile, err := os.OpenFile(dstDir+"/"+fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()
	cfgFile, err := ds.files.Open("data/" + fileName)
	if err != nil {
		return "", err
	}
	defer cfgFile.Close()
	_, err = io.Copy(dstFile, cfgFile)
	return "", err
}

func (ds *Set) Run() {
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
		spinner.Start()

		switch step.Command {
		case CreateDir:
			out, err = ds.execCreateDir(step.Params[0])
		case AppendFile:
			out, err = ds.execAppendFile(step.Params[0], step.Params[1])
		case AddGroup:
			out, err = ds.execAddGroup(step.Params[0])
		case Replace:
			out, err = ds.execReplace(step.Params[0], step.Params[1])
		case CopyFile:
			out, err = ds.execCopyFile(step.Params[0], step.Params[1])
		case ChangeOwner:
			out, err = ds.execChangeOwner(step.Params[0], step.Params[1])
		case ChangePerm:
			out, err = ds.execChangePerm(step.Params[0], step.Params[1])
		case AssignGroups:
			out, err = ds.execAssignGroups(step.Params[0], step.Params[1])
		case ReloadSysctl:
			out, err = ds.execReloadSysctl()
		case UpdateRepos:
			out, err = ds.execUpdateRepos()
		case UpgradePackages:
			out, err = ds.execUpgradePackages()
		case InstallPackage:
			out, err = ds.execInstallPackage(step.Params[0])
		case EnableService:
			out, err = ds.execEnableService(step.Params[0])
		case UnzipFile:
			out, err = ds.execUnzipFile(step.Params[0], step.Params[1])
		case Run:
			out, err = ds.execRun(step.Params[0])
		case Untar:
			out, err = ds.execUntar(step.Params[0], step.Params[1])
		case CloneRepo:
			out, err = ds.execCloneRepo(step.Params[0], step.Params[1], step.Params[2])
		case AddUser:
			out, err = ds.execAddUser(step.Params[0])
		case SetPass:
			out, err = ds.execSetPass(step.Params[0], step.Params[1])
		case Download:
			out, err = ds.execDownload(step.Params[0], step.Params[1])
		case PrimaryGroup:
			out, err = ds.execPrimaryGroup(step.Params[0], step.Params[1])
		case ReloadUnits:
			out, err = ds.execReloadUnits()
		case AddRepoKey:
			out, err = ds.execAddRepoKey(step.Params[0], step.Params[1])
		}
		if err != nil {
			step.Status.ErrLvl = 1
			step.Status.Message = err.Error()
			ds.Steps[i] = step
			ds.saveStats()
			//fmt.Println("STATUS:", spinner.Status().String())
			spinner.StopFailMessage(step.Desc + ": " + err.Error())
			spinner.StopFail()
			log.Fatal(out)
		} else {
			step.Complete = true
			ds.Steps[i] = step
			ds.saveStats()
			spinner.Stop()
		}
	}
}
