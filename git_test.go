package main

import (
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
