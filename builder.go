package main

import (
	"encoding/json"
	"fmt"
	"github.com/hoisie/mustache"
	"github.com/likexian/simplejson"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var githubDomain string = "https://api.github.com"

func main() {
	createHooks()
	serve()
}

func init() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/hooks/push", pushHandler)
	http.HandleFunc("/hooks/pull_request", pullRequestHandler)
	http.HandleFunc("/builds", buildsHandler)
	serveFile("/scripts/jquery-2.0.3.min.js", "public/scripts/jquery-2.0.3.min.js")
	serveFile("/scripts/build.js", "public/scripts/build.js")
	serveFile("/styles/build.css", "public/styles/build.css")
	serveFile("/styles/bootstrap.min.css", "public/styles/bootstrap.min.css")
}

func serve() {
	err := http.ListenAndServe(":"+CurrentConfiguration().Port, nil)
	if err != nil {
		panic(err)
	}
}

func serveFile(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	body := mustache.RenderFile("views/home.mustache")
	w.Write([]byte(body))
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	push, err := simplejson.Loads(string(body))
	if err != nil {
		fmt.Println("Error parsing push")
		fmt.Println(err)
		return
	}

	ref, _ := push.Get("ref").String()
	owner, _ := push.Get("repository").Get("owner").Get("name").String()
	name, _ := push.Get("repository").Get("name").String()
	refParts := strings.Split(ref, "/")
	sha, _ := push.Get("head_commit").Get("id").String()
	build := NewBuild(
		owner,
		name,
		refParts[len(refParts)-1],
		sha,
	)
	build.start()
}

func pullRequestHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	pullRequest, err := simplejson.Loads(string(body))
	if err != nil {
		fmt.Println("Error parsing pull request")
		fmt.Println(err)
		return
	}

	action, _ := pullRequest.Get("action").String()
	if action == "closed" {
		return
	}

	fullName, _ := pullRequest.Get("repository").Get("full_name").String()
	ref, _ := pullRequest.Get("pull_request").Get("head").Get("ref").String()
	sha, _ := pullRequest.Get("pull_request").Get("head").Get("sha").String()
	build := NewBuild(
		strings.Split(fullName, "/")[0],
		strings.Split(fullName, "/")[1],
		ref,
		sha,
	)
	build.start()
}

func buildsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(AllBuilds())
	w.Write(b)
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
			if response.StatusCode == 401 {
				fmt.Println("Auth Token appears to be invalid")
				os.Exit(1)
			}
		}
	}
}
