package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/hoisie/mustache"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var launcher BuildLauncher = &Builder{}
var database Database = &PostgresDatabase{}

type BuildLauncher interface {
	LaunchBuild(owner string, repo string, ref string, sha string, githubURL string, commits []Commit) error
}

type Builder struct {
}

func (builder *Builder) LaunchBuild(owner string, repo string, ref string, sha string, githubURL string, commits []Commit) error {
	repository := database.FindRepository(owner, repo)
	if repository == nil {
		return errors.New(fmt.Sprintf("Couldn't find access token to build %v/%v\n", owner, repo))
	}

	build := &Build{
		Owner:      owner,
		Repository: repo,
		Ref:        ref,
		Sha:        sha,
		GithubUrl:  githubURL,
		Commits:    commits,
	}
	err := database.CreateBuild(repository, build)
	if err != nil {
		return err
	}
	build.start()
	return nil
}

func defaultViewContext(r *http.Request) map[string]interface{} {
	account := currentAccount(r)

	context := map[string]interface{}{
		"client_id": configuration.GithubClientID,
		"logged_in": (account != nil),
	}
	return context
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	context := defaultViewContext(r)
	context["css"] = map[string]string{
		"name": "home.css",
	}
	context["js"] = map[string]string{
		"name": "home.js",
	}
	body := mustache.RenderFileInLayout("views/home.mustache", "views/layout.mustache", context)
	w.Write([]byte(body))
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	account := currentAccount(r)
	if account == nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	context := defaultViewContext(r)
	body := mustache.RenderFileInLayout("views/settings.mustache", "views/layout.mustache", context)
	w.Write([]byte(body))
}

func buildOutputHandler(w http.ResponseWriter, r *http.Request) {
	context := defaultViewContext(r)
	context["css"] = map[string]string{
		"name": "build_output.css",
	}
	context["js"] = map[string]string{
		"name": "build_output.js",
	}
	context["build_id"] = r.URL.Query().Get(":id")
	body := mustache.RenderFileInLayout("views/build_output.mustache", "views/layout.mustache", context)
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

	err = launcher.LaunchBuild(
		owner,
		name,
		strings.Replace(ref, "refs/heads/", "", -1),
		sha,
		githubURL,
		commits,
	)
	if err != nil {
		fmt.Println(err)
	}
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

	err = launcher.LaunchBuild(
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
}

func buildsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(database.AllBuilds(currentAccount(r)))
	w.Write(b)
}

func buildOutputRawHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, _ := strconv.Atoi(r.URL.Query().Get(":id"))
	start, _ := strconv.Atoi(r.URL.Query().Get("start"))
	for _, build := range database.AllBuilds(currentAccount(r)) {
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
	http.SetCookie(w, &http.Cookie{Name: "account_id", Value: "empty", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "token", Value: "empty", MaxAge: -1})
	http.Redirect(w, r, "/", 302)
}

func addRepositoryHandler(w http.ResponseWriter, r *http.Request) {
	account := currentAccount(r)

	if account != nil {
		owner := r.PostFormValue("owner")
		repository := r.PostFormValue("repository")

		err := git.CreateHooks(account.AccessToken, owner, repository)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}

		err = database.AddRepositoryToAccount(account, &Repository{
			Owner:      owner,
			Repository: repository,
		})
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
	}
}

func githubLoginHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	accessToken, err := git.GetAccessToken(
		configuration.GithubClientID,
		configuration.GithubClientSecret,
		code)
	if err != nil {
		fmt.Println(err)
		return
	}

	githubUserID, err := git.GetUserID(accessToken)
	if err != nil {
		fmt.Println(err)
		return
	}

	account := &Account{
		Id:          githubUserID,
		AccessToken: accessToken,
	}

	err = database.CreateAccount(account)
	if err != nil {
		return
	}

	login, err := database.CreateLoginForAccount(account)
	if err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "account_id", Value: strconv.Itoa(githubUserID)})
	http.SetCookie(w, &http.Cookie{Name: "token", Value: login.Token})
	http.Redirect(w, r, "/", 302)
}
