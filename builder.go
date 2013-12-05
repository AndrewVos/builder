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

func main() {
	token := retrieveAuthToken()
	createHook(token)
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

func createHook(token string) {
	// ["push", "pull_request"]
	// http://developer.github.com/v3/repos/hooks/
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
