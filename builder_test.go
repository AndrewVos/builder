package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func setup() {
	builderJson := `{
      "AuthToken": "lolsszz",
      "Host": "http://example.org",
      "Port": "1212",
      "Repositories": [
        {"Owner": "AndrewVos", "Repository": "builder"}
      ]
    }`
	builderJson = strings.TrimSpace(builderJson)
	ioutil.WriteFile("builder.json", []byte(builderJson), 0700)
}

func cleanup() {
	os.RemoveAll("builds")
	os.Remove("build_results.json")
	os.Remove("builder.json")
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

func TestCreatesHooks(t *testing.T) {
	setup()
	defer cleanup()

	var paths []string
	var bodies []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		b, _ := ioutil.ReadAll(r.Body)
		bodies = append(bodies, string(b))
	}))
	defer ts.Close()
	githubDomain = ts.URL

	supportedEvents := []string{"push", "pull_request"}
	createHooks()

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
        "url": "http://example.org:1212/hooks/` + event + `",
        "content_type": "json"
      }
    }`
		if bodies[i] != expectedBody {
			t.Errorf("Didn't post expected body\nExpected:\n%v\nActual:\n%v", expectedBody, bodies[i])
		}
	}
}

func TestRedPush(t *testing.T) {
	setup()
	defer cleanup()

	postToHooks("test-data/red_push.json", "push")

	build := allBuilds()[0]
	if build.Success {
		t.Errorf("Build should have failed!")
	}
	buildOutput, _ := ioutil.ReadFile(build.LogPath())
	if expected := "FAILING BUILD"; strings.Contains(string(buildOutput), expected) == false {
		t.Errorf("Expected log to contain %q. Got:\n%v", expected, string(buildOutput))
	}
}

func TestGreenPush(t *testing.T) {
	setup()
	defer cleanup()

	postToHooks("test-data/green_push.json", "push")

	build := allBuilds()[0]
	if build.Success == false {
		t.Errorf("Build should have succeeded!")
	}
	buildOutput, _ := ioutil.ReadFile(build.LogPath())
	if expected := "SUCCESSFUL BUILD"; strings.Contains(string(buildOutput), expected) == false {
		t.Errorf("Expected log to contain %q. Got:\n%v", expected, string(buildOutput))
	}
}

func TestRedPullRequest(t *testing.T) {
	setup()
	defer cleanup()

	postToHooks("test-data/red_pull_request.json", "pull_request")

	build := allBuilds()[0]
	if build.Success {
		t.Errorf("Build should have failed!")
	}
	buildOutput, _ := ioutil.ReadFile(build.LogPath())
	if expected := "FAILING BUILD"; strings.Contains(string(buildOutput), expected) == false {
		t.Errorf("Expected log to contain %q. Got:\n%v", expected, string(buildOutput))
	}
}

func TestGreenPullRequest(t *testing.T) {
	setup()
	defer cleanup()

	postToHooks("test-data/green_pull_request.json", "pull_request")

	build := allBuilds()[0]
	if build.Success == false {
		t.Errorf("Build should have succeeded!")
	}
	buildOutput, _ := ioutil.ReadFile(build.LogPath())
	if expected := "SUCCESSFUL BUILD"; strings.Contains(string(buildOutput), expected) == false {
		t.Errorf("Expected log to contain %q. Got:\n%v", expected, string(buildOutput))
	}
}

func TestClosedPullRequest(t *testing.T) {
	setup()
	defer cleanup()

	postToHooks("test-data/closed_pull_request.json", "pull_request")

	if len(allBuilds()) > 0 {
		t.Errorf("Erm, probably shouldn't build a closed pull request")
	}
}
