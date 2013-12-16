package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/kr/pty"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type Build struct {
	Owner    string
	Repo     string
	Ref      string
	SHA      string
	Complete bool
	Success  bool
}

func (build *Build) save() {
	path := "build_results.json"

	var newBuilds []*Build
	for _, b := range AllBuilds() {
		if b.ID() != build.ID() {
			newBuilds = append(newBuilds, build)
		}
	}
	newBuilds = append(newBuilds, build)

	marshalled, _ := json.Marshal(newBuilds)
	err := ioutil.WriteFile(path, marshalled, 0700)
	if err != nil {
		fmt.Println(err)
	}
}

func (build *Build) ID() string {
	hash := md5.New()
	io.WriteString(hash, build.Ref)
	io.WriteString(hash, build.SHA)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (build *Build) start() {
	build.save()
	err := os.MkdirAll(build.Path(), 0700)
	if err != nil {
		fmt.Println(err)
		return
	}
	output, _ := os.Create(build.LogPath())
	defer output.Close()
	build.checkout(output)
	build.execute(output)
}

func (build *Build) checkout(output *os.File) {
	url := "https://" + CurrentConfiguration().AuthToken + "@github.com/" + build.Owner + "/" + build.Repo

	cmd := exec.Command("git", "clone", "--depth=50", "--branch", build.Ref, url, build.SourcePath())
	cmd.Stdout = output
	cmd.Stderr = output

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error cloning repository")
		fmt.Println(err)
		os.Exit(1)
	}

	cmd = exec.Command("git", "checkout", build.SHA)
	cmd.Dir = build.SourcePath()
	cmd.Stdout = output
	cmd.Stderr = output

	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (build *Build) execute(output *os.File) {
	cmd := exec.Command("bash", "./Builderfile")
	cmd.Dir = build.SourcePath()
	cmd.Stdout = output
	cmd.Stderr = output
	f, err := pty.Start(cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	io.Copy(output, f)

	success := true
	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
		success = false
	}

	build.Complete = true
	build.Success = success
	build.save()
}

func (b *Build) Path() string {
	return "builds/" + b.ID()
}

func (b *Build) LogPath() string {
	return b.Path() + "/output.log"
}

func (b *Build) SourcePath() string {
	return b.Path() + "/source"
}

func AllBuilds() []*Build {
	path := "build_results.json"
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return []*Build{}
	}
	var allBuilds []*Build
	err = json.Unmarshal(b, &allBuilds)
	return allBuilds
}
