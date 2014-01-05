package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

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
