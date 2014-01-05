package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestClosedPullRequest(t *testing.T) {
	setup("green")
	defer cleanup()

	postToHooks("test-data/closed_pull_request.json", "pull_request")

	if len(database.AllBuilds()) > 0 {
		t.Errorf("Erm, probably shouldn't build a closed pull request")
	}
}

func TestDeleteBranch(t *testing.T) {
	setup("green")
	defer cleanup()

	postToHooks("test-data/delete_branch_push.json", "push")

	if len(database.AllBuilds()) > 0 {
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

	build := database.AllBuilds()[0]

	expectedLines := []string{
		"BUILDER_BUILD_URL=" + build.Url,
		"BUILDER_BUILD_ID=" + strconv.Itoa(build.Id),
		"BUILDER_BUILD_OWNER=" + build.Owner,
		"BUILDER_BUILD_REPO=" + build.Repository,
		"BUILDER_BUILD_REF=" + build.Ref,
		"BUILDER_BUILD_SHA=" + build.Sha,
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

	build := database.AllBuilds()[0]
	outputFile := "data/builds/" + strconv.Itoa(build.Id) + "/hook-output"
	b, _ := ioutil.ReadFile(outputFile)
	dir, _ := os.Getwd()
	expected := dir + `/data/builds/` + strconv.Itoa(build.Id) + `
BUILDER_BUILD_RESULT=pass
BUILDER_BUILD_URL=http://localhost:1212/build_output?id=` + strconv.Itoa(build.Id) + `
BUILDER_BUILD_ID=` + strconv.Itoa(build.Id) + `
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

	build := database.AllBuilds()[0]
	expected := "https://api.github.com/repos/AndrewVos/builder-test-green-repo/pulls/2"
	if build.GithubUrl != expected {
		t.Errorf("Expected Github URL:\n%v\nGot:\n%v\n", expected, build.GithubUrl)
	}
}

func TestStoresPushInfo(t *testing.T) {
	setup("green")
	defer cleanup()

	postToHooks("test-data/green_push.json", "push")

	build := database.AllBuilds()[0]

	commits := []Commit{
		Commit{Sha: "92a9437adf4ac6f0114552e5149d0598fdbf0355", Message: "empty", Url: "https://github.com/AndrewVos/builder-test-green-repo/commit/92a9437adf4ac6f0114552e5149d0598fdbf0355"},
		Commit{Sha: "576be25d7e3d5320e92472d5734b50b17c1822e0", Message: "output something", Url: "https://github.com/AndrewVos/builder-test-green-repo/commit/576be25d7e3d5320e92472d5734b50b17c1822e0"},
	}

	if len(build.Commits) != 2 {
		t.Fatalf("Expected two commits, but got %d\n", len(build.Commits))
	}

	for i, expectedCommit := range commits {
		if build.Commits[i].Sha != expectedCommit.Sha {
			t.Errorf("Expected commit to have Sha:\n%vGot:\n%v\n", expectedCommit.Sha, build.Commits[i].Sha)
		}
		if build.Commits[i].Message != expectedCommit.Message {
			t.Errorf("Expected commit to have Message:\n%vGot:\n%v\n", expectedCommit.Message, build.Commits[i].Message)
		}
		if build.Commits[i].Url != expectedCommit.Url {
			t.Errorf("Expected commit to have Url:\n%vGot:\n%v\n", expectedCommit.Url, build.Commits[i].Url)
		}
	}

	expectedGithubUrl := "https://github.com/AndrewVos/builder-test-green-repo/compare/da46166aa120...576be25d7e3d"
	if build.GithubUrl != expectedGithubUrl {
		t.Errorf("Expected Github URL:\n%v\nGot:\n%v\n", expectedGithubUrl, build.GithubUrl)
	}
}
