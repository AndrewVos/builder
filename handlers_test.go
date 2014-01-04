package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

type FakeBuildLauncher struct {
	values  map[string]interface{}
	commits []Commit
}

func (bl *FakeBuildLauncher) Do(build *Build) {}
func (fbl *FakeBuildLauncher) LaunchBuild(owner string, repo string, ref string, sha string, githubURL string, commits []Commit) error {
	fbl.values = map[string]interface{}{
		"owner":     owner,
		"repo":      repo,
		"ref":       ref,
		"sha":       sha,
		"githubURL": githubURL,
	}
	fbl.commits = commits
	return nil
}

func createFakeRequest(bodyPath string) *http.Request {
	b, _ := ioutil.ReadFile(bodyPath)
	reader := bytes.NewReader(b)
	return &http.Request{Body: ioutil.NopCloser(reader)}
}

func withFakeLauncher(block func(fbl *FakeBuildLauncher)) {
	oldLauncher := launcher
	fbl := &FakeBuildLauncher{}
	launcher = fbl
	block(fbl)
	launcher = oldLauncher
}

func TestPushHandlerLaunchesBuildWithCorrectValues(t *testing.T) {
	withFakeLauncher(func(fbl *FakeBuildLauncher) {
		pushHandler(nil, createFakeRequest("test-data/green_push.json"))

		expectedValues := map[string]interface{}{
			"owner":     "AndrewVos",
			"repo":      "builder-test-green-repo",
			"ref":       "master",
			"sha":       "576be25d7e3d5320e92472d5734b50b17c1822e0",
			"githubURL": "https://github.com/AndrewVos/builder-test-green-repo/compare/da46166aa120...576be25d7e3d",
		}

		for field, expected := range expectedValues {
			if fbl.values[field] != expected {
				t.Errorf("Expected field %v to be:\n%v\nActual:\n%v", field, expected, fbl.values[field])
			}
		}

		expectedCommits := []Commit{
			Commit{Sha: "92a9437adf4ac6f0114552e5149d0598fdbf0355", Message: "empty", Url: "https://github.com/AndrewVos/builder-test-green-repo/commit/92a9437adf4ac6f0114552e5149d0598fdbf0355"},
			Commit{Sha: "576be25d7e3d5320e92472d5734b50b17c1822e0", Message: "output something", Url: "https://github.com/AndrewVos/builder-test-green-repo/commit/576be25d7e3d5320e92472d5734b50b17c1822e0"},
		}

		for i, expected := range expectedCommits {
			if fbl.commits[i] != expected {
				t.Errorf("Expected commit %d to be:\n%v\nActual:\n%v\n", i, expected, fbl.commits[i])
			}
		}
	})
}

func TestPullRequestHandlerLaunchesBuildWithCorrectValues(t *testing.T) {
	withFakeLauncher(func(fbl *FakeBuildLauncher) {
		pullRequestHandler(nil, createFakeRequest("test-data/green_pull_request.json"))

		expectedValues := map[string]interface{}{
			"owner":     "AndrewVos",
			"repo":      "builder-test-green-repo",
			"ref":       "pool-request",
			"sha":       "7f39d6495acae9db022cc20e7f0d940158e0337d",
			"githubURL": "https://api.github.com/repos/AndrewVos/builder-test-green-repo/pulls/2",
		}

		for field, expected := range expectedValues {
			if fbl.values[field] != expected {
				t.Errorf("Expected field %v to be:\n%v\nActual:\n%v", field, expected, fbl.values[field])
			}
		}
	})
}
