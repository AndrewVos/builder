package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)

type FakeGit struct {
	FakeRepo              string
	createHooksParameters map[string]interface{}
}

func (g *FakeGit) Retrieve(log io.Writer, url string, path string, branch string, sha string) error {
	files, _ := ioutil.ReadDir("test-repos/" + g.FakeRepo)
	os.MkdirAll(path, 0700)
	for _, file := range files {
		b, _ := ioutil.ReadFile("test-repos/" + g.FakeRepo + "/" + file.Name())
		ioutil.WriteFile(path+"/"+file.Name(), b, 0700)
	}
	return nil
}

func (g *FakeGit) CreateHooks(accessToken string, owner string, repo string) error {
	g.createHooksParameters = map[string]interface{}{
		"accessToken": accessToken,
		"owner":       owner,
		"repository":  repo,
	}
	return nil
}

func postToHooks(path string, event string) {
	b, _ := ioutil.ReadFile(path)
	request, _ := http.NewRequest("POST", "/hooks/"+event, nil)
	request.Body = ioutil.NopCloser(strings.NewReader(string(b)))
	w := httptest.NewRecorder()
	if event == "push" {
		pushHandler(w, request)
	} else if event == "pull_request" {
		pullRequestHandler(w, request)
	}
}

func setup(fakeRepo string) {
	os.Mkdir("data", 0700)
	os.Mkdir("data/hooks", 0700)

	if fakeRepo != "" {
		git = &FakeGit{FakeRepo: fakeRepo}
	}

	database := &PostgresDatabase{}
	database.SaveGithubBuild(&GithubBuild{AccessToken: "hello", Owner: "AndrewVos", Repository: "builder-test-green-repo"})
	database.SaveGithubBuild(&GithubBuild{AccessToken: "hello", Owner: "AndrewVos", Repository: "builder-test-red-repo"})
}

func cleanup() {
	db, _ := connect()
	db.Query("DELETE FROM github_builds").Run()
	db.Query("DELETE FROM builds").Run()
	db.Query("DELETE FROM commits").Run()
	os.RemoveAll("data")
}
