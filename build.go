package main

import (
	"errors"
	"fmt"
	"github.com/kr/pty"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

type Build struct {
	Id           int
	RepositoryId int
	Url          string
	Owner        string
	Repository   string
	Ref          string
	Sha          string
	Complete     bool
	Success      bool
	Result       string
	GithubUrl    string
	Commits      []Commit
}

type Commit struct {
	Id      int
	BuildId int
	Sha     string
	Message string
	Url     string
}

func (build *Build) start() {
	err := os.MkdirAll(build.Path(), 0700)
	if err != nil {
		build.fail()
		return
	}

	output, err := os.Create(build.LogPath())
	if err != nil {
		build.fail()
		return
	}
	defer output.Close()

	err = build.checkout(output)
	if err != nil {
		build.fail()
		return
	}

	err = build.execute(output)
	if err != nil {
		build.fail()
		return
	}
	build.pass()
}

func (build *Build) checkout(output *os.File) error {
	repository := database.FindRepository(build.Owner, build.Repository)
	if repository == nil {
		return errors.New("Don't have access to build this project")
	}
	url := "https://" + repository.AccessToken + "@github.com/" + build.Owner + "/" + build.Repository

	err := git.Retrieve(output, url, build.SourcePath(), build.Ref, build.Sha)
	if err != nil {
		fmt.Fprintln(output, err)
		return err
	}

	return nil
}

func (build *Build) environs() []string {
	return []string{
		"BUILDER_BUILD_RESULT=" + build.Result,
		"BUILDER_BUILD_URL=" + build.Url,
		"BUILDER_BUILD_ID=" + strconv.Itoa(build.Id),
		"BUILDER_BUILD_OWNER=" + build.Owner,
		"BUILDER_BUILD_REPO=" + build.Repository,
		"BUILDER_BUILD_REF=" + build.Ref,
		"BUILDER_BUILD_SHA=" + build.Sha,
	}
}

func (build *Build) execute(output *os.File) error {
	cmd := exec.Command("bash", "./Builderfile")
	cmd.Dir = build.SourcePath()
	cmd.Stdout = output
	cmd.Stderr = output

	customEnv := build.environs()

	for _, c := range os.Environ() {
		customEnv = append(customEnv, c)
	}
	cmd.Env = customEnv

	f, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	io.Copy(output, f)

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (build *Build) pass() {
	build.Complete = true
	build.Success = true
	build.Result = "pass"
	database.SaveBuild(build)
	build.executeHooks()
}

func (build *Build) fail() {
	build.Complete = true
	build.Success = false
	build.Result = "fail"
	database.SaveBuild(build)
	build.executeHooks()
}

func (build *Build) executeHooks() {
	hooks, _ := ioutil.ReadDir("data/hooks")
	for _, file := range hooks {
		cmd := exec.Command("bash", "../../../data/hooks/"+file.Name())
		cmd.Dir = build.Path()

		customEnv := build.environs()

		for _, c := range os.Environ() {
			customEnv = append(customEnv, c)
		}
		cmd.Env = customEnv
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(output))
	}
}

func (b *Build) Path() string {
	return "data/builds/" + strconv.Itoa(b.Id)
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
