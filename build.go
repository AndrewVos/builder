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
	"time"
)

type Build struct {
	ID       string
	Owner    string
	Repo     string
	Ref      string
	SHA      string
	Complete bool
	Success  bool
}

func NewBuild(owner string, repo string, ref string, sha string) *Build {
	build := &Build{
		Owner: owner,
		Repo:  repo,
		Ref:   ref,
		SHA:   sha,
	}

	hash := md5.New()
	io.WriteString(hash, build.Ref)
	io.WriteString(hash, build.SHA)
	build.ID = fmt.Sprintf("%v-%x", time.Now().Unix(), hash.Sum(nil))

	return build
}

func (build *Build) save() {
	path := "build_results.json"

	allBuilds := AllBuilds()
	added := false

	for i, b := range allBuilds {
		if b.ID == build.ID {
			allBuilds[i] = build
			added = true
		}
	}
	if !added {
		allBuilds = append(allBuilds, build)
	}

	marshalled, err := json.Marshal(allBuilds)
	if err != nil {
		fmt.Println(err)
	}
	err = ioutil.WriteFile(path, marshalled, 0700)
	if err != nil {
		fmt.Println(err)
	}
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

	customEnv := []string{
		"BUILDER_BUILD_ID=" + build.ID,
		"BUILDER_BUILD_OWNER=" + build.Owner,
		"BUILDER_BUILD_REPO=" + build.Repo,
		"BUILDER_BUILD_REF=" + build.Ref,
		"BUILDER_BUILD_SHA=" + build.SHA,
	}
	for _, c := range os.Environ() {
		customEnv = append(customEnv, c)
	}
	cmd.Env = customEnv

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
	return "builds/" + b.ID
}

func (b *Build) LogPath() string {
	return b.Path() + "/output.log"
}

func (b *Build) SourcePath() string {
	return b.Path() + "/source"
}

func (build *Build) ReadOutput() string {
	b, err := ioutil.ReadFile(build.LogPath())
	if err != nil {
		fmt.Println(err)
	}
	return string(b)
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
