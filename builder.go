package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/howeyc/gopass"
	"github.com/kr/pty"
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

func (b *Build) Save() {
	path := buildID(b) + ".json"
	marshalled, _ := json.Marshal(b)

	err := ioutil.WriteFile(path, marshalled, 0644)
	if err != nil {
		fmt.Println(err)
	}
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
	Host         string
	Port         string
	Repositories []Repository
}

func NewConfigurationFromFile(path string) *Configuration {
	c := &Configuration{}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading config file")
		fmt.Println(err)
		os.Exit(1)
	}
	json.Unmarshal(b, c)
	return c
}

type Repository struct {
	Owner      string
	Repository string
}

var configuration *Configuration

func main() {
	configuration = NewConfigurationFromFile("builder.json")
	token := retrieveAuthToken()
	createHooks(token)
	serve()
}

type GithubAuthorization struct {
	Token string `json:"token"`
}

func retrieveAuthToken() string {
	if _, err := os.Stat("AUTH_TOKEN"); os.IsNotExist(err) {
		var username string
		fmt.Print("Github username: ")
		fmt.Scanln(&username)
		fmt.Print("Github password: ")
		password := gopass.GetPasswd()

		url := "https://api.github.com/authorizations"
		body := strings.NewReader(`{"scopes":["repo"]}`)
		client := &http.Client{}
		request, _ := http.NewRequest("POST", url, body)
		request.SetBasicAuth(username, string(password))
		response, err := client.Do(request)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		responseBody, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()

		var authorization GithubAuthorization
		err = json.Unmarshal(responseBody, &authorization)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = ioutil.WriteFile("AUTH_TOKEN", []byte(authorization.Token), 0644)
		if err != nil {
			panic(err)
		}
	}

	token, err := ioutil.ReadFile("AUTH_TOKEN")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return string(token)
}

func createHooks(token string) {
	for _, repo := range configuration.Repositories {
		url := "https://api.github.com/repos/" + repo.Owner + "/" + repo.Repository + "/hooks?access_token=" + token
		body := `
    {
      "name": "web",
      "active": true,
      "events": [
        "push",
        "pull_request"
      ],
      "config": {
        "url": "` + configuration.Host + ":" + configuration.Port + `/hook",
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
		responseBody, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()
		fmt.Println(string(responseBody))
	}
}

func serve() {
	http.HandleFunc("/hook", hookHandler)
	http.ListenAndServe(":"+configuration.Port, nil)
}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	fmt.Println(string(body))
	var push GithubPushEvent
	json.Unmarshal(body, &push)

	build := &Build{
		Owner: push.Repository.Owner.Name,
		Repo:  push.Repository.Name,
		Ref:   push.Ref,
		SHA:   push.HeadCommit.ID,
	}
	build.Save()
	os.MkdirAll(build.Path(), 0600)
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

// func ensureHookComesFromGithub(request) {
//   ips = get("https://api.github.com/meta")
//   return request.ip in ips
// }

func checkout(build *Build, output *os.File) {
	branch := strings.Split(build.Ref, "/")[2]

	url := "https://" + retrieveAuthToken() + "@github.com/" + build.Owner + "/" + build.Repo

	cmd := exec.Command("git", "clone", "--depth=50", "--branch", branch, url, build.SourcePath())
	cmd.Stdout = output
	cmd.Stderr = output

	err := cmd.Run()
	if err != nil {
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

	fmt.Println("build complete")
	success := true
	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
		success = false
	}

	build.Complete = true
	build.Success = success
	build.Save()
}
