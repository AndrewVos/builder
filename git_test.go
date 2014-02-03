package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreatesPushAndPullRequestHooks(t *testing.T) {
	var paths []string
	var bodies []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		b, _ := ioutil.ReadAll(r.Body)
		bodies = append(bodies, string(b))
	}))
	oldDomain := githubDomain
	githubDomain = ts.URL

	defer func() {
		githubDomain = oldDomain
		ts.Close()
	}()

	supportedEvents := []string{"push", "pull_request"}
	git := Git{}
	git.CreateHooks("lolsszz", "AndrewVos", "builder")

	for i, event := range supportedEvents {
		expectedPath := "/repos/AndrewVos/builder/hooks?access_token=lolsszz"
		if paths[i] != expectedPath {
			t.Errorf("Got wrong post address\nExpected: %v\nActual: %v", expectedPath, paths[i])
		}
		expectedBody := `{
      "name": "web",
      "active": true,
      "events": [ "` + event + `" ],
      "config": {
        "url": "http://localhost:1212/hooks/` + event + `",
        "content_type": "json"
      }
    }`
		if bodies[i] != expectedBody {
			t.Errorf("Didn't post expected body\nExpected:\n%v\nActual:\n%v", expectedBody, bodies[i])
		}
	}
}

func TestCanTellIfARepositoryIsPrivate(t *testing.T) {
	git := Git{}

	status := 0
	serverThatReturnsStatus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
	oldDomain := githubDomain
	githubDomain = serverThatReturnsStatus.URL
	defer func() {
		githubDomain = oldDomain
		serverThatReturnsStatus.Close()
	}()

	status = 200
	if git.IsRepositoryPrivate("bla", "reponame") {
		t.Errorf("repository isn't actually private")
	}

	status = 404
	if git.IsRepositoryPrivate("blaaaa", "ergh") == false {
		t.Errorf("repository returned 404, so it should be private")
	}
}

func withFakedGithubApiDomain(domain string, block func()) {
	oldDomain := githubDomain
	githubDomain = domain
	block()
	githubDomain = oldDomain
}

func TestListsRepoCollaborators(t *testing.T) {
	git := Git{}

	response := `
    [
      { "login": "AndrewVos", "id": 363618 },
      { "login": "andrewvo", "id": 1605821 }
    ]
  `

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedUrl := fmt.Sprintf("/repos/%v/%v/collaborators?access_token=%v", "owner1", "repo3", "TOKEN")
		if r.URL.RequestURI() == expectedUrl {
			w.Write([]byte(response))
		}
	}))
	defer ts.Close()

	withFakedGithubApiDomain(ts.URL, func() {
		collaborators := git.RepositoryCollaborators("TOKEN", "owner1", "repo3")
		if len(collaborators) != 2 {
			t.Fatalf("Expected to have 2 collaborators, not %d\n", len(collaborators))
		}
		if collaborators[0].Login != "AndrewVos" || collaborators[0].Id != 363618 {
			t.Errorf("Collaborator 0 was wrong. Got:\n%+v", collaborators[0])
		}
		if collaborators[1].Login != "andrewvo" || collaborators[1].Id != 1605821 {
			t.Errorf("Collaborator 1 was wrong. Got:\n%+v", collaborators[1])
		}
	})
}
