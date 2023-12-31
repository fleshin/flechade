package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/fleshin/flechade/run"
	git "gopkg.in/src-d/go-git.v4"
)

func showVersion() {
	fmt.Println("flechade - customize your linux")
	fmt.Println("Version:", run.GetVer())
	fmt.Println("")
}

func runFromLocal(cfgFS embed.FS) {
	targetDir := "/tmp/flechade-default"
	err := os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		log.Fatal("unable to create directory: ", targetDir)
	}
	fs.WalkDir(cfgFS, "data", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if !d.IsDir() {
			//fmt.Println(path)
			last := strings.LastIndex(path, "/")
			fname := path[last:]
			dstFile, err := os.OpenFile(targetDir+fname, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer dstFile.Close()
			cfgFile, err := cfgFS.Open(path)
			if err != nil {
				return err
			}
			defer cfgFile.Close()
			_, err = io.Copy(dstFile, cfgFile)
			if err != nil {
				return err
			}
		}
		return nil
	})
	runFromDir(targetDir)
}

func runFromDir(dataDir string) {
	set, err := run.LoadSetFromDir(dataDir)
	if err != nil {
		log.Fatal(err)
	}
	set.Run()
	fmt.Println("Setup complete. Enjoy!")
}

func runFromUrl(repoUrl string) {
	tgtDir := "/tmp/flechade-repo"
	err := os.RemoveAll(tgtDir)
	if err != nil {
		log.Fatal(err)
	}
	_, err = git.PlainClone(tgtDir, false, &git.CloneOptions{
		URL: repoUrl,
	})
	if err != nil {
		log.Fatal(err)
	}
	runFromDir(tgtDir)
}

func contPrevRun() {
	set := new(run.Set)
	err := set.Load()
	if err != nil {
		log.Fatal(err)
	}
	set.Run()
}
