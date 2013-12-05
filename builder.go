package main

import (
	"encoding/json"
	"fmt"
	"github.com/howeyc/gopass"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

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
}

// func hookHandler() {
//   * https://help.github.com/articles/post-receive-hooks
// if ensureHookComesFromGithub(request) == false {
// 	return
// }
// buildPath = generateBuildPath()
// push = json.Unmarshal(request.body)
// checkout(buildPath, push)
// build(buildPath)
// }

// func buildHandler() {
// }

// func ensureHookComesFromGithub(request) {
//   ips = get("https://api.github.com/meta")
//   return request.ip in ips
// }

// func checkout(path string, push) {
//   branch = strings.Split(push.ref, "/"))[-1]
//   run("mkdir " + path)
//   run("git clone --depth=50 --quiet --branch "+push.branch+ "https://github.com/"+push.owner + "/" + push.repo + ".git")
//   run("git checkout -fq " + push.headCommit.id)
// }

// func build(path string) {
//   run("cd path")
//   run("bash Builderfile")
// }

// func run(directory string, command string) {
//   "cd directory"
//   i = exec(command)
//   i.output = directory + "/output.log"
// }
