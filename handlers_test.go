package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type FakeBuildLauncher struct {
	values        map[string]interface{}
	commits       []Commit
	launchedBuild bool
}

func (fbl *FakeBuildLauncher) LaunchBuild(owner string, repo string, ref string, sha string, githubURL string, commits []Commit) error {
	fbl.launchedBuild = true
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
		pushHandler(httptest.NewRecorder(), createFakeRequest("test-data/green_push.json"))

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

func TestPushHandlerIgnoresDeletedBranches(t *testing.T) {
	withFakeLauncher(func(fbl *FakeBuildLauncher) {
		pushHandler(httptest.NewRecorder(), createFakeRequest("test-data/delete_branch_push.json"))
		if fbl.launchedBuild {
			t.Error("Shouldn't build deleted branch pushes")
		}
	})
}

func TestPullRequestHandlerIgnoresClosedPullRequest(t *testing.T) {
	withFakeLauncher(func(fbl *FakeBuildLauncher) {
		pullRequestHandler(httptest.NewRecorder(), createFakeRequest("test-data/closed_pull_request.json"))
		if fbl.launchedBuild {
			t.Error("Shouldn't build closed pull requests")
		}
	})
}

func TestPullRequestHandlerLaunchesBuildWithCorrectValues(t *testing.T) {
	withFakeLauncher(func(fbl *FakeBuildLauncher) {
		pullRequestHandler(httptest.NewRecorder(), createFakeRequest("test-data/green_pull_request.json"))

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

func TestAddRepositoryHandlerCreatesHooksAndRepository(t *testing.T) {
	formValues := url.Values{}
	formValues.Set("owner", "RepoOwnerrr")
	formValues.Set("repository", "RailsTurboLinks")

	fakeDatabase.FindAccountByIdToReturn = &Account{
		Id:          3232,
		AccessToken: "sdfwef",
	}

	r, _ := http.NewRequest("", "", nil)
	r.AddCookie(&http.Cookie{Name: "account_id", Value: "123"})
	r.AddCookie(&http.Cookie{Name: "token", Value: "nothing"})
	r.PostForm = formValues
	addRepositoryHandler(httptest.NewRecorder(), r)

	expectedValues := map[string]interface{}{
		"accessToken": "sdfwef",
		"owner":       "RepoOwnerrr",
		"repository":  "RailsTurboLinks",
	}

	for field, expectedValue := range expectedValues {
		if actual := fakeGit.createHooksParameters[field]; actual != expectedValue {
			t.Errorf("Expected create hook parameter %q to be %q, but was %q\n", field, expectedValue, actual)
		}
	}

	if fakeDatabase.SavedRepository.Owner != expectedValues["owner"] {
		t.Errorf("Expected Owner to be %q, but was %q\n", expectedValues["owner"], fakeDatabase.SavedRepository.Owner)
	}
	if fakeDatabase.SavedRepository.Repository != expectedValues["repository"] {
		t.Errorf("Expected Repository to be %q, but was %q\n", expectedValues["repository"], fakeDatabase.SavedRepository.Repository)
	}
}

func TestGithubLoginHandlerCreatesNewAccount(t *testing.T) {
	resetFakeDatabase()
	resetFakeGit()
	fakeGit.AccessTokenToReturn = "some-access-token-123"
	fakeGit.UserIdToReturn = 56733
	fakeDatabase.LoginToReturn = &Login{AccountId: 121212, Token: "tokennn"}

	r, _ := http.NewRequest("GET", "http://bla.com/?code=QUERY_CODE", nil)
	githubLoginHandler(httptest.NewRecorder(), r)
	if fakeDatabase.CreatedAccount == nil {
		t.Fatal("Expected an account to be created")
	}
	if fakeDatabase.CreatedAccount.Id != 56733 {
		t.Fatalf("Github user ID wasn't stored")
	}
	if fakeDatabase.CreatedAccount.AccessToken != "some-access-token-123" {
		t.Fatalf("Access Token wasn't stored")
	}
}

func TestGithubLoginHandlerSetsCookieWithValidLogin(t *testing.T) {
	resetFakeDatabase()
	resetFakeGit()
	fakeDatabase.LoginToReturn = &Login{AccountId: 121212, Token: "tokennn"}

	r, _ := http.NewRequest("GET", "http://bla.com/?code=QUERY_CODE", nil)
	w := httptest.NewRecorder()

	fakeGit.AccessTokenToReturn = "some-access-token-123"
	fakeGit.UserIdToReturn = 56733

	githubLoginHandler(w, r)

	if w.Header()["Set-Cookie"][0] != "account_id=56733" {
		t.Errorf("Cookie account_id wasn't set properly, got %v", w.Header()["Set-Cookie"][0])
	}
	if w.Header()["Set-Cookie"][1] != "token=tokennn" {
		t.Errorf("Cookie token wasn't set properly")
	}
}
