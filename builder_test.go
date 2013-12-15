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

func postToHooks(path string) {
	b, _ := ioutil.ReadFile(path)
	request, _ := http.NewRequest("POST", "/hooks", nil)
	request.Body = ioutil.NopCloser(strings.NewReader(string(b)))
	w := httptest.NewRecorder()
	hookHandler(w, request)
}

func TestCreatesPushHook(t *testing.T) {
	setup()
	defer cleanup()

	path := ""
	body := ""
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.RequestURI()
		b, _ := ioutil.ReadAll(r.Body)
		body = string(b)
	}))
	defer ts.Close()
	githubDomain = ts.URL

	createHooks()

	expectedPath := "/repos/AndrewVos/builder/hooks?access_token=lolsszz"
	if path != expectedPath {
		t.Errorf("Got wrong post address\nExpected: %v\nActual: %v", expectedPath, path)
	}
	expectedBody := `{
      "name": "web",
      "active": true,
      "events": [
        "push",
        "pull_request"
      ],
      "config": {
        "url": "http://example.org:1212/hook",
        "content_type": "json"
      }
    }
  `
	expectedBody = strings.TrimSpace(expectedBody)
	if body != expectedBody {
		t.Errorf("Didn't post expected body\nExpected:\n%v\nActual:\n%v", expectedBody, body)
	}
}

func TestRedPush(t *testing.T) {
	setup()
	defer cleanup()

	postToHooks("test-data/red_push.json")

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

	postToHooks("test-data/green_push.json")

	build := allBuilds()[0]
	if build.Success == false {
		t.Errorf("Build should have succeeded!")
	}
	buildOutput, _ := ioutil.ReadFile(build.LogPath())
	if expected := "SUCCESSFUL BUILD"; strings.Contains(string(buildOutput), expected) == false {
		t.Errorf("Expected log to contain %q. Got:\n%v", expected, string(buildOutput))
	}
}
