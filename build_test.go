package main

import (
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
)

func TestBuildUrl(t *testing.T) {
	withFakeDatabase(func(fdb *FakeDatabase) {
		database := &PostgresDatabase{}
		githubBuild := &GithubBuild{Owner: "bla", Repository: "repooo"}
		database.SaveGithubBuild(githubBuild)
		build := &Build{
			Owner:      "bla",
			Repository: "repooo",
		}
		database.CreateBuild(githubBuild, build)
		expected := "http://localhost:1212/build_output?id=" + strconv.Itoa(build.Id)
		if build.Url != expected {
			t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, build.Url)
		}
	})
}

func TestBuildUrlPort80(t *testing.T) {
	setup("")
	defer cleanup()

	oldPort := configuration.Port
	configuration.Port = "80"
	defer func() { configuration.Port = oldPort }()

	githubBuild := &GithubBuild{Owner: "bla", Repository: "repooo"}
	database.SaveGithubBuild(githubBuild)
	build := &Build{
		Owner:      "bla",
		Repository: "repooo",
	}
	database.CreateBuild(githubBuild, build)
	expected := "http://localhost/build_output?id=" + strconv.Itoa(build.Id)
	if build.Url != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, build.Url)
	}
}

func TestFailingBuild(t *testing.T) {
	withFakeDatabase(func(fdb *FakeDatabase) {
		setup("red")
		ghb := &GithubBuild{Owner: "some-owner", Repository: "some-repo"}
		fdb.GithubBuild = ghb

		database.SaveGithubBuild(ghb)

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
	})
}

func TestPassingBuild(t *testing.T) {
	withFakeDatabase(func(fdb *FakeDatabase) {
		setup("green")
		ghb := &GithubBuild{Owner: "some-owner", Repository: "some-repo"}
		fdb.GithubBuild = ghb
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
	})
}

func TestOutputEnvirons(t *testing.T) {
	withFakeDatabase(func(fdb *FakeDatabase) {
		setup("environs")
		ghb := &GithubBuild{Owner: "some-owner", Repository: "some-repo"}
		fdb.GithubBuild = ghb
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
	})
}
