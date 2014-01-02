package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCreatesHooks(t *testing.T) {
	setup("")
	defer cleanup()

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
	createHooks("lolsszz", "AndrewVos", "builder")

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

func TestRedPush(t *testing.T) {
	setup("red")
	defer cleanup()

	postToHooks("test-data/red_push.json", "push")

	build := AllBuilds()[0]
	if build.Success {
		t.Errorf("Build should have failed!")
	}
	buildOutput, _ := ioutil.ReadFile(build.LogPath())
	if expected := "FAILING BUILD"; strings.Contains(string(buildOutput), expected) == false {
		t.Errorf("Expected log to contain %q. Got:\n%v", expected, string(buildOutput))
	}
}

func TestGreenPush(t *testing.T) {
	setup("green")
	defer cleanup()

	postToHooks("test-data/green_push.json", "push")

	build := AllBuilds()[0]
	if build.Success == false {
		t.Errorf("Build should have succeeded!")
	}
	buildOutput, _ := ioutil.ReadFile(build.LogPath())
	if expected := "SUCCESSFUL BUILD"; strings.Contains(string(buildOutput), expected) == false {
		t.Errorf("Expected log to contain %q. Got:\n%v", expected, string(buildOutput))
	}
}

func TestRedPullRequest(t *testing.T) {
	setup("red")
	defer cleanup()

	postToHooks("test-data/red_pull_request.json", "pull_request")

	build := AllBuilds()[0]
	if build.Success {
		t.Errorf("Build should have failed!")
	}
	buildOutput, _ := ioutil.ReadFile(build.LogPath())
	if expected := "FAILING BUILD"; strings.Contains(string(buildOutput), expected) == false {
		t.Errorf("Expected log to contain %q. Got:\n%v", expected, string(buildOutput))
	}
}

func TestGreenPullRequest(t *testing.T) {
	setup("green")
	defer cleanup()

	postToHooks("test-data/green_pull_request.json", "pull_request")

	build := AllBuilds()[0]
	if build.Success == false {
		t.Errorf("Build should have succeeded!")
	}
	buildOutput, _ := ioutil.ReadFile(build.LogPath())
	if expected := "SUCCESSFUL BUILD"; strings.Contains(string(buildOutput), expected) == false {
		t.Errorf("Expected log to contain %q. Got:\n%v", expected, string(buildOutput))
	}
}

func TestClosedPullRequest(t *testing.T) {
	setup("green")
	defer cleanup()

	postToHooks("test-data/closed_pull_request.json", "pull_request")

	if len(AllBuilds()) > 0 {
		t.Errorf("Erm, probably shouldn't build a closed pull request")
	}
}

func TestDeleteBranch(t *testing.T) {
	setup("green")
	defer cleanup()

	postToHooks("test-data/delete_branch_push.json", "push")

	if len(AllBuilds()) > 0 {
		t.Errorf("Erm, probably shouldn't build a delete branch push")
	}
}

func TestFakeSomeBuilds(t *testing.T) {
	if os.Getenv("FAKE_BUILDS") != "" {
		setup("")
		defer cleanup()

		go func() {
			serve()
		}()
		for {
			postToHooks("test-data/red_pull_request.json", "pull_request")
			postToHooks("test-data/green_pull_request.json", "pull_request")
			postToHooks("test-data/red_push.json", "push")
			postToHooks("test-data/green_push.json", "push")
			postToHooks("test-data/slow_pull_request.json", "pull_request")
			time.Sleep(1000000 * time.Second)
		}
	}
}

func TestOutputEnvirons(t *testing.T) {
	setup("environs")
	defer cleanup()

	postToHooks("test-data/output_environs_push.json", "push")

	build := AllBuilds()[0]

	expectedLines := []string{
		"BUILDER_BUILD_URL=" + build.URL,
		"BUILDER_BUILD_ID=" + build.ID,
		"BUILDER_BUILD_OWNER=" + build.Owner,
		"BUILDER_BUILD_REPO=" + build.Repo,
		"BUILDER_BUILD_REF=" + build.Ref,
		"BUILDER_BUILD_SHA=" + build.SHA,
	}
	actual := build.ReadOutput()

	for _, expected := range expectedLines {
		if strings.Contains(actual, expected) == false {
			t.Errorf("Expected build output to contain:\n%v\nGot this instead:\n%v", expected, actual)
		}
	}
}

func TestExecutesHooksWithEnvirons(t *testing.T) {
	setup("green")
	defer cleanup()

	hook := `
    #!/usr/bin/env bash

    pwd >> hook-output
    echo BUILDER_BUILD_RESULT=$BUILDER_BUILD_RESULT >> hook-output
    echo BUILDER_BUILD_URL=$BUILDER_BUILD_URL >> hook-output
    echo BUILDER_BUILD_ID=$BUILDER_BUILD_ID >> hook-output
    echo BUILDER_BUILD_OWNER=$BUILDER_BUILD_OWNER >> hook-output
    echo BUILDER_BUILD_REPO=$BUILDER_BUILD_REPO >> hook-output
    echo BUILDER_BUILD_REF=$BUILDER_BUILD_REF >> hook-output
    echo BUILDER_BUILD_SHA=$BUILDER_BUILD_SHA >> hook-output
  `
	hook = strings.TrimSpace(hook)

	ioutil.WriteFile("data/hooks/test-hook", []byte(hook), 0700)
	postToHooks("test-data/green_pull_request.json", "pull_request")

	build := AllBuilds()[0]
	outputFile := "data/builds/" + build.ID + "/hook-output"
	b, _ := ioutil.ReadFile(outputFile)
	dir, _ := os.Getwd()
	expected := dir + `/data/builds/` + build.ID + `
BUILDER_BUILD_RESULT=pass
BUILDER_BUILD_URL=http://localhost:1212/build_output?id=` + build.ID + `
BUILDER_BUILD_ID=` + build.ID + `
BUILDER_BUILD_OWNER=AndrewVos
BUILDER_BUILD_REPO=builder-test-green-repo
BUILDER_BUILD_REF=pool-request
BUILDER_BUILD_SHA=7f39d6495acae9db022cc20e7f0d940158e0337d`

	actual := strings.TrimSpace(string(b))
	if actual != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, actual)
	}
}

func TestStoresPullRequestInfo(t *testing.T) {
	setup("green")
	defer cleanup()

	postToHooks("test-data/green_pull_request.json", "pull_request")

	build := AllBuilds()[0]
	expected := "https://api.github.com/repos/AndrewVos/builder-test-green-repo/pulls/2"
	if build.GithubURL != expected {
		t.Errorf("Expected Github URL:\n%v\nGot:\n%v\n", expected, build.GithubURL)
	}
}

func TestStoresPushInfo(t *testing.T) {
	setup("green")
	defer cleanup()

	postToHooks("test-data/green_push.json", "push")

	build := AllBuilds()[0]

	commits := []Commit{
		Commit{SHA: "92a9437adf4ac6f0114552e5149d0598fdbf0355", Message: "empty", URL: "https://github.com/AndrewVos/builder-test-green-repo/commit/92a9437adf4ac6f0114552e5149d0598fdbf0355"},
		Commit{SHA: "576be25d7e3d5320e92472d5734b50b17c1822e0", Message: "output something", URL: "https://github.com/AndrewVos/builder-test-green-repo/commit/576be25d7e3d5320e92472d5734b50b17c1822e0"},
	}

	if len(build.Commits) != 2 {
		t.Fatalf("Expected two commits")
	}

	for i, expectedCommit := range commits {
		if build.Commits[i] != expectedCommit {
			t.Errorf("Expected commit:\n%vGot:\n%v\n", expectedCommit, build.Commits[i])
		}
	}

	expectedGithubURL := "https://github.com/AndrewVos/builder-test-green-repo/compare/da46166aa120...576be25d7e3d"
	if build.GithubURL != expectedGithubURL {
		t.Errorf("Expected Github URL:\n%v\nGot:\n%v\n", expectedGithubURL, build.GithubURL)
	}
}
