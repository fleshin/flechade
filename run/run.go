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
	"time"

	"github.com/theckman/yacspin"
	"gopkg.in/yaml.v3"
)

var Commands map[string]func(*Set, ...string) (string, error)
var version string

func init() {
	version = "0.0.5"
	Commands = make(map[string]func(*Set, ...string) (string, error))
	LoadCommands()
}

func SetCommand(n string, f func(*Set, ...string) (string, error)) {
	Commands[n] = f
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
	//if !slices.Contains(Commands, cmdId) {
	_, ok := Commands[cmdId]
	if !ok {
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

		out, err = Commands[step.Command](ds, step.Params...)

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
