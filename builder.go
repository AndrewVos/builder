package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/howeyc/gopass"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

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
	build(push)
}

func build(push GithubPushEvent) {
	os.MkdirAll(buildPath(push), 0600)
	output, _ := os.Create(logPath(push))
	defer output.Close()
	checkout(push, output)
	execute(push, output)
}

func sourcePath(push GithubPushEvent) string {
	return buildPath(push) + "/source"
}

func logPath(push GithubPushEvent) string {
	return buildPath(push) + "/output.log"
}

func resultPath(push GithubPushEvent) string {
	return buildPath(push) + "/result.json"
}

func buildPath(push GithubPushEvent) string {
	hash := md5.New()
	io.WriteString(hash, push.Ref)
	io.WriteString(hash, push.HeadCommit.ID)
	return "builds/" + fmt.Sprintf("%x", hash.Sum(nil))
}

// func ensureHookComesFromGithub(request) {
//   ips = get("https://api.github.com/meta")
//   return request.ip in ips
// }

func checkout(push GithubPushEvent, output *os.File) {
	branch := strings.Split(push.Ref, "/")[2]

	url := strings.Replace(push.Repository.URL, "https://", "https://"+retrieveAuthToken()+"@", -1)

	cmd := exec.Command("git", "clone", "--depth=50", "--branch", branch, url, sourcePath(push))
	cmd.Stdout = output
	cmd.Stderr = output

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmd = exec.Command("git", "checkout", push.HeadCommit.ID)
	cmd.Dir = sourcePath(push)
	cmd.Stdout = output
	cmd.Stderr = output

	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execute(push GithubPushEvent, output *os.File) {
	cmd := exec.Command("ssh", "-t", "-t", "localhost", "cd "+sourcePath(push)+";./Builderfile")
	cmd.Dir = sourcePath(push)
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmd.Wait()

	fmt.Println("build complete")
	result := &BuildResult{
		Success: cmd.ProcessState.Success(),
	}
	b, _ := json.Marshal(result)
	ioutil.WriteFile(resultPath(push), b, 0600)
}

type BuildResult struct {
	Success bool
}
