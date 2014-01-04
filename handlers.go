package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/hoisie/mustache"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	_, loggedIn := authenticated(r)

	context := map[string]interface{}{
		"clientID": configuration.GithubClientID,
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
			Sha:     m["id"].(string),
			Message: m["message"].(string),
			Url:     m["url"].(string),
		}
		commits = append(commits, commit)
	}

	build, err := CreateBuild(
		owner,
		name,
		strings.Replace(ref, "refs/heads/", "", -1),
		sha,
		githubURL,
		commits,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
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

	build, err := CreateBuild(
		strings.Split(fullName, "/")[0],
		strings.Split(fullName, "/")[1],
		ref,
		sha,
		githubURL,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	build.start()
}

func buildsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(AllBuilds())
	w.Write(b)
}

func buildOutputHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	body := mustache.RenderFile("views/build_output.mustache", map[string]string{"build_id": id})
	w.Write([]byte(body))
}

func buildOutputRawHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	start, _ := strconv.Atoi(r.URL.Query().Get("start"))
	for _, build := range AllBuilds() {
		if build.Id == id {
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

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "github_access_token", Value: "empty", MaxAge: -1})
	http.Redirect(w, r, "/", 302)
}

func addRepositoryHandler(w http.ResponseWriter, r *http.Request) {
	cookie, loggedIn := authenticated(r)
	if loggedIn {
		ghb := GithubBuild{
			AccessToken: cookie.Value,
			Owner:       r.PostFormValue("owner"),
			Repository:  r.PostFormValue("repository"),
		}

		err := createHooks(ghb.AccessToken, ghb.Owner, ghb.Repository)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}

		err = ghb.Save()
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
	}
}

func githubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	body, _ := json.Marshal(map[string]interface{}{
		"client_id":     configuration.GithubClientID,
		"client_secret": configuration.GithubClientSecret,
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
