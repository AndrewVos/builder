package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
)

func cleanDataDirectory() {
	os.RemoveAll("data")
}

func TestBuildUrl(t *testing.T) {
	database := &PostgresDatabase{}
	repository := &Repository{Owner: "bla", Repository: "repooo"}
	database.SaveRepository(repository)
	build := &Build{
		Owner:      "bla",
		Repository: "repooo",
	}
	database.CreateBuild(repository, build)
	expected := "http://localhost:1212/build_output?id=" + strconv.Itoa(build.Id)
	if build.Url != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, build.Url)
	}
}

func TestBuildUrlPort80(t *testing.T) {
	database := &PostgresDatabase{}

	oldPort := configuration.Port
	configuration.Port = "80"
	defer func() { configuration.Port = oldPort }()

	repository := &Repository{Owner: "bla", Repository: "repooo"}
	database.SaveRepository(repository)
	build := &Build{
		Owner:      "bla",
		Repository: "repooo",
	}
	database.CreateBuild(repository, build)
	expected := "http://localhost/build_output?id=" + strconv.Itoa(build.Id)
	if build.Url != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, build.Url)
	}
}

func TestFailingBuild(t *testing.T) {
	defer cleanDataDirectory()

	fakeGit.FakeRepo = "red"
	repository := &Repository{Owner: "some-owner", Repository: "some-repo"}
	fakeDatabase.SavedRepository = repository

	build := &Build{Owner: "some-owner", Repository: "some-repo"}

	build.start()

	if !build.Complete {
		t.Error("Build should be complete")
	}
	if build.Success {
		t.Error("Build should not be successful")
	}
	if build.Result != "fail" {
		t.Error("Build should have failed")
	}

	buildOutput, _ := ioutil.ReadFile(build.LogPath())
	if expected := "FAILING BUILD"; strings.Contains(string(buildOutput), expected) == false {
		t.Errorf("Expected log to contain %q. Got:\n%v", expected, string(buildOutput))
	}
}

func TestPassingBuild(t *testing.T) {
	defer cleanDataDirectory()

	fakeGit.FakeRepo = "green"
	repository := &Repository{Owner: "some-owner", Repository: "some-repo"}
	fakeDatabase.SavedRepository = repository
	build := &Build{Owner: "some-owner", Repository: "some-repo"}

	build.start()

	if !build.Complete {
		t.Error("Build should be complete")
	}
	if !build.Success {
		t.Error("Build should be successful")
	}
	if build.Result != "pass" {
		t.Error("Build should have passed")
	}

	buildOutput, _ := ioutil.ReadFile(build.LogPath())
	if expected := "SUCCESSFUL BUILD"; strings.Contains(string(buildOutput), expected) == false {
		t.Errorf("Expected log to contain %q. Got:\n%v", expected, string(buildOutput))
	}
}

func TestOutputEnvirons(t *testing.T) {
	defer cleanDataDirectory()

	fakeGit.FakeRepo = "environs"
	repository := &Repository{Owner: "some-owner", Repository: "some-repo"}
	fakeDatabase.SavedRepository = repository
	build := &Build{
		Url:        " sdsdfsd",
		Id:         23,
		Owner:      "some-owner",
		Repository: "some-repo",
		Ref:        "sdfw23233",
		Sha:        "ewf2f",
	}

	build.start()

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
