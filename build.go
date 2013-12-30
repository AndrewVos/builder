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
	ID        string
	URL       string
	Owner     string
	Repo      string
	Ref       string
	SHA       string
	Complete  bool
	Success   bool
	Result    string
	GithubURL string
}

func init() {
	for _, build := range AllBuilds() {
		if build.Complete == false {
			build.fail()
		}
	}
}

func NewBuild(owner string, repo string, ref string, sha string, githubURL string) *Build {
	build := &Build{
		Owner:     owner,
		Repo:      repo,
		Ref:       ref,
		SHA:       sha,
		Result:    "incomplete",
		GithubURL: githubURL,
	}

	hash := md5.New()
	io.WriteString(hash, build.Ref)
	io.WriteString(hash, build.SHA)
	build.ID = fmt.Sprintf("%v-%x", time.Now().Unix(), hash.Sum(nil))

	build.URL = CurrentConfiguration().Host
	if CurrentConfiguration().Port != "80" {
		build.URL += ":" + CurrentConfiguration().Port
	}
	build.URL += "/build_output?id=" + build.ID

	return build
}

func BuildResultsPath() string {
	return "data/build_results.json"
}

func (build *Build) save() {
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
	err = ioutil.WriteFile(BuildResultsPath(), marshalled, 0700)
	if err != nil {
		fmt.Println(err)
	}
}

func (build *Build) start() {
	build.save()
	err := os.MkdirAll(build.Path(), 0700)
	if err != nil {
		fmt.Println(err)
		build.fail()
		return
	}
	output, _ := os.Create(build.LogPath())
	defer output.Close()

	err = build.checkout(output)
	if err != nil {
		fmt.Println(err)
		build.fail()
		return
	}

	err = build.execute(output)
	if err != nil {
		fmt.Println(err)
		build.fail()
		return
	}
	build.pass()
}

func (build *Build) checkout(output *os.File) error {
	url := "https://" + CurrentConfiguration().AuthToken + "@github.com/" + build.Owner + "/" + build.Repo

	err := git.Retrieve(output, url, build.SourcePath(), build.Ref, build.SHA)
	if err != nil {
		return err
	}

	return nil
}

func (build *Build) environs() []string {
	return []string{
		"BUILDER_BUILD_RESULT=" + build.Result,
		"BUILDER_BUILD_URL=" + build.URL,
		"BUILDER_BUILD_ID=" + build.ID,
		"BUILDER_BUILD_OWNER=" + build.Owner,
		"BUILDER_BUILD_REPO=" + build.Repo,
		"BUILDER_BUILD_REF=" + build.Ref,
		"BUILDER_BUILD_SHA=" + build.SHA,
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
	build.save()
	build.executeHooks()
}

func (build *Build) fail() {
	build.Complete = true
	build.Success = false
	build.Result = "fail"
	build.save()
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
	return "data/builds/" + b.ID
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

func FindBuildById(id string) *Build {
	for _, build := range AllBuilds() {
		if build.ID == id {
			return build
		}
	}
	return nil
}

func AllBuilds() []*Build {
	b, err := ioutil.ReadFile(BuildResultsPath())
	if err != nil {
		return []*Build{}
	}
	var allBuilds []*Build
	err = json.Unmarshal(b, &allBuilds)
	return allBuilds
}
