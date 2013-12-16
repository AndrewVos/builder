package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/kr/pty"
	"github.com/likexian/simplejson"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Build struct {
	Owner    string
	Repo     string
	Ref      string
	SHA      string
	Complete bool
	Success  bool
}

func (build *Build) Save() {
	path := "build_results.json"

	var newBuilds []*Build
	for _, b := range allBuilds() {
		if buildID(b) != buildID(build) {
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

func allBuilds() []*Build {
	path := "build_results.json"
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return []*Build{}
	}
	var allBuilds []*Build
	err = json.Unmarshal(b, &allBuilds)
	return allBuilds
}

func (b *Build) Path() string {
	return "builds/" + buildID(b)
}

func (b *Build) LogPath() string {
	return b.Path() + "/output.log"
}

func (b *Build) SourcePath() string {
	return b.Path() + "/source"
}

type GithubPushEvent struct {
	Ref        string
	HeadCommit GithubCommit `json:"head_commit"`
	Repository GithubRepository
}

type GithubRepository struct {
	Name  string
	URL   string
	Owner GithubOwner
}

type GithubOwner struct {
	Name string
}

type GithubCommit struct {
	ID string
}

type Configuration struct {
	AuthToken    string
	Host         string
	Port         string
	Repositories []Repository
}

func CurrentConfiguration() Configuration {
	b, err := ioutil.ReadFile("builder.json")
	if err != nil {
		fmt.Println("Error reading config file")
		fmt.Println(err)
	}

	var c Configuration
	err = json.Unmarshal(b, &c)
	if err != nil {
		fmt.Println("Error parsing config file")
		fmt.Println(err)
	}
	return c
}

type Repository struct {
	Owner      string
	Repository string
}

var githubDomain string = "https://api.github.com"

func main() {
	createHooks()
	serve()
}

func createHooks() {
	for _, repo := range CurrentConfiguration().Repositories {
		url := githubDomain + "/repos/" + repo.Owner + "/" + repo.Repository + "/hooks?access_token=" + CurrentConfiguration().AuthToken

		supportedEvents := []string{"push", "pull_request"}
		for _, event := range supportedEvents {
			body := `{
      "name": "web",
      "active": true,
      "events": [ "` + event + `" ],
      "config": {
        "url": "` + CurrentConfiguration().Host + ":" + CurrentConfiguration().Port + `/hooks/` + event + `",
        "content_type": "json"
      }
    }`

			client := &http.Client{}
			request, _ := http.NewRequest("POST", url, strings.NewReader(body))
			response, err := client.Do(request)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			response.Body.Close()
		}
	}
}

func init() {
	http.HandleFunc("/hooks/push", pushHandler)
	http.HandleFunc("/hooks/pull_request", pullRequestHandler)
}

func serve() {
	http.ListenAndServe(":"+CurrentConfiguration().Port, nil)
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var push GithubPushEvent
	json.Unmarshal(body, &push)

	refParts := strings.Split(push.Ref, "/")
	build := &Build{
		Owner: push.Repository.Owner.Name,
		Repo:  push.Repository.Name,
		Ref:   refParts[len(refParts)-1],
		SHA:   push.HeadCommit.ID,
	}
	startBuild(build)
}

func pullRequestHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	pullRequest, err := simplejson.Loads(string(body))
	if err != nil {
		fmt.Println("Error parsing pull request")
		fmt.Println(err)
	}

	action, _ := pullRequest.Get("action").String()
	if action == "closed" {
		return
	}

	fullName, _ := pullRequest.Get("repository").Get("full_name").String()
	ref, _ := pullRequest.Get("pull_request").Get("head").Get("ref").String()
	sha, _ := pullRequest.Get("pull_request").Get("head").Get("sha").String()
	build := &Build{
		Owner: strings.Split(fullName, "/")[0],
		Repo:  strings.Split(fullName, "/")[1],
		Ref:   ref,
		SHA:   sha,
	}
	startBuild(build)
}

func startBuild(build *Build) {
	build.Save()
	err := os.MkdirAll(build.Path(), 0700)
	if err != nil {
		fmt.Println(err)
		return
	}
	output, _ := os.Create(build.LogPath())
	defer output.Close()
	checkout(build, output)
	execute(build, output)
}

func buildID(build *Build) string {
	hash := md5.New()
	io.WriteString(hash, build.Ref)
	io.WriteString(hash, build.SHA)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func checkout(build *Build, output *os.File) {
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

func execute(build *Build, output *os.File) {
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
	build.Save()
}
