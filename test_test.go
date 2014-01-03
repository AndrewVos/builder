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
	FakeRepo string
}

func (git FakeGit) Retrieve(log io.Writer, url string, path string, branch string, sha string) error {
	files, _ := ioutil.ReadDir("test-repos/" + git.FakeRepo)
	os.MkdirAll(path, 0700)
	for _, file := range files {
		b, _ := ioutil.ReadFile("test-repos/" + git.FakeRepo + "/" + file.Name())
		ioutil.WriteFile(path+"/"+file.Name(), b, 0700)
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
		git = FakeGit{FakeRepo: fakeRepo}
	}

	GithubBuild{AccessToken: "hello", RepositoryOwner: "AndrewVos", RepositoryName: "builder-test-green-repo"}.Save()
	GithubBuild{AccessToken: "hello", RepositoryOwner: "AndrewVos", RepositoryName: "builder-test-red-repo"}.Save()
}

func cleanup() {
	db, _ := connect()
	db.Exec("DELETE FROM github_builds")
	os.RemoveAll("data")
	git = Git{}
}
