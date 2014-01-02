package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/hoisie/mustache"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var githubDomain string = "https://api.github.com"
var git GitTool

func main() {
	createHooks()
	serve()
}

func init() {
	git = Git{}
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/hooks/push", pushHandler)
	http.HandleFunc("/hooks/pull_request", pullRequestHandler)
	http.HandleFunc("/builds", buildsHandler)
	http.HandleFunc("/build_output", buildOutputHandler)
	http.HandleFunc("/build_output_raw", buildOutputRawHandler)
	http.HandleFunc("/github_callback", githubCallbackHandler)
	http.HandleFunc("/logout", logoutHandler)

	serveFile("public/scripts/jquery-2.0.3.min.js")
	serveFile("public/scripts/build.js")
	serveFile("public/scripts/build_output.js")
	serveFile("public/styles/common.css")
	serveFile("public/styles/build.css")
	serveFile("public/styles/build_output.css")
	serveFile("public/styles/bootstrap.min.css")
}

func serve() {
	err := http.ListenAndServe(":"+CurrentConfiguration().Port, nil)
	if err != nil {
		panic(err)
	}
}

func serveFile(filename string) {
	pattern := strings.Replace(filename, "public", "", 1)
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	})
}

func authenticated(r *http.Request) bool {
	_, err := r.Cookie("github_access_token")
	if err != nil {
		return false
	}
	return true
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	loggedIn := authenticated(r)

	context := map[string]interface{}{
		"clientID": os.Getenv("GITHUB_CLIENT_ID"),
		"loggedIn": loggedIn,
	}
	body := mustache.RenderFile("views/home.mustache", context)
	w.Write([]byte(body))
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	push, err := simplejson.NewJson(body)
	if err != nil {
		fmt.Println("Error parsing push")
		fmt.Println(err)
		return
	}

	deleted, err := push.Get("deleted").Bool()
	if err == nil && deleted {
		return
	}

	ref, _ := push.Get("ref").String()
	owner, _ := push.Get("repository").Get("owner").Get("name").String()
	name, _ := push.Get("repository").Get("name").String()
	sha, _ := push.Get("head_commit").Get("id").String()
	githubURL, _ := push.Get("compare").String()

	jsonCommits, err := push.Get("commits").Array()
	var commits []Commit
	for _, c := range jsonCommits {
		m := c.(map[string]interface{})
		commit := Commit{
			SHA:     m["id"].(string),
			Message: m["message"].(string),
			URL:     m["url"].(string),
		}
		commits = append(commits, commit)
	}

	build := NewBuild(
		owner,
		name,
		strings.Replace(ref, "refs/heads/", "", -1),
		sha,
		githubURL,
		commits,
	)
	build.start()
}

func pullRequestHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	pullRequest, err := simplejson.NewJson(body)
	if err != nil {
		fmt.Println("Error parsing pull request")
		fmt.Println(err)
		return
	}

	action, _ := pullRequest.Get("action").String()
	if action != "opened" {
		return
	}

	fullName, _ := pullRequest.Get("repository").Get("full_name").String()
	ref, _ := pullRequest.Get("pull_request").Get("head").Get("ref").String()
	sha, _ := pullRequest.Get("pull_request").Get("head").Get("sha").String()
	githubURL, _ := pullRequest.Get("pull_request").Get("_links").Get("self").Get("href").String()

	build := NewBuild(
		strings.Split(fullName, "/")[0],
		strings.Split(fullName, "/")[1],
		ref,
		sha,
		githubURL,
		nil,
	)
	build.start()
}

func buildsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(AllBuilds())
	w.Write(b)
}

func buildOutputRawHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	start, _ := strconv.Atoi(r.URL.Query().Get("start"))
	for _, build := range AllBuilds() {
		if build.ID == id {
			raw := build.ReadOutput()
			raw = raw[start:]
			converted := ""
			if len(raw) != 0 {
				converted = AnsiToHtml(raw)
			}
			output := map[string]interface{}{
				"length": len(raw),
				"output": converted,
			}
			b, _ := json.Marshal(output)
			w.Write(b)
			return
		}
	}
}

func buildOutputHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	body := mustache.RenderFile("views/build_output.mustache", map[string]string{"build_id": id})
	w.Write([]byte(body))
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "github_access_token", Value: "empty", MaxAge: -1})
	http.Redirect(w, r, "/", 302)
}

func githubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	body, _ := json.Marshal(map[string]interface{}{
		"client_id":     os.Getenv("GITHUB_CLIENT_ID"),
		"client_secret": os.Getenv("GITHUB_CLIENT_SECRET"),
		"code":          code,
	})

	client := &http.Client{}
	request, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewReader(body))

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer response.Body.Close()
	body, _ = ioutil.ReadAll(response.Body)
	var accessTokenResponse map[string]string
	json.Unmarshal(body, &accessTokenResponse)
	http.SetCookie(w, &http.Cookie{Name: "github_access_token", Value: accessTokenResponse["access_token"]})
	http.Redirect(w, r, "/", 302)
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
